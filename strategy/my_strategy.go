package strategy

import "github.com/tarao/prisoners-switch/rule"

// MyNewStrategy returns a new strategy
func MyNewStrategy() rule.Strategy {
	return &myStrategy{}
}

type myStrategy struct {
}

func (s *myStrategy) NewPrisoner(number int, shout chan rule.Shout) rule.Prisoner {
	return &prisoner{shout: shout, iAmCounter: number == 0}
}

type prisoner struct {
	iAmCounter    bool
	isInitialized bool
	counter       int

	haveSeenOn bool
	switched   bool
	shout      chan rule.Shout
}

func (p *prisoner) Enter(room rule.Room) {
	if rule.TotalPrisoners == 1 {
		// Am I the only one?
		p.shout <- rule.Triumph
		return
	}
	aIsOn := room.TakeSwitchA().State()
	if !p.iAmCounter {
		if p.switched {
			return;
		}
		if aIsOn {
			p.haveSeenOn = true
		} else {
			if p.haveSeenOn {
				room.TakeSwitchA().Toggle()
				p.switched = true
			}
		}
		return
	}
	// counter
	if (!p.isInitialized) {
		// make sure that intial state is 1!!
		if (aIsOn) {
			// intial state was 1
			room.TakeSwitchA().Toggle()
		} else {
			/// intial state was 0
			p.counter -= 1
			room.TakeSwitchA().Toggle()
		}
		p.isInitialized = true
		return
	}
	if (aIsOn) {
		p.counter += 1;
		if p.counter == rule.TotalPrisoners-1 {
			p.shout <- rule.Triumph
			return
		}
		room.TakeSwitchA().Toggle()
	} else {
		p.counter -= 1;
		room.TakeSwitchA().Toggle()
	}
}
