package strategy

import "github.com/tarao/prisoners-switch/rule"

// MyNewStrategy returns a new strategy
func MyNewStrategy() rule.Strategy {
	return &myStrategy{}
}

type myStrategy struct {
}

func (s *myStrategy) NewPrisoner(number int, shout chan rule.Shout) rule.Prisoner {
	return &prisoner{ shout: shout, iAmCounter: number == 0}
}

type prisoner struct {
	iAmCounter   bool

	// 数え役用
	isInitialized bool
	counterA      int
	counterB      int
	shout         chan rule.Shout

	// その他用
	switchedA bool
	switchedB bool
}

/*
カウント役と、それ以外でわける。カウント役は
counterA 01
counterB 10
の数をカウントする。自分以外の囚人はTotalPrisoners-1人いるので、counterがそこまで進んだら終了。
ただし、カウント役に最初に回ってきた時に01だった場合には、それが囚人によるものか初期配置によるものかどうか確定できない。
そのため、その場合にはcounterAを１減らす。

スイッチ一個だとあまり工夫の余地がなくて、同じ考えで2(n-1)回カウントすれば良い。

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
		if aIsOn || bIsOn {
			// 00以外の状態からは何もしない
			return
		}
		if !p.switchedA {
			// 1st time
			room.TakeSwitchA().Toggle()
			p.switchedA = true
			return
		}
		if !p.switchedB {
			// 2nd time
			room.TakeSwitchB().Toggle()
			p.switchedB = true
			return
		}
		// 3rd time and further
		return
	}
	// 数え役の処理
	if !p.isInitialized {
		// 初回アクセスの処理。
		if aIsOn && !bIsOn {
			// この組み合わせの場合、初期配置によるものか初回カウントか区別できない
			// カウントの回数を１回補正する
			p.counterA -= 1;
		}
		if aIsOn {
			room.TakeSwitchA().Toggle()
		}
		if bIsOn {
			room.TakeSwitchB().Toggle()
		}
		p.isInitialized = true;
		return;
	}
	if (aIsOn) {
		p.counterA += 1;
		if (p.counterA == rule.TotalPrisoners-1) {
			p.shout <- rule.Triumph
			return
		}
		room.TakeSwitchA().Toggle()
		return
	}
	if (bIsOn) {
		p.counterB += 1;
		if (p.counterB == rule.TotalPrisoners-1) {
			p.shout <- rule.Triumph
			return
		}
		room.TakeSwitchB().Toggle()
		return
	}

}
