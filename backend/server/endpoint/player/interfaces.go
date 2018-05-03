package player

import (
	"github.com/benoleary/ilutulestikud/backend/player"
)

// StateCollection defines what a struct should do to allow a Handler from
// github.com/benoleary/ilutulestikud/backend/endpoint/player to read and write
// player states.
type StateCollection interface {
	// All should return a slice of all the players in the collection. The order is not
	// mandated, and may even change with repeated calls to the same unchanged collection
	// (analogously to the entry set of a standard Golang map, for example), though of
	// course an implementation may order the slice consistently.
	All() []player.ReadonlyState

	// Get should return a read-only state for the identified player.
	Get(playerIdentifier string) (player.ReadonlyState, error)

	// AvailableChatColors should return the chat colors available to the collection.
	AvailableChatColors() []string

	// Add should add a new player to the collection, defined by the given arguments.
	Add(playerName string, chatColor string) error

	// UpdateColor should update the given player with the given chat color.
	UpdateColor(playerName string, chatColor string) error

	// Reset should reset the players to the initial set.
	Reset()
}
