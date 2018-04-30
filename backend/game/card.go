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

type CardDeck struct {
	cardsInDeck []ReadonlyCard
}

// DrawFromTop returns the first card in the deck and removes the reference
// to it from the deck.
func (cardDeck *CardDeck) DrawFromTop() (*ReadonlyCard, error) {
	if len(cardDeck.cardsInDeck) <= 0 {
		return nil, fmt.Errorf("No cards left to draw")
	}

	drawnCard := cardDeck.cardsInDeck[0]
	cardDeck.cardsInDeck[0] = nil
	cardDeck.cardsInDeck = cardDeck.cardsInDeck[1:]

	return drawnCard
}

// Ruleset has to manipulate this.
type DiscardArea struct {
	discardedCards map[string][]ReadonlyCard
}

// Ruleset has to manipulate this.
func (discardArea DiscardArea) AddToPile(discardedCard ReadonlyCard) {
	colorPile, _ := discardArea.discardedCards[discardedCard.ColorSuit()]
	sort.Sort(BySequenceIndex(colorPile))
	discardArea.discardedCards[discardedCard.ColorSuit()] =
		append(colorPile, discardedCard)
}

// Ruleset has to manipulate this.
type PlayedArea struct {
	playedCards map[string][]ReadonlyCard
}

func NewDeckAndAreas(sourceCardset []ReadonlyCard) (*CardDeck, *DiscardArea, *PlayedArea) {
	copyCardset := make([]ReadonlyCard, len(sourceCardset))
	copy(copyCardset, sourceCardset)

	cardDeck := &CardDeck{
		cardsInDeck: copyCardset,
	}

	discardArea := &DiscardArea{
		discardedCards: make(map[string][]ReadonlyCard, 0),
	}

	playedArea := &PlayedArea{
		discardedCards: make(map[string][]ReadonlyCard, 0),
	}

	return cardDeck, discardArea, playedArea
}

OK, I need:
PlayedArea (stores ordered lists of cards per suit (only allows increasing sequences))
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
