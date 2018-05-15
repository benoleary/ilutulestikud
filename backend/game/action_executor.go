package game

import (
	"fmt"

	"github.com/benoleary/ilutulestikud/backend/player"
)

// ActionExecutor encapsulates the write functions on a game's state
// which update the state based on player actions, according to the
// game's ruleset.
type ActionExecutor struct {
	gameRuleset  Ruleset
	gameState    ReadAndWriteState
	actingPlayer player.ReadonlyState
}

// ExecutorForPlayer creates a ActionExecutor around the given game
// state if the given player is a participant, returning a pointer to
// the executor. If the player is not a participant, it returns nil
// along with an error.
func ExecutorForPlayer(
	stateOfGame ReadAndWriteState,
	actingPlayer player.ReadonlyState) (*ActionExecutor, error) {
	gameParticipants := stateOfGame.Read().PlayerNames()
	for _, gameParticipant := range gameParticipants {
		if gameParticipant == actingPlayer.Name() {
			actionExecutor :=
				&ActionExecutor{
					gameRuleset:  stateOfGame.Read().Ruleset(),
					gameState:    stateOfGame,
					actingPlayer: actingPlayer,
				}

			return actionExecutor, nil
		}
	}

	// If we have not yet returned a pointer, then the player was not
	// a participant.
	notFoundError :=
		fmt.Errorf(
			"No player with name %v is a participant in game %v",
			actingPlayer.Name(),
			stateOfGame.Read().Name())

	return nil, notFoundError
}

// RecordChatMessage records the given chat message from the acting player,
// or returns an error.
func (actionExecutor *ActionExecutor) RecordChatMessage(chatMessage string) {
	actionExecutor.gameState.RecordChatMessage(actionExecutor.actingPlayer, chatMessage)
}
