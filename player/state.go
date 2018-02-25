package player

import (
	"sync"

	"github.com/benoleary/ilutulestikud/backendjson"
)

// State encapsulates all the state that the back-end needs to know about a player.
type State struct {
	Name            string
	Color           string
	mutualExclusion sync.Mutex
}

// NewState creates a new State object with name and color from the given arguments
// in that order, and returns a pointer to it.
func NewState(nameForNewPlayer string, colorForNewPlayer string) *State {
	return &State{
		Name:            nameForNewPlayer,
		Color:           colorForNewPlayer,
		mutualExclusion: sync.Mutex{},
	}
}

// UpdateNonEmptyStrings over-writes all non-name string attributes of this
// state with those from updaterReference unless the string in updaterReference
// is empty.
func (state *State) UpdateNonEmptyStrings(updaterReference backendjson.PlayerState) {
	// It would be more efficient to only lock if we go into an if statement,
	// but then multiple if statements would be less efficient, and there would
	// be a mutex in each if statement.
	state.mutualExclusion.Lock()

	if updaterReference.Color != "" {
		state.Color = updaterReference.Color
	}

	state.mutualExclusion.Unlock()
}

// ForBackend writes the relevant parts of the state into the JSON object for the front-end.
func (state *State) ForBackend() backendjson.PlayerState {
	return backendjson.PlayerState{
		Name:  state.Name,
		Color: state.Color,
	}
}
