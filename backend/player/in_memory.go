package player

import (
	"fmt"
	"sync"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// InMemoryCollection stores inMemoryState objects as State objects. The
// players are mapped to by their identifiers.
type InMemoryCollection struct {
	mutualExclusion     sync.Mutex
	playerStates        map[string]*inMemoryState
	nameToIdentifier    endpoint.NameToIdentifier
	initialPlayerNames  map[string]bool
	availableChatColors []string
	numberOfChatColors  int
}

// NewInMemoryCollection creates a Collection around a map of players created
// from the given initial player names, with colors according to the available
// chat colors.
func NewInMemoryCollection(
	nameToIdentifier endpoint.NameToIdentifier,
	initialPlayerNames []string,
	availableChatColors []string) *InMemoryCollection {
	numberOfPlayers := len(initialPlayerNames)
	initialPlayerNameSet := make(map[string]bool, numberOfPlayers)
	for _, initialPlayerName := range initialPlayerNames {
		initialPlayerNameSet[initialPlayerName] = true
	}

	numberOfColors := len(availableChatColors)
	deepCopyOfChatColors := make([]string, numberOfColors)
	copy(deepCopyOfChatColors, availableChatColors)

	newCollection := &InMemoryCollection{
		mutualExclusion:     sync.Mutex{},
		playerStates:        make(map[string]*inMemoryState, numberOfPlayers),
		nameToIdentifier:    nameToIdentifier,
		initialPlayerNames:  initialPlayerNameSet,
		availableChatColors: deepCopyOfChatColors,
		numberOfChatColors:  numberOfColors,
	}

	for playerCount := 0; playerCount < numberOfPlayers; playerCount++ {
		newCollection.addWithDefaultColorWithoutCounting(
			initialPlayerNames[playerCount],
			playerCount,
			numberOfColors)
	}

	return newCollection
}

// Add creates a new inMemoryState object with name and color parsed from the given
// endpoint.PlayerState (choosing a default color if the endpoint.PlayerState did not
// provide one), and adds a reference to it into the collection.
func (inMemoryCollection *InMemoryCollection) Add(
	endpointPlayer endpoint.PlayerState) (string, error) {
	if endpointPlayer.Color == "" {
		return inMemoryCollection.addWithDefaultColor(endpointPlayer.Name)
	}

	return inMemoryCollection.addWithGivenColor(endpointPlayer.Name, endpointPlayer.Color)
}

// UpdateFromPresentAttributes updates the player identified by the endpoint.PlayerState
// by over-writing all non-name string attributes with those from updaterReference, except
// for strings in updaterReference which are empty strings. It uses a mutex to ensure thread
// safety.
func (inMemoryCollection *InMemoryCollection) UpdateFromPresentAttributes(
	updaterReference endpoint.PlayerState) error {
	playerToUpdate, playerExists := inMemoryCollection.playerStates[updaterReference.Identifier]

	if !playerExists {
		return fmt.Errorf(
			"No player with identifier %v is registered",
			updaterReference.Identifier)
	}

	// It would be more efficient to only lock if we go into an if statement,
	// but then multiple if statements would be less efficient, and there would
	// be a mutex in each if statement.
	inMemoryCollection.mutualExclusion.Lock()

	if updaterReference.Color != "" {
		playerToUpdate.color = updaterReference.Color
	}

	inMemoryCollection.mutualExclusion.Unlock()

	return nil
}

// Get returns the ReadOnly corresponding to the given player identifier if it exists
// already along with an error which is nil if there was no problem. If the player does
// not exist, a non-nil error is returned along with a nil ReadOnly.
func (inMemoryCollection *InMemoryCollection) Get(playerIdentifier string) (ReadOnly, error) {
	playerState, playerExists := inMemoryCollection.playerStates[playerIdentifier]
	if !playerExists {
		return nil, fmt.Errorf(
			"No player with identifier %v is registered",
			playerIdentifier)
	}

	return playerState, nil
}

// All returns a slice of all the ReadOnly instances in the collection, ordered in the
// random way the iteration over the entries of a Golang map normally is.
func (inMemoryCollection *InMemoryCollection) All() []ReadOnly {
	playerList := make([]ReadOnly, 0, len(inMemoryCollection.playerStates))
	for _, playerState := range inMemoryCollection.playerStates {
		playerList = append(playerList, playerState)
	}

	return playerList
}

// Reset removes all players which are not among the initial players.
// It does not restore any initial players who have been removed as
// there is no possibility to remove them anyway.
func (inMemoryCollection *InMemoryCollection) Reset() {
	playersToRemove := make([]string, 0)
	for _, playerState := range inMemoryCollection.playerStates {
		if !inMemoryCollection.initialPlayerNames[playerState.Name()] {
			playersToRemove = append(playersToRemove, playerState.Identifier())
		}
	}

	for _, playerToRemove := range playersToRemove {
		delete(inMemoryCollection.playerStates, playerToRemove)
	}
}

// AvailableChatColors returns the chat colors which are allowed for players, as
// a full deep copy of the internal slice.
func (inMemoryCollection *InMemoryCollection) AvailableChatColors() []string {
	deepCopyOfChatColors := make([]string, inMemoryCollection.numberOfChatColors)
	copy(deepCopyOfChatColors, inMemoryCollection.availableChatColors)
	return deepCopyOfChatColors
}

// addWithGivenColor creates a new inMemoryState object with the given name and
// given color, and adds a reference to it into the collection.
func (inMemoryCollection *InMemoryCollection) addWithGivenColor(
	playerName string,
	chatColor string) (string, error) {
	playerIdentifier := inMemoryCollection.nameToIdentifier.Identifier(playerName)

	_, playerExists := inMemoryCollection.playerStates[playerIdentifier]

	if playerExists {
		return "", fmt.Errorf("Player %v already exists", playerName)
	}

	inMemoryCollection.mutualExclusion.Lock()

	inMemoryCollection.playerStates[playerIdentifier] = &inMemoryState{
		mutualExclusion: sync.Mutex{},
		identifier:      playerIdentifier,
		name:            playerName,
		color:           chatColor,
	}

	inMemoryCollection.mutualExclusion.Unlock()

	return playerIdentifier, nil
}

// addWithDefaultColor chooses a default chat color for the given new player name
// based on the given number of existing players and the given number of available
// chat colors, and then adds the new player as Add would.
func (inMemoryCollection *InMemoryCollection) addWithDefaultColor(
	playerName string) (string, error) {
	return inMemoryCollection.addWithDefaultColorWithoutCounting(
		playerName,
		len(inMemoryCollection.playerStates),
		inMemoryCollection.numberOfChatColors)
}

// addWithDefaultColor chooses a default chat color for the given new player name
// based on the given number of existing players and the given number of available
// chat colors, and then adds the new player as Add would.
func (inMemoryCollection *InMemoryCollection) addWithDefaultColorWithoutCounting(
	playerName string,
	playerCount int,
	numberOfColors int) (string, error) {
	chatColor := inMemoryCollection.availableChatColors[playerCount%numberOfColors]
	return inMemoryCollection.addWithGivenColor(playerName, chatColor)
}

// inMemoryState encapsulates all the state that the backend needs to know about a player,
// using a mutex to ensure that updates are thread-safe.
type inMemoryState struct {
	mutualExclusion sync.Mutex
	identifier      string
	name            string
	color           string
}

// Identifier returns the private identifier field.
func (playerState *inMemoryState) Identifier() string {
	return playerState.identifier
}

// Name returns the private name field.
func (playerState *inMemoryState) Name() string {
	return playerState.name
}

// Color returns the private color field.
func (playerState *inMemoryState) Color() string {
	return playerState.color
}
