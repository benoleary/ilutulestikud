package player

import (
	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// State defines the interface for structs which should encapsulate the state of a player.
type State interface {
	// Name should return the name of the player as known to the games.
	Name() string

	// Color should return the color that the player uses for chat messages.
	Color() string

	// UpdateFromPresentAttributes should over-write all non-name string attributes of this
	// state with those from updaterReference unless the string in updaterReference is empty.
	// This should be thread-safe and if the instance is persisted by a Collection, this
	// function should ensure that the updated version gets persisted, again in a thread-safe
	// way.
	UpdateFromPresentAttributes(updaterReference endpoint.PlayerState)
}

// ForBackend writes the relevant parts of the state into the JSON object for the frontend.
func ForBackend(state State) endpoint.PlayerState {
	return endpoint.PlayerState{
		Name:  state.Name(),
		Color: state.Color(),
	}
}

// Collection defines the interface for structs which should be able to create objects
// implementing the State interface out of instances of endpoint.PlayerState, which
// normally represents de-serialized JSON from the frontend. It is also responsible
// for maintaining the list of chat colors available for player states.
type Collection interface {
	// Add should add an element to the collection which is a new object implementing
	// the State interface with information given by the endpoint.PlayerState object.
	Add(playerInformation endpoint.PlayerState)

	// Get should return the State corresponding to the given player name if it exists
	// already (or else nil) along with whether the State exists, analogously to a
	// standard Golang map.
	Get(playerName string) (State, bool)

	// All should return a slice of all the State instances in the collection. The order
	// is not mandated, and may even change with repeated calls to the same unchanged
	// Collection (analogously to the entry set of a standard Golang map, for example),
	// though of course an implementation may order the slice consistently.
	All() []State

	// Reset should remove all players which are not among the initial players, and
	// restore any initial players who have been removed.
	Reset()

	// AvailableChatColors should return the chat colors which are allowed for players.
	AvailableChatColors() []string
}
