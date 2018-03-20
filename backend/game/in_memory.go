package game

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/benoleary/ilutulestikud/backend/chat"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// InMemoryCollection stores inMemoryState objects as State objects. The games are
// mapped to by their identifiers, as are the players.
type InMemoryCollection struct {
	mutualExclusion  sync.Mutex
	gameStates       map[string]State
	nameToIdentifier endpoint.NameToIdentifier
	gamesWithPlayers map[string][]State
}

// NewInMemoryCollection creates a Collection around a map of games.
func NewInMemoryCollection(nameToIdentifier endpoint.NameToIdentifier) *InMemoryCollection {
	return &InMemoryCollection{
		mutualExclusion:  sync.Mutex{},
		gameStates:       make(map[string]State, 1),
		nameToIdentifier: nameToIdentifier,
		gamesWithPlayers: make(map[string][]State, 0),
	}
}

// Add should add an element to the collection which is a new object implementing
// the State interface with information given by the endpoint.GameDefinition object.
func (inMemoryCollection *InMemoryCollection) Add(
	gameDefinition endpoint.GameDefinition,
	playerCollection player.Collection) (string, error) {
	if gameDefinition.GameName == "" {
		return "", fmt.Errorf("Game must have a name")
	}

	gameIdentifier := inMemoryCollection.nameToIdentifier.Identifier(gameDefinition.GameName)
	_, gameExists := inMemoryCollection.gameStates[gameIdentifier]

	if gameExists {
		return "", fmt.Errorf("Game %v already exists", gameDefinition.GameName)
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

	playerStates := make([]player.ReadOnly, numberOfPlayers)
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

	newGame :=
		NewInMemoryState(
			gameIdentifier,
			gameDefinition.GameName,
			gameRuleset,
			playerStates)

	inMemoryCollection.mutualExclusion.Lock()

	inMemoryCollection.gameStates[gameIdentifier] = newGame

	for _, playerName := range gameDefinition.PlayerIdentifiers {
		existingGamesWithPlayer := inMemoryCollection.gamesWithPlayers[playerName]
		inMemoryCollection.gamesWithPlayers[playerName] = append(existingGamesWithPlayer, newGame)
	}

	inMemoryCollection.mutualExclusion.Unlock()
	return gameIdentifier, nil
}

// Get should return the State corresponding to the given game identifier if it
// exists already (or else nil) along with whether the State exists, analogously
// to a standard Golang map.
func (inMemoryCollection *InMemoryCollection) Get(gameIdentifier string) (State, bool) {
	gameState, gameExists := inMemoryCollection.gameStates[gameIdentifier]
	return gameState, gameExists
}

// All should return a slice of all the State instances in the collection which
// have the given player as a participant. The order is not mandated, and may even
// change with repeated calls to the same unchanged Collection (analogously to the
// entry set of a standard Golang map, for example), though of course an
// implementation may order the slice consistently.
func (inMemoryCollection *InMemoryCollection) All(playerName string) []State {
	return inMemoryCollection.gamesWithPlayers[playerName]
}

// inMemoryState is a struct meant to encapsulate all the state required for a single game to function.
type inMemoryState struct {
	mutualExclusion      sync.Mutex
	gameIdentifier       string
	gameName             string
	gameRuleset          Ruleset
	creationTime         time.Time
	participatingPlayers []player.ReadOnly
	chatLog              *chat.Log
	turnNumber           int
	currentScore         int
	numberOfReadyHints   int
	numberOfMistakesMade int
	undrawnDeck          []Card
}

// NewInMemoryState creates a new game given the required information.
func NewInMemoryState(
	gameIdentifier string,
	gameName string,
	gameRuleset Ruleset,
	playerStates []player.ReadOnly) State {
	return NewInMemoryStateWithGivenSeed(
		gameIdentifier,
		gameName,
		gameRuleset,
		playerStates,
		time.Now().UnixNano())
}

// NewInMemoryStateWithGivenSeed creates a new game given the required information,
// using the given seed for the random number generator used to shuffle the deck
// initially.
func NewInMemoryStateWithGivenSeed(
	gameIdentifier string,
	gameName string,
	gameRuleset Ruleset,
	playerStates []player.ReadOnly, randomNumberSeed int64) State {
	randomNumberGenerator := rand.New(rand.NewSource(randomNumberSeed))

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

	return &inMemoryState{
		mutualExclusion:      sync.Mutex{},
		gameIdentifier:       gameIdentifier,
		gameName:             gameName,
		gameRuleset:          gameRuleset,
		creationTime:         time.Now(),
		participatingPlayers: playerStates,
		chatLog:              chat.NewLog(),
		turnNumber:           1,
		numberOfReadyHints:   MaximumNumberOfHints,
		numberOfMistakesMade: 0,
		undrawnDeck:          shuffledDeck,
	}
}

// Identifier returns the private gameIdentifier field.
func (gameState *inMemoryState) Identifier() string {
	return gameState.gameIdentifier
}

// Name returns the value of the private gameName string.
func (gameState *inMemoryState) Name() string {
	return gameState.gameName
}

// Ruleset returns the ruleset for the game.
func (gameState *inMemoryState) Ruleset() Ruleset {
	return gameState.gameRuleset
}

// Name returns a slice of the private participatingPlayers array.
func (gameState *inMemoryState) Players() []player.ReadOnly {
	return gameState.participatingPlayers
}

// Name returns the value of the private turnNumber int.
func (gameState *inMemoryState) Turn() int {
	return gameState.turnNumber
}

// CreationTime returns the value of the private time object describing the time at
// which the state was created.
func (gameState *inMemoryState) CreationTime() time.Time {
	return gameState.creationTime
}

// HasPlayerAsParticipant returns true if the given player identifier matches
// the identifier of any of the game's participating players.
// This could be done with using a map[string]bool for player identifier mapped
// to whether or not a participant, but it's more effort to set up the map than
// would be gained in performance here.
func (gameState *inMemoryState) HasPlayerAsParticipant(playerIdentifier string) bool {
	for _, participatingPlayer := range gameState.participatingPlayers {
		if participatingPlayer.Identifier() == playerIdentifier {
			return true
		}
	}

	return false
}

// PerformAction should perform the given action for its player or return an error,
// but right now it only performs the action to record a chat message.
func (gameState *inMemoryState) PerformAction(
	actingPlayer player.ReadOnly, playerAction endpoint.PlayerAction) error {
	if playerAction.ActionType == "chat" {
		gameState.chatLog.AppendNewMessage(
			actingPlayer.Name(),
			actingPlayer.Color(),
			playerAction.ChatMessage)
		return nil
	}

	return fmt.Errorf("Unknown action: %v", playerAction.ActionType)
}

// ChatLog returns the chat log of the game at the current moment.
func (gameState *inMemoryState) ChatLog() *chat.Log {
	return gameState.chatLog
}

// Score returns the total score of the cards which have been correctly played in the
// game so far.
func (gameState *inMemoryState) Score() int {
	return gameState.currentScore
}

// NumberOfReadyHints returns the total number of hints which are available to be
// played.
func (gameState *inMemoryState) NumberOfReadyHints() int {
	return gameState.numberOfReadyHints
}

// NumberOfMistakesMade returns the total number of cards which have been played
// incorrectly.
func (gameState *inMemoryState) NumberOfMistakesMade() int {
	return gameState.numberOfMistakesMade
}
