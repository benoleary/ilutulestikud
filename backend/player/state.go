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
	// state with those from updaterReference unless the string in updaterReference
	// is empty.
	UpdateFromPresentAttributes(updaterReference endpoint.PlayerState)
}

// ForBackend writes the relevant parts of the state into the JSON object for the frontend.
func ForBackend(state State) endpoint.PlayerState {
	return endpoint.PlayerState{
		Name:  state.Name(),
		Color: state.Color(),
	}
}

// Factory defines the interface for structs which should be able to create objects
// implementing the State interface out of instances of endpoint.PlayerState, which
// normally represents de-serialized JSON from the frontend.
type Factory interface {
	// Create should return an object implementing the State interface which
	// corresponds to the information given by the endpoint.PlayerState object.
	Create(playerInformation endpoint.PlayerState) State
}
