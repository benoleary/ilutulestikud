package game

import (
	"sort"
	"time"

	"github.com/benoleary/ilutulestikud/backend/chat"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// State defines the interface for structs which should encapsulate the state of a single game.
type State interface {
	// Identifier should return the identifier of the game for interaction between frontend
	// and backend.
	Identifier() string

	// Name should return the name of the game as known to the players.
	Name() string

	// Ruleset should return the ruleset for the game.
	Ruleset() Ruleset

	// Players should return the list of players participating in the game, in the order in
	// which they have their first turns.
	Players() []player.ReadOnly

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

	// PerformAction should perform the given action for its player or return an error.
	PerformAction(actingPlayer player.ReadOnly, playerAction endpoint.PlayerAction) error

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

// ForPlayer writes the relevant parts of the state of the game as should be known by the given
// player into the relevant JSON object for the frontend.
func ForPlayer(state State, playerIdentifier string) endpoint.GameView {
	// The remaining attributes of the endpoint.GameView require some calculation based on the
	// game's ruleset.
	return endpoint.GameView{
		ChatLog:                      state.ChatLog().ForFrontend(),
		ScoreSoFar:                   state.Score(),
		NumberOfReadyHints:           state.NumberOfReadyHints(),
		NumberOfSpentHints:           MaximumNumberOfHints - state.NumberOfReadyHints(),
		NumberOfMistakesStillAllowed: MaximumNumberOfMistakesAllowed - state.NumberOfMistakesMade(),
		NumberOfMistakesMade:         state.NumberOfMistakesMade(),
	}
}

// Collection defines the interface for structs which should be able to create objects
// implementing the State interface encapsulating the state information for individual
// games, and for tracking the objects by their identifier, which is the game name.
type Collection interface {
	// Add should add an element to the collection which is a new object implementing
	// the State interface with information given by the endpoint.GameDefinition object,
	// and return the identifier of the newly-created game, along with an error which
	// of course should be nil if there was no problem.
	// The given player collection should be used as the source of player states to be
	// matched to names given in the game definition. It should return an error if a
	// game with the given name already exists, or if the definition includes invalid
	// players.
	Add(gameDefinition endpoint.GameDefinition,
		playerCollection player.Collection) (string, error)

	// Get should return the State corresponding to the given game identifier if it
	// exists already (or else nil) along with whether the State exists, analogously to
	// a standard Golang map.
	Get(gameIdentifier string) (State, bool)

	// All should return a slice of all the State instances in the collection which
	// have the given player as a participant. The order is not mandated, and may even
	// change with repeated calls to the same unchanged Collection (analogously to the
	// entry set of a standard Golang map, for example), though of course an
	// implementation may order the slice consistently.
	All(playerIdentifier string) []State
}

// ByCreationTime implements sort interface for []State based on the creationTime field.
type ByCreationTime []State

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

// TurnSummariesForFrontend writes the turn summary information for each game which has
// the given player into the relevant JSON object for the frontend.
func TurnSummariesForFrontend(collection Collection, playerIdentifier string) endpoint.TurnSummaryList {
	gameList := collection.All(playerIdentifier)

	sort.Sort(ByCreationTime(gameList))

	numberOfGamesWithPlayer := len(gameList)

	turnSummaries := make([]endpoint.TurnSummary, numberOfGamesWithPlayer)
	for gameIndex := 0; gameIndex < numberOfGamesWithPlayer; gameIndex++ {
		nameOfGame := gameList[gameIndex].Name()
		gameTurn := gameList[gameIndex].Turn()

		gameParticipants := gameList[gameIndex].Players()
		numberOfParticipants := len(gameParticipants)

		playerNamesInTurnOrder := make([]string, numberOfParticipants)

		turnsUntilPlayer := 0
		for playerIndex := 0; playerIndex < numberOfParticipants; playerIndex++ {
			// Game turns begin with 1 rather than 0, so this sets the player names in order,
			// wrapping index back to 0 when at the end of the list.
			// E.g. turn 3, 5 players: playerNamesInTurnOrder will start with
			// gameParticipants[2], then [3], then [4], then [0], then [1].
			playerInTurnOrder :=
				gameParticipants[(playerIndex+gameTurn-1)%numberOfParticipants]
			playerNamesInTurnOrder[playerIndex] =
				playerInTurnOrder.Name()

			if playerIdentifier == playerInTurnOrder.Identifier() {
				turnsUntilPlayer = playerIndex
			}
		}

		turnSummaries[gameIndex] = endpoint.TurnSummary{
			GameIdentifier:             gameList[gameIndex].Identifier(),
			GameName:                   nameOfGame,
			RulesetDescription:         gameList[gameIndex].Ruleset().FrontendDescription(),
			CreationTimestampInSeconds: gameList[gameIndex].CreationTime().Unix(),
			TurnNumber:                 gameTurn,
			PlayerNamesInNextTurnOrder: playerNamesInTurnOrder,
			IsPlayerTurn:               turnsUntilPlayer == 0,
		}
	}

	return endpoint.TurnSummaryList{TurnSummaries: turnSummaries}
}
