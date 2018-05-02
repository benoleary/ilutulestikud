package game

import (
	"math/rand"
)

// ReadonlyCard should encapsulate the read-only state of a single card.
type ReadonlyCard interface {
	// ColorSuit should be the suit of the card, which should be contained
	// in the set of color suits of the game's ruleset.
	ColorSuit() string

	// SequenceIndex should be the sequence index of the card, which should
	// be contained in the set of sequence indices of the game's ruleset.
	SequenceIndex() int
}

// InferredCard encapsulates the information known to a player about a card
// held by that player.
type InferredCard struct {
	underlyingCard  ReadonlyCard
	possibleColors  []string
	possibleIndices []int
}

// ShuffleInPlace shuffles the given cards in place (using the Fisher-Yates
// algorithm).
func ShuffleInPlace(cardsToShuffle []ReadonlyCard, randomSeed int64) {
	randomNumberGenerator := rand.New(rand.NewSource(randomSeed))

	// Good ol' Fisher-Yates!
	numberOfUnshuffledCards := len(cardsToShuffle)
	for numberOfUnshuffledCards > 0 {
		indexToMove := randomNumberGenerator.Intn(numberOfUnshuffledCards)

		// We decrement now so that we can use it as the index of the destination
		// of the card chosen to be moved.
		numberOfUnshuffledCards -= 1
		cardsToShuffle[numberOfUnshuffledCards], cardsToShuffle[indexToMove] =
			cardsToShuffle[indexToMove], cardsToShuffle[numberOfUnshuffledCards]
	}
}

// BySequenceIndex implements sort interface for []ReadonlyCard based on the return
// from its SequenceIndex(), ignoring its ColorSuit(). It is exported for ease of
// testing.
type BySequenceIndex []ReadonlyCard

// Len implements part of the sort interface for BySequenceIndex.
func (bySequenceIndex BySequenceIndex) Len() int {
	return len(bySequenceIndex)
}

// Swap implements part of the sort interface for BySequenceIndex.
func (bySequenceIndex BySequenceIndex) Swap(firstIndex int, secondIndex int) {
	bySequenceIndex[firstIndex], bySequenceIndex[secondIndex] =
		bySequenceIndex[secondIndex], bySequenceIndex[firstIndex]
}

// Less implements part of the sort interface for BySequenceIndex.
func (bySequenceIndex BySequenceIndex) Less(firstIndex int, secondIndex int) bool {
	return bySequenceIndex[firstIndex].SequenceIndex() < bySequenceIndex[secondIndex].SequenceIndex()
}

type simpleCard struct {
	colorSuit     string
	sequenceIndex int
}

func (singleCard *simpleCard) ColorSuit() string {
	return singleCard.colorSuit
}

func (singleCard *simpleCard) SequenceIndex() int {
	return singleCard.sequenceIndex
}
