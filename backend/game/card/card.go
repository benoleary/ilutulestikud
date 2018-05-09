package card

import (
	"math/rand"
)

// Readonly encapsulates the read-only state of a single card.
type Readonly struct {
	colorSuit     string
	sequenceIndex int
}

// NewReadonly returns a new Readonly card.
func NewReadonly(
	colorSuit string,
	sequenceIndex int) Readonly {
	return Readonly{
		colorSuit:     colorSuit,
		sequenceIndex: sequenceIndex,
	}
}

// ErrorReadonly returns a card signalling that there was an error.
func ErrorReadonly() Readonly {
	return NewReadonly("error", -1)
}

// ColorSuit returns the suit of the card, which should be contained
// in the set of color suits of the game's ruleset.
func (readonlyCard *Readonly) ColorSuit() string {
	return readonlyCard.colorSuit
}

// SequenceIndex returns the sequence index of the card, which should
// be contained in the set of sequence indices of the game's ruleset.
func (readonlyCard *Readonly) SequenceIndex() int {
	return readonlyCard.sequenceIndex
}

// Inferred encapsulates the information known to a player about a card
// held by that player.
type Inferred struct {
	underlyingCard  Readonly
	possibleColors  []string
	possibleIndices []int
}

// PossibleColors returns the color suits which this card could have and
// have not yet been eliminated by hints .
func (inferredCard *Inferred) PossibleColors() []string {
	return inferredCard.possibleColors
}

// PossibleIndices returns the sequence indices which this card could have
// and have not yet been eliminated by hints .
func (inferredCard *Inferred) PossibleIndices() []int {
	return inferredCard.possibleIndices
}

// ShuffleInPlace shuffles the given cards in place (using the Fisher-Yates
// algorithm).
func ShuffleInPlace(cardsToShuffle []Readonly, randomSeed int64) {
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
