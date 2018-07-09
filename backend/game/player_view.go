package game

import (
	"fmt"

	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
)

// PlayerView encapsulates the functions on a game's read-only state
// which provide the information available to a particular player for
// that state. It has a few quantities which should be derived on
// construction.
type PlayerView struct {
	gameState               ReadonlyState
	playerProvider          ReadonlyPlayerProvider
	gameParticipants        []string
	numberOfParticipants    int
	playerName              string
	gameRuleset             Ruleset
	colorSuits              []string
	distinctPossibleIndices []int
	playedCards             [][]card.Readonly
}

// ViewOnStateForPlayer creates a PlayerView around the given game
// state if the given player is a participant, returning a pointer to
// the view. If the player is not a participant, it returns nil
// along with an error.
func ViewOnStateForPlayer(
	stateOfGame ReadonlyState,
	playerProvider ReadonlyPlayerProvider,
	nameOfPlayer string) (ViewForPlayer, error) {
	participantsInGame := stateOfGame.PlayerNames()
	for _, gameParticipant := range participantsInGame {
		if gameParticipant == nameOfPlayer {
			numberOfPlayers := len(participantsInGame)
			rulesetOfGame := stateOfGame.Ruleset()
			gameColorSuits := rulesetOfGame.ColorSuits()
			distinctPossibleIndices := rulesetOfGame.DistinctPossibleIndices()

			numberOfSuits := len(gameColorSuits)
			playedCards := make([][]card.Readonly, numberOfSuits)
			for suitIndex := 0; suitIndex < numberOfSuits; suitIndex++ {
				suitColor := gameColorSuits[suitIndex]

				cardsPlayedForSuit := stateOfGame.PlayedForColor(suitColor)

				playedCards[suitIndex] = cardsPlayedForSuit
			}

			playerView :=
				&PlayerView{
					gameState:               stateOfGame,
					playerProvider:          playerProvider,
					gameParticipants:        participantsInGame,
					numberOfParticipants:    numberOfPlayers,
					playerName:              nameOfPlayer,
					gameRuleset:             rulesetOfGame,
					colorSuits:              gameColorSuits,
					distinctPossibleIndices: distinctPossibleIndices,
					playedCards:             playedCards,
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

// GameIsFinished returns true if the game is finished because either too many
// mistakes have been made, or if there are no more cards to draw and the turn
// is for a player with a hand which is smaller than the full hand size for the
// ruleset (which indicates that all players have played a card after the deck
// was exhausted).
func (playerView *PlayerView) GameIsFinished() (bool, error) {
	if playerView.gameOverBecauseOfMistakes() {
		return true, nil
	}

	if playerView.DeckSize() > 0 {
		return false, nil
	}

	fullHandSize :=
		playerView.gameRuleset.NumberOfCardsInPlayerHand(playerView.numberOfParticipants)

	currentPlayer := playerView.gameParticipants[playerView.playerIndexForTurn(0)]
	inferredHand, errorFromState := playerView.gameState.InferredHand(currentPlayer)

	if errorFromState != nil {
		return true, errorFromState
	}

	return (len(inferredHand) < fullHandSize), nil
}

// CurrentTurnOrder returns the names of the participants of the game in the
// order which their next turns are in, along with the index of the viewing
// player in that list.
func (playerView *PlayerView) CurrentTurnOrder() ([]string, int) {
	playerNamesInTurnOrder := make([]string, playerView.numberOfParticipants)
	playerIndexInTurnOrder := -1

	for turnsAfterCurrent := 0; turnsAfterCurrent < playerView.numberOfParticipants; turnsAfterCurrent++ {
		indexInOriginalOrder := playerView.playerIndexForTurn(turnsAfterCurrent)
		playerInTurnOrder := playerView.gameParticipants[indexInOriginalOrder]
		playerNamesInTurnOrder[turnsAfterCurrent] = playerInTurnOrder

		if playerView.playerName == playerInTurnOrder {
			playerIndexInTurnOrder = turnsAfterCurrent
		}
	}

	return playerNamesInTurnOrder, playerIndexInTurnOrder
}

// Turn just wraps around the read-only game state's Turn function.
func (playerView *PlayerView) Turn() int {
	return playerView.gameState.Turn()
}

// Score derives the score from the cards in the played area.
func (playerView *PlayerView) Score() int {
	if playerView.gameOverBecauseOfMistakes() {
		return 0
	}

	scoreSoFar := 0
	for _, playedPile := range playerView.playedCards {
		for _, playedCard := range playedPile {
			scoreSoFar += playerView.gameRuleset.PointsForCard(playedCard)
		}
	}

	return scoreSoFar
}

// NumberOfReadyHints just wraps around the read-only game state's
// NumberOfReadyHints function.
func (playerView *PlayerView) NumberOfReadyHints() int {
	return playerView.gameState.NumberOfReadyHints()
}

// MaximumNumberOfHints just wraps around the game's ruleset's maximum
// number of hints.
func (playerView *PlayerView) MaximumNumberOfHints() int {
	return playerView.gameState.Ruleset().MaximumNumberOfHints()
}

// ColorsAvailableAsHint just wraps around the function returning the
// color suits available for hints from the game's ruleset.
func (playerView *PlayerView) ColorsAvailableAsHint() []string {
	return playerView.gameState.Ruleset().ColorsAvailableAsHint()
}

// IndicesAvailableAsHint just wraps around the function returning the
// sequence indices available for hints from the game's ruleset.
func (playerView *PlayerView) IndicesAvailableAsHint() []int {
	return playerView.gameState.Ruleset().IndicesAvailableAsHint()
}

// NumberOfMistakesMade just wraps around the read-only game state's
// NumberOfMistakesMade function.
func (playerView *PlayerView) NumberOfMistakesMade() int {
	return playerView.gameState.NumberOfMistakesMade()
}

// NumberOfMistakesIndicatingGameOver just wraps around the game's
// ruleset's NumberOfMistakesIndicatingGameOver.
func (playerView *PlayerView) NumberOfMistakesIndicatingGameOver() int {
	return playerView.gameState.Ruleset().NumberOfMistakesIndicatingGameOver()
}

// DeckSize just wraps around the read-only game state's DeckSize function.
func (playerView *PlayerView) DeckSize() int {
	return playerView.gameState.DeckSize()
}

// PlayedCards lists the cards in play, in slices per suit.
func (playerView *PlayerView) PlayedCards() [][]card.Readonly {
	return playerView.playedCards
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

// VisibleHand returns the cards held by the given player along with the chat color for
// that player, or nil and a string which will be ignored and an error if the player
// cannot see the cards.
func (playerView *PlayerView) VisibleHand(playerName string) ([]card.Readonly, string, error) {
	if playerName == playerView.playerName {
		return nil, "no color because of error", fmt.Errorf("Player is not allowed to view own hand")
	}

	playerState, errorFromPlayerProvider := playerView.playerProvider.Get(playerName)
	if errorFromPlayerProvider != nil {
		return nil, "no color because of error", errorFromPlayerProvider
	}

	visibleCards, errorFromGameState := playerView.gameState.VisibleHand(playerName)

	return visibleCards, playerState.Color(), errorFromGameState
}

// KnowledgeOfOwnHand returns the knowledge about the player's own cards which
// was inferred directly from the hints officially given so far.
func (playerView *PlayerView) KnowledgeOfOwnHand() ([]card.Inferred, error) {
	return playerView.gameState.InferredHand(playerView.playerName)
}

func (playerView *PlayerView) playerIndexForTurn(turnsAfterCurrent int) int {
	// Game turn indices begin with 1 rather than 0, so this returns the
	// index of the player for the turn with the given offset from the current turn,
	// wrapping back to 0 when the turn index is greater than the number of players.
	// E.g. turn 3 for a game with 5 players: called with 0, 1, 2, 3, 4, the return
	// values will be 2, then 3, then 4, then 0, then 1.
	turnIndexFromZero := (turnsAfterCurrent + playerView.gameState.Turn() - 1)
	return turnIndexFromZero % playerView.numberOfParticipants
}

func (playerView *PlayerView) gameOverBecauseOfMistakes() bool {
	return playerView.NumberOfMistakesMade() >=
		playerView.NumberOfMistakesIndicatingGameOver()
}
