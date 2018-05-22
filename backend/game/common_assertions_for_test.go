package game_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game/card"
)

func assertReadonlyCardSlicesMatch(
	testIdentifier string,
	unitTest *testing.T,
	actualCards []card.Readonly,
	expectedCards []card.Readonly) {
	numberOfExpectedCards := len(expectedCards)

	if len(actualCards) != numberOfExpectedCards {
		unitTest.Fatalf(
			testIdentifier+
				"/Card slice lengths do not match - actual: %v; expected %v",
			actualCards,
			expectedCards)
	}

	for cardIndex := 0; cardIndex < numberOfExpectedCards; cardIndex++ {
		actualCard := actualCards[cardIndex]
		expectedCard := expectedCards[cardIndex]
		if (actualCard.ColorSuit() != expectedCard.ColorSuit()) ||
			(actualCard.SequenceIndex() != expectedCard.SequenceIndex()) {
			unitTest.Fatalf(
				testIdentifier+
					"/Card slices do not match in element %v - actual: %v; expected %v"+
					" - actual array %v; expected array %v",
				cardIndex,
				actualCard,
				expectedCard,
				actualCards,
				expectedCards)
		}
	}
}

func assertInHandCardSlicesMatch(
	testIdentifier string,
	unitTest *testing.T,
	actualCards []card.InHand,
	expectedCards []card.InHand) {
	numberOfExpectedCards := len(expectedCards)

	if len(actualCards) != numberOfExpectedCards {
		unitTest.Fatalf(
			testIdentifier+
				"/Card slice lengths do not match - actual: %v; expected %v",
			actualCards,
			expectedCards)
	}

	for cardIndex := 0; cardIndex < numberOfExpectedCards; cardIndex++ {
		actualCard := actualCards[cardIndex]
		expectedCard := expectedCards[cardIndex]

		actualReadonly := actualCard.Readonly
		expectedReadonly := expectedCard.Readonly

		if (actualReadonly.ColorSuit() != expectedReadonly.ColorSuit()) ||
			(actualReadonly.SequenceIndex() != expectedReadonly.SequenceIndex()) {
			unitTest.Fatalf(
				testIdentifier+
					"/Card slices do not match in element %v - actual: %v; expected %v"+
					" - actual array %v; expected array %v",
				cardIndex,
				actualCard,
				expectedCard,
				actualCards,
				expectedCards)
		}

		assertInferredCardPossibilitiesCorrect(
			testIdentifier,
			unitTest,
			actualCard.Inferred,
			expectedCard.PossibleColors(),
			expectedCard.PossibleIndices())
	}
}

func assertInferredCardPossibilitiesCorrect(
	testIdentifier string,
	unitTest *testing.T,
	actualCard card.Inferred,
	expectedColors []string,
	expectedIndices []int) {

	// We compare the possible colors and indices as sets. Then it is
	// sufficient to check that the lengths are the same and that every
	// actual value is found in the map of expected values.
	if (len(actualCard.PossibleColors()) != len(expectedColors)) ||
		(len(actualCard.PossibleIndices()) != len(expectedIndices)) {
		unitTest.Fatalf(
			testIdentifier+
				"/inferred card %v did not match expected colors %v and indices %v",
			actualCard,
			expectedColors,
			expectedIndices)
	}

	expectedColorMap := make(map[string]bool)
	for _, expectedColor := range expectedColors {
		if expectedColorMap[expectedColor] {
			unitTest.Fatalf(
				testIdentifier+
					"/expected colors %v had duplicate(s)",
				expectedColors)
		}

		expectedColorMap[expectedColor] = true
	}

	for _, actualColor := range actualCard.PossibleColors() {
		if !expectedColorMap[actualColor] {
			unitTest.Fatalf(
				testIdentifier+
					"/inferred card %v did not match expected colors %v and indices %v",
				actualCard,
				expectedColors,
				expectedIndices)
		}
	}

	expectedIndexMap := make(map[int]bool)
	for _, expectedIndex := range expectedIndices {
		if expectedIndexMap[expectedIndex] {
			unitTest.Fatalf(
				testIdentifier+
					"/expected indices %v had duplicate(s)",
				expectedIndices)
		}

		expectedIndexMap[expectedIndex] = true
	}

	for _, actualIndex := range actualCard.PossibleIndices() {
		if !expectedIndexMap[actualIndex] {
			unitTest.Fatalf(
				testIdentifier+
					"/inferred card %v did not match expected colors %v and indices %v",
				actualCard,
				expectedColors,
				expectedIndices)
		}
	}
}
