package player

import ()

type State struct {
	name string
	id   string
}

func (state *State) Name() string {
	return state.name
}

func (state *State) Id() string {
	return state.id
}

// CreateNew with two arguments creates a new State object with name and ID
// from the given arguments in that order. The ID is not a particularly well
// thought-through notion so far, and is mainly still here to make the State
// object less trivial.
func CreateByNameAndId(nameForNewPlayer string, idForNewPlayer string) State {
	return State{nameForNewPlayer, idForNewPlayer}
}

// CreateNew with one argument creates a new State object with name and ID
// both equal to the given argument. The ID is not a particularly well
// thought-through notion so far, and is mainly still here to make the State
// object less trivial.
func CreateByNameOnly(nameForNewPlayer string) State {
	return State{nameForNewPlayer, nameForNewPlayer}
}
