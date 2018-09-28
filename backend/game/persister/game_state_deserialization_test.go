package persister_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/persister"
)

var testRuleset = game.NewStandardWithoutRainbow()

func TestPlayedCardsDeserializedCorrectly(unitTest *testing.T) {
	firstColor := "first color"
	secondColor := "second color"
	thirdColor := "third color"
	expectedPlayedCards := []card.Defined{
		card.Defined{
			ColorSuit:     firstColor,
			SequenceIndex: 1,
		},
		card.Defined{
			ColorSuit:     firstColor,
			SequenceIndex: 2,
		},
		card.Defined{
			ColorSuit:     firstColor,
			SequenceIndex: 3,
		},
		card.Defined{
			ColorSuit:     secondColor,
			SequenceIndex: 1,
		},
		card.Defined{
			ColorSuit:     thirdColor,
			SequenceIndex: 1,
		},
		card.Defined{
			ColorSuit:     thirdColor,
			SequenceIndex: 2,
		},
	}

	numberOfExpectedPlayedCards := len(expectedPlayedCards)
	serializablePart :=
		persister.NewSerializableState("deserializing played cards test", 0, nil, testRuleset, nil, nil)
	serializablePart.PlayedCards = expectedPlayedCards

	deserializedState :=
		persister.CreateDeserializedState(serializablePart, testRuleset)

	for cardIndex := 0; cardIndex < numberOfExpectedPlayedCards; cardIndex++ {
		expectedCard := expectedPlayedCards[cardIndex]
		serializablePartCard := serializablePart.PlayedCards[cardIndex]

		if serializablePartCard != expectedCard {
			unitTest.Fatalf(
				"Wrapping in DeserializedState altered SerializedState.PlayedCards:"+
					" expected %v, actual %v",
				expectedPlayedCards,
				serializablePart.PlayedCards)
		}

		deserializedCardsForColor :=
			deserializedState.PlayedForColor(expectedCard.ColorSuit)

		var deserializedCard card.Defined

		switch expectedCard.ColorSuit {
		case firstColor:
			deserializedCard = deserializedCardsForColor[cardIndex]
		case secondColor:
			deserializedCard = deserializedCardsForColor[cardIndex-3]
		case thirdColor:
			deserializedCard = deserializedCardsForColor[cardIndex-4]
		default:
			deserializedCard = card.Defined{}
			unitTest.Fatalf(
				"expected card had color %v not in switch statement",
				expectedCard)
		}

		if deserializedCard != expectedCard {
			unitTest.Fatalf(
				"DeserializedState had wrong played cards: expected %v, actual %v",
				expectedPlayedCards,
				deserializedState.PlayedCardsForColor)
		}
	}
}

func TestDiscardedCardsDeserializedCorrectly(unitTest *testing.T) {
	expectedDiscardedCards := []card.Defined{
		card.Defined{
			ColorSuit:     "first color",
			SequenceIndex: 2,
		},
		card.Defined{
			ColorSuit:     "first color",
			SequenceIndex: 1,
		},
		card.Defined{
			ColorSuit:     "second color",
			SequenceIndex: 1,
		},
		card.Defined{
			ColorSuit:     "third color",
			SequenceIndex: 1,
		},
		card.Defined{
			ColorSuit:     "third color",
			SequenceIndex: 2,
		},
		card.Defined{
			ColorSuit:     "first color",
			SequenceIndex: 3,
		},
	}

	numberOfExpectedDiscardedCards := len(expectedDiscardedCards)
	serializablePart :=
		persister.NewSerializableState("deserializing discarded cards test", 0, nil, testRuleset, nil, nil)
	serializablePart.DiscardedCards = expectedDiscardedCards

	deserializedState :=
		persister.CreateDeserializedState(serializablePart, testRuleset)

	for cardIndex := 0; cardIndex < numberOfExpectedDiscardedCards; cardIndex++ {
		expectedCard := expectedDiscardedCards[cardIndex]
		serializablePartCard := serializablePart.DiscardedCards[cardIndex]

		if serializablePartCard != expectedCard {
			unitTest.Fatalf(
				"Wrapping in DeserializedState altered SerializedState.DiscardedCards:"+
					" expected %v, actual %v",
				expectedDiscardedCards,
				serializablePart.DiscardedCards)
		}

		deserializedNumberOfCards :=
			deserializedState.NumberOfDiscardedCards(
				expectedCard.ColorSuit,
				expectedCard.SequenceIndex)

		if deserializedNumberOfCards != 1 {
			unitTest.Fatalf(
				"DeserializedState had wrong discarded cards: expected 1 card %v, actual number %v",
				expectedDiscardedCards,
				deserializedState.NumbersOfDiscardedCards)
		}
	}

	if len(deserializedState.NumbersOfDiscardedCards) != numberOfExpectedDiscardedCards {
		unitTest.Fatalf(
			"DeserializedState had wrong discarded card numbers: expected %v, actual %v",
			expectedDiscardedCards,
			deserializedState.NumbersOfDiscardedCards)
	}
}
