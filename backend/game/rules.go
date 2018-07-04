package game

import (
	"fmt"

	"github.com/benoleary/ilutulestikud/backend/game/card"
)

// This file contains some implementations of the interface for rulesets.
// It also contains a system mapping ints to implementations, which is here
// as it is both used by the endpoint-handling code in its communication
// with the frontend, and is relevant to serializing the game state.

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
		return NewStandardWithoutRainbow(), nil
	case WithRainbowAsSeparateIdentifier:
		return NewRainbowAsSeparateSuit(), nil
	case WithRainbowAsCompoundIdentifier:
		return NewRainbowAsCompoundSuit(), nil
	default:
		return nil, fmt.Errorf("Ruleset identifier %v not recognized", rulesetIdentifier)
	}
}

// standardWithoutRainbowRuleset represents the standard ruleset, which
// does not include the rainbow color suit.
type standardWithoutRainbowRuleset struct {
	indicesWithRepetition []int
	distinctIndices       []int
}

// NewStandardWithoutRainbow creates a new standardWithoutRainbowRuleset
// with the indices set up correctly.
func NewStandardWithoutRainbow() Ruleset {
	explicitStandardWithoutRainbow := createStandardWithoutRainbow()
	return &explicitStandardWithoutRainbow
}

// NewStandardWithoutRainbow creates a new standardWithoutRainbowRuleset
// with the indices set up correctly.
func createStandardWithoutRainbow() standardWithoutRainbowRuleset {
	indicesWithRepetition := []int{1, 1, 1, 2, 2, 3, 3, 4, 4, 5}

	// The indices are fixed, and the list of their distinct values
	// could easily be initialized by hand. However, even though writing
	// more code is also error-prone, it is fun to automate consistency.
	// Also, we could just create a map and then use it to create the
	// list of distinct indices afterwards, but the order is not
	// guaranteed in that case, and I would rather preserve the order.
	alreadyAdded := make(map[int]bool, 0)
	distinctIndices := make([]int, 0)

	for _, indexWithRepetition := range indicesWithRepetition {
		if !alreadyAdded[indexWithRepetition] {
			distinctIndices = append(distinctIndices, indexWithRepetition)
			alreadyAdded[indexWithRepetition] = true
		}
	}

	return standardWithoutRainbowRuleset{
		indicesWithRepetition: indicesWithRepetition,
		distinctIndices:       distinctIndices,
	}
}

// FrontendDescription describes the standard ruleset.
func (standardRuleset *standardWithoutRainbowRuleset) FrontendDescription() string {
	return "standard (without rainbow cards)"
}

// CopyOfFullCardset returns an array populated with every card which should be present
// for a game under the ruleset, including duplicates.
func (standardRuleset *standardWithoutRainbowRuleset) CopyOfFullCardset() []card.Readonly {
	colorSuits := standardRuleset.ColorSuits()
	numberOfColors := len(colorSuits)
	numberOfIndices := len(standardRuleset.indicesWithRepetition)
	numberOfCardsInDeck := numberOfColors * numberOfIndices
	fullCardset := make([]card.Readonly, 0, numberOfCardsInDeck)

	for _, colorSuit := range colorSuits {
		for _, sequenceIndex := range standardRuleset.indicesWithRepetition {
			fullCardset = append(fullCardset, card.NewReadonly(colorSuit, sequenceIndex))
		}
	}

	return fullCardset
}

// NumberOfCardsInPlayerHand returns the number of cards held in a player's
// hand, dependent on the number of players in the game.
func (standardRuleset *standardWithoutRainbowRuleset) NumberOfCardsInPlayerHand(
	numberOfPlayers int) int {
	if numberOfPlayers <= 3 {
		return 5
	}

	return 4
}

// ColorSuits returns the set of colors used as suits.
func (standardRuleset *standardWithoutRainbowRuleset) ColorSuits() []string {
	return []string{
		"red",
		"green",
		"blue",
		"yellow",
		"white",
	}
}

// DistinctPossibleIndices returns all the distinct indices for the cards
// across all suits of the ruleset.
func (standardRuleset *standardWithoutRainbowRuleset) DistinctPossibleIndices() []int {
	return standardRuleset.distinctIndices
}

// MinimumNumberOfPlayers returns the minimum number of players needed for a game.
func (standardRuleset *standardWithoutRainbowRuleset) MinimumNumberOfPlayers() int {
	return 2
}

// MaximumNumberOfPlayers returns the maximum number of players allowed for a game.
func (standardRuleset *standardWithoutRainbowRuleset) MaximumNumberOfPlayers() int {
	return 5
}

// MaximumNumberOfHints returns the maximum number of hints which can be available at
// any instant.
func (standardRuleset *standardWithoutRainbowRuleset) MaximumNumberOfHints() int {
	return 8
}

