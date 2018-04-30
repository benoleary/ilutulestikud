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

OK, I need:
CardDeck (gives out card at [0], sets own [0] to nil, re-sets slice over array, something for when empty)
DiscardArea (stores ordered lists of cards per suit (does sorting), ensures only ruleset suits allowed)
PlayedArea (stores ordered lists of cards per suit (only allows increasing sequences), ensures only ruleset suits allowed, in charge of whether play is legal)
PlayerHand (stores inferred cards (shown when viewer is not holder), gives out ReadonlyCard in exchange for substitute (in charge of wrapping in InferredCard), something when out of cards in deck)
InferredCard (has ReadonlyCard, has list of possible suits, has list of possible indices)

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
