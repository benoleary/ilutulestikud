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
	mutualExclusion       sync.Mutex
	randomNumberGenerator *rand.Rand
	gameStates            map[string]readAndWriteState
	nameToIdentifier      endpoint.NameToIdentifier
	gamesWithPlayers      map[string][]ReadonlyState
}

// NewInMemoryCollection creates a Collection around a map of games.
func NewInMemoryCollection(nameToIdentifier endpoint.NameToIdentifier) *InMemoryCollection {
	return &InMemoryCollection{
		mutualExclusion:       sync.Mutex{},
		randomNumberGenerator: rand.New(rand.NewSource(time.Now().Unix())),
		gameStates:            make(map[string]readAndWriteState, 1),
		nameToIdentifier:      nameToIdentifier,
		gamesWithPlayers:      make(map[string][]ReadonlyState, 0),
	}
}

// randomSeed provides an int64 which can be used as a seed for the
// rand.NewSource(...) function.
func (inMemoryCollection *InMemoryCollection) randomSeed() int64 {
	return inMemoryCollection.randomNumberGenerator.Int63()
}

// addGame adds an element to the collection which is a new object implementing
// the readAndWriteState interface from the given arguments, and returns the
// identifier of the newly-created game, along with an error which of course is
// nil if there was no problem. It returns an error if a game with the given name
// already exists.
func (inMemoryCollection *InMemoryCollection) addGame(
	gameName string,
	gameRuleset Ruleset,
	playerStates []player.ReadonlyState,
	initialShuffle []Card) (string, error) {
	if gameName == "" {
		return "", fmt.Errorf("Game must have a name")
	}

	gameIdentifier := inMemoryCollection.nameToIdentifier.Identifier(gameName)
	_, gameExists := inMemoryCollection.gameStates[gameIdentifier]

	if gameExists {
		return "", fmt.Errorf("Game %v already exists", gameName)
	}

	newGame :=
		newInMemoryState(
			gameIdentifier,
			gameName,
			gameRuleset,
			playerStates,
			initialShuffle)

	inMemoryCollection.mutualExclusion.Lock()

	inMemoryCollection.gameStates[gameIdentifier] = newGame

	for _, playerState := range playerStates {
		playerName := playerState.Name()
		existingGamesWithPlayer := inMemoryCollection.gamesWithPlayers[playerName]
		inMemoryCollection.gamesWithPlayers[playerName] =
			append(existingGamesWithPlayer, newGame.read())
	}

	inMemoryCollection.mutualExclusion.Unlock()
	return gameIdentifier, nil
}

// readAllWithPlayer returns a slice of all the ReadonlyState instances in the collection
// which have the given player as a participant. The order is not consistent with repeated
// calls, as it is defined by the entry set of a standard Golang map.
func (inMemoryCollection *InMemoryCollection) readAllWithPlayer(
	playerIdentifier string) []ReadonlyState {
	return inMemoryCollection.gamesWithPlayers[playerIdentifier]
}

// ReadGame returns the ReadonlyState corresponding to the given game identifier if
// it exists already (or else nil) along with whether the game exists, analogously
// to a standard Golang map.
func (inMemoryCollection *InMemoryCollection) readAndWriteGame(
	gameIdentifier string) (readAndWriteState, bool) {
	gameState, gameExists := inMemoryCollection.gameStates[gameIdentifier]
	return gameState, gameExists
}

// inMemoryState is a struct meant to encapsulate all the state required for a single game to function.
type inMemoryState struct {
	mutualExclusion      sync.Mutex
	gameIdentifier       string
	gameName             string
	gameRuleset          Ruleset
	creationTime         time.Time
	participatingPlayers []player.ReadonlyState
	chatLog              *chat.Log
	turnNumber           int
	currentScore         int
	numberOfReadyHints   int
	numberOfMistakesMade int
	undrawnDeck          []Card
}

// newInMemoryState creates a new game given the required information,
// using the given seed for the random number generator used to shuffle the deck
// initially.
func newInMemoryState(
	gameIdentifier string,
	gameName string,
	gameRuleset Ruleset,
	playerStates []player.ReadonlyState,
	shuffledDeck []Card) readAndWriteState {
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

// Read returns the gameState itself as a read-only object for the purposes of reading properties.
func (gameState *inMemoryState) read() ReadonlyState {
	return gameState
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
func (gameState *inMemoryState) Players() []player.ReadonlyState {
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

// HasPlayerAsParticipant returns true if the given player name matches the name of
// any of the game's participating players.
// This could be done with using a map[string]bool for player name mapped to whether
// or not the player is a participant, but it's more effort to set up the map than
// would be gained in performance here.
func (gameState *inMemoryState) HasPlayerAsParticipant(playerName string) bool {
	for _, participatingPlayer := range gameState.participatingPlayers {
		if participatingPlayer.Name() == playerName {
			return true
		}
	}

	return false
}

// performAction should perform the given action for its player or return an error,
// but right now it only performs the action to record a chat message.
func (gameState *inMemoryState) performAction(
	actingPlayer player.ReadonlyState, playerAction endpoint.PlayerAction) error {
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
