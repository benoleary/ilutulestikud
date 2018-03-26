package player

import (
	"fmt"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// StateCollection wraps around a player.StatePersister to encapsulate logic acting on
// the functions of the interface. It also has the responsibility of maintaining the
// list of chat colors which are available to players, and providing default colors if
// player definitions do not contain specific colors.
type StateCollection struct {
	statePersister      StatePersister
	availableChatColors []string
}

// NewCollection creates a new StateCollection around the given StatePersister and list
// of chat colors, giving default colors to the initial players.
func NewCollection(
	statePersister StatePersister,
	initialPlayerNames []string,
	availableColors []string) *StateCollection {
	numberOfColors := len(availableColors)
	deepCopyOfChatColors := make([]string, numberOfColors)
	copy(deepCopyOfChatColors, availableColors)

	newCollection := &StateCollection{
		statePersister:      statePersister,
		availableChatColors: deepCopyOfChatColors,
	}

	for _, initialPlayerName := range initialPlayerNames {
		newCollection.Add(endpoint.PlayerState{
			Name: initialPlayerName,
		})
	}

	return newCollection
}

// Add ensures that the player definition has a chat color before wrapping around
// the add function of the internal collection.
func (stateCollection StateCollection) Add(
	playerInformation endpoint.PlayerState) (string, error) {
	if playerInformation.Name == "" {
		return "", fmt.Errorf("Player must have a name")
	}

	if playerInformation.Color == "" {
		playerCount := len(stateCollection.statePersister.all())
		numberOfColors := len(stateCollection.availableChatColors)
		playerInformation.Color = stateCollection.availableChatColors[playerCount%numberOfColors]
	}

	return stateCollection.statePersister.add(playerInformation)
}

// UpdateFromPresentAttributes just wraps around the updateFromPresentAttributes
// function of the internal collection.
func (stateCollection StateCollection) UpdateFromPresentAttributes(
	updaterReference endpoint.PlayerState) error {
	return stateCollection.statePersister.updateFromPresentAttributes(updaterReference)
}

// Get just wraps around the get function of the internal collection.
func (stateCollection StateCollection) Get(playerIdentifier string) (ReadonlyState, error) {
	return stateCollection.statePersister.get(playerIdentifier)
}

// All just wraps around the all function of the internal collection.
func (stateCollection StateCollection) All() []ReadonlyState {
	return stateCollection.statePersister.all()
}

// Reset just wraps around the reset of the internal collection.
func (stateCollection StateCollection) Reset() {
	stateCollection.statePersister.reset()
}

// RegisteredPlayersForEndpoint writes relevant parts of the collection's players
// into the JSON object for the frontend as a list of player objects as its
// "Players" attribute. The order of the players may not be consistent with repeated
// calls, as the order of All is not guaranteed to be consistent.
func (stateCollection StateCollection) RegisteredPlayersForEndpoint() endpoint.PlayerList {
	playerStates := stateCollection.statePersister.all()
	playerList := make([]endpoint.PlayerState, 0, len(playerStates))
	for _, playerState := range playerStates {
		playerList = append(playerList, endpoint.PlayerState{
			Identifier: playerState.Identifier(),
			Name:       playerState.Name(),
			Color:      playerState.Color(),
		})
	}

	return endpoint.PlayerList{
		Players: playerList,
	}
}

// AvailableChatColorsForEndpoint writes the chat colors available to the collection
// into the JSON object for the frontend.
func (stateCollection StateCollection) AvailableChatColorsForEndpoint() endpoint.ChatColorList {
	return endpoint.ChatColorList{
		Colors: stateCollection.availableChatColors,
	}
}
