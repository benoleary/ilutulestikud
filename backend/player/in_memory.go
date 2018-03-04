package player

import (
	"sync"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// ThreadsafeFactory creates threadsafeState objects as State objects.
type ThreadsafeFactory struct {
}

// Create creates a new threadsafeState object with name and color from the given arguments
// in that order, and returns a pointer to it.
func (threadsafeFactory *ThreadsafeFactory) Create(endpointPlayer endpoint.PlayerState) State {
	return &threadsafeState{
		name:            endpointPlayer.Name,
		color:           endpointPlayer.Color,
		mutualExclusion: sync.Mutex{},
	}
}

// threadsafeState encapsulates all the state that the backend needs to know about a player,
// using a mutex to ensure that updates are thread-safe.
type threadsafeState struct {
	name            string
	color           string
	mutualExclusion sync.Mutex
}

// Name returns the private name field.
func (playerState *threadsafeState) Name() string {
	return playerState.name
}

// Color returns the private color field.
func (playerState *threadsafeState) Color() string {
	return playerState.color
}

// UpdateFromPresentAttributes over-writes all non-name string attributes of this
// state with those from updaterReference unless the string in updaterReference
// is empty.
func (playerState *threadsafeState) UpdateFromPresentAttributes(updaterReference endpoint.PlayerState) {
	// It would be more efficient to only lock if we go into an if statement,
	// but then multiple if statements would be less efficient, and there would
	// be a mutex in each if statement.
	playerState.mutualExclusion.Lock()

	if updaterReference.Color != "" {
		playerState.color = updaterReference.Color
	}

	playerState.mutualExclusion.Unlock()
}
