package game

import (
	"github.com/benoleary/ilutulestikud/backend/game"
)

// StateCollection defines what a struct should do to allow a Handler from
// github.com/benoleary/ilutulestikud/backend/endpoint/game to read and write
// game states.
type StateCollection interface {
	// ViewState should return a view around the read-only game state corresponding
	// to the given name as seen by the given player. If the game does not exist or
	// the player is not a participant, it should return an error.
	ViewState(gameName string, playerName string) (*game.PlayerView, error)

	// ViewAllWithPlayer should return a slice of read-only views on all the games in
	// the collection which have the given player as a participant. It should return an
	// error if there is a problem wrapping any of the read-only game states in a view.
	// The order is not mandated, and may even change with repeated calls to the same
	// unchanged collection (analogously to the entry set of a standard Golang map, for
	// example), though of course an implementation may order the slice consistently.
	ViewAllWithPlayer(playerName string) ([]*game.PlayerView, error)

	// ExecuteAction should return an executor around the read-and-write game state
	// corresponding to the given name, for actions by the given player for the given,
	// or should return an error.
	ExecuteAction(gameName string, playerName string) (*game.ActionExecutor, error)

	// AddNew should add a new game to the collection based on the given arguments.
	AddNew(
		gameName string,
		gameRuleset game.Ruleset,
		playerNames []string) error
}
