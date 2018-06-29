package game_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game/card"

	"github.com/benoleary/ilutulestikud/backend/game"
)

func TestInvalidIdentifierProducesError(unitTest *testing.T) {
	invalidIdentifier := -1
	invalidRuleset, errorFromGet :=
		game.RulesetFromIdentifier(invalidIdentifier)

	if errorFromGet == nil {
		unitTest.Fatalf(
			"RulesetFromIdentifier(%v) %+v produced nil error",
			invalidIdentifier,
			invalidRuleset)
	}
}

func TestAvailableRulesetsHaveUniqueDescriptions(unitTest *testing.T) {
	descriptionMap := make(map[string]bool, 0)

	for _, validIdentifier := range game.ValidRulesetIdentifiers() {
		validRuleset, errorFromGet :=
			game.RulesetFromIdentifier(validIdentifier)

		if errorFromGet != nil {
			unitTest.Fatalf(
				"RulesetFromIdentifier(%v) produced error %v",
				validIdentifier,
				errorFromGet)
		}

		frontendDescription := validRuleset.FrontendDescription()

		if descriptionMap[frontendDescription] {
			unitTest.Fatalf(
				"RulesetFromIdentifier(%v) produced previously-seen description %v",
				validIdentifier,
				frontendDescription)
		}

		descriptionMap[frontendDescription] = true
	}
}

func TestStandardHintsAndMistakesAreValid(unitTest *testing.T) {
	standardRuleset := game.NewStandardWithoutRainbow()

	if standardRuleset.MaximumNumberOfHints() <= 0 {
		unitTest.Fatalf(
			"standard ruleset allows %v hints as maximum",
			standardRuleset.MaximumNumberOfHints())
	}

	if standardRuleset.NumberOfMistakesIndicatingGameOver() <= 0 {
		unitTest.Fatalf(
			"standard ruleset allows %v mistakes as maximum",
			standardRuleset.NumberOfMistakesIndicatingGameOver()-1)
	}
}

func TestAllCardsPresentInCompoundRainbow(unitTest *testing.T) {
	rainbowRuleset := game.NewRainbowAsCompoundSuit()

	colorSuitMap := make(map[string]int, 0)

	for _, colorSuit := range rainbowRuleset.ColorSuits() {
		colorSuitMap[colorSuit] = 0
	}

	for _, cardInDeck := range rainbowRuleset.CopyOfFullCardset() {
		countOfSuitUntilNow, isValidSuit :=
			colorSuitMap[cardInDeck.ColorSuit()]

		if !isValidSuit {
			unitTest.Fatalf(
				"found unexpected card %+v in deck",
				cardInDeck)
		}

		colorSuitMap[cardInDeck.ColorSuit()] = countOfSuitUntilNow + 1

		cardValue := rainbowRuleset.PointsForCard(cardInDeck)
		if ((cardInDeck.SequenceIndex() < 5) && (cardValue != cardInDeck.SequenceIndex())) ||
			((cardInDeck.SequenceIndex() >= 5) && (cardValue != (2 * cardInDeck.SequenceIndex()))) {
			unitTest.Fatalf(
				"card %+v has points value %v",
				cardInDeck,
				cardValue)
		}

		hintsForPlayingCard := rainbowRuleset.HintsForPlayingCard(cardInDeck)
		if ((cardInDeck.SequenceIndex() < 5) && (hintsForPlayingCard != 0)) ||
			((cardInDeck.SequenceIndex() >= 5) && (hintsForPlayingCard != 1)) {
			unitTest.Fatalf(
				"card %+v gives %v hints when successfully played",
				cardInDeck,
				hintsForPlayingCard)
		}
	}

	for colorSuit, countForSuit := range colorSuitMap {
		if countForSuit <= 0 {
			unitTest.Fatalf(
				"found no card for color %v",
				colorSuit)
		}
	}
}

func TestStandardRejectsWrongInitialCardInSequence(unitTest *testing.T) {
	standardRuleset := game.NewStandardWithoutRainbow()
	testColor := standardRuleset.ColorSuits()[0]
	playedCards := []card.Readonly{}

	for _, possibleIndex := range standardRuleset.DistinctPossibleIndices() {
		candidateCard := card.NewReadonly(testColor, possibleIndex)
		isPlayable := standardRuleset.IsCardPlayable(candidateCard, playedCards)
		if isPlayable && (possibleIndex != 1) {
			unitTest.Fatalf(
				"standard ruleset allows %v to be played as initial card of sequence",
				candidateCard)
		}
	}
}

func TestStandardRejectsWrongCardInNonemptySequence(unitTest *testing.T) {
	standardRuleset := game.NewStandardWithoutRainbow()
	testColor := standardRuleset.ColorSuits()[0]
	possibleIndices := standardRuleset.DistinctPossibleIndices()
	topmostCard := card.NewReadonly(testColor, possibleIndices[1])
	playedCards := []card.Readonly{
		card.NewReadonly(testColor, possibleIndices[0]),
		topmostCard,
	}

	for _, possibleIndex := range standardRuleset.DistinctPossibleIndices() {
		candidateCard := card.NewReadonly(testColor, possibleIndex)
		isPlayable := standardRuleset.IsCardPlayable(candidateCard, playedCards)
		if isPlayable && (possibleIndex != (topmostCard.SequenceIndex() + 1)) {
			unitTest.Fatalf(
				"standard ruleset allows %v to be played onto %v",
				candidateCard,
				playedCards)
		}
	}
}
