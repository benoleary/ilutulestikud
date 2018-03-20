package player

import (
	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// ReadOnly defines the interface for structs which should encapsulate the state of a player.
type ReadOnly interface {
	// Identifier should return the identifier of the player for interaction between
	// frontend and backend.
	Identifier() string

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

// Collection defines the interface for structs which should be able to create objects
// implementing the State interface out of instances of endpoint.PlayerState, which
// normally represents de-serialized JSON from the frontend, and for tracking the objects
// by their identifier, which is the player name. It is also responsible for maintaining
// the list of chat colors available for player states.
type Collection interface {
	// Add should add an element to the collection which is a new object implementing
	// the State interface with information given by the endpoint.PlayerState object,
	// and return the identifier of the newly-created player, along with an error which
	// of course should be nil if there was no problem.
	// It should return an error if the player already exists.
	Add(playerInformation endpoint.PlayerState) (string, error)

	// Get should return the State corresponding to the given player identifier if it
	// exists already (or else nil) along with whether the State exists, analogously to
	// a standard Golang map.
	Get(playerIdentifier string) (ReadOnly, bool)

	// All should return a slice of all the State instances in the collection. The order
	// is not mandated, and may even change with repeated calls to the same unchanged
	// Collection (analogously to the entry set of a standard Golang map, for example),
	// though of course an implementation may order the slice consistently.
	All() []ReadOnly

	// Reset should remove all players which are not among the initial players, and
	// restore any initial players who have been removed.
	Reset()

	// AvailableChatColors should return the chat colors which are allowed for players.
	AvailableChatColors() []string
}

// ForEndpoint writes relevant parts of the collection's states into the JSON object for
// the frontend as a list of player objects as its "Players" attribute. The order of the
// players may not be consistent with repeated calls, as the order of All is not guaranteed
// to be consistent
func ForEndpoint(collection Collection) endpoint.PlayerList {
	playerStates := collection.All()
	playerList := make([]endpoint.PlayerState, 0, len(playerStates))
	for _, playerState := range playerStates {
		playerList = append(playerList, endpoint.PlayerState{
			Identifier: playerState.Identifier(),
			Name:       playerState.Name(),
			Color:      playerState.Color(),
		})
	}

	return endpoint.PlayerList{Players: playerList}
}

// AvailableChatColorsForEndpoint writes the chat colors available to the given collection
// into the JSON object for the frontend.
func AvailableChatColorsForEndpoint(collection Collection) endpoint.ChatColorList {
	return endpoint.ChatColorList{Colors: collection.AvailableChatColors()}
}
