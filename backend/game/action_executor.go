package game

import (
	"context"
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
	creationContext context.Context,
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
func (actionExecutor *ActionExecutor) RecordChatMessage(
	executionContext context.Context,
	chatMessage string) error {
	return actionExecutor.gameState.RecordChatMessage(
		executionContext,
		actionExecutor.actingPlayer,
		chatMessage)
}

// TakeTurnByDiscarding enacts a turn by discarding the indicated card from the hand
// of the acting player, or returns an error if it was not possible.
func (actionExecutor *ActionExecutor) TakeTurnByDiscarding(
	executionContext context.Context,
	indexInHand int) error {
	// First we must determine if the player is allowed to take an action.
	discardedCard, errorFromHand :=
		actionExecutor.cardFromHandIfTurnElseError(indexInHand)

	if errorFromHand != nil {
		return errorFromHand
	}

	gameReadState := actionExecutor.gameState.Read()
	gameRuleset := gameReadState.Ruleset()
	replacementCard :=
		card.Inferred{
			PossibleColors:  gameRuleset.ColorSuits(),
			PossibleIndices: gameRuleset.DistinctPossibleIndices(),
		}

	actionMessage :=
		fmt.Sprintf(
			"discards card %v %v",
			discardedCard.ColorSuit,
			discardedCard.SequenceIndex)

	// The logic for determining how many hints to provide per discarded card could be
	// given over to the ruleset, but it's not much of an issue.
	numberOfHintsToAdd := 0
	if gameReadState.NumberOfReadyHints() < gameRuleset.MaximumNumberOfHints() {
		numberOfHintsToAdd = 1
	}

	return actionExecutor.gameState.EnactTurnByDiscardingAndReplacing(
		executionContext,
		actionMessage,
		actionExecutor.actingPlayer,
		indexInHand,
		replacementCard,
		numberOfHintsToAdd,
		0)
}

// TakeTurnByPlaying enacts a turn by attempting to play the indicated card from the hand
// of the acting player, resulting in the card going into the played area or into the
// discard pile while causing a mistake, or returns an error if it was not possible.
func (actionExecutor *ActionExecutor) TakeTurnByPlaying(
	executionContext context.Context,
	indexInHand int) error {
	// First we must determine if the player is allowed to take an action.
	selectedCard, errorFromHand :=
		actionExecutor.cardFromHandIfTurnElseError(indexInHand)

	if errorFromHand != nil {
		return errorFromHand
	}

	gameReadState := actionExecutor.gameState.Read()
	gameRuleset := gameReadState.Ruleset()
	replacementCard :=
		card.Inferred{
			PossibleColors:  gameRuleset.ColorSuits(),
			PossibleIndices: gameRuleset.DistinctPossibleIndices(),
		}

	playedCards := gameReadState.PlayedForColor(selectedCard.ColorSuit)

	if !gameRuleset.IsCardPlayable(selectedCard, playedCards) {
		actionMessage :=
			fmt.Sprintf(
				"mistakenly tries to play card %v %v",
				selectedCard.ColorSuit,
				selectedCard.SequenceIndex)

		return actionExecutor.gameState.EnactTurnByDiscardingAndReplacing(
			executionContext,
			actionMessage,
			actionExecutor.actingPlayer,
			indexInHand,
			replacementCard,
			0,
			1)
	}

	actionMessage :=
		fmt.Sprintf(
			"successfully plays card %v %v",
			selectedCard.ColorSuit,
			selectedCard.SequenceIndex)

	numberOfHintsToAdd := gameRuleset.HintsForPlayingCard(selectedCard)
	maximumNumberOfHintsWhichCouldBeAdded :=
		gameRuleset.MaximumNumberOfHints() - gameReadState.NumberOfReadyHints()
	if numberOfHintsToAdd > maximumNumberOfHintsWhichCouldBeAdded {
		numberOfHintsToAdd = maximumNumberOfHintsWhichCouldBeAdded
	}

	return actionExecutor.gameState.EnactTurnByPlayingAndReplacing(
		executionContext,
		actionMessage,
		actionExecutor.actingPlayer,
		indexInHand,
		replacementCard,
		numberOfHintsToAdd)
}

// TakeTurnByHintingColor enacts a turn by giving a hint to the receiving player
// about a color suit with respect to the receiver's hand, or return an error if
// it was not possible.
func (actionExecutor *ActionExecutor) TakeTurnByHintingColor(
	executionContext context.Context,
	receivingPlayer string,
	hintedColor string) error {
	visibleHandOfReceiver, inferredHandOfReceiverBeforeHint, errorFromHand :=
		actionExecutor.handOfHintReceiver(receivingPlayer)

	if errorFromHand != nil {
		return errorFromHand
	}

	inferredHandOfReceiverAfterHint :=
		actionExecutor.gameState.Read().Ruleset().AfterColorHint(
			inferredHandOfReceiverBeforeHint,
			visibleHandOfReceiver,
			hintedColor)

	actionMessage :=
		fmt.Sprintf(
			"gives hint to %v about color %v",
			receivingPlayer,
			hintedColor)

	return actionExecutor.gameState.EnactTurnByUpdatingHandWithHint(
		executionContext,
		actionMessage,
		actionExecutor.actingPlayer,
		receivingPlayer,
		inferredHandOfReceiverAfterHint,
		1)
}

// TakeTurnByHintingIndex enacts a turn by giving a hint to the receiving player
// about a sequence index with respect to the receiver's hand, or return an error
// if it was not possible.
func (actionExecutor *ActionExecutor) TakeTurnByHintingIndex(
	executionContext context.Context,
	receivingPlayer string,
	hintedIndex int) error {
	visibleHandOfReceiver, inferredHandOfReceiverBeforeHint, errorFromHand :=
		actionExecutor.handOfHintReceiver(receivingPlayer)

	if errorFromHand != nil {
		return errorFromHand
	}

	inferredHandOfReceiverAfterHint :=
		actionExecutor.gameState.Read().Ruleset().AfterIndexHint(
			inferredHandOfReceiverBeforeHint,
			visibleHandOfReceiver,
			hintedIndex)

	actionMessage :=
		fmt.Sprintf(
			"gives hint to %v about number %v",
			receivingPlayer,
			hintedIndex)

	return actionExecutor.gameState.EnactTurnByUpdatingHandWithHint(
		executionContext,
		actionMessage,
		actionExecutor.actingPlayer,
		receivingPlayer,
		inferredHandOfReceiverAfterHint,
		1)
}

func (actionExecutor *ActionExecutor) handOfHintReceiver(
	receivingPlayer string) ([]card.Defined, []card.Inferred, error) {
	if receivingPlayer == actionExecutor.actingPlayer.Name() {
		errorToReturn :=
			fmt.Errorf(
				"Player %v cannot give a hint to self",
				actionExecutor.actingPlayer.Name())
		return nil, nil, errorToReturn
	}

	readonlyGame := actionExecutor.gameState.Read()
	if readonlyGame.NumberOfReadyHints() <= 0 {
		return nil, nil, fmt.Errorf("No hints available to use")
	}

	// First we must determine if the player is allowed to take an action,
	// though we do not need to see the hand - it just has to be found to
	// determine if the game is not yet over.
	_, errorFromHinterHand := actionExecutor.playerHandIfTurnElseError()

	if errorFromHinterHand != nil {
		return nil, nil, errorFromHinterHand
	}

	receiverKnowledgeOfOwnHand, errorFromReceiverInferredHand :=
		readonlyGame.InferredHand(receivingPlayer)

	if errorFromReceiverInferredHand != nil {
		return nil, nil, errorFromReceiverInferredHand
	}

	visibleHandOfReceiver, errorFromReceiverVisibleHand :=
		readonlyGame.VisibleHand(receivingPlayer)

	if errorFromReceiverVisibleHand != nil {
		return nil, nil, errorFromReceiverVisibleHand
	}

	return visibleHandOfReceiver, receiverKnowledgeOfOwnHand, nil
}

func (actionExecutor *ActionExecutor) playerHandIfTurnElseError() ([]card.Defined, error) {
	gameReadState := actionExecutor.gameState.Read()
	if IsFinished(gameReadState) {
		return nil, fmt.Errorf("Game is finished, cannot take turn")
	}

	// The turn number starts from 1.
	indexOfPlayerForCurrentTurn :=
		(gameReadState.Turn() - 1) % len(actionExecutor.gameParticipants)
	playerForCurrentTurn := actionExecutor.gameParticipants[indexOfPlayerForCurrentTurn]

	if playerForCurrentTurn != actionExecutor.actingPlayer.Name() {
		errorToReturn :=
			fmt.Errorf(
				"Player %v is not the current player (%v) so cannot take a turn",
				actionExecutor.actingPlayer.Name(),
				playerForCurrentTurn)

		return nil, errorToReturn
	}

	playerHand, errorFromVisibleHand :=
		gameReadState.VisibleHand(actionExecutor.actingPlayer.Name())

	if errorFromVisibleHand != nil {
		errorToReturn :=
			fmt.Errorf(
				"Unable to retrieve hand or player %v because of error %v",
				actionExecutor.actingPlayer.Name(),
				errorFromVisibleHand)

		return nil, errorToReturn
	}

	return playerHand, nil
}

func (actionExecutor *ActionExecutor) cardFromHandIfTurnElseError(
	indexInHand int) (card.Defined, error) {
	playerHand, errorFromGettingHand := actionExecutor.playerHandIfTurnElseError()

	if errorFromGettingHand != nil {
		invalidCard := card.Defined{ColorSuit: errorFromGettingHand.Error(), SequenceIndex: -1}
		return invalidCard, errorFromGettingHand
	}

	handSize := len(playerHand)
	if (indexInHand < 0) || (indexInHand >= handSize) {
		errorFromOutOfRange := fmt.Errorf(
			"Index %v is out of the acceptable range %v to %v of the player's hand",
			indexInHand,
			0,
			handSize)

		invalidCard := card.Defined{ColorSuit: errorFromOutOfRange.Error(), SequenceIndex: -1}
		return invalidCard, errorFromOutOfRange
	}

	return playerHand[indexInHand], nil
}
