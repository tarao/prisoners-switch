package rule

// TotalPrisoners is the total number of prisoners
const TotalPrisoners = 100

// Shout is a type which describes what a prisoner can say to the game master
type Shout struct{}

// Triumph is the only value of Shout
var Triumph Shout = struct{}{}

// SwitchButton describes what a prisoner can do for a switch button in the room
type SwitchButton interface {
	Name() string
	State() bool
	Toggle() bool
}

// Room describes what is in the room
type Room interface {
	TakeSwitchA() SwitchButton
	TakeSwitchB() SwitchButton
}

// Prisoner describes what a prisoner can do
type Prisoner interface {
	Enter(room Room)
}

// Strategy describes how a prisoner behaves
type Strategy interface {
	// NewPrisoner Creates a new instance of a prisoner, who is
	// distinguished by other prisoners by its number.
	//
	// Note that this method will be called multiple times against the
	// same number for new games.
	//
	// A prisoner should keep shout channel if he wants to assert that
	// he wins.
	NewPrisoner(number int, shout chan Shout) Prisoner
}
