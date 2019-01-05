package game

import (
	"fmt"
	"io"
)

// LogLevel distingushes event log level
type LogLevel int

const (
	// LogSilent logs nothing
	LogSilent LogLevel = iota

	// LogResult logs only the result of the game
	LogResult

	// LogChanges logs only change to the game state
	LogChanges

	// LogAll logs all events on the game
	LogAll

	// LogDebug logs everything including debug output
	LogDebug
)

// Logger logs events in the game
type Logger struct {
	Game     string
	LogLevel LogLevel
	Writer   io.Writer
}

func (l *Logger) printResult(r *Result) {
	result := "FAIL"
	if r.Success {
		result = "SUCCESS"
	}

	switches := "no switch"
	if r.UsedSwitches == 1 {
		switches = "a switch"
	} else if r.UsedSwitches > 1 {
		switches = fmt.Sprintf("%d switches", r.UsedSwitches)
	}

	if l.LogLevel >= LogResult {
		fmt.Fprintf(l.Writer, "[Game %s] %s in %d steps with %s\n", l.Game, result, r.Steps, switches)
	}
}

func (l *Logger) printChange(name string, state interface{}, operator string) {
	if l.LogLevel >= LogChanges {
		fmt.Fprintf(l.Writer, "[Game %s] %s set to %v by %s\n", l.Game, name, state, operator)
	}
}

func (l *Logger) printInfo(msg string) {
	if l.LogLevel >= LogAll {
		fmt.Fprintf(l.Writer, "[Game %s] %s\n", l.Game, msg)
	}
}

func (l *Logger) printDebugInfo(msg string) {
	if l.LogLevel >= LogDebug {
		fmt.Fprintf(l.Writer, "[Game %s] %s\n", l.Game, msg)
	}
}
