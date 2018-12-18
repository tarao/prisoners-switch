package strategy

import (
	"github.com/tarao/prisoners-switch/rule"
)

// MyNewStrategy returns a new strategy
func MyNewStrategy() rule.Strategy {
	return &myStrategy{}
}

type myStrategy struct {
}

func (s *myStrategy) NewPrisoner(number int, shout chan rule.Shout) rule.Prisoner {
	return &prisoner{shout: shout}
}

type prisoner struct {
	shout              chan rule.Shout
	defaultSwitchCount int
	toggled            bool
}

func (p *prisoner) Enter(room rule.Room) {
	// shout triumph when you think it's ready

	aState := room.TakeSwitchA().State()
	bState := room.TakeSwitchB().State()

	if !p.toggled {
		incrementSwitch(room)
		p.toggled = true
	}

	if !aState && !bState {
		p.defaultSwitchCount++
	} else {
		p.defaultSwitchCount = 0
	}

	if p.defaultSwitchCount > 1200 {
		p.shout <- rule.Triumph
	}
}

func incrementSwitch(room rule.Room) {
	bState := room.TakeSwitchB().State()

	if !bState {
		room.TakeSwitchB().Toggle()
		return
	}

	room.TakeSwitchA().Toggle()
	room.TakeSwitchB().Toggle()
}
