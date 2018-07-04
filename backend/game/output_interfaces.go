package game

import (
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
)

// Ruleset should encapsulate the set of rules for a game as functions.
type Ruleset interface {
	// FrontendDescription should describe the ruleset succintly enough for the frontend.
	FrontendDescription() string

	// CopyOfFullCardset should return a new array populated with every card which should
	// be present for a game under the ruleset, including duplicates.
	CopyOfFullCardset() []card.Readonly

	// NumberOfCardsInPlayerHand should return the number of cards held
	// in a player's hand, dependent on the number of players in the game.
	NumberOfCardsInPlayerHand(numberOfPlayers int) int

	// ColorSuits should return the set of colors used as suits.
	ColorSuits() []string

	// DistinctPossibleIndices should return all the distinct indices for the cards across
	// all suits, as it is used to set up the initial state of inferred cards.
	DistinctPossibleIndices() []int

	// MinimumNumberOfPlayers should return the minimum number of players needed for a game.
	MinimumNumberOfPlayers() int

	// MaximumNumberOfPlayers should return the maximum number of players allowed for a game.
	MaximumNumberOfPlayers() int

	// MaximumNumberOfHints should return the maximum number of hints which can be available
	// at any instant.
	MaximumNumberOfHints() int

	// NumberOfMistakesIndicatingGameOver should return the number of mistakes which indicates
	// that the game is over with the players having zero score.
	NumberOfMistakesIndicatingGameOver() int

	// IsCardPlayable should return true if the given card can be played onto the given
	// sequence of cards already played in the cards's suit.
	IsCardPlayable(cardToPlay card.Readonly, cardsAlreadyPlayedInSuit []card.Readonly) bool

	// HintsForPlayingCard should return the number of hints to refresh upon successfully
	// playing the given card.
	HintsForPlayingCard(cardToEvaluate card.Readonly) int

	// PointsPerCard should return the points value of the given card.
	PointsForCard(cardToEvaluate card.Readonly) int
}

// ViewForPlayer should encapsulate functions to view the state of the game as seen by a
// particular player.
type ViewForPlayer interface {
	// GameName should just wrap around the read-only game state's Name function.
	GameName() string

	// RulesetDescription should return the description given by the ruleset of the game.
	RulesetDescription() string

	// ChatLog should return the chat log of the read-only game state.
	ChatLog() []message.Readonly

	// ActionLog should return the action log of the read-only game state.
	ActionLog() []message.Readonly

	// GameIsFinished should return true if the game is finished.
	GameIsFinished() (bool, error)

	// CurrentTurnOrder should return the names of the participants of the game in the
	// order which their next turns are in, along with the index of the viewing
	// player in that list.
	CurrentTurnOrder() ([]string, int)

	// Turn should just wrap around the read-only game state's Turn function.
	Turn() int

	// Score should derive the score from the cards in the played area.
	Score() int

	// NumberOfReadyHints should just wrap around the read-only game state's
	// NumberOfReadyHints function.
	NumberOfReadyHints() int

	// MaximumNumberOfHints should just wrap around the game's ruleset's maximum
	// number of hints.
	MaximumNumberOfHints() int

	// NumberOfMistakesMade should just wrap around the read-only game state's
	// NumberOfMistakesMade function.
	NumberOfMistakesMade() int

	// NumberOfMistakesIndicatingGameOver should just wrap around the game's
	// ruleset's NumberOfMistakesIndicatingGameOver.
	NumberOfMistakesIndicatingGameOver() int

	// DeckSize should just wrap around the read-only game state's DeckSize function.
	DeckSize() int

	// PlayedCards should list the cards in play, in slices per suit.
	PlayedCards() [][]card.Readonly

	// DiscardedCards should list the discarded cards, ordered by suit first then by index.
	DiscardedCards() []card.Readonly

	// VisibleHand should return the cards held by the given player along with the chat
	// color for that player, or nil and a string which will be ignored and an error if the
	// player cannot see the cards.
	VisibleHand(playerName string) ([]card.Readonly, string, error)

	// KnowledgeOfOwnHand should return the knowledge about the player's own cards which
	// was inferred directly from the hints officially given so far.
	KnowledgeOfOwnHand() ([]card.Inferred, error)
}

// ExecutorForPlayer should encapsulate functions to execute actions by a particular player
// on the state of the game.
type ExecutorForPlayer interface {
	// RecordChatMessage should record the given chat message from the acting player, or
	// return an error if it was not possible.
	RecordChatMessage(chatMessage string) error

	// TakeTurnByDiscarding should enact a turn by discarding the indicated card from the
	// hand of the acting player, or return an error if it was not possible.
	TakeTurnByDiscarding(indexInHand int) error

	// TakeTurnByPlaying should enact a turn by attempting to play the indicated card from
	// the hand of the acting player, resulting in the card going into the played area or
	// into the discard pile while causing a mistake, or return an error if it was not
	// possible.
	TakeTurnByPlaying(indexInHand int) error
}

// PlayerNameWithHand is a struct to keep the initial hand of a player with the name,
// so that the player names can be passed in turn order with the hands kept with the
// holding player. (Using a map of player names to slices of cards would not preserve
// the order of the player names, but a slice of these structs does.)
type PlayerNameWithHand struct {
	PlayerName  string
	InitialHand []card.InHand
}