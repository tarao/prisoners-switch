package strategy

import "github.com/tarao/prisoners-switch/rule"

// MyNewStrategy returns a new strategy
func MyNewStrategy() rule.Strategy {
	return &myStrategy{}
}

type myStrategy struct {
}

// Each prisoner receives unique number between 0 and rule.TotalPrisoners -1. Order is not guaranteed.
func (s *myStrategy) NewPrisoner(number int, shout chan rule.Shout) rule.Prisoner {
	return &prisoner{
		shout:       shout,
		isCollector: number == 0,
		remaining:   rule.TotalPrisoners - 1, // number of reporters not reporting yet
	}
}

type prisoner struct {
	shout       chan rule.Shout
	isCollector bool
	initialized bool

	// collector fields.
	remaining int

	// reporter fields.
	initialState         int
	initialStateModified bool
	reported             bool
}

func (p *prisoner) Enter(room rule.Room) {
	if p.isCollector {
		p.collectorEnter(room)
	} else {
		p.reporterEnter(room)
	}
}

func (p *prisoner) reporterEnter(room rule.Room) {
	c := getCounter(room)
	if !p.initialized {
		p.initialState = c
		p.initialized = true
		return
	}
	if !p.initialStateModified && p.initialState == c {
		return
	}
	p.initialStateModified = true

	if !p.reported && c < 1 {
		setCounter(room, 1)
		p.reported = true
	}
}

func (p *prisoner) collectorEnter(room rule.Room) {
	c := getCounter(room)
	defer func() {
		if c == 0 {
			// change switch state to notify reporters that the collector is ready.
			setCounter(room, 1)
			p.remaining++
		} else {
			setCounter(room, 0)
		}
	}()

	if !p.initialized {
		p.initialized = true
		return
	}
	p.remaining -= c
	if p.remaining == 0 {
		p.shout <- rule.Triumph
	}
}

func getCounter(room rule.Room) int {
	sa := room.TakeSwitchA()
	c := 0
	if sa.State() {
		c += 1
	}
	return c
}

func setCounter(room rule.Room, c int) {
	sa := room.TakeSwitchA()
	a := c != 0
	if sa.State() != a {
		sa.Toggle()
	}
}
