package game_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game/card"
)

func assertCardsMatch(
	testIdentifier string,
	unitTest *testing.T,
	actualCards []card.Readonly,
	expectedCards []card.Readonly) {
	numberOfExpectedCards := len(expectedCards)

	if len(actualCards) != numberOfExpectedCards {
		unitTest.Errorf(
			testIdentifier+"/Card array lengths do not match - actual: %v; expected %v",
			actualCards,
			expectedCards)
	}

	for cardIndex := 0; cardIndex < numberOfExpectedCards; cardIndex++ {
		actualCard := actualCards[cardIndex]
		expectedCard := expectedCards[cardIndex]
		if (actualCard.ColorSuit() != expectedCard.ColorSuit()) ||
			(actualCard.SequenceIndex() != expectedCard.SequenceIndex()) {
			unitTest.Errorf(
				testIdentifier+"/Card arrays do not match in element %v - actual: %v; expected %v - actual array %v; expected array %v",
				cardIndex,
				actualCard,
				expectedCard,
				actualCards,
				expectedCards)
		}
	}
}
