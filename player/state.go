package player

import ()

// State encapsulates all the state that the back-end needs to know about a player.
type State struct {
	Name  string
	Color string
}

// NewState creates a new State object with name and color from the given arguments
// in that order, and returns a pointer to it.
func NewState(nameForNewPlayer string, colorForNewPlayer string) *State {
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
