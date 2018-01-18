package player

import ()

type State struct {
	Name  string
	Color string
}

// CreateNew with two arguments creates a new State object with name and ID
// from the given arguments in that order. The ID is not a particularly well
// thought-through notion so far, and is mainly still here to make the State
// object less trivial.
func CreateByNameAndColor(nameForNewPlayer string, colorForNewPlayer string) State {
	return State{nameForNewPlayer, colorForNewPlayer}
}
