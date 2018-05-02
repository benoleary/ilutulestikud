package game

import (
	"fmt"
)

// This file contains the definition of an interface for rulesets,
// along with some implementations of the interface.

// Ruleset should encapsulate the set of rules for a game as functions.
type Ruleset interface {
	// FrontendDescription should describe the ruleset succintly enough for the frontend.
	FrontendDescription() string

	// FullCardset should return an array populated with every card which should be present
	// for a game under the ruleset, including duplicates.
	FullCardset() []ReadonlyCard

	// NumberOfCardsInPlayerHand should return the number of cards held
	// in a player's hand, dependent on the number of players in the game.
	NumberOfCardsInPlayerHand(numberOfPlayers int) int

	// ColorSuits should return the set of colors used as suits.
	ColorSuits() []string

	// SequenceIndices returns all the indices for the cards, per card so
	// including repetitions of indices, as they should be played per suit.
	SequenceIndices() []int

	// MinimumNumberOfPlayers should return the minimum number of players needed for a game.
	MinimumNumberOfPlayers() int

	// MaximumNumberOfPlayers should return the maximum number of players allowed for a game.
	MaximumNumberOfPlayers() int

	// MaximumNumberOfHints should return the maximum number of hints which can be available
	// at any instant.
	MaximumNumberOfHints() int

	// MaximumNumberOfMistakesAllowed should return the maximum number of mistakes which can
	// be made without the game ending (i.e. the game ends on the next mistake after that).
	MaximumNumberOfMistakesAllowed() int
}

const (
	// NoRulesetChosen denotes 0 as no ruleset chosen, as a missing JSON identifier
	// will end up as 0.
	NoRulesetChosen = iota

	// StandardWithoutRainbowIdentifier is the identifier of the standard ruleset.
	StandardWithoutRainbowIdentifier = iota

	// WithRainbowAsSeparateIdentifier is the identifier of the ruleset with rainbow
	// cards added as a separate suit.
	WithRainbowAsSeparateIdentifier = iota

	// WithRainbowAsCompoundIdentifier is the identifier of the ruleset with rainbow
	// cards added as a special suit which counts as all the others.
	WithRainbowAsCompoundIdentifier = iota
)

// ValidRulesetIdentifiers returns the list of identifiers of valid rulesets.
func ValidRulesetIdentifiers() []int {
	return []int{
		StandardWithoutRainbowIdentifier,
		WithRainbowAsSeparateIdentifier,
		WithRainbowAsCompoundIdentifier,
	}
}

// RulesetFromIdentifier returns the appropriate ruleset for the identifier.
func RulesetFromIdentifier(rulesetIdentifier int) (Ruleset, error) {
	switch rulesetIdentifier {
	case StandardWithoutRainbowIdentifier:
		return &StandardWithoutRainbowRuleset{}, nil
	case WithRainbowAsSeparateIdentifier:
		return &RainbowAsSeparateSuitRuleset{
			BasisRules: &StandardWithoutRainbowRuleset{},
		}, nil
	case WithRainbowAsCompoundIdentifier:
		return &RainbowAsCompoundSuitRuleset{
			BasisRainbow: &RainbowAsSeparateSuitRuleset{
				BasisRules: &StandardWithoutRainbowRuleset{},
			},
		}, nil
	default:
		return nil, fmt.Errorf("Ruleset identifier %v not recognized", rulesetIdentifier)
	}
}

// StandardWithoutRainbowRuleset represents the standard ruleset, which does
// not include the rainbow color suit.
type StandardWithoutRainbowRuleset struct {
}

// FrontendDescription describes the standard ruleset.
func (standardRuleset *StandardWithoutRainbowRuleset) FrontendDescription() string {
	return "standard (without rainbow cards)"
}

// FullCardset returns an array populated with every card which should be present
// for a game under the ruleset, including duplicates.
func (standardRuleset *StandardWithoutRainbowRuleset) FullCardset() []ReadonlyCard {
	colorSuits := standardRuleset.ColorSuits()
	numberOfColors := len(colorSuits)
	sequenceIndices := standardRuleset.SequenceIndices()
	numberOfIndices := len(sequenceIndices)
	fullCardset := make([]ReadonlyCard, 0, numberOfColors*numberOfIndices)

	for _, colorSuit := range colorSuits {
		for _, sequenceIndex := range sequenceIndices {
			fullCardset = append(fullCardset, &simpleCard{
				colorSuit:     colorSuit,
				sequenceIndex: sequenceIndex,
			})
		}
	}

	return fullCardset
}

