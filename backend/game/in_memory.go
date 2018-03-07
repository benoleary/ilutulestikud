package game

import (
	"fmt"
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
	playerCollection player.Collection) error {
	gameIdentifier := inMemoryCollection.nameToIdentifier.Identifier(gameDefinition.Name)
	_, gameExists := inMemoryCollection.gameStates[gameIdentifier]

	if gameExists {
		return fmt.Errorf("Game %v already exists", gameDefinition.Name)
	}

	numberOfPlayers := len(gameDefinition.Players)
	playerStates := make([]player.State, numberOfPlayers)
	for playerIndex := 0; playerIndex < numberOfPlayers; playerIndex++ {
		playerState, playerExists := playerCollection.Get(gameDefinition.Players[playerIndex])

		if !playerExists {
			return fmt.Errorf("Player %v is not registered", gameDefinition.Players[playerIndex])
		}

		playerStates[playerIndex] = playerState
	}

	newGame := &inMemoryState{
		mutualExclusion:      sync.Mutex{},
		gameIdentifier:       gameIdentifier,
		gameName:             gameDefinition.Name,
		participatingPlayers: playerStates,
		turnNumber:           1,
		chatLog:              chat.NewLog(),
	}

	inMemoryCollection.mutualExclusion.Lock()

	inMemoryCollection.gameStates[gameIdentifier] = newGame

	for _, playerName := range gameDefinition.Players {
		existingGamesWithPlayer := inMemoryCollection.gamesWithPlayers[playerName]
		inMemoryCollection.gamesWithPlayers[playerName] = append(existingGamesWithPlayer, newGame)
	}

	inMemoryCollection.mutualExclusion.Unlock()
	return nil
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
	creationTime         time.Time
	participatingPlayers []player.State
	turnNumber           int
	chatLog              *chat.Log
}

// Identifier returns the private gameIdentifier field.
func (gameState *inMemoryState) Identifier() string {
	return gameState.gameIdentifier
}

// Name returns the value of the private gameName string.
func (gameState *inMemoryState) Name() string {
	return gameState.gameName
}

// Name returns a slice of the private participatingPlayers array.
func (gameState *inMemoryState) Players() []player.State {
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

// ChatLog returns the chat log of the game at the current moment.
func (gameState *inMemoryState) ChatLog() *chat.Log {
	return gameState.chatLog
}

// PerformAction should perform the given action for its player or return an error,
// but right now it only performs the action to record a chat message.
func (gameState *inMemoryState) PerformAction(
	actingPlayer player.State,
	playerAction endpoint.PlayerAction) error {
	if playerAction.Action == "chat" {
		gameState.chatLog.AppendNewMessage(
			actingPlayer.Name(),
			actingPlayer.Color(),
			playerAction.ChatMessage)
	}

	return nil
}