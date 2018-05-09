package game

import (
	"time"

	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/chat"
	"github.com/benoleary/ilutulestikud/backend/player"
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

// ReadonlyState defines the interface for structs which should provide read-only
// information which can completely describe the state of a game.
type ReadonlyState interface {
	// Name should return the name of the game as known to the players.
	Name() string

	// Ruleset should return the ruleset for the game.
	Ruleset() Ruleset

	// Players should return the list of players participating in the game, in the
	// order in which they have their first turns.
	Players() []player.ReadonlyState

	// CreationTime should return the time object describing the time at which the
	// state was created.
	CreationTime() time.Time

	// ChatLog should return the chat log of the game at the current moment.
	ChatLog() *chat.Log

	// Turn should given the number of the turn (with the first turn being 1 rather
	// than 0) which is the current turn in the game (assuming 1 turn per player,
	// not 1 turn being when all players have acted and play returns to the first
	// player).
	Turn() int

	// Score should return the total score of the cards which have been correctly
	// played in the game so far.
	Score() int

	// NumberOfReadyHints should return the total number of hints which are available
	// to be played.
	NumberOfReadyHints() int

	// NumberOfMistakesMade should return the total number of cards which have been
	// played incorrectly.
	NumberOfMistakesMade() int

	// DeckSize should return the number of cards left to draw from the deck.
	DeckSize() int

	// LastPlayedForColor should return the last card which has been played correctly
	// for the given color suit along with whether any card has been played in that
	// suit so far, analogously to how a Go map works.
	LastPlayedForColor(colorSuit string) (card.Readonly, bool)

	// NumberOfDiscardedCards should return the number of cards with the given suit
	// and index which were discarded or played incorrectly.
	NumberOfDiscardedCards(colorSuit string, sequenceIndex int) int

	// VisibleCardInHand should return the card held by the given player in the given
	// position.
	VisibleCardInHand(holdingPlayerName string, indexInHand int) (card.Readonly, error)

	// InferredCardInHand should return the inferred information about the card held by
	// the given player in the given position in their hand.
	InferredCardInHand(holdingPlayerName string, indexInHand int) (card.Inferred, error)
}

// ReadAndWriteState defines the interface for structs which should encapsulate the
// state of a single game.
type ReadAndWriteState interface {
	// Read should return the state as a read-only object for the purposes of reading
	// properties.
	Read() ReadonlyState

	// RecordChatMessage should record a chat message from the given player.
	RecordChatMessage(actingPlayer player.ReadonlyState, chatMessage string) error

	// DrawCard should return the top-most card of the deck, or a card representing an
	// error  along with an actual Golang error if there are no cards left.
	DrawCard() (card.Readonly, error)

	// ReplaceCardInHand should replace the card at the given index in the hand of the
	// given player with the given replacement card, and return the card which has just
	// been replaced.
	ReplaceCardInHand(
		holdingPlayerName string,
		indexInHand int,
		replacementCard card.Readonly) (card.Readonly, error)

	// AddCardToPlayedSequence should add the given card to the appropriate sequence of
	// played cards.
	AddCardToPlayedSequence(playedCard card.Readonly) error

	// AddCardToDiscardPile should add the given card to the pile of discarded cards.
	AddCardToDiscardPile(discardedCard card.Readonly) error
}

// StatePersister defines the interface for structs which should be able to create
// objects implementing the ReadAndWriteState interface encapsulating the state
// information for individual games, and for tracking the games by their name.
type StatePersister interface {
	// RandomSeed should provide an int64 which can be used as a seed for the
	// rand.NewSource(...) function.
	RandomSeed() int64

	// ReadAndWriteGame should return the ReadAndWriteState corresponding to the given
	// game name, or nil with an error if it does not exist.
	ReadAndWriteGame(gameName string) (ReadAndWriteState, error)

	// ReadAllWithPlayer should return a slice of all the games in the collection which
	// have the given player as a participant, where each game is given as a
	// ReadonlyState instance.
	// The order is not mandated, and may even change with repeated calls to the same
	// unchanged persister (analogously to the entry set of a standard Golang map, for
	// example), though of course an implementation may order the slice consistently.
	ReadAllWithPlayer(playerName string) []ReadonlyState

	// AddGame should add an element to the collection which is a new object implementing
	// the readAndWriteState interface from the given argument. It should return an error
	// if a game with the given name already exists.
	AddGame(
		gameName string,
		gameRuleset Ruleset,
		playerStates []player.ReadonlyState,
		initialDeck []card.Readonly) error
}
