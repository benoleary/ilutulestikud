package game

import (
	"fmt"
	"math/rand"
	"sort"
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

// StateCollection defines the interface for structs which should be able to create objects
// implementing the readAndWriteState interface encapsulating the state information for
// individual games, and for tracking the objects by their identifier, which is an encoded
// form of the game name.
type StateCollection interface {
	// randomSeed should provide an int64 which can be used as a seed for the
	// rand.NewSource(...) function.
	randomSeed() int64

	// addGame should add an element to the collection which is a new object implementing
	// the readAndWriteState interface from the given arguments, and return the identifier
	// of the newly-created game, along with an error which of course should be nil if
	// there was no problem. It should return an error if a game with the given name
	// already exists.
	addGame(
		gameName string,
		gameRuleset Ruleset,
		playerStates []player.ReadonlyState,
		initialShuffle []Card) (string, error)

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

// ReadState returns the ReadonlyState corresponding to the given game identifier if
// it exists in the given collection already (or else nil) along with whether the game
// exists, analogously to a standard Golang map.
func ReadState(gameCollection StateCollection, gameIdentifier string) (ReadonlyState, bool) {
	gameState, gameExists := gameCollection.readAndWriteGame(gameIdentifier)

	if gameState == nil {
		return nil, false
	}

	return gameState.read(), gameExists
}

// PerformAction finds the given game and performs the given action for its player,
// or returns an error.
func PerformAction(
	gameCollection StateCollection, playerCollection player.StatePersister,
	playerAction endpoint.PlayerAction) error {
	actingPlayer, playeridentificationError :=
		playerCollection.Get(playerAction.PlayerIdentifier)

	if playeridentificationError != nil {
		return playeridentificationError
	}

	gameState, isFound := gameCollection.readAndWriteGame(playerAction.GameIdentifier)

	if !isFound {
		return fmt.Errorf(
			"Game %v does not exist, cannot perform action from player %v",
			playerAction.GameIdentifier,
			playerAction.PlayerIdentifier)
	}

	if !gameState.read().HasPlayerAsParticipant(playerAction.PlayerIdentifier) {
		return fmt.Errorf(
			"Player %v is not a participant in game %v",
			playerAction.PlayerIdentifier,
			playerAction.GameIdentifier)
	}

	return gameState.performAction(actingPlayer, playerAction)
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

// AddNew prepares a new shuffled deck using a random seed taken from the given
// collection, and uses it to create a new game in the given collection from the
// given definition. It returns an error if a game with the given name already
// exists, or if the definition includes invalid players.
func AddNew(
	gameDefinition endpoint.GameDefinition,
	gameCollection StateCollection,
	playerCollection player.StatePersister) (string, error) {
	if gameCollection == nil {
		return "Error", fmt.Errorf("Cannot create a game in a nil collection")
	}

	return AddNewWithGivenRandomSeed(
		gameDefinition,
		gameCollection,
		playerCollection,
		gameCollection.randomSeed())
}

// AddNewWithGivenRandomSeed prepares a new shuffled deck using the given seed for
// a random number generator, and uses it to create a new game in the given collection
// from the given definition. It returns an error if a game with the given name already
// exists, or if the definition includes invalid players.
func AddNewWithGivenRandomSeed(
	gameDefinition endpoint.GameDefinition,
	gameCollection StateCollection,
	playerCollection player.StatePersister,
	randomSeed int64) (string, error) {
	if gameDefinition.GameName == "" {
		return "", fmt.Errorf("Game must have a name")
	}

	gameRuleset, unknownRulesetError := RulesetFromIdentifier(gameDefinition.RulesetIdentifier)
	if unknownRulesetError != nil {
		return "", fmt.Errorf(
			"Problem identifying ruleset from identifier %v; error is: %v",
			gameDefinition.RulesetIdentifier,
			unknownRulesetError)
	}

	// A nil slice still has a length of 0, so this is OK.
	numberOfPlayers := len(gameDefinition.PlayerIdentifiers)

	if numberOfPlayers < gameRuleset.MinimumNumberOfPlayers() {
		return "", fmt.Errorf(
			"Game must have at least %v players",
			gameRuleset.MinimumNumberOfPlayers())
	}

	if numberOfPlayers > gameRuleset.MaximumNumberOfPlayers() {
		return "", fmt.Errorf(
			"Game must have no more than %v players",
			gameRuleset.MaximumNumberOfPlayers())
	}

	playerIdentifiers := make(map[string]bool, 0)

	playerStates := make([]player.ReadonlyState, numberOfPlayers)
	for playerIndex := 0; playerIndex < numberOfPlayers; playerIndex++ {
		playerIdentifier := gameDefinition.PlayerIdentifiers[playerIndex]
		playerState, identificationError := playerCollection.Get(playerIdentifier)

		if identificationError != nil {
			return "", identificationError
		}

		if playerIdentifiers[playerIdentifier] {
			return "", fmt.Errorf(
				"Player with identifier %v appears more than once in the list of players",
				playerIdentifier)
		}

		playerIdentifiers[playerIdentifier] = true

		playerStates[playerIndex] = playerState
	}

	randomNumberGenerator := rand.New(rand.NewSource(randomSeed))

	shuffledDeck := gameRuleset.FullCardset()

	numberOfCards := len(shuffledDeck)

	// This is probably excessive.
	numberOfShuffles := 8 * numberOfCards

	for shuffleCount := 0; shuffleCount < numberOfShuffles; shuffleCount++ {
		firstShuffleIndex := randomNumberGenerator.Intn(numberOfCards)
		secondShuffleIndex := randomNumberGenerator.Intn(numberOfCards)
		shuffledDeck[firstShuffleIndex], shuffledDeck[secondShuffleIndex] =
			shuffledDeck[secondShuffleIndex], shuffledDeck[firstShuffleIndex]
	}

	return gameCollection.addGame(
		gameDefinition.GameName,
		gameRuleset,
		playerStates,
		shuffledDeck)
}

// TurnSummariesForFrontend writes the turn summary information for each game which has
// the given player into the relevant JSON object for the frontend.
func TurnSummariesForFrontend(gameCollection StateCollection, playerIdentifier string) endpoint.TurnSummaryList {
	gameList := gameCollection.readAllWithPlayer(playerIdentifier)

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
