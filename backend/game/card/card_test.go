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
			"NewReadonly(%v, %v) produced unexpected %v",
			testColor,
			testIndex,
			validReadonly)
	}
}

func TestNewErrorReadonly(unitTest *testing.T) {
	validReadonly := card.ErrorReadonly()

	if (validReadonly.ColorSuit() != "error") ||
		(validReadonly.SequenceIndex() != -1) {
		unitTest.Fatalf(
			"ErrorReadonly() produced unexpected %v",
			validReadonly)
	}
}
