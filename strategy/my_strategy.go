package strategy

import "github.com/tarao/prisoners-switch/rule"
	
const kCount = 2000

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
	cnt int
	done bool
	initA bool
	initB bool
}

func (p *prisoner) Enter(room rule.Room) {
	if p.done {
		return
	}
	if p.cnt == 0 {
		p.initA = room.TakeSwitchA().State()
		p.initB = room.TakeSwitchB().State()
	}
	p.cnt++
	if p.cnt < kCount {
		return
	}
	if p.initA == room.TakeSwitchA().State() && p.initB == room.TakeSwitchB().State() {
		// probably the first
		room.TakeSwitchA().Toggle()
		p.done = true
		return
	}
	if p.initB == room.TakeSwitchB().State() {
		// probably the second
		room.TakeSwitchB().Toggle()
		p.done = true
		return
	}
	// probably the third
	p.shout <- rule.Triumph
}
