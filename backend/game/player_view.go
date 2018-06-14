package game

import (
	"fmt"

	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
)

// PlayerView encapsulates the functions on a game's read-only state
// which provide the information available to a particular player for
// that state.
type PlayerView struct {
	gameState               ReadonlyState
	gameParticipants        []string
	numberOfParticipants    int
	playerName              string
	gameRuleset             Ruleset
	colorSuits              []string
	numberOfSuits           int
	distinctPossibleIndices []int
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
			distinctPossibleIndices := rulesetOfGame.DistinctPossibleIndices()

			playerView :=
				&PlayerView{
					gameState:               stateOfGame,
					gameParticipants:        participantsInGame,
					numberOfParticipants:    numberOfPlayers,
					playerName:              nameOfPlayer,
					gameRuleset:             rulesetOfGame,
					colorSuits:              gameColorSuits,
					numberOfSuits:           len(gameColorSuits),
					distinctPossibleIndices: distinctPossibleIndices,
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

// ChatLog just wraps around the read-only game state's ChatLog function.
func (playerView *PlayerView) ChatLog() []message.Readonly {
	return playerView.gameState.ChatLog()
}

// ActionLog just wraps around the read-only game state's ActionLog function.
func (playerView *PlayerView) ActionLog() []message.Readonly {
	return playerView.gameState.ActionLog()
}

// CurrentTurnOrder returns the names of the participants of the game in the
// order which their next turns are in, along with the index of the viewing
// player in that list.
func (playerView *PlayerView) CurrentTurnOrder() ([]string, int) {
	playerNamesInTurnOrder := make([]string, playerView.numberOfParticipants)

	gameTurn := playerView.gameState.Turn()
	playerIndexInTurnOrder := -1

	for currentTurnIndex := 0; currentTurnIndex < playerView.numberOfParticipants; currentTurnIndex++ {
		// Game turns begin with 1 rather than 0, so this sets the player names in order,
		// wrapping index back to 0 when at the end of the list.
		// E.g. turn 3, 5 players: playerNamesInTurnOrder will start with
		// gameParticipants[2], then [3], then [4], then [0], then [1].
		indexInOriginalOrder := (currentTurnIndex + gameTurn - 1) % playerView.numberOfParticipants
		playerInTurnOrder := playerView.gameParticipants[indexInOriginalOrder]
		playerNamesInTurnOrder[currentTurnIndex] = playerInTurnOrder

		if playerView.playerName == playerInTurnOrder {
			playerIndexInTurnOrder = currentTurnIndex
		}
	}

	return playerNamesInTurnOrder, playerIndexInTurnOrder
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
	return maximumNumber - playerView.NumberOfReadyHints()
}

// NumberOfMistakesStillAllowed just subtracts the read-only game state's
// NumberOfMistakesMade function's return value from the constant maximum.
func (playerView *PlayerView) NumberOfMistakesStillAllowed() int {
	maximumNumber := playerView.gameState.Ruleset().MaximumNumberOfMistakesAllowed()
	return maximumNumber - playerView.NumberOfMistakesMade()
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

// PlayedCards lists the cards in play, in slices per suit.
func (playerView *PlayerView) PlayedCards() [][]card.Readonly {
	playedCards := make([][]card.Readonly, playerView.numberOfSuits)

	for suitIndex := 0; suitIndex < playerView.numberOfSuits; suitIndex++ {
		suitColor := playerView.colorSuits[suitIndex]

		cardsPlayedForSuit :=
			playerView.gameState.PlayedForColor(suitColor)

		playedCards[suitIndex] = cardsPlayedForSuit
	}

	return playedCards
}

// DiscardedCards lists the discarded cards, ordered by suit first then by index.
func (playerView *PlayerView) DiscardedCards() []card.Readonly {
	discardedCards := make([]card.Readonly, 0)

	for _, colorSuit := range playerView.colorSuits {
		for _, sequenceIndex := range playerView.distinctPossibleIndices {
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

	return playerView.gameState.VisibleHand(playerName)
}

// KnowledgeOfOwnHand returns the knowledge about the player's own cards which
// was inferred directly from the hints officially given so far.
func (playerView *PlayerView) KnowledgeOfOwnHand() ([]card.Inferred, error) {
	return playerView.gameState.InferredHand(playerView.playerName)
}
