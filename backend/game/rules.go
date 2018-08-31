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

// BackendIdentifier returns the identifier for the standard ruleset.
func (standardRuleset *standardWithoutRainbowRuleset) BackendIdentifier() int {
	return StandardWithoutRainbowIdentifier
}

// FrontendDescription describes the standard ruleset.
func (standardRuleset *standardWithoutRainbowRuleset) FrontendDescription() string {
	return "standard (without rainbow cards)"
}

// CopyOfFullCardset returns an array populated with every card which should be present
// for a game under the ruleset, including duplicates.
func (standardRuleset *standardWithoutRainbowRuleset) CopyOfFullCardset() []card.Defined {
	colorSuits := standardRuleset.ColorSuits()
	numberOfColors := len(colorSuits)
	numberOfIndices := len(standardRuleset.indicesWithRepetition)
	numberOfCardsInDeck := numberOfColors * numberOfIndices
	fullCardset := make([]card.Defined, 0, numberOfCardsInDeck)

	for _, colorSuit := range colorSuits {
		for _, sequenceIndex := range standardRuleset.indicesWithRepetition {
			fullCardset =
				append(
					fullCardset,
					card.Defined{
						ColorSuit:     colorSuit,
						SequenceIndex: sequenceIndex,
					})
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

// ColorsAvailableAsHint just returns all the suits under the standard rules.
func (standardRuleset *standardWithoutRainbowRuleset) ColorsAvailableAsHint() []string {
	return standardRuleset.ColorSuits()
}

// IndicesAvailableAsHint just returns all the indices under the standard rules.
func (standardRuleset *standardWithoutRainbowRuleset) IndicesAvailableAsHint() []int {
	return standardRuleset.DistinctPossibleIndices()
}

// AfterColorHint returns the knowledge about a hand that a player has after applying
// the given hint about color to the given knowledge about the hand prior to the hint.
// In this case, if the color of the card matches the color of the hint, the
// possibilities list is reduced to a single element which is that color; otherwise
// the hinted color is removed from the possibilities list if it is on it.
func (standardRuleset *standardWithoutRainbowRuleset) AfterColorHint(
	knowledgeBeforeHint []card.Inferred,
	cardsInHand []card.Defined,
	hintedColor string) []card.Inferred {
	handSize := len(cardsInHand)
	knowledgeAfterHint := make([]card.Inferred, handSize)
	for indexInHand := 0; indexInHand < handSize; indexInHand++ {
		colorOfCard := cardsInHand[indexInHand].ColorSuit

		var replacementColors []string

		if colorOfCard == hintedColor {
			replacementColors = []string{colorOfCard}
		} else {
			originalColors :=
				knowledgeBeforeHint[indexInHand].PossibleColors
			replacementColors = nil

			for _, possibleColor := range originalColors {
				if possibleColor == hintedColor {
					continue
				}

				replacementColors = append(replacementColors, possibleColor)
			}
		}

		knowledgeAfterHint[indexInHand] =
			card.Inferred{
				PossibleColors:  replacementColors,
				PossibleIndices: knowledgeBeforeHint[indexInHand].PossibleIndices,
			}
	}

	return knowledgeAfterHint
}

// AfterIndexHint should return the knowledge about a hand that a player has after applying
// the given hint about index to the given knowledge about the hand prior to the hint.
// In this case, if the sequence index of the card matches that of the hint, the
// possibilities list is reduced to a single element which is that index; otherwise
// the hinted index is removed from the possibilities list if it is on it.
func (standardRuleset *standardWithoutRainbowRuleset) AfterIndexHint(
	knowledgeBeforeHint []card.Inferred,
	cardsInHand []card.Defined,
	hintedIndex int) []card.Inferred {
	handSize := len(cardsInHand)
	knowledgeAfterHint := make([]card.Inferred, handSize)
	for indexInHand := 0; indexInHand < handSize; indexInHand++ {
		sequenceIndexOfCard := cardsInHand[indexInHand].SequenceIndex

		var replacementIndices []int

		if sequenceIndexOfCard == hintedIndex {
			replacementIndices = []int{sequenceIndexOfCard}
		} else {
			originalSequenceIndices :=
				knowledgeBeforeHint[indexInHand].PossibleIndices

			for _, possibleIndex := range originalSequenceIndices {
				if possibleIndex == hintedIndex {
					continue
				}

				replacementIndices = append(replacementIndices, possibleIndex)
			}
		}

		knowledgeAfterHint[indexInHand] =
			card.Inferred{
				PossibleColors:  knowledgeBeforeHint[indexInHand].PossibleColors,
				PossibleIndices: replacementIndices,
			}
	}

	return knowledgeAfterHint
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
	cardToPlay card.Defined,
	cardsAlreadyPlayedInSuit []card.Defined) bool {
	numberOfCardsPlayedInSuit := len(cardsAlreadyPlayedInSuit)
	if numberOfCardsPlayedInSuit <= 0 {
		return cardToPlay.SequenceIndex == 1
	}

	topmostPlayedCard := cardsAlreadyPlayedInSuit[numberOfCardsPlayedInSuit-1]
	return cardToPlay.SequenceIndex == (topmostPlayedCard.SequenceIndex + 1)
}

// HintsForPlayingCard returns the number of hints to refresh upon successfully
// playing the given card.
func (standardRuleset *standardWithoutRainbowRuleset) HintsForPlayingCard(
	cardToEvaluate card.Defined) int {
	if cardToEvaluate.SequenceIndex >= 5 {
		return 1
	}

	return 0
}

// PointsPerCard returns the points value of the given card.
func (standardRuleset *standardWithoutRainbowRuleset) PointsForCard(
	cardToEvaluate card.Defined) int {
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

// BackendIdentifier returns the identifier for the ruleset with rainbow cards which
// are a separate suit which behaves like the standard suits.
func (separateRainbow *RainbowAsSeparateSuitRuleset) BackendIdentifier() int {
	return WithRainbowAsSeparateIdentifier
}

// FrontendDescription describes the ruleset with rainbow cards which are a separate
// suit which behaves like the standard suits.
func (separateRainbow *RainbowAsSeparateSuitRuleset) FrontendDescription() string {
	return "with rainbow, hints as separate color"
}

// CopyOfFullCardset returns an array populated with every card which should be present
// for a game under the ruleset, including duplicates.
func (separateRainbow *RainbowAsSeparateSuitRuleset) CopyOfFullCardset() []card.Defined {
	fullCardset := separateRainbow.standardWithoutRainbowRuleset.CopyOfFullCardset()

	for _, sequenceIndex := range separateRainbow.indicesWithRepetition {
		fullCardset =
			append(
				fullCardset,
				card.Defined{
					ColorSuit: RainbowSuit, SequenceIndex: sequenceIndex,
				})
	}

	return fullCardset
}

// ColorSuits returns the set of colors used as suits.
func (separateRainbow *RainbowAsSeparateSuitRuleset) ColorSuits() []string {
	return append(separateRainbow.standardWithoutRainbowRuleset.ColorSuits(), RainbowSuit)
}

// ColorsAvailableAsHint just returns all the suits under the rainbow-as-extra-suit rules.
func (separateRainbow *RainbowAsSeparateSuitRuleset) ColorsAvailableAsHint() []string {
	return separateRainbow.ColorSuits()
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

// BackendIdentifier returns the identifier for the ruleset with rainbow cards which are a
// separate suit which behaves differently from the standard suits, in the sense that it is
// not directly a color which can be given as a hint, but rather every hint for a standard
// color will also identify a rainbow card as a card of that standard color.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) BackendIdentifier() int {
	return WithRainbowAsCompoundIdentifier
}

// FrontendDescription describes the ruleset with rainbow cards which are a separate suit
// which behaves differently from the standard suits, in the sense that it is not directly
// a color which can be given as a hint, but rather every hint for a standard color will
// also identify a rainbow card as a card of that standard color.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) FrontendDescription() string {
	return "with rainbow, hints as every color"
}

// ColorsAvailableAsHint returns all the suits of the standard ruleset (i.e. without
// rainbow) under the rainbow-as-compound-for-hints rules.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) ColorsAvailableAsHint() []string {
	return compoundRainbow.standardWithoutRainbowRuleset.ColorSuits()
}

// AfterColorHint returns the knowledge about a hand that a player has after applying
// the given hint about color to the given knowledge about the hand prior to the hint.
// Under this ruleset, if the card is a rainbow card, it should count as "marked" when
// any color hint is given. This means that any card "marked" by a hint could be either
// the color of the hint or a rainbow card. If a card is "marked" by hints for two
// different colors, then it is inferred to be a rainbow card. If it is "ignored" by a
// hint, then it is inferred as not that color and also not a rainbow card.
func (compoundRainbow *RainbowAsCompoundSuitRuleset) AfterColorHint(
	knowledgeBeforeHint []card.Inferred,
	cardsInHand []card.Defined,
	hintedColor string) []card.Inferred {
	handSize := len(cardsInHand)
	knowledgeAfterHint := make([]card.Inferred, handSize)
	for indexInHand := 0; indexInHand < handSize; indexInHand++ {
		colorOfCard := cardsInHand[indexInHand].ColorSuit
		originalColors :=
			knowledgeBeforeHint[indexInHand].PossibleColors
		replacementColors := []string{}

		if (colorOfCard == RainbowSuit) || (colorOfCard == hintedColor) {
			// A hint in this ruleset can never be for rainbow directly, so we can
			// remove all possibilities other than the hinted color and the rainbow
			// color. In either case, the other option (rainbow if the card is not
			// rainbow, or the hinted color if the card is rainbow) may have been
			// already ruled out by another hint, or the card's suit may end up as
			// the only possibility left after this hint.
			for _, possibleColor := range originalColors {
				if (possibleColor == hintedColor) || (possibleColor == RainbowSuit) {
					replacementColors = append(replacementColors, possibleColor)
				}
			}
		} else {
			// If the card is "ignored" by this hint, the player can infer that it is
			// not the hinted color and not a rainbow card.
			for _, possibleColor := range originalColors {
				if (possibleColor == hintedColor) || (possibleColor == RainbowSuit) {
					continue
				}

				replacementColors = append(replacementColors, possibleColor)
			}
		}

		knowledgeAfterHint[indexInHand] =
			card.Inferred{
				PossibleColors:  replacementColors,
				PossibleIndices: knowledgeBeforeHint[indexInHand].PossibleIndices,
			}
	}

	return knowledgeAfterHint
}
