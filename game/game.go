package game

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/tarao/prisoners-switch/rule"
)

type switchButton struct {
	name  string
	state bool
	used  uint64
}

func newSwitchButton(name string) switchButton {
	return switchButton{
		name:  name,
		state: rand.Int31n(2) == 0,
	}
}

func (btn *switchButton) Name() string {
	return btn.name
}

func (btn *switchButton) State() bool {
	return btn.state
}

func (btn *switchButton) Toggle() bool {
	btn.state = !btn.state
	return btn.state
}

type loggableSwitchButton struct {
	*switchButton
	logger         *Logger
	prisonerNumber int
}

func (btn *loggableSwitchButton) Toggle() bool {
	state := btn.switchButton.Toggle()
	btn.logger.printChange(
		fmt.Sprintf("switch %s", btn.name),
		state,
		fmt.Sprintf("prisoner #%d", btn.prisonerNumber),
	)
	return state
}

type room struct {
	btnA switchButton
	btnB switchButton
}

func (r *room) prepareEnteringBy(p *prisonerState, logger *Logger) *loggableRoom {
	return &loggableRoom{
		room:           *r, // copy
		logger:         logger,
		prisonerNumber: p.number,
	}
}

type loggableRoom struct {
	room           room
	logger         *Logger
	prisonerNumber int
}

func (r *loggableRoom) TakeSwitchA() rule.SwitchButton {
	r.room.btnA.used = 1
	return &loggableSwitchButton{
		&r.room.btnA,
		r.logger,
		r.prisonerNumber,
	}
}

func (r *loggableRoom) TakeSwitchB() rule.SwitchButton {
	r.room.btnB.used = 1
	return &loggableSwitchButton{
		&r.room.btnB,
		r.logger,
		r.prisonerNumber,
	}
}

type prisonerState struct {
	number   int
	entered  int
	prisoner rule.Prisoner
}

func newPrisonerState(
	strategy rule.Strategy,
	number int,
	finish chan rule.Shout,
) *prisonerState {
	shout := make(chan rule.Shout, 1)
	prisoner := strategy.NewPrisoner(number, shout)
	go func() {
		s := <-shout
		finish <- s
	}()

	return &prisonerState{
		number:   number,
		entered:  0,
		prisoner: prisoner,
	}
}

const maxSteps = 100000

// Result describes the result of a game
type Result struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	Steps        uint64 `json:"steps"`
	UsedSwitches uint64 `json:"used_switches"`
	Score        uint64 `json:"score"`
}

func (r *Result) calcScore() *Result {
	if !r.Success {
		r.Score = 0
		return r
	}

	r.Score = maxSteps - r.Steps
	if r.Steps > maxSteps {
		r.Score = 0
	}

	// a bonus for using only one switch
	if r.Score > 0 && r.UsedSwitches <= 1 {
		r.Score += maxSteps
	}

	return r
}

// Merge merges another result into this result
func (r *Result) Merge(other *Result) *Result {
	r.Steps += other.Steps
	if r.UsedSwitches < other.UsedSwitches {
		r.UsedSwitches = other.UsedSwitches
	}
	r.Score += other.Score

	if !other.Success {
		r.Success = false
		r.Message = other.Message
		r.Score = 0
	}

	if r.Message == "" {
		r.Message = other.Message
	}

	return r
}

// Game is the interface of the game
type Game interface {
	Start(strategy rule.Strategy, numPrisoners int) <-chan bool
	Stop()
	Result() *Result
}

func newGame(logger *Logger) *game {
	return &game{
		room: room{
			btnA: newSwitchButton("A"),
			btnB: newSwitchButton("B"),
		},
		states: make([]*prisonerState, 0),
		logger: logger,
	}
}

type game struct {
	mu      sync.RWMutex
	room    room
	states  []*prisonerState
	steps   uint64
	success uint64
	logger  *Logger
	stopped int32
}

func (g *game) initialize(strategy rule.Strategy, numPrisoners int) <-chan bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	shout := make(chan rule.Shout, 1)
	for i := 0; i < numPrisoners; i++ {
		g.states = append(g.states, newPrisonerState(strategy, i, shout))
	}

	finish := make(chan bool, 1)
	go func() {
		<-shout

		r := g.check()
		if r {
			atomic.AddUint64(&g.success, 1)
		}
		g.logger.printResult(g.Result())

		g.Stop()
		finish <- r
	}()

	return finish
}

func (g *game) check() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, p := range g.states {
		if p.entered < 1 {
			return false
		}
	}
	return true
}

func (g *game) letEnter(p *prisonerState) {
	atomic.AddUint64(&g.steps, 1)

	g.mu.Lock()
	p.entered++
	g.mu.Unlock()

	g.logger.printInfo(fmt.Sprintf("prisoner #%d is entering the room", p.number))

	r := g.room.prepareEnteringBy(p, g.logger)
	p.prisoner.Enter(r)

	g.mu.Lock()
	g.room = r.room // copy
	g.mu.Unlock()
}

func (g *game) Success() bool {
	return atomic.LoadUint64(&g.success) > 0
}

func (g *game) Steps() uint64 {
	return atomic.LoadUint64(&g.steps)
}

func (g *game) UsedSwitches() uint64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return atomic.LoadUint64(&g.room.btnA.used) + atomic.LoadUint64(&g.room.btnB.used)
}

func (g *game) Result() *Result {
	msg := "All games passed"
	if !g.Success() {
		msg = "Some game failed"
	}
	r := &Result{
		Success:      g.Success(),
		Message:      msg,
		Steps:        g.Steps(),
		UsedSwitches: g.UsedSwitches(),
	}
	return r.calcScore()
}

func (g *game) Stop() {
	atomic.AddInt32(&g.stopped, 1)
}

func (g *game) IsStopped() bool {
	return atomic.LoadInt32(&g.stopped) > 0
}

// NewFairGame creates a new game in which prisoner can win with some strategy
func NewFairGame(logger *Logger) Game {
	return &fairGame{newGame(logger)}
}

type fairGame struct {
	*game
}

func (g *fairGame) Start(strategy rule.Strategy, numPrisoners int) <-chan bool {
	ch := g.initialize(strategy, numPrisoners)

	go func() {
		for {
			p := g.states[rand.Int31n(int32(len(g.states)))]
			g.letEnter(p)
			g.logger.printDebugInfo(fmt.Sprintf("switch A: %v", g.room.btnA.state))
			g.logger.printDebugInfo(fmt.Sprintf("switch B: %v", g.room.btnB.state))
			if g.Success() || g.IsStopped() {
				break
			}
		}
	}()

	return ch
}

// NewMortalGame creates a new game in which prisoner can never win
func NewMortalGame(logger *Logger) Game {
	return &mortalGame{newGame(logger), -1}
}

type mortalGame struct {
	*game
	skippedPrisoner int32
}

func (g *mortalGame) Start(strategy rule.Strategy, numPrisoners int) <-chan bool {
	result := make(chan bool, 2)

	ch := g.initialize(strategy, numPrisoners)
	g.skippedPrisoner = rand.Int31n(int32(len(g.states)))

	go func() {
		for {
			i := rand.Int31n(int32(len(g.states)))
			if i != g.skippedPrisoner {
				p := g.states[i]
				g.letEnter(p)
				g.logger.printDebugInfo(fmt.Sprintf("switch A: %v", g.room.btnA.state))
				g.logger.printDebugInfo(fmt.Sprintf("switch B: %v", g.room.btnB.state))

				if g.IsStopped() {
					result <- true
					break
				}
			}
		}
	}()

	go func() {
		r := <-ch
		result <- r
	}()

	return result
}
