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
	return &prisoner{shout: shout}
}

type prisoner struct {
	shout chan rule.Shout
}

func (p *prisoner) Enter(room rule.Room) {
	// shout triumph when you think it's ready
	p.shout <- rule.Triumph
}
