package game

import (
	"fmt"

	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/chat"
)

// PlayerView encapsulates the functions on a game's read-only state
// which provide the information available to a particular player for
// that state.
type PlayerView struct {
	gameState            ReadonlyState
	gameParticipants     []string
	numberOfParticipants int
	playerName           string
	gameRuleset          Ruleset
	colorSuits           []string
	numberOfSuits        int
	sequenceIndices      []int
	handSize             int
}

// ViewOnStateForPlayer creates a PlayerView around the given game
// state if the given player is a participant, returning a pointer to
// the view. If the player is not a participant, it returns nil
// along with an error.
func ViewOnStateForPlayer(
	stateOfGame ReadonlyState,
	nameOfPlayer string) (ViewForPlayer, error) {
	participantsInGame := stateOfGame.PlayerNames()
	for _, gameParticipant := range participantsInGame {
		if gameParticipant == nameOfPlayer {
			numberOfPlayers := len(participantsInGame)
			rulesetOfGame := stateOfGame.Ruleset()
			gameColorSuits := rulesetOfGame.ColorSuits()

			playerView :=
				&PlayerView{
					gameState:            stateOfGame,
					gameParticipants:     participantsInGame,
					numberOfParticipants: numberOfPlayers,
					playerName:           nameOfPlayer,
					gameRuleset:          rulesetOfGame,
					colorSuits:           gameColorSuits,
					numberOfSuits:        len(gameColorSuits),
					sequenceIndices:      rulesetOfGame.SequenceIndices(),
					handSize:             rulesetOfGame.NumberOfCardsInPlayerHand(numberOfPlayers),
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

// RulesetDescription returns the description given by the ruleset of the game.
func (playerView *PlayerView) RulesetDescription() string {
	return playerView.gameState.Ruleset().FrontendDescription()
}

// SortedChatLog sorts the read-only game state's ChatLog and returns the sorted log.
func (playerView *PlayerView) SortedChatLog() []chat.Message {
	return playerView.gameState.ChatLog().Sorted()
}

// CurrentTurnOrder returns the names of the participants of the game in the
// order which their next turns are in, along with true if the view is for
// the first player in that list or false otherwise.
func (playerView *PlayerView) CurrentTurnOrder() ([]string, bool) {
	playerNamesInTurnOrder := make([]string, playerView.numberOfParticipants)

	gameTurn := playerView.gameState.Turn()
	isPlayerTurn := false
	for playerIndex := 0; playerIndex < playerView.numberOfParticipants; playerIndex++ {
		// Game turns begin with 1 rather than 0, so this sets the player names in order,
		// wrapping index back to 0 when at the end of the list.
		// E.g. turn 3, 5 players: playerNamesInTurnOrder will start with
		// gameParticipants[2], then [3], then [4], then [0], then [1].
		playerIndex := (playerIndex + gameTurn - 1) % playerView.numberOfParticipants
		playerInTurnOrder := playerView.gameParticipants[playerIndex]
		playerNamesInTurnOrder[playerIndex] = playerInTurnOrder

		if playerView.playerName == playerInTurnOrder {
			isPlayerTurn = true
		}
	}

	return playerNamesInTurnOrder, isPlayerTurn
}

// Turn just wraps around the read-only game state's Turn function.
func (playerView *PlayerView) Turn() int {
	return playerView.gameState.Turn()
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

// DeckSize just wraps around the read-only game state's DeckSize function.
func (playerView *PlayerView) DeckSize() int {
	return playerView.gameState.DeckSize()
}

// TopmostPlayedCards lists the top-most cards in play for each suit, leaving out
// any color suits which have no cards in play yet.
func (playerView *PlayerView) TopmostPlayedCards() []card.Readonly {
	topmostCards := make([]card.Readonly, 0)

	for suitIndex := 0; suitIndex < playerView.numberOfSuits; suitIndex++ {
		suitColor := playerView.colorSuits[suitIndex]

		topmostCardOfSuit, hasSuitAnyPlayedCards :=
			playerView.gameState.LastPlayedForColor(suitColor)

			// If a card has been played for the suit, we add it to the list.
		if hasSuitAnyPlayedCards {
			topmostCards = append(topmostCards, topmostCardOfSuit)
		}
	}

	return topmostCards
}

// DiscardedCards lists the discarded cards, ordered by suit first then by index.
func (playerView *PlayerView) DiscardedCards() []card.Readonly {
	discardedCards := make([]card.Readonly, 0)

	for _, colorSuit := range playerView.colorSuits {
		for _, sequenceIndex := range playerView.sequenceIndices {
			numberOfDiscardedCopies :=
				playerView.gameState.NumberOfDiscardedCards(colorSuit, sequenceIndex)
			discardedCard := card.NewReadonly(colorSuit, sequenceIndex)

			for copiesCount := 0; copiesCount < numberOfDiscardedCopies; copiesCount++ {
				discardedCards = append(discardedCards, discardedCard)
			}
		}
	}

	return discardedCards
}

// VisibleHand returns the cards held by the given player, or nil and an error if
// the player cannot see the cards.
func (playerView *PlayerView) VisibleHand(playerName string) ([]card.Readonly, error) {
	if playerName == playerView.playerName {
		return nil, fmt.Errorf("Player is not allowed to view own hand")
	}

	playerHand := make([]card.Readonly, playerView.handSize)

	for indexInHand := 0; indexInHand < playerView.handSize; indexInHand++ {
		visibleCard, errorFromView :=
			playerView.gameState.VisibleCardInHand(playerName, indexInHand)

		if errorFromView != nil {
			return nil, errorFromView
		}

		playerHand[indexInHand] = visibleCard
	}

	return playerHand, nil
}

// KnowledgeOfOwnHand returns the knowledge about the player's own cards which
// was inferred directly from the hints officially given so far.
func (playerView *PlayerView) KnowledgeOfOwnHand() ([]card.Inferred, error) {
	playerHand := make([]card.Inferred, playerView.handSize)

	for indexInHand := 0; indexInHand < playerView.handSize; indexInHand++ {
		inferredCard, errorFromInferral :=
			playerView.gameState.InferredCardInHand(playerView.playerName, indexInHand)

		if errorFromInferral != nil {
			return nil, errorFromInferral
		}

		playerHand[indexInHand] = inferredCard
	}

	return playerHand, nil
}
