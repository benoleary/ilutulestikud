package card_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game/card"
)

func TestNewValidReadonly(unitTest *testing.T) {
	testColor := "test color"
	testIndex := 16
	validReadonly := card.NewReadonly(testColor, testIndex)

	if (validReadonly.ColorSuit() != testColor) ||
		(validReadonly.SequenceIndex() != testIndex) {
		unitTest.Fatalf(
			"NewReadonly(%v, %v) produced unexpected %+v",
			testColor,
			testIndex,
			validReadonly)
	}
}

func TestNewErrorReadonly(unitTest *testing.T) {
	errorReadonly := card.ErrorReadonly()

	if (errorReadonly.ColorSuit() != "error") ||
		(errorReadonly.SequenceIndex() != -1) {
		unitTest.Fatalf(
			"ErrorReadonly() produced unexpected %+v",
			errorReadonly)
	}
}

func TestNewValidInferred(unitTest *testing.T) {
	testColors :=
		[]string{
			"test color",
			"another color",
		}

	testIndices := []int{
		16,
		23,
		0,
	}

	validInferred := card.NewInferred(testColors, testIndices)

	cardColors := validInferred.PossibleColors()
	numberOfExpectedColors := len(testColors)

	if len(cardColors) != numberOfExpectedColors {
		unitTest.Fatalf(
			"NewInferred(%v, %v) produced unexpected %+v",
			testColors,
			testIndices,
			validInferred)
	}

	for colorIndex := 0; colorIndex < numberOfExpectedColors; colorIndex++ {
		if cardColors[colorIndex] != testColors[colorIndex] {
			unitTest.Fatalf(
				"NewInferred(%v, %v) produced unexpected %+v",
				testColors,
				testIndices,
				validInferred)
		}
	}

	cardIndices := validInferred.PossibleIndices()
	numberOfExpectedIndices := len(testIndices)

	if len(cardIndices) != numberOfExpectedIndices {
		unitTest.Fatalf(
			"NewInferred(%v, %v) produced unexpected %+v",
			testColors,
			testIndices,
			validInferred)
	}

	for indexIndex := 0; indexIndex < numberOfExpectedIndices; indexIndex++ {
		if cardIndices[indexIndex] != testIndices[indexIndex] {
			unitTest.Fatalf(
				"NewInferred(%v, %v) produced unexpected %+v",
				testColors,
				testIndices,
				validInferred)
		}
	}
}

func TestNewErrorInferred(unitTest *testing.T) {
	errorInferred := card.ErrorInferred()

	if (errorInferred.PossibleColors() != nil) ||
		(errorInferred.PossibleIndices() != nil) {
		unitTest.Fatalf(
			"ErrorInferred() produced unexpected %+v",
			errorInferred)
	}
}

func TestShuffleReordersAndKeepsCorrectCards(unitTest *testing.T) {
	comparisonDeck :=
		[]card.Readonly{
			card.NewReadonly("a", 6),
			card.NewReadonly("b", 5),
			card.NewReadonly("c", 4),
			card.NewReadonly("d", 3),
			card.NewReadonly("e", 2),
			card.NewReadonly("f", 1),
		}

	expectedNumberOfCards := len(comparisonDeck)

	comparisonMap := make(map[card.Readonly]bool, expectedNumberOfCards)
	shuffledDeck := make([]card.Readonly, 0, expectedNumberOfCards)

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
