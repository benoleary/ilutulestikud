package persister

import (
	"github.com/benoleary/ilutulestikud/backend/game"
)

// DeserializedState is a struct meant to encapsulate all the state
// required for a single game to function, along with having de-serialized
// the ruleset from its identifier.
type DeserializedState struct {
	SerializableState
	deserializedRuleset game.Ruleset
}

// Ruleset returns the ruleset for the game.
func (gameState *DeserializedState) Ruleset() game.Ruleset {
	return gameState.deserializedRuleset
}

// Read returns the game state itself as a read-only object for the
// purposes of reading properties.
func (gameState *DeserializedState) Read() game.ReadonlyState {
	return gameState
}
