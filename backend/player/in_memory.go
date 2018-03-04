package player

import (
	"sync"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// InMemoryCollection stores inMemoryState objects as State objects.
type InMemoryCollection struct {
	mutualExclusion     sync.Mutex
	playerStates        map[string]State
	initialPlayerNames  map[string]bool
	availableChatColors []string
	numberOfChatColors  int
}

// NewInMemoryCollection creates a Collection around a map of players created
// from the given initial player names, with colors according to the available
// chat colors.
func NewInMemoryCollection(
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
		playerStates:        make(map[string]State, numberOfPlayers),
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
func (inMemoryCollection *InMemoryCollection) Add(endpointPlayer endpoint.PlayerState) {
	if endpointPlayer.Color == "" {
		inMemoryCollection.addWithDefaultColor(endpointPlayer.Name)
	} else {
		inMemoryCollection.addWithGivenColor(endpointPlayer.Name, endpointPlayer.Color)
	}
}

// Get returns the State corresponding to the given player name if it exists already
// (or else nil) along with whether the State exists, analogously to a standard Golang
// map.
func (inMemoryCollection *InMemoryCollection) Get(playerName string) (State, bool) {
	playerState, playerExists := inMemoryCollection.playerStates[playerName]
	return playerState, playerExists
}

// All returns a slice of all the State instances in the collection, ordered in the
// random way the iteration over the entries of a Golang map normally is.
func (inMemoryCollection *InMemoryCollection) All() []State {
	playerList := make([]State, 0, len(inMemoryCollection.playerStates))
	for _, playerState := range inMemoryCollection.playerStates {
		playerList = append(playerList, playerState)
	}

	return playerList
}

// Reset removes all players which are not among the initial players, and
// restores any initial players who have been removed.
func (inMemoryCollection *InMemoryCollection) Reset() {
	playersToRemove := make([]string, 0)
	for _, playerState := range inMemoryCollection.playerStates {
		if !inMemoryCollection.initialPlayerNames[playerState.Name()] {
			playersToRemove = append(playersToRemove, playerState.Name())
		}
	}

	for _, playerToRemove := range playersToRemove {
		delete(inMemoryCollection.playerStates, playerToRemove)
	}

	// We assume that we never change the initial player "set", so we can
	// ignore the value bool and just use the keys.
	for initialPlayerName := range inMemoryCollection.initialPlayerNames {
		_, initialPlayerExists := inMemoryCollection.playerStates[initialPlayerName]
		if !initialPlayerExists {
			inMemoryCollection.addWithDefaultColor(initialPlayerName)
		}
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
	chatColor string) {
	inMemoryCollection.mutualExclusion.Lock()

	inMemoryCollection.playerStates[playerName] =
		&inMemoryState{
			mutualExclusion: sync.Mutex{},
			name:            playerName,
			color:           chatColor,
		}

	inMemoryCollection.mutualExclusion.Unlock()
}

// addWithDefaultColor chooses a default chat color for the given new player name
// based on the given number of existing players and the given number of available
// chat colors, and then adds the new player as Add would.
func (inMemoryCollection *InMemoryCollection) addWithDefaultColor(
	playerName string) {
	inMemoryCollection.addWithDefaultColorWithoutCounting(
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
	numberOfColors int) {
	chatColor := inMemoryCollection.availableChatColors[playerCount%numberOfColors]
	inMemoryCollection.addWithGivenColor(playerName, chatColor)
}

// inMemoryState encapsulates all the state that the backend needs to know about a player,
// using a mutex to ensure that updates are thread-safe.
type inMemoryState struct {
	mutualExclusion sync.Mutex
	name            string
	color           string
}

// Name returns the private name field.
func (playerState *inMemoryState) Name() string {
	return playerState.name
}

// Color returns the private color field.
func (playerState *inMemoryState) Color() string {
	return playerState.color
}

// UpdateFromPresentAttributes over-writes all non-name string attributes of this
// state with those from updaterReference unless the string in updaterReference
// is empty. It uses a mutex to ensure thread safety, and since the InMemoryCollection
// does not persist inMemoryState instances outside of its map of interfaces,
// there is no issue with persistence.
func (playerState *inMemoryState) UpdateFromPresentAttributes(updaterReference endpoint.PlayerState) {
	// It would be more efficient to only lock if we go into an if statement,
	// but then multiple if statements would be less efficient, and there would
	// be a mutex in each if statement.
	playerState.mutualExclusion.Lock()

	if updaterReference.Color != "" {
		playerState.color = updaterReference.Color
	}

	playerState.mutualExclusion.Unlock()
}
