package player

import (
	"sync"

	"github.com/benoleary/ilutulestikud/backendjson"
)

// State defines the interface for structs which should encapsulate the state of a player.
type State interface {
	Name() string

	Color() string

	UpdateNonEmptyStrings(updaterReference backendjson.PlayerState)
}

// ForBackend writes the relevant parts of the state into the JSON object for the front-end.
func ForBackend(state State) backendjson.PlayerState {
	return backendjson.PlayerState{
		Name:  state.Name(),
		Color: state.Color(),
	}
}

// OriginalState encapsulates all the state that the back-end needs to know about a player.
type OriginalState struct {
	name            string
	color           string
	mutualExclusion sync.Mutex
}

// NewState creates a new State object with name and color from the given arguments
// in that order, and returns a pointer to it.
func NewState(nameForNewPlayer string, colorForNewPlayer string) *OriginalState {
	return &OriginalState{
		name:            nameForNewPlayer,
		color:           colorForNewPlayer,
		mutualExclusion: sync.Mutex{},
	}
}

// Name returns the private name field.
func (state *OriginalState) Name() string {
	return state.name
}

// Color returns the private color field.
func (state *OriginalState) Color() string {
	return state.color
}

// UpdateNonEmptyStrings over-writes all non-name string attributes of this
// state with those from updaterReference unless the string in updaterReference
// is empty.
func (state *OriginalState) UpdateNonEmptyStrings(updaterReference backendjson.PlayerState) {
	// It would be more efficient to only lock if we go into an if statement,
	// but then multiple if statements would be less efficient, and there would
	// be a mutex in each if statement.
	state.mutualExclusion.Lock()

	if updaterReference.Color != "" {
		state.color = updaterReference.Color
	}

	state.mutualExclusion.Unlock()
}
