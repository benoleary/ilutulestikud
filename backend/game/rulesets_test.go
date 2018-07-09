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

		// All the official rulesets give one point per card successfully played,
		// no matter which card.
		if rainbowRuleset.PointsForCard(cardInDeck) != 1 {
			unitTest.Fatalf(
				"card %+v has points value %v",
				cardInDeck,
				rainbowRuleset.PointsForCard(cardInDeck))
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

func TestAllIndicesAreAvailableForHintsInRelevantRulesets(unitTest *testing.T) {
	rulesetsWithAllIndicesForHints :=
		[]game.Ruleset{
			game.NewStandardWithoutRainbow(),
			game.NewRainbowAsSeparateSuit(),
			game.NewRainbowAsCompoundSuit(),
		}

	for _, rulesetWithAllIndicesForHints := range rulesetsWithAllIndicesForHints {
		possibleIndices := rulesetWithAllIndicesForHints.DistinctPossibleIndices()
		expectedNumberOfIndicesForHints := len(possibleIndices)

		indicesForHints := rulesetWithAllIndicesForHints.IndicesAvailableAsHint()

		if len(indicesForHints) != expectedNumberOfIndicesForHints {
			unitTest.Fatalf(
				"ruleset %v has indices %v for cards, but %v for hints",
				rulesetWithAllIndicesForHints.FrontendDescription(),
				possibleIndices,
				indicesForHints)
		}

		remainingIndices := make(map[int]int, expectedNumberOfIndicesForHints)
		for _, possibleIndex := range possibleIndices {
			countForThisIndex := remainingIndices[possibleIndex]
			if countForThisIndex > 0 {
				unitTest.Fatalf(
					"ruleset %v has repeated index %v in possible indices %v",
					rulesetWithAllIndicesForHints.FrontendDescription(),
					possibleIndex,
					possibleIndices)
			}

			remainingIndices[possibleIndex] = countForThisIndex + 1
		}

		for _, indexForHints := range indicesForHints {
			remainingIndices[indexForHints] = remainingIndices[indexForHints] - 1
		}

		// Since the possible indices are unique and the lengths match if we are at
		// this point, it suffices to check that each possible index is found in the
		// hint indices.
		for _, countForRemainingIndex := range remainingIndices {
			if countForRemainingIndex != 0 {
				unitTest.Fatalf(
					"ruleset %v remaining indices %v",
					rulesetWithAllIndicesForHints.FrontendDescription(),
					remainingIndices)
			}
		}
	}
}

func TestAllColorsAreAvailableForHintsInStandardAndRainbowAsExtra(unitTest *testing.T) {
	rulesetsWithAllColorsForHints :=
		[]game.Ruleset{
			game.NewStandardWithoutRainbow(),
			game.NewRainbowAsSeparateSuit(),
		}

	for _, rulesetWithAllColorsForHints := range rulesetsWithAllColorsForHints {
		colorsForCards := rulesetWithAllColorsForHints.ColorSuits()
		expectedNumberOfColorsForHints := len(colorsForCards)

		colorsForHints := rulesetWithAllColorsForHints.ColorsAvailableAsHint()

		if len(colorsForHints) != expectedNumberOfColorsForHints {
			unitTest.Fatalf(
				"ruleset %v has colors %v for cards, but %v for hints",
				rulesetWithAllColorsForHints.FrontendDescription(),
				colorsForCards,
				colorsForHints)
		}

		remainingColors := make(map[string]int, expectedNumberOfColorsForHints)
		for _, colorForCards := range colorsForCards {
			countForThisColor := remainingColors[colorForCards]
			if countForThisColor > 0 {
				unitTest.Fatalf(
					"ruleset %v has repeated color %v in possible colors %v",
					rulesetWithAllColorsForHints.FrontendDescription(),
					colorForCards,
					colorsForCards)
			}

			remainingColors[colorForCards] = countForThisColor + 1
		}

		for _, colorForHints := range colorsForHints {
			remainingColors[colorForHints] = remainingColors[colorForHints] - 1
		}

		// Since the possible colors are unique and the lengths match if we are at
		// this point, it suffices to check that each possible color is found in the
		// hint colors.
		for _, countForRemainingColor := range remainingColors {
			if countForRemainingColor != 0 {
				unitTest.Fatalf(
					"ruleset %v remaining colors %v",
					rulesetWithAllColorsForHints.FrontendDescription(),
					remainingColors)
			}
		}
	}
}

func TestRainbowAsCompoundSuitDoesNotHaveRainbowForHints(unitTest *testing.T) {
	compoundRainbowRuleset := game.NewRainbowAsCompoundSuit()

	colorsForCards := compoundRainbowRuleset.ColorSuits()

	// We expect there to be one less color mentioned as available for hints.
	expectedNumberOfColorsForHints := len(colorsForCards) - 1

	colorsForHints := compoundRainbowRuleset.ColorsAvailableAsHint()

	if len(colorsForHints) != expectedNumberOfColorsForHints {
		unitTest.Fatalf(
			"ruleset %v has colors %v for cards, but %v for hints",
			compoundRainbowRuleset.FrontendDescription(),
			colorsForCards,
			colorsForHints)
	}

	remainingColors := make(map[string]int, expectedNumberOfColorsForHints)
	for _, colorForCards := range colorsForCards {
		countForThisColor := remainingColors[colorForCards]
		if countForThisColor > 0 {
			unitTest.Fatalf(
				"ruleset %v has repeated color %v in possible colors %v",
				compoundRainbowRuleset.FrontendDescription(),
				colorForCards,
				colorsForCards)
		}

		remainingColors[colorForCards] = countForThisColor + 1
	}

	for _, colorForHints := range colorsForHints {
		remainingColors[colorForHints] = remainingColors[colorForHints] - 1
	}

	// Since the possible colors are unique and the lengths match if we are at
	// this point, it suffices to check that each possible color is found in the
	// hint colors - with the exception of the rainbow suit!
	for remainingColor, countForRemainingColor := range remainingColors {
		if remainingColor == game.RainbowSuit {
			if countForRemainingColor != 1 {
				unitTest.Fatalf(
					"rainbow color not correct for ruleset %v: colors for cards %v, colors for hints %v",
					compoundRainbowRuleset.FrontendDescription(),
					colorsForCards,
					colorsForHints)
			}
		} else {
			if countForRemainingColor != 0 {
				unitTest.Fatalf(
					"ruleset %v remaining colors %v",
					compoundRainbowRuleset.FrontendDescription(),
					remainingColors)
			}
		}
	}
}
