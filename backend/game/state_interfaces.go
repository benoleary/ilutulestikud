package game

import (
	"time"

	"github.com/benoleary/ilutulestikud/backend/chat"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// ReadonlyState defines the interface for structs which should provide read-only information
// which can completely describe the state of a game.
type ReadonlyState interface {
	// Name should return the name of the game as known to the players.
	Name() string

	// Ruleset should return the ruleset for the game.
	Ruleset() Ruleset

	// Players should return the list of players participating in the game, in the order in
	// which they have their first turns.
	Players() []player.ReadonlyState

	// Turn should given the number of the turn (with thfirst turn being 1 rather than 0) which
	// is the current turn in the game (assuming 1 turn per player, not 1 turn being when all
	// players have acted and play returns to the first player).
	Turn() int

	// CreationTime should return the time object describing the time at which the state
	// was created.
	CreationTime() time.Time

	// ChatLog should return the chat log of the game at the current moment.
	ChatLog() *chat.Log

	// Score should return the total score of the cards which have been correctly played in
	// the game so far.
	Score() int

	// NumberOfReadyHints should return the total number of hints which are available to be
	// played.
	NumberOfReadyHints() int

	// NumberOfMistakesMade should return the total number of cards which have been played
	// incorrectly.
	NumberOfMistakesMade() int

	// DeckSize should return the number of cards left to draw from the deck.
	DeckSize() int

	// LastPlayedForColor should return the last card which has been played correctly for
	// the given color suit.
	LastPlayedForColor(colorSuit string) ReadonlyCard

	// NumberOfDiscardedCards should return the number of cards with the given suit and index
	// which were discarded or played incorrectly.
	NumberOfDiscardedCards(colorSuit string, sequenceIndex int) int

	// VisibleCardInHand should return the card held by the given player in the given position.
	VisibleCardInHand(holdingPlayerName string, indexInHand int) ReadonlyCard

	// InferredCardInHand should return the inferred information about the card held by the
	// given player in the given position.
	InferredCardInHand(holdingPlayerName string, indexInHand int) InferredCard
}

// ReadAndWriteState defines the interface for structs which should encapsulate the state of
// a single game.
type ReadAndWriteState interface {
	// Read should return the state as a read-only object for the purposes of reading
	// properties.
	Read() ReadonlyState

	// RecordChatMessage should record a chat message from the given player.
	RecordChatMessage(actingPlayer player.ReadonlyState, chatMessage string) error

	// ReplaceCardInHand should replace the card at the given index in the hand of the given
	// player with the given replacement card, and return the card which has just been
	// replaced.
	ReplaceCardInHand(
		holdingPlayerName string,
		indexInHand int,
		replacementCard ReadonlyCard) (ReadonlyCard, error)

	// AddCardToPlayedSequence should add the given card to the appropriate sequence of played
	// cards.
	AddCardToPlayedSequence(discardedCard ReadonlyCard) error

	// AddCardToDiscardPile should add the given card to the pile of discarded cards.
	AddCardToDiscardPile(discardedCard ReadonlyCard) error
}

// StatePersister defines the interface for structs which should be able to create objects
// implementing the readAndWriteState interface encapsulating the state information for
// individual games, and for tracking the games by their name.
type StatePersister interface {
	// randomSeed should provide an int64 which can be used as a seed for the
	// rand.NewSource(...) function.
	randomSeed() int64

	// addGame should add an element to the collection which is a new object implementing
	// the readAndWriteState interface from the given argument. It should return an error
	// if a game with the given name already exists.
	addGame(
		gameName string,
		gameRuleset Ruleset,
		playerStates []player.ReadonlyState,
		initialShuffle []ReadonlyCard) error

	// readAllWithPlayer should return a slice of all the games in the collection which
	// have the given player as a participant, where each game is given as a ReadonlyState
	// instance.
	// The order is not mandated, and may even change with repeated calls to the same
	// unchanged persister (analogously to the entry set of a standard Golang map, for
	// example), though of course an implementation may order the slice consistently.
	readAllWithPlayer(playerName string) []ReadonlyState

	// readAndWriteGame should return the readAndWriteState corresponding to the given game
	// name if it exists already (or else nil) along with whether the game exists,
	// analogously to a standard Golang map.
	readAndWriteGame(gameName string) (readAndWriteState, bool)
}
