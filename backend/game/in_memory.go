package game

import (
	"sync"
	"time"

	"github.com/benoleary/ilutulestikud/backend/chat"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// InMemoryCollection stores inMemoryState objects as State objects.
type InMemoryCollection struct {
	mutualExclusion  sync.Mutex
	gameStates       map[string]State
	gamesWithPlayers map[string][]State
}

// NewInMemoryCollection creates a Collection around a map of games.
func NewInMemoryCollection() *InMemoryCollection {
	return &InMemoryCollection{
		mutualExclusion:  sync.Mutex{},
		gameStates:       make(map[string]State, 1),
		gamesWithPlayers: make(map[string][]State, 0),
	}
}

// Add should add an element to the collection which is a new object implementing
// the State interface with information given by the endpoint.GameDefinition object.
func (inMemoryCollection *InMemoryCollection) Add(
	gameDefinition endpoint.GameDefinition,
	playerCollection player.Collection) {
	numberOfPlayers := len(gameDefinition.Players)
	playerStates := make([]player.State, numberOfPlayers)
	for playerIndex := 0; playerIndex < numberOfPlayers; playerIndex++ {
		playerStates[playerIndex], _ = playerCollection.Get(gameDefinition.Players[playerIndex])
	}

	newGame := &inMemoryState{
		mutualExclusion:      sync.Mutex{},
		gameName:             gameDefinition.Name,
		participatingPlayers: playerStates,
		turnNumber:           1,
		chatLog:              chat.NewLog(),
	}

	inMemoryCollection.mutualExclusion.Lock()

	inMemoryCollection.gameStates[gameDefinition.Name] = newGame

	for _, playerName := range gameDefinition.Players {
		existingGamesWithPlayer := inMemoryCollection.gamesWithPlayers[playerName]
		inMemoryCollection.gamesWithPlayers[playerName] = append(existingGamesWithPlayer, newGame)
	}

	inMemoryCollection.mutualExclusion.Unlock()
}

// Get should return the State corresponding to the given game name if it exists
// already (or else nil) along with whether the State exists, analogously to a
// standard Golang map.
func (inMemoryCollection *InMemoryCollection) Get(gameName string) (State, bool) {
	gameState, gameExists := inMemoryCollection.gameStates[gameName]
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
	gameName             string
	creationTime         time.Time
	participatingPlayers []player.State
	turnNumber           int
	chatLog              *chat.Log
}

// Name returns the value of the private name string.
func (gameState *inMemoryState) Name() string {
	return gameState.gameName
}

// Name returns the value of the private name string.
func (gameState *inMemoryState) Players() []player.State {
	return gameState.participatingPlayers
}

// Name returns the value of the private name string.
func (gameState *inMemoryState) Turn() int {
	return gameState.turnNumber
}

// CreationTime returns the value of the private time object describing the time at
// which the state was created.
func (gameState *inMemoryState) CreationTime() time.Time {
	return gameState.creationTime
}

// HasPlayerAsParticipant returns true if the given player name matches
// the name of any of the game's participating players.
// This could be done with using a map[string]bool for player name mapped
// to whether or not a participant, but it's more effort to set up the map
// than would be gained in performance here.
func (gameState *inMemoryState) HasPlayerAsParticipant(playerName string) bool {
	for _, participatingPlayer := range gameState.participatingPlayers {
		if participatingPlayer.Name() == playerName {
			return true
		}
	}

	return false
}

// ChatLog returns the chat log of the game at the current moment.
func (gameState *inMemoryState) ChatLog() *chat.Log {
	return gameState.chatLog
}

// RecordPlayerChatMessage adds the given new message to the end of the chat log
// and removes the oldest message from the top.
func (gameState *inMemoryState) RecordPlayerChatMessage(chattingPlayer player.State, chatMessage string) {
	gameState.chatLog.AppendNewMessage(chattingPlayer.Name(), chattingPlayer.Color(), chatMessage)
}
