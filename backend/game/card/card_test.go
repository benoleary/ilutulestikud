package card_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game/card"
)

func TestShuffleReordersAndKeepsCorrectCards(unitTest *testing.T) {
	comparisonDeck :=
		[]card.Defined{
			card.Defined{ColorSuit: "a", SequenceIndex: 6},
			card.Defined{ColorSuit: "b", SequenceIndex: 5},
			card.Defined{ColorSuit: "c", SequenceIndex: 4},
			card.Defined{ColorSuit: "d", SequenceIndex: 3},
			card.Defined{ColorSuit: "e", SequenceIndex: 2},
			card.Defined{ColorSuit: "f", SequenceIndex: 1},
		}

	expectedNumberOfCards := len(comparisonDeck)

	comparisonMap := make(map[card.Defined]bool, expectedNumberOfCards)
	shuffledDeck := make([]card.Defined, 0, expectedNumberOfCards)

	for _, cardInDeck := range comparisonDeck {
		shuffledDeck = append(shuffledDeck, cardInDeck)
		comparisonMap[cardInDeck] = true
	}

	seedForTest := int64(7)
	card.ShuffleInPlace(shuffledDeck, seedForTest)

	if len(shuffledDeck) != expectedNumberOfCards {
		unitTest.Fatalf(
			"shuffled deck %+v did not have expected number of cards %v",
			shuffledDeck,
			expectedNumberOfCards)
	}

	numberOfCardsInSamePosition := 0
	for comparisonIndex := 0; comparisonIndex < expectedNumberOfCards; comparisonIndex++ {
		cardFromShuffledDeck := shuffledDeck[comparisonIndex]

		if !comparisonMap[cardFromShuffledDeck] {
			unitTest.Fatalf(
				"shuffled deck %+v had unexpected card %+v not in expected list %+v",
				shuffledDeck,
				cardFromShuffledDeck,
				comparisonDeck)
		}

		if cardFromShuffledDeck == comparisonDeck[comparisonIndex] {
			numberOfCardsInSamePosition++
		}
	}

	// The chosen seed actually leaves no card in the same position, but we allow
	// for some statistical "bad luck" for our test.
	if numberOfCardsInSamePosition > 2 {
		unitTest.Fatalf(
			"shuffled deck %+v left %v cards in the same position as in the original list %+v",
			shuffledDeck,
			numberOfCardsInSamePosition,
			comparisonDeck)
	}
}
