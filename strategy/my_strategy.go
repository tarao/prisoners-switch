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

type History int

const (
	Initial History = iota
	AoffBoff
	AonBoff
	AoffBon
	AonBon
	Ready     // seen state change, but could not switch
	Switched  // already switched
)

type prisoner struct {
	iAmCounter    bool
	isInitialized bool

	counter int
	shout   chan rule.Shout

	history History
}

/*
数え役は必ず状態を変化させる。(00->01, それ以外は00へ)
その他は、「状態変化」するまで待機、状態変化したらインクリメントを実行。
この方法だと、カウンターの2ビットが活かせる
 */

func (p *prisoner) Enter(room rule.Room) {
	if rule.TotalPrisoners == 1 {
		// An I the only one?
		p.shout <- rule.Triumph
		return
	}

	aIsOn := room.TakeSwitchA().State()
	bIsOn := room.TakeSwitchB().State()
	if !p.iAmCounter {
		if p.history == Switched {
			// already switched
			return;
		}
		if p.history == Initial {
			// capture
			if aIsOn && bIsOn {
				p.history = AonBon
			} else if aIsOn {
				p.history = AonBoff
			} else if bIsOn {
				p.history = AoffBon
			} else {
				p.history = AoffBoff
			}
			return
		}
		if aIsOn && bIsOn {
			if p.history != AonBon {
				p.history = Ready
			}
		} else if aIsOn {
			if p.history != AonBoff {
				p.history = Ready
			}
		} else if bIsOn {
			if p.history != AoffBon {
				p.history = Ready
			}
		} else {
			if p.history != AoffBoff {
				p.history = Ready
			}
		}
		if (p.history != Ready) {
			return
		}
		// have seen change. ready to move
		if aIsOn && bIsOn {
			return; // cannot increment
		} else if aIsOn { // 01 -> 10
			room.TakeSwitchA().Toggle()
			room.TakeSwitchB().Toggle()
		} else { // 10 -> 11 , 00 -> 01
			room.TakeSwitchA().Toggle()
		}
		p.history = Switched
		return
	}
	count := 0;
	if (aIsOn) {
		count += 1;
	}
	if (bIsOn) {
		count += 2;
	}
	if !p.isInitialized {
		// ignore current count
		count = 0;
		p.isInitialized = true
	}
	p.counter += count;
	if (p.counter == rule.TotalPrisoners-1) {
		p.shout <- rule.Triumph
		return
	}

	if !aIsOn && !bIsOn { //00 -> 01
		p.counter -= 1;
		room.TakeSwitchA().Toggle()
		return
	}
	if (aIsOn) {
		room.TakeSwitchA().Toggle()
	}
	if (bIsOn) {
		room.TakeSwitchB().Toggle()
	}
}
