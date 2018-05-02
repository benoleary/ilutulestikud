package game

import (
	"fmt"

	"github.com/benoleary/ilutulestikud/backend/chat"
)

// PlayerView encapsulates the functions on a game's read-only state
// which provide the information available to a particular player for
// that state.
type PlayerView struct {
	gameState  ReadonlyState
	playerName string
}

// ViewForPlayer creates a PlayerView around the given game state if
// the given player is a participant, returning a pointer to the view.
// If the player is not a participant, it returns nil along with an
// error.
func ViewForPlayer(
	stateOfGame ReadonlyState,
	nameOfPlayer string) (*PlayerView, error) {
	gameParticipants := stateOfGame.Players()
	for _, gameParticipant := range gameParticipants {
		if gameParticipant.Name() == nameOfPlayer {
			playerView :=
				&PlayerView{
					gameState:  stateOfGame,
					playerName: nameOfPlayer,
				}

			return playerView, nil
		}
	}

	// If we have not yet returned a pointer, then the player was not a
	// participant.
	notFoundError :=
		fmt.Errorf(
			"No player with name %v is a participant in game %v",
			nameOfPlayer,
			stateOfGame.Name())

	return nil, notFoundError
}

// GameName just wraps around the read-only game state's Name function.
func (playerView *PlayerView) GameName() string {
	return playerView.gameState.Name()
}

// SortedChatLog sorts the read-only game state's ChatLog and returns the sorted log.
func (playerView *PlayerView) SortedChatLog() []chat.Message {
	return playerView.gameState.ChatLog().Sorted()
}

// Score just wraps around the read-only game state's Score function.
func (playerView *PlayerView) Score() int {
	return playerView.gameState.Score()
}

// NumberOfReadyHints just wraps around the read-only game state's
// NumberOfReadyHints function.
func (playerView *PlayerView) NumberOfReadyHints() int {
	return playerView.gameState.NumberOfReadyHints()
}

// NumberOfSpentHints just subtracts the read-only game state's
// NumberOfReadyHints function's return value from the constant maximum.
func (playerView *PlayerView) NumberOfSpentHints() int {
	maximumNumber := playerView.gameState.Ruleset().MaximumNumberOfHints()
	return maximumNumber - playerView.gameState.NumberOfReadyHints()
}

// NumberOfMistakesStillAllowed just subtracts the read-only game state's
// NumberOfMistakesMade function's return value from the constant maximum.
func (playerView *PlayerView) NumberOfMistakesStillAllowed() int {
	maximumNumber := playerView.gameState.Ruleset().MaximumNumberOfHints()
	return maximumNumber - playerView.gameState.NumberOfMistakesMade()
}

// NumberOfMistakesMade just wraps around the read-only game state's
// NumberOfMistakesMade function.
func (playerView *PlayerView) NumberOfMistakesMade() int {
	return playerView.gameState.NumberOfMistakesMade()
}

// CurrentTurnOrder returns the names of the participants of the game in the
// order which their next turns are in, along with true if the view is for
// the first player in that list or false otherwise.
func (playerView *PlayerView) CurrentTurnOrder() ([]string, bool) {
	gameParticipants := playerView.gameState.Players()
	numberOfParticipants := len(gameParticipants)

	playerNamesInTurnOrder := make([]string, numberOfParticipants)

	gameTurn := playerView.gameState.Turn()
	isPlayerTurn := false
	for playerIndex := 0; playerIndex < numberOfParticipants; playerIndex++ {
		// Game turns begin with 1 rather than 0, so this sets the player names in order,
		// wrapping index back to 0 when at the end of the list.
		// E.g. turn 3, 5 players: playerNamesInTurnOrder will start with
		// gameParticipants[2], then [3], then [4], then [0], then [1].
		playerInTurnOrder :=
			gameParticipants[(playerIndex+gameTurn-1)%numberOfParticipants]
		playerNamesInTurnOrder[playerIndex] =
			playerInTurnOrder.Name()

		if playerView.playerName == playerInTurnOrder.Name() {
			isPlayerTurn = true
		}
	}

	return playerNamesInTurnOrder, isPlayerTurn
}
