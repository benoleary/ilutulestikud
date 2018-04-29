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

// OrderedCardset bundles typical operations on ordered sets of cards which
// represent the actions typically performed on physical sets of cards for
// the game.
type OrderedCardset struct {
	cardList []ReadonlyCard
}

// ReplaceCard should do something.
func (orderedCardset OrderedCardset) ReplaceCard(cardIndex int, cardToInsert ReadonlyCard) (ReadonlyCard, error) {
	// something
}

// RemoveCard removes the card with the given index from the list and returns
// it, or nil and an error are not enough cards.
func (orderedCardset OrderedCardset) RemoveCard(cardIndex int) (ReadonlyCard, error) {
	// something
}

// ShuffleCards shuffles the cards in place (using the Fisher-Yates
// algorithm).
func (orderedCardset OrderedCardset) ShuffleCards(randomSeed int64) {
	randomNumberGenerator := rand.New(rand.NewSource(randomSeed))
	cardsToShuffle := orderedCardset.cardList

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
