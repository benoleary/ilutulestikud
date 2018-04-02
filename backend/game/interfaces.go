package game

import (
	"time"

	"github.com/benoleary/ilutulestikud/backend/chat"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// ReadonlyState defines the interface for structs which should provide read-only information
// which can completely describe the state of a game.
type ReadonlyState interface {
	// Identifier should return the identifier of the game for interaction between frontend
	// and backend.
	Identifier() string

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

	// HasPlayerAsParticipant should return true if the given player identifier matches
	// the identifier of any of the game's participating players.
	HasPlayerAsParticipant(playerIdentifier string) bool

	// ChatLog should return the chat log of the game at the current moment.
	ChatLog() *chat.Log

	// Score should return the total score of the cards which have been correctly played in the
	// game so far.
	Score() int

	// NumberOfReadyHints should return the total number of hints which are available to be
	// played.
	NumberOfReadyHints() int

	// NumberOfMistakesMade should return the total number of cards which have been played
	// incorrectly.
	NumberOfMistakesMade() int
}

// readAndWriteState defines the interface for structs which should encapsulate the state of
// a single game.
type readAndWriteState interface {
	// Read should return the state as a read-only object for the purposes of reading
	// properties.
	read() ReadonlyState

	// performAction should perform the given action for its player or return an error.
	performAction(actingPlayer player.ReadonlyState, playerAction endpoint.PlayerAction) error
}

// ForPlayer writes the relevant parts of the state of the game as should be known by the given
// player into the relevant JSON object for the frontend.
func ForPlayer(readonlyState ReadonlyState, playerIdentifier string) endpoint.GameView {
	// The remaining attributes of the endpoint.GameView require some calculation based on the
	// game's ruleset.
	return endpoint.GameView{
		ChatLog:                      readonlyState.ChatLog().ForFrontend(),
		ScoreSoFar:                   readonlyState.Score(),
		NumberOfReadyHints:           readonlyState.NumberOfReadyHints(),
		NumberOfSpentHints:           MaximumNumberOfHints - readonlyState.NumberOfReadyHints(),
		NumberOfMistakesStillAllowed: MaximumNumberOfMistakesAllowed - readonlyState.NumberOfMistakesMade(),
		NumberOfMistakesMade:         readonlyState.NumberOfMistakesMade(),
	}
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
		initialShuffle []Card) error

	// readAllWithPlayer should return a slice of all the games in the collection which
	// have the given player as a participant, where each game is given as a ReadonlyState
	// instance.
	// The order is not mandated, and may even change with repeated calls to the same
	// unchanged Collection (analogously to the entry set of a standard Golang map, for
	// example), though of course an implementation may order the slice consistently.
	readAllWithPlayer(playerIdentifier string) []ReadonlyState

	// readAndWriteGame should return the readAndWriteState corresponding to the given game
	// identifier if it exists already (or else nil) along with whether the game exists,
	// analogously to a standard Golang map.
	readAndWriteGame(gameIdentifier string) (readAndWriteState, bool)
}

// ByCreationTime implements sort interface for []ReadonlyState based on the return
// from its CreationTime().
type ByCreationTime []ReadonlyState

// Len implements part of the sort interface for ByCreationTime.
func (byCreationTime ByCreationTime) Len() int {
	return len(byCreationTime)
}

// Swap implements part of the sort interface for ByCreationTime.
func (byCreationTime ByCreationTime) Swap(firstIndex int, secondIndex int) {
	byCreationTime[firstIndex], byCreationTime[secondIndex] =
		byCreationTime[secondIndex], byCreationTime[firstIndex]
}

// Less implements part of the sort interface for ByCreationTime.
func (byCreationTime ByCreationTime) Less(firstIndex int, secondIndex int) bool {
	return byCreationTime[firstIndex].CreationTime().Before(
		byCreationTime[secondIndex].CreationTime())
}
