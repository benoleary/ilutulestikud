package player

import (
	"fmt"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// StateHandler wraps around a player.StatePersister to encapsulate logic acting on the
// functions of the interface. It also has the responsibility of maintaining the list of
// chat colors which are available to players, and providing default colors if player
// definitions do not contain specific colors.
type StateHandler struct {
	statePersister      StatePersister
	availableChatColors []string
}

// NewStateHandler creates a new StateHandler around the given StatePersister and list
// of chat colors, giving default colors to the initial players.
func NewStateHandler(
	statePersister StatePersister,
	initialPlayerNames []string,
	availableColors []string) *StateHandler {
	numberOfColors := len(availableColors)
	deepCopyOfChatColors := make([]string, numberOfColors)
	copy(deepCopyOfChatColors, availableColors)

	newHandler := &StateHandler{
		statePersister:      statePersister,
		availableChatColors: deepCopyOfChatColors,
	}

	for _, initialPlayerName := range initialPlayerNames {
		newHandler.Add(endpoint.PlayerState{
			Name: initialPlayerName,
		})
	}

	return newHandler
}

// Add ensures that the player definition has a chat color before wrapping around
// the add function of the internal collection.
func (stateHandler StateHandler) Add(
	playerInformation endpoint.PlayerState) (string, error) {
	if playerInformation.Name == "" {
		return "", fmt.Errorf("Player must have a name")
	}

	if playerInformation.Color == "" {
		playerCount := len(stateHandler.statePersister.all())
		numberOfColors := len(stateHandler.availableChatColors)
		playerInformation.Color = stateHandler.availableChatColors[playerCount%numberOfColors]
	}

	return stateHandler.statePersister.add(playerInformation)
}

// UpdateFromPresentAttributes just wraps around the updateFromPresentAttributes
// function of the internal collection.
func (stateHandler StateHandler) UpdateFromPresentAttributes(
	updaterReference endpoint.PlayerState) error {
	return stateHandler.statePersister.updateFromPresentAttributes(updaterReference)
}

// Get just wraps around the get function of the internal collection.
func (stateHandler StateHandler) Get(playerIdentifier string) (ReadonlyState, error) {
	return stateHandler.statePersister.get(playerIdentifier)
}

// All just wraps around the all function of the internal collection.
func (stateHandler StateHandler) All() []ReadonlyState {
	return stateHandler.statePersister.all()
}

// Reset just wraps around the reset of the internal collection.
func (stateHandler StateHandler) Reset() {
	stateHandler.statePersister.reset()
}

// RegisteredPlayersForEndpoint writes relevant parts of the handler's collection's players
// into the JSON object for the frontend as a list of player objects as its "Players"
// attribute. he order of the players may not be consistent with repeated calls, as the
// order of All is not guaranteed to be consistent.
func (stateHandler StateHandler) RegisteredPlayersForEndpoint() endpoint.PlayerList {
	playerStates := stateHandler.statePersister.all()
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

// AvailableChatColorsForEndpoint writes the chat colors available to the handler's
// collection into the JSON object for the frontend.
func (stateHandler StateHandler) AvailableChatColorsForEndpoint() endpoint.ChatColorList {
	return endpoint.ChatColorList{
		Colors: stateHandler.availableChatColors,
	}
}
