package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/tarao/prisoners-switch/game"
	"github.com/tarao/prisoners-switch/rule"
	"github.com/tarao/prisoners-switch/strategy"
)

func main() {
	var logLevelName string
	var patient bool
	flag.StringVar(&logLevelName, "log-level", "result", "specify events to be logged.\n\"silent\", \"result\", \"changes\", \"all\", \"debug\"")
	flag.BoolVar(&patient, "patient", false, "wait until every game finishes or times out\nand ignore failure in mortal games.")
	flag.Parse()

	logLevel := game.LogResult
	switch logLevelName {
	case "debug":
		logLevel = game.LogDebug
	case "all":
		logLevel = game.LogAll
	case "changes":
		logLevel = game.LogChanges
	case "result":
		logLevel = game.LogResult
	case "silent":
		logLevel = game.LogSilent
	}

	totalGames := 100

	games := make([]game.Game, 0)
	fairGames := make([]game.Game, 0)
	mortalGames := make([]game.Game, 0)

	for i := 0; i < totalGames; i++ {
		g := game.NewFairGame(&game.Logger{
			Game:     fmt.Sprintf("#%d", i+1),
			LogLevel: logLevel,
			Writer:   os.Stderr,
		})
		games = append(games, g)
		fairGames = append(fairGames, g)
	}

	for i := 0; i < totalGames; i++ {
		g := game.NewMortalGame(&game.Logger{
			Game:     fmt.Sprintf("#%d", i+len(fairGames)+1),
			LogLevel: logLevel,
			Writer:   os.Stderr,
		})
		games = append(games, g)
		mortalGames = append(mortalGames, g)
	}

	s := strategy.MyNewStrategy()
	strategies := newShuffledStrategies(len(games))

	done := make(chan struct{})

	go func() {
		forEachGame(fairGames, func(g game.Game) {
			resumableStrategy := newResumableStrategy(s)
			strategies.append(resumableStrategy)
			success := <-g.Start(resumableStrategy, rule.TotalPrisoners)
			if !success && !patient {
				exitByGames(fairGames)
			}
		})
		done <- struct{}{}
	}()

	go func() {
		forEachGame(mortalGames, func(g game.Game) {
			resumableStrategy := newResumableStrategy(s)
			strategies.append(resumableStrategy)
			success := <-g.Start(resumableStrategy, rule.TotalPrisoners)
			if !success && !patient {
				exit(false, g.Result().Message, fairGames)
			}
		})
	}()

	strategies.yield(rule.TotalPrisoners)

	select {
	case <-done:
		for _, g := range mortalGames {
			g.Stop()
		}
		exitByGames(fairGames)
	case <-time.After(300 * time.Second):
		exit(false, "Timed out", fairGames)
	}
}

func gameResult(games []game.Game) *game.Result {
	result := &game.Result{Success: true}
	for _, g := range games {
		result.Merge(g.Result())
	}
	result.Score /= uint64(len(games))
	return result
}

func exitByResult(result *game.Result) {
	jsonResult, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	os.Stdout.Write(jsonResult)

	if result.Success {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func exitByGames(games []game.Game) {
	exitByResult(gameResult(games))
}

func exit(success bool, msg string, games []game.Game) {
	result := &game.Result{
		Success: success,
		Message: msg,
	}
	result.Merge(gameResult(games))

	exitByResult(result)
}

func forEachGame(gs []game.Game, f func(g game.Game)) {
	wg := &sync.WaitGroup{}
	for _, g := range gs {
		g := g
		wg.Add(1)
		go func() {
			f(g)
			wg.Done()
		}()
	}
	wg.Wait()
}

func newResumableStrategy(s rule.Strategy) *resumableStrategy {
	return &resumableStrategy{
		strategy: s,
		ch:       make(chan struct{}),
	}
}

type resumableStrategy struct {
	strategy rule.Strategy
	ch       chan struct{}
}

func (s *resumableStrategy) yield() {
	s.ch <- struct{}{}
}

func (s *resumableStrategy) NewPrisoner(number int, shout chan rule.Shout) rule.Prisoner {
	<-s.ch
	return s.strategy.NewPrisoner(number, shout)
}

func newShuffledStrategies(size int) *shuffledStrategies {
	wg := &sync.WaitGroup{}
	wg.Add(size)
	return &shuffledStrategies{
		strategies: make([]*resumableStrategy, 0, size),
		wg:         wg,
	}
}

type shuffledStrategies struct {
	strategies []*resumableStrategy
	mu         sync.RWMutex
	wg         *sync.WaitGroup
}

func (ss *shuffledStrategies) append(s *resumableStrategy) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.strategies = append(ss.strategies, s)
	ss.wg.Done()
}

func (ss *shuffledStrategies) shuffle() {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	for i := len(ss.strategies) - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		ss.strategies[i], ss.strategies[j] = ss.strategies[j], ss.strategies[i]
	}
}

func (ss *shuffledStrategies) ready() {
	ss.wg.Wait()
}

func (ss *shuffledStrategies) yieldOnce() {
	ss.shuffle()

	ss.mu.RLock()
	defer ss.mu.RUnlock()

	for _, s := range ss.strategies {
		s.yield()
	}
}

func (ss *shuffledStrategies) yield(n int) {
	ss.ready()
	for i := 0; i < n; i++ {
		ss.yieldOnce()
	}
}
