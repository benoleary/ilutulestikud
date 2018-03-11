package game

// This file contains some constants and functions for setting up the
// game according to the standard ruleset. It could become an interface
// if there is ever a call for allowing the players to choose some
// rules when setting up a game.

// MinimumNumberOfPlayers is the minimum number of players for a game.
const MinimumNumberOfPlayers = 2

// MaximumNumberOfPlayers is the maximum number of players for a game.
const MaximumNumberOfPlayers = 5

// MaximumNumberOfHints is the maximum number of hints which can be
// ready to use in a game at any time.
const MaximumNumberOfHints = 8

// MaximumNumberOfMistakesAllowed is the maximum number of mistakes
// that can be made without the game ending (i.e. the game ends on the
// third mistake).
const MaximumNumberOfMistakesAllowed = 2

// NumberOfCardsInPlayerHand is the number of cards held in a player's
// hand, dependent on the number of players in the game.
func NumberOfCardsInPlayerHand(numberOfPlayers int) int {
	if numberOfPlayers <= 3 {
		return 5
	}

	return 4
}

// ColorSuits returns the set of colors used as suits.
func ColorSuits(includeRainbow bool) []string {
	basicSuits := []string{
		"red",
		"green",
		"blue",
		"yellow",
		"white",
	}

	if includeRainbow {
		return append(basicSuits, "rainbow")
	}

	return basicSuits
}

// SequenceIndices returns all the indices for the cards, per card so
// including repetitions of indices, as they should be played per suit.
func SequenceIndices() []int {
	return []int{1, 1, 1, 2, 2, 3, 3, 4, 4, 5}
}

// PointsPerCard returns the points value of a card with the given
// sequence index.
func PointsPerCard(cardSequenceIndex int) int {
	if cardSequenceIndex >= 5 {
		return 2 * cardSequenceIndex
	}

	return cardSequenceIndex
}
