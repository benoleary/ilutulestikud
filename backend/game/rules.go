package game

import (
	"fmt"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// This file contains the definition of an interface for rulesets,
// along with some implementations of the interface, and some
// interface-independent constants which are common to all the
// rulesets. These could of course be replaced by interface
// functions which return the same constants across the various
// implementations, but that's just busywork for no gain.

// MaximumNumberOfHints is the maximum number of hints which can be
// ready to use in a game at any time.
const MaximumNumberOfHints = 8

// MaximumNumberOfMistakesAllowed is the maximum number of mistakes
// that can be made without the game ending (i.e. the game ends on the
// third mistake).
const MaximumNumberOfMistakesAllowed = 2

// Ruleset should encapsulate the set of rules for a game as functions.
type Ruleset interface {
	// FrontendDescription should describe the ruleset succintly enough for the frontend.
	FrontendDescription() string

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
}

const (
	// We denote 0 as no ruleset chosen as a missing JSON identifier will end up as 0.
	noRulesetChosen        = iota
	standardWithoutRainbow = iota
	withRainbowAsSeparate  = iota
	withRainbowAsCompound  = iota
)

// AvailableRulesets returns the list of identifiers with descriptions
// representing the rulesets which are available for creating games.
func AvailableRulesets() ([]endpoint.SelectableRuleset, error) {
	availableRulesets := make([]endpoint.SelectableRuleset, 0)
	for rulesetIndex := standardWithoutRainbow; rulesetIndex <= withRainbowAsCompound; rulesetIndex++ {
		availableRuleset, indexError := GetRuleset(rulesetIndex)
		if indexError != nil {
			return nil, fmt.Errorf(
				"AvailableRulesets() managed to get an error when iterating over rulesets: %v",
				indexError)
		}

		availableRulesets = append(
			availableRulesets,
			endpoint.SelectableRuleset{
				Identifier:             rulesetIndex,
				Description:            availableRuleset.FrontendDescription(),
				MinimumNumberOfPlayers: availableRuleset.MinimumNumberOfPlayers(),
				MaximumNumberOfPlayers: availableRuleset.MaximumNumberOfPlayers(),
			})
	}

	return availableRulesets, nil
}

// GetRuleset returns the appropriate ruleset for the identifier.
func GetRuleset(rulesetIdentifier int) (Ruleset, error) {
	switch rulesetIdentifier {
	case standardWithoutRainbow:
		return &StandardWithoutRainbowRuleset{}, nil
	case withRainbowAsSeparate:
		return &RainbowAsSeparateSuitRuleset{
			BasisRules: &StandardWithoutRainbowRuleset{},
		}, nil
	case withRainbowAsCompound:
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

// RainbowAsSeparateSuitRuleset represents the ruleset which include sthe rainbow
// color suit as just another suit which is separate for the purposes of hints as
// well.
type RainbowAsSeparateSuitRuleset struct {
	BasisRules *StandardWithoutRainbowRuleset
}

// FrontendDescription describes the ruleset with rainbow cards which are a separate suit
// which behaves like the standard suits.
func (separateRainbow *RainbowAsSeparateSuitRuleset) FrontendDescription() string {
	return "with rainbow, hints as separate color"
}

// NumberOfCardsInPlayerHand returns the number of cards held in a player's
// hand, dependent on the number of players in the game.
func (separateRainbow *RainbowAsSeparateSuitRuleset) NumberOfCardsInPlayerHand(
	numberOfPlayers int) int {
	return separateRainbow.BasisRules.NumberOfCardsInPlayerHand(numberOfPlayers)
}

// ColorSuits returns the set of colors used as suits.
func (separateRainbow *RainbowAsSeparateSuitRuleset) ColorSuits() []string {
	return append(separateRainbow.BasisRules.ColorSuits(), "rainbow")
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

// RainbowAsCompoundSuitRuleset represents the ruleset which includes the rainbow
// color suit as another suit which, however, counts as all the other suits for
// hints. Most of the functions are the same.
type RainbowAsCompoundSuitRuleset struct {
	BasisRainbow *RainbowAsSeparateSuitRuleset
}

// FrontendDescription describes the ruleset with rainbow cards which are a separate suit
// which behaves differently from the standard suits, in the sense that it is not directly a color which
// can be given as a hint, but rather every hint for a standard color will also identify a rainbow card
// as a card of that standard color.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) FrontendDescription() string {
	return "with rainbow, hints as every color"
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
