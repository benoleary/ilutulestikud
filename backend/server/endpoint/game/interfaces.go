package game

import (
	"context"

	"github.com/benoleary/ilutulestikud/backend/game"
)

// StateCollection defines what a struct should do to allow a Handler from
// github.com/benoleary/ilutulestikud/backend/endpoint/game to read and write
// game states.
type StateCollection interface {
	// ViewState should return a view around the read-only game state corresponding
	// to the given name as seen by the given player. If the game does not exist or
	// the player is not a participant, it should return an error.
	ViewState(
		executionContext context.Context,
		gameName string,
		playerName string) (game.ViewForPlayer, error)

	// ViewAllWithPlayer should return a slice of read-only views on all the games in
	// the collection which have the given player as a participant. It should return an
	// error if there is a problem wrapping any of the read-only game states in a view.
	// The order is not mandated, and may even change with repeated calls to the same
	// unchanged collection (analogously to the entry set of a standard Golang map, for
	// example), though of course an implementation may order the slice consistently.
	ViewAllWithPlayer(
		executionContext context.Context,
		playerName string) ([]game.ViewForPlayer, error)

	// ExecuteAction should return an executor around the read-and-write game state
	// corresponding to the given name, for actions by the given player for the given,
	// or should return an error.
	ExecuteAction(
		executionContext context.Context,
		gameName string,
		playerName string) (game.ExecutorForPlayer, error)

	// AddNew should add a new game to the collection based on the given arguments.
	AddNew(
		executionContext context.Context,
		gameName string,
		gameRuleset game.Ruleset,
		playerNames []string) error

	// RemoveGameFromListForPlayer should remove the given player from the given game in
	// the sense that the game will no longer show up in the result of
	// ReadAllWithPlayer(playerName). It should return an error if the player is not a
	// participant of the game, as well as in general I/O errors and so on.
	RemoveGameFromListForPlayer(
		executionContext context.Context,
		gameName string,
		playerName string) error

	// Delete should delete the given game from the collection.
	Delete(executionContext context.Context, gameName string) error
}