// NumberOfCardsInPlayerHand returns the number of cards held in a player's
// hand, dependent on the number of players in the game.
func (standardRuleset *StandardWithoutRainbowRuleset) NumberOfCardsInPlayerHand(
	numberOfPlayers int) int {
	if numberOfPlayers <= 3 {
		return 5
	}

	return 4
}

// ColorSuits returns the set of colors used as suits.
func (standardRuleset *StandardWithoutRainbowRuleset) ColorSuits() []string {
	return []string{
		"red",
		"green",
		"blue",
		"yellow",
		"white",
	}
}

// SequenceIndices returns all the indices for the cards, per card so
// including repetitions of indices, as they should be played per suit.
func (standardRuleset *StandardWithoutRainbowRuleset) SequenceIndices() []int {
	return []int{1, 1, 1, 2, 2, 3, 3, 4, 4, 5}
}

// PointsPerCard returns the points value of a card with the given
// sequence index.
func (standardRuleset *StandardWithoutRainbowRuleset) PointsPerCard(
	cardSequenceIndex int) int {
	if cardSequenceIndex >= 5 {
		return 2 * cardSequenceIndex
	}

	return cardSequenceIndex
}

// MinimumNumberOfPlayers returns the minimum number of players needed for a game.
func (standardRuleset *StandardWithoutRainbowRuleset) MinimumNumberOfPlayers() int {
	return 2
}

// MaximumNumberOfPlayers returns the maximum number of players allowed for a game.
func (standardRuleset *StandardWithoutRainbowRuleset) MaximumNumberOfPlayers() int {
	return 5
}

// MaximumNumberOfHints returns the maximum number of hints which can be available at
// any instant.
func (standardRuleset *StandardWithoutRainbowRuleset) MaximumNumberOfHints() int {
	return 8
}

// MaximumNumberOfMistakesAllowed should return the maximum number of mistakes which
// can be made without the game ending (i.e. the game ends on the next mistake after
// that).
func (standardRuleset *StandardWithoutRainbowRuleset) MaximumNumberOfMistakesAllowed() int {
	return 2
}

// RainbowSuit gives the name of the special suit for the variation rulesets. It is
// exported for ease of testing.
const RainbowSuit = "rainbow"

// RainbowAsSeparateSuitRuleset represents the ruleset which includes the rainbow
// color suit as just another suit which is separate for the purposes of hints as
// well.
type RainbowAsSeparateSuitRuleset struct {
	BasisRules *StandardWithoutRainbowRuleset
}

// FrontendDescription describes the ruleset with rainbow cards which are a separate
// suit which behaves like the standard suits.
func (separateRainbow *RainbowAsSeparateSuitRuleset) FrontendDescription() string {
	return "with rainbow, hints as separate color"
}

// FullCardset returns an array populated with every card which should be present
// for a game under the ruleset, including duplicates.
func (separateRainbow *RainbowAsSeparateSuitRuleset) FullCardset() []ReadonlyCard {
	fullCardset := separateRainbow.BasisRules.FullCardset()
	sequenceIndices := separateRainbow.SequenceIndices()

	for _, sequenceIndex := range sequenceIndices {
		fullCardset = append(fullCardset, &simpleCard{
			colorSuit:     RainbowSuit,
			sequenceIndex: sequenceIndex,
		})
	}

	return fullCardset
}

// NumberOfCardsInPlayerHand returns the number of cards held in a player's
// hand, dependent on the number of players in the game.
func (separateRainbow *RainbowAsSeparateSuitRuleset) NumberOfCardsInPlayerHand(
	numberOfPlayers int) int {
	return separateRainbow.BasisRules.NumberOfCardsInPlayerHand(numberOfPlayers)
}

// ColorSuits returns the set of colors used as suits.
func (separateRainbow *RainbowAsSeparateSuitRuleset) ColorSuits() []string {
	return append(separateRainbow.BasisRules.ColorSuits(), RainbowSuit)
}

// SequenceIndices returns all the indices for the cards, per card so
// including repetitions of indices, as they should be played per suit.
func (separateRainbow *RainbowAsSeparateSuitRuleset) SequenceIndices() []int {
	return separateRainbow.BasisRules.SequenceIndices()
}

