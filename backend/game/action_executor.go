package game

import (
	"fmt"
)

// ActionExecutor encapsulates the write functions on a game's state
// which update the state based on player actions.
type ActionExecutor struct {
	gameState  readAndWriteState
	playerName string
}

// ExecutorForPlayer creates a ActionExecutor around the given game
// state if the given player is a participant, returning a pointer to
// the executor. If the player is not a participant, it returns nil
// along with an error.
func ExecutorForPlayer(
	stateOfGame readAndWriteState,
	nameOfPlayer string) (*ActionExecutor, error) {
	gameParticipants := stateOfGame.read().Players()
	for _, gameParticipant := range gameParticipants {
		if gameParticipant.Name() == nameOfPlayer {
			actionExecutor :=
				&ActionExecutor{
					gameState:  stateOfGame,
					playerName: nameOfPlayer,
				}

			return actionExecutor, nil
		}
	}

	// If we have not yet returned a pointer, then the player was not
	// a participant.
	notFoundError :=
		fmt.Errorf(
			"No player with name %v is a participant in game %v",
			nameOfPlayer,
			stateOfGame.read().Name())

	return nil, notFoundError
}
