package player

import ()

type State struct {
	name  string
	color string
}

func (state *State) Name() string {
	return state.name
}

func (state *State) Color() string {
	return state.color
}

// CreateNew with two arguments creates a new State object with name and ID
// from the given arguments in that order. The ID is not a particularly well
// thought-through notion so far, and is mainly still here to make the State
// object less trivial.
func CreateByNameAndColor(nameForNewPlayer string, colorForNewPlayer string) State {
	return State{nameForNewPlayer, colorForNewPlayer}
}
