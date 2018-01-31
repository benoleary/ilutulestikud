package player

import ()

type State struct {
	Name  string
	Color string
}

// CreateNew with two arguments creates a new State object with name and color
// from the given arguments in that order, and returns a pointer to it.
func CreateByNameAndColor(nameForNewPlayer string, colorForNewPlayer string) *State {
	return &State{nameForNewPlayer, colorForNewPlayer}
}

// UpdateNonEmptyStrings over-writes all non-name string attributes of this
// state with those from updaterReference unless the string in updaterReference
// is empty.
func (state *State) UpdateNonEmptyStrings(updaterReference *State) {
	if updaterReference.Color != "" {
		state.Color = updaterReference.Color
	}
}
