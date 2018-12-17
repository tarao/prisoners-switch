package strategy

import "github.com/tarao/prisoners-switch/rule"

// MyNewStrategy returns a new strategy
func MyNewStrategy() rule.Strategy {
	return &myStrategy{}
}

type myStrategy struct {
}

func (s *myStrategy) NewPrisoner(number int, shout chan rule.Shout) rule.Prisoner {
	return &prisoner{number: number, shout: shout}
}

type prisoner struct {
	shout    chan rule.Shout
	number   int
	entered  bool
	finished bool
}

func (p *prisoner) Enter(room rule.Room) {
	if !p.entered {
		p.entered = true
		theCounter.count(p.number)
	}

	if p.number == 0 && !p.finished && theCounter.isFinished() {
		p.finished = true
		p.shout <- rule.Triumph
	}
}

// a global counter
var theCounter = newCounter(100)

func newCounter(size int) *counter {
	return &counter{flags: make([]int, size), ch: make(chan struct{}, 1)}
}

type counter struct {
	flags []int
	ch    chan struct{}
}

func (c *counter) lock() {
	c.ch <- struct{}{}
}

func (c *counter) unlock() {
	<-c.ch
}

func (c *counter) count(number int) {
	c.lock()
	defer c.unlock()

	c.flags[number] = c.flags[number] + 1
}

func (c *counter) isFinished() bool {
	c.lock()
	defer c.unlock()

	for _, c := range c.flags {
		if c < 100 {
			return false
		}
	}
	return true
}
