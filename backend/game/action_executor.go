package game

import (
	"fmt"

	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// ActionExecutor encapsulates the write functions on a game's state
// which update the state based on player actions, according to the
// game's ruleset.
type ActionExecutor struct {
	gameRuleset      Ruleset
	gameState        ReadAndWriteState
	actingPlayer     player.ReadonlyState
	gameParticipants []string
}

// ExecutorOfActionsForPlayer creates a ActionExecutor around the
// given game state if the given player is a participant, returning a
// pointer to the executor. If the player is not a participant, it
// returns nil along with an error.
func ExecutorOfActionsForPlayer(
	stateOfGame ReadAndWriteState,
	actingPlayer player.ReadonlyState) (ExecutorForPlayer, error) {
	gameParticipants := stateOfGame.Read().PlayerNames()

	for _, gameParticipant := range gameParticipants {
		if gameParticipant == actingPlayer.Name() {
			actionExecutor :=
				&ActionExecutor{
					gameRuleset:      stateOfGame.Read().Ruleset(),
					gameState:        stateOfGame,
					actingPlayer:     actingPlayer,
					gameParticipants: gameParticipants,
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
func (actionExecutor *ActionExecutor) RecordChatMessage(chatMessage string) error {
	return actionExecutor.gameState.RecordChatMessage(actionExecutor.actingPlayer, chatMessage)
}

// TakeTurnByDiscarding enacts a turn by discarding the indicated card from the hand
// of the acting player, or return an error if it was not possible.
func (actionExecutor *ActionExecutor) TakeTurnByDiscarding(indexInHandToDiscard int) error {
	// First we must determine if the player is allowed to take an action.
	playerHand, handSize, errorFromHand := actionExecutor.handIfTurnElseError()

	if errorFromHand != nil {
		return errorFromHand
	}

	if (indexInHandToDiscard < 0) || (indexInHandToDiscard >= handSize) {
		return fmt.Errorf(
			"Index %v is out of the acceptable range %v to %v of the player's hand",
			indexInHandToDiscard,
			0,
			handSize)
	}

	numberOfParticipants := len(actionExecutor.gameParticipants)
	gameRuleset := actionExecutor.gameState.Read().Ruleset()

	if handSize < gameRuleset.NumberOfCardsInPlayerHand(numberOfParticipants) {
		return fmt.Errorf(
			"Player %v could not discard card because their last turn was already taken",
			actionExecutor.actingPlayer.Name())
	}

	discardedCard := playerHand[indexInHandToDiscard]

	actionMessage :=
		fmt.Sprintf(
			"discards card %v %v",
			discardedCard.ColorSuit(),
			discardedCard.SequenceIndex())

	replacementCard :=
		card.NewInferred(gameRuleset.ColorSuits(), gameRuleset.DistinctPossibleIndices())

	numberOfHintsToAdd := 0
	if actionExecutor.gameState.Read().NumberOfReadyHints() < gameRuleset.MaximumNumberOfHints() {
		numberOfHintsToAdd = 1
	}

	return actionExecutor.gameState.EnactTurnByDiscardingAndReplacing(
		actionMessage,
		actionExecutor.actingPlayer,
		indexInHandToDiscard,
		replacementCard,
		numberOfHintsToAdd,
		0)
}

func (actionExecutor *ActionExecutor) handIfTurnElseError() ([]card.Readonly, int, error) {
	turnNumber := actionExecutor.gameState.Read().Turn()
	indexOfPlayerForCurrentTurn := turnNumber % len(actionExecutor.gameParticipants)
	playerForCurrentTurn := actionExecutor.gameParticipants[indexOfPlayerForCurrentTurn]

	if playerForCurrentTurn != actionExecutor.actingPlayer.Name() {
		errorToReturn :=
			fmt.Errorf(
				"Player %v is not the current player so cannot take a turn",
				actionExecutor.actingPlayer.Name())
		return nil, -1, errorToReturn
	}

	playerHand, errorFromVisibleHand :=
		actionExecutor.gameState.Read().VisibleHand(actionExecutor.actingPlayer.Name())

	if errorFromVisibleHand != nil {
		errorToReturn :=
			fmt.Errorf(
				"Unable to retrieve hand or player %v because of error %v",
				actionExecutor.actingPlayer.Name(),
				errorFromVisibleHand)
		return nil, -1, errorToReturn
	}

	handSize := len(playerHand)
	numberOfParticipants := len(actionExecutor.gameParticipants)
	gameRuleset := actionExecutor.gameState.Read().Ruleset()

	if handSize < gameRuleset.NumberOfCardsInPlayerHand(numberOfParticipants) {
		errorToReturn :=
			fmt.Errorf(
				"Player %v could not take a turn because their last turn was already taken",
				actionExecutor.actingPlayer.Name())
		return nil, handSize, errorToReturn
	}

	return playerHand, handSize, nil
}
