package game

import (
	"context"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// ReadonlyPlayerProvider defines an interface for structs to provide
// player.ReadonlyStates for given player names.
type ReadonlyPlayerProvider interface {
	Get(
		executionContext context.Context,
		playerName string) (player.ReadonlyState, error)
}

// ReadonlyState defines the interface for structs which should provide read-only
// information which can completely describe the state of a game.
type ReadonlyState interface {
	// Name should return the name of the game as known to the players.
	Name() string

	// Ruleset should return the ruleset for the game.
	Ruleset() Ruleset

	// PlayerNames should return the list of players participating in the game, in
	// the order in which they have their first turns.
	PlayerNames() []string

	// CreationTime should return the time object describing the time at which the
	// state was created.
	CreationTime() time.Time

	// ChatLog should return the chat log of the game at the current moment.
	ChatLog() []message.FromPlayer

	// ActionLog should return the action log of the game at the current moment.
	ActionLog() []message.FromPlayer

	// Turn should given the number of the turn (with the first turn being 1 rather
	// than 0) which is the current turn in the game (assuming 1 turn per player,
	// not 1 turn being when all players have acted and play returns to the first
	// player).
	Turn() int

	// TurnsTakenWithEmptyDeck should return the number of turns which have been taken
	// since the turn which drew the last card from the deck.
	TurnsTakenWithEmptyDeck() int

	// NumberOfReadyHints should return the total number of hints which are available
	// to be played.
	NumberOfReadyHints() int

	// NumberOfMistakesMade should return the total number of cards which have been
	// played incorrectly.
	NumberOfMistakesMade() int

	// DeckSize should return the number of cards left to draw from the deck.
	DeckSize() int

	// PlayedForColor should return the cards, in order, which have been played
	// correctly for the given color suit.
	PlayedForColor(colorSuit string) []card.Defined

	// NumberOfDiscardedCards should return the number of cards with the given suit
	// and index which were discarded or played incorrectly.
	NumberOfDiscardedCards(colorSuit string, sequenceIndex int) int

	// VisibleHand should return the card helds by the given player.
	VisibleHand(holdingPlayerName string) ([]card.Defined, error)

	// InferredHand should return the inferred information about the cards held by
	// the given player.
	InferredHand(holdingPlayerName string) ([]card.Inferred, error)
}

// ReadAndWriteState defines the interface for structs which should encapsulate the
// state of a single game.
type ReadAndWriteState interface {
	// Read should return the state as a read-only object for the purposes of reading
	// properties.
	Read() ReadonlyState

	// RecordChatMessage should record a chat message from the given player.
	RecordChatMessage(
		executionContext context.Context,
		actingPlayer player.ReadonlyState,
		chatMessage string) error

	// EnactTurnByDiscardingAndReplacing should increment the turn number and move the
	// card in the acting player's hand at the given index into the discard pile, and
	// replace it in the player's hand with the next card from the deck, bundled with
	// the given knowledge about the new card from the deck which the player should
	// have (which should always be that any color suit is possible and any sequence
	// index is possible). If there is no card to draw from the deck, it should
	// increment the number of turns taken with an empty deck of replacing the card in
	// the hand. It should also add the given numbers to the counts of available hints
	// and mistakes made respectively.
	EnactTurnByDiscardingAndReplacing(
		executionContext context.Context,
		actionMessage string,
		actingPlayer player.ReadonlyState,
		indexInHand int,
		knowledgeOfDrawnCard card.Inferred,
		numberOfReadyHintsToAdd int,
		numberOfMistakesMadeToAdd int) error

	// EnactTurnByPlayingAndReplacing should increment the turn number and move the
	// card in the acting player's hand at the given index into the appropriate color
	// sequence, and replace it in the player's hand with the next card from the deck,
	// bundled with the given knowledge about the new card from the deck which the
	// player should have (which should always be that any color suit is possible and
	// any sequence index is possible). If there is no card to draw from the deck, it
	// should increment the number of turns taken with an empty deck of replacing the
	// card in the hand. It should also add the given number of hints to the count of
	// ready hints available (such as when playing the end of sequence gives a bonus
	// hint).
	EnactTurnByPlayingAndReplacing(
		executionContext context.Context,
		actionMessage string,
		actingPlayer player.ReadonlyState,
		indexInHand int,
		knowledgeOfDrawnCard card.Inferred,
		numberOfReadyHintsToAdd int) error

	// EnactTurnByUpdatingHandWithHint should increment the turn number and replace
	// the given player's inferred hand with the given inferred hand, while also
	// decrementing the number of available hints appropriately. If the deck is empty,
	// this function should also increment the number of turns taken with an empty
	// deck.
	EnactTurnByUpdatingHandWithHint(
		executionContext context.Context,
		actionMessage string,
		actingPlayer player.ReadonlyState,
		receivingPlayerName string,
		updatedReceiverKnowledgeOfOwnHand []card.Inferred,
		numberOfReadyHintsToSubtract int) error
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
	ReadAndWriteGame(
		executionContext context.Context,
		gameName string) (ReadAndWriteState, error)

	// ReadAllWithPlayer should return a slice of all the games in the collection which
	// have the given player as a participant, where each game is given as a
	// ReadonlyState instance.
	// The order is not mandated, and may even change with repeated calls to the same
	// unchanged persister (analogously to the entry set of a standard Golang map, for
	// example), though of course an implementation may order the slice consistently.
	ReadAllWithPlayer(
		executionContext context.Context,
		playerName string) ([]ReadonlyState, error)

	// AddGame should add an element to the collection which is a new object implementing
	// the ReadAndWriteState interface from the given argument. It should return an error
	// if a game with the given name already exists.
	AddGame(
		executionContext context.Context,
		gameName string,
		chatLogLength int,
		initialActionLog []message.FromPlayer,
		gameRuleset Ruleset,
		playersInTurnOrderWithInitialHands []PlayerNameWithHand,
		initialDeck []card.Defined) error

	// RemoveGameFromListForPlayer should remove the given player from the given game in
	// the sense that the game will no longer show up in the result of
	// ReadAllWithPlayer(playerName). It should return an error if the player is not a
	// participant of the game, as well as in general I/O errors and so on.
	RemoveGameFromListForPlayer(
		executionContext context.Context,
		gameName string,
		playerName string) error

	// Delete should delete the given game from the persistence store.
	Delete(executionContext context.Context, gameName string) error
}
