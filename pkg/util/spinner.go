package util

import (
	"fmt"
	"os"
)

//var spinChars = `-/|\`
var spinChars = []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'}

// Spinner keeps track of a message and a spin character. It can be used to indicate progress in between
// processing work items by repeatedly calling Tick()
type Spinner struct {
	message string
	i       int
}

// NewSpinner creates a spinner with the provided message. This message will be printed together with the current spin
// character. It may be empty
func NewSpinner(msg string) *Spinner {
	return &Spinner{message: msg}
}

// TickWithUpdate prints the spinner message to stdout and advances the spinner by one tick
func (s *Spinner) TickWithUpdate(update string) {
	fmt.Printf("%s %c - %s \r", s.message, spinChars[s.i], update)
	s.i = (s.i + 1) % len(spinChars)
}

// Tick prints the spinner character to stdout and advances the spinner by one tick
func (s *Spinner) Tick() {
	fmt.Printf("%s %c\r", s.message, spinChars[s.i])
	s.i = (s.i + 1) % len(spinChars)
}

// IsTTY indicates whether the current stdout is a TTY
func IsTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