// PointsPerCard returns the points value of a card with the given
// sequence index.
func (separateRainbow *RainbowAsSeparateSuitRuleset) PointsPerCard(
	cardSequenceIndex int) int {
	return separateRainbow.BasisRules.PointsPerCard(cardSequenceIndex)
}

// MinimumNumberOfPlayers returns the minimum number of players needed for a game.
func (separateRainbow *RainbowAsSeparateSuitRuleset) MinimumNumberOfPlayers() int {
	return separateRainbow.BasisRules.MinimumNumberOfPlayers()
}

// MaximumNumberOfPlayers returns the maximum number of players allowed for a game.
func (separateRainbow *RainbowAsSeparateSuitRuleset) MaximumNumberOfPlayers() int {
	return separateRainbow.BasisRules.MaximumNumberOfPlayers()
}

// MaximumNumberOfHints returns the maximum number of hints which can be available at
// any instant.
func (separateRainbow *RainbowAsSeparateSuitRuleset) MaximumNumberOfHints() int {
	return separateRainbow.BasisRules.MaximumNumberOfHints()
}

// MaximumNumberOfMistakesAllowed returns the maximum number of mistakes which can be
// made without the game ending (i.e. the game ends on the next mistake after
// that).
func (separateRainbow *RainbowAsSeparateSuitRuleset) MaximumNumberOfMistakesAllowed() int {
	return separateRainbow.BasisRules.MaximumNumberOfMistakesAllowed()
}

// RainbowAsCompoundSuitRuleset represents the ruleset which includes the rainbow
// color suit as another suit which, however, counts as all the other suits for
// hints. Most of the functions are the same.
type RainbowAsCompoundSuitRuleset struct {
	BasisRainbow *RainbowAsSeparateSuitRuleset
}

// FrontendDescription describes the ruleset with rainbow cards which are a separate suit
// which behaves differently from the standard suits, in the sense that it is not directly
// a color which can be given as a hint, but rather every hint for a standard color will
// also identify a rainbow card as a card of that standard color.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) FrontendDescription() string {
	return "with rainbow, hints as every color"
}

// FullCardset returns an array populated with every card which should be present
// for a game under the ruleset, including duplicates.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) FullCardset() []ReadonlyCard {
	return compoundRainbow.BasisRainbow.FullCardset()
}

// NumberOfCardsInPlayerHand returns the number of cards held in a player's
// hand, dependent on the number of players in the game.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) NumberOfCardsInPlayerHand(
	numberOfPlayers int) int {
	return compoundRainbow.BasisRainbow.NumberOfCardsInPlayerHand(numberOfPlayers)
}

// ColorSuits returns the set of colors used as suits.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) ColorSuits() []string {
	return compoundRainbow.BasisRainbow.ColorSuits()
}

// SequenceIndices returns all the indices for the cards, per card so
// including repetitions of indices, as they should be played per suit.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) SequenceIndices() []int {
	return compoundRainbow.BasisRainbow.SequenceIndices()
}

// PointsPerCard returns the points value of a card with the given
// sequence index.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) PointsPerCard(
	cardSequenceIndex int) int {
	return compoundRainbow.BasisRainbow.PointsPerCard(cardSequenceIndex)
}

// MinimumNumberOfPlayers returns the minimum number of players needed for a game.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) MinimumNumberOfPlayers() int {
	return compoundRainbow.BasisRainbow.MinimumNumberOfPlayers()
}

// MaximumNumberOfPlayers returns the maximum number of players allowed for a game.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) MaximumNumberOfPlayers() int {
	return compoundRainbow.BasisRainbow.MaximumNumberOfPlayers()
}

// MaximumNumberOfHints returns the maximum number of hints which can be available at
// any instant.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) MaximumNumberOfHints() int {
	return compoundRainbow.BasisRainbow.MaximumNumberOfHints()
}

// MaximumNumberOfMistakesAllowed returns the maximum number of mistakes which can be
// made without the game ending (i.e. the game ends on the next mistake after
// that).
func (compoundRainbow *RainbowAsCompoundSuitRuleset) MaximumNumberOfMistakesAllowed() int {
	return compoundRainbow.BasisRainbow.MaximumNumberOfMistakesAllowed()
}
