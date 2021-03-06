package player

import (
	"context"

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
	All(executionContext context.Context) ([]player.ReadonlyState, error)

	// Get should return a read-only state for the given player.
	Get(executionContext context.Context, playerName string) (player.ReadonlyState, error)

	// AvailableChatColors should return the chat colors available to the collection.
	AvailableChatColors(executionContext context.Context) []string

	// Add should add a new player to the collection, defined by the given arguments.
	Add(executionContext context.Context, playerName string, chatColor string) error

	// UpdateColor should update the given player with the given chat color.
	UpdateColor(executionContext context.Context, playerName string, chatColor string) error

	// Delete should delete the given player from the collection.
	Delete(executionContext context.Context, playerName string) error
}
