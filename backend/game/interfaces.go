package game

import (
	"time"

	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// ReadonlyPlayerProvider defines an interface for structs to provide
// player.ReadonlyStates for given player names.
type ReadonlyPlayerProvider interface {
	Get(playerName string) (player.ReadonlyState, error)
}

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

	// CurrentTurnOrder should return the names of the participants of the game in the
	// order which their next turns are in, along with the index of the viewing
	// player in that list.
	CurrentTurnOrder() ([]string, int)

	// Turn should just wrap around the read-only game state's Turn function.
	Turn() int

	// Score should just wrap around the read-only game state's Score function.
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
	TakeTurnByDiscarding(indexInHandToDiscard int) error
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
	ChatLog() []message.Readonly

	// ActionLog should return the action log of the game at the current moment.
	ActionLog() []message.Readonly

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

	// PlayedForColor should return the cards, in order, which have been played
	// correctly for the given color suit.
	PlayedForColor(colorSuit string) []card.Readonly

	// NumberOfDiscardedCards should return the number of cards with the given suit
	// and index which were discarded or played incorrectly.
	NumberOfDiscardedCards(colorSuit string, sequenceIndex int) int

	// VisibleHand should return the card helds by the given player.
	VisibleHand(holdingPlayerName string) ([]card.Readonly, error)

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
	RecordChatMessage(actingPlayer player.ReadonlyState, chatMessage string) error

	// EnactTurnByDiscardingAndReplacing should increment the turn number and move the
	// card in the acting player's hand at the given index into the discard pile, and
	// replace it in the player's hand with the next card from the deck, bundled with
	// the given knowledge about the new card from the deck which the player should
	// have (which should always be that any color suit is possible and any sequence
	// index is possible). It should also add the given numbers to the counts of
	// available hints and mistakes made respectively.
	EnactTurnByDiscardingAndReplacing(
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
	// any sequence index is possible). It should also add the given number of hints
	// to the count of ready hints available (such as when playing the end of sequence
	// gives a bonus hint).
	EnactTurnByPlayingAndReplacing(
		actionMessage string,
		actingPlayer player.ReadonlyState,
		indexInHand int,
		knowledgeOfDrawnCard card.Inferred,
		numberOfReadyHintsToAdd int) error
}

// PlayerNameWithHand is a struct to keep the initial hand of a player with the name,
// so that the player names can be passed in turn order with the hands kept with the
// holding player. (Using a map of player names to slices of cards would not preserve
// the order of the player names, but a slice of these structs does.)
type PlayerNameWithHand struct {
	PlayerName  string
	InitialHand []card.InHand
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
	// the ReadAndWriteState interface from the given argument. It should return an error
	// if a game with the given name already exists.
	AddGame(
		gameName string,
		chatLogLength int,
		initialActionLog []message.Readonly,
		gameRuleset Ruleset,
		playersInTurnOrderWithInitialHands []PlayerNameWithHand,
		initialDeck []card.Readonly) error
}
