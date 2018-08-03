package card

import (
	"math/rand"
)

// Defined encapsulates the state of a single card which should be treated
// as read-only, which in practical terms means alwayys passing by value.
// It has to be an exported struct with only exported data members so
// that it serializes easily.
type Defined struct {
	ColorSuit     string
	SequenceIndex int
}

// Inferred encapsulates the information known to a player about a card
// held by that player. It has to be an exported struct with only exported
// data members so that it serializes easily.
type Inferred struct {
	PossibleColors  []string
	PossibleIndices []int
}

// InHand bundles together a card with the information about it known to
// the player holding it.
type InHand struct {
	Defined
	Inferred
}

// ShuffleInPlace shuffles the given cards in place (using the Fisher-Yates
// algorithm).
func ShuffleInPlace(cardsToShuffle []Defined, randomSeed int64) {
	randomNumberGenerator := rand.New(rand.NewSource(randomSeed))

	// Good ol' Fisher-Yates!
	numberOfUnshuffledCards := len(cardsToShuffle)
	for numberOfUnshuffledCards > 0 {
		indexToMove := randomNumberGenerator.Intn(numberOfUnshuffledCards)

		// We decrement now so that we can use it as the index of the destination
		// of the card chosen to be moved.
		numberOfUnshuffledCards--
		cardsToShuffle[numberOfUnshuffledCards], cardsToShuffle[indexToMove] =
			cardsToShuffle[indexToMove], cardsToShuffle[numberOfUnshuffledCards]
	}
}