// NumberOfMistakesIndicatingGameOver returns the number of mistakes which indicates
// that the game is over with the players having zero score.
func (standardRuleset *standardWithoutRainbowRuleset) NumberOfMistakesIndicatingGameOver() int {
	return 3
}

// IsCardPlayable returns true if the given card has a value exactly one greater than
// the last card in the given sequence of cards already played in the cards's suit if
// the slice is not empty, or true if the sequence is empty and the card's value is
// one, or false otherwise.
func (standardRuleset *standardWithoutRainbowRuleset) IsCardPlayable(
	cardToPlay card.Readonly,
	cardsAlreadyPlayedInSuit []card.Readonly) bool {
	numberOfCardsPlayedInSuit := len(cardsAlreadyPlayedInSuit)
	if numberOfCardsPlayedInSuit <= 0 {
		return cardToPlay.SequenceIndex() == 1
	}

	topmostPlayedCard := cardsAlreadyPlayedInSuit[numberOfCardsPlayedInSuit-1]
	return cardToPlay.SequenceIndex() == (topmostPlayedCard.SequenceIndex() + 1)
}

// HintsForPlayingCard returns the number of hints to refresh upon successfully
// playing the given card.
func (standardRuleset *standardWithoutRainbowRuleset) HintsForPlayingCard(
	cardToEvaluate card.Readonly) int {
	if cardToEvaluate.SequenceIndex() >= 5 {
		return 1
	}

	return 0
}

// PointsPerCard returns the points value of the given card.
func (standardRuleset *standardWithoutRainbowRuleset) PointsForCard(
	cardToEvaluate card.Readonly) int {
	// Every card is worth the same in the standard rules.
	return 1
}

// RainbowSuit gives the name of the special suit for the variation rulesets. It is
// exported for ease of testing.
const RainbowSuit = "rainbow"

// RainbowAsSeparateSuitRuleset represents the ruleset which includes the rainbow
// color suit as just another suit which is separate for the purposes of hints as
// well.
type RainbowAsSeparateSuitRuleset struct {
	standardWithoutRainbowRuleset
}

// NewRainbowAsSeparateSuit creates a new RainbowAsSeparateSuitRuleset
// with the indices set up correctly.
func NewRainbowAsSeparateSuit() *RainbowAsSeparateSuitRuleset {
	return &RainbowAsSeparateSuitRuleset{
		// The indices for the rainbow suit are the same as the basic suit.
		standardWithoutRainbowRuleset: createStandardWithoutRainbow(),
	}
}

// FrontendDescription describes the ruleset with rainbow cards which are a separate
// suit which behaves like the standard suits.
func (separateRainbow *RainbowAsSeparateSuitRuleset) FrontendDescription() string {
	return "with rainbow, hints as separate color"
}

// CopyOfFullCardset returns an array populated with every card which should be present
// for a game under the ruleset, including duplicates.
func (separateRainbow *RainbowAsSeparateSuitRuleset) CopyOfFullCardset() []card.Readonly {
	fullCardset := separateRainbow.standardWithoutRainbowRuleset.CopyOfFullCardset()

	for _, sequenceIndex := range separateRainbow.indicesWithRepetition {
		fullCardset = append(fullCardset, card.NewReadonly(RainbowSuit, sequenceIndex))
	}

	return fullCardset
}

// ColorSuits returns the set of colors used as suits.
func (separateRainbow *RainbowAsSeparateSuitRuleset) ColorSuits() []string {
	return append(separateRainbow.standardWithoutRainbowRuleset.ColorSuits(), RainbowSuit)
}

// RainbowAsCompoundSuitRuleset represents the ruleset which includes the rainbow
// color suit as another suit which, however, counts as all the other suits for
// hints. Most of the functions are the same.
type RainbowAsCompoundSuitRuleset struct {
	RainbowAsSeparateSuitRuleset
}

// NewRainbowAsCompoundSuit creates a new RainbowAsCompoundSuitRuleset
// with the indices set up correctly.
func NewRainbowAsCompoundSuit() *RainbowAsCompoundSuitRuleset {
	return &RainbowAsCompoundSuitRuleset{
		// The indices for the rainbow suit no matter how hints act on the rainbow cards.
		RainbowAsSeparateSuitRuleset: RainbowAsSeparateSuitRuleset{
			standardWithoutRainbowRuleset: createStandardWithoutRainbow(),
		},
	}
}

// FrontendDescription describes the ruleset with rainbow cards which are a separate suit
// which behaves differently from the standard suits, in the sense that it is not directly
// a color which can be given as a hint, but rather every hint for a standard color will
// also identify a rainbow card as a card of that standard color.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) FrontendDescription() string {
	return "with rainbow, hints as every color"
}
