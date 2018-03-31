package player

import (
	"fmt"
)

// StateCollection wraps around a player.StatePersister to encapsulate logic acting on
// the functions of the interface. It also has the responsibility of maintaining the
// list of chat colors which are available to players, and providing default colors if
// player definitions do not contain specific colors.
type StateCollection struct {
	statePersister      StatePersister
	initialPlayerNames  []string
	availableChatColors []string
}

// NewCollection creates a new StateCollection around the given StatePersister and list
// of chat colors, giving default colors to the initial players.
func NewCollection(
	statePersister StatePersister,
	initialPlayerNames []string,
	availableColors []string) *StateCollection {
	newCollection := &StateCollection{
		statePersister:      statePersister,
		initialPlayerNames:  initialPlayerNames,
		availableChatColors: deepCopyStringSlice(availableColors),
	}

	newCollection.addInitialPlayers()

	return newCollection
}

// AvailableChatColors returns a deep copy of state collection's chat color slice.
func (stateCollection *StateCollection) AvailableChatColors() []string {
	return deepCopyStringSlice(stateCollection.availableChatColors)
}

// Add ensures that the player definition has a chat color before wrapping around
// the add function of the internal collection.
func (stateCollection *StateCollection) Add(
	playerName string,
	chatColor string) error {
	if playerName == "" {
		return fmt.Errorf("Player must have a name")
	}

	if chatColor == "" {
		playerCount := len(stateCollection.statePersister.all())
		numberOfColors := len(stateCollection.availableChatColors)
		chatColor = stateCollection.availableChatColors[playerCount%numberOfColors]
	}

	return stateCollection.statePersister.add(playerName, chatColor)
}

// UpdateColor just wraps around the updateColor function of the internal collection.
func (stateCollection *StateCollection) UpdateColor(
	playerName string,
	chatColor string) error {
	return stateCollection.statePersister.updateColor(playerName, chatColor)
}

// All just wraps around the all function of the internal collection.
func (stateCollection *StateCollection) All() []ReadonlyState {
	return stateCollection.statePersister.all()
}

// Get just wraps around the get function of the internal collection.
func (stateCollection *StateCollection) Get(playerIdentifier string) (ReadonlyState, error) {
	return stateCollection.statePersister.get(playerIdentifier)
}

// Reset calls the reset of the internal collection then adds the initial players again.
func (stateCollection *StateCollection) Reset() {
	stateCollection.statePersister.reset()
	stateCollection.addInitialPlayers()
}

func deepCopyStringSlice(stringsToCopy []string) []string {
	deepCopy := make([]string, len(stringsToCopy))
	copy(deepCopy, stringsToCopy)

	return deepCopy
}

func (stateCollection *StateCollection) addInitialPlayers() {
	for _, initialPlayerName := range stateCollection.initialPlayerNames {
		stateCollection.Add(initialPlayerName, "")
	}
}
