package player

import (
	"fmt"
)

// StateCollection wraps around a player.StatePersister to encapsulate logic acting on
// the functions of the interface. It also has the responsibility of maintaining the
// list of chat colors which are available to players, and providing default colors if
// player definitions do not contain specific colors.
type StateCollection struct {
	statePersister     StatePersister
	initialPlayerNames []string
	chatColorSlice     []string
	chatColorMap       map[string]bool
	numberOfColors     int
}

// NewCollection creates a new StateCollection around the given StatePersister and list
// of chat colors, giving default colors to the initial players. It returns nil and an
// error if given no chat colors.
func NewCollection(
	statePersister StatePersister,
	initialPlayerNames []string,
	availableColors []string) (*StateCollection, error) {
	if len(availableColors) <= 0 {
		return nil, fmt.Errorf("Chat color list must have at least one color")
	}

	// We keep a map of colors to validity to both remove duplicate colors and
	// to make it easy to check if a color is valid when updating players.
	colorMap := make(map[string]bool, 0)
	for _, chatColor := range availableColors {
		colorMap[chatColor] = true
	}

	uniqueColors := make([]string, 0)

	for uniqueColor, isColorAvailable := range colorMap {
		if isColorAvailable {
			uniqueColors = append(uniqueColors, uniqueColor)
		}
	}

	newCollection :=
		&StateCollection{
			statePersister:     statePersister,
			initialPlayerNames: initialPlayerNames,
			chatColorSlice:     uniqueColors,
			chatColorMap:       colorMap,
			numberOfColors:     len(uniqueColors),
		}

	newCollection.addInitialPlayers()

	return newCollection, nil
}

// All just wraps around the All function of the internal persistence store.
func (stateCollection *StateCollection) All() ([]ReadonlyState, error) {
	return stateCollection.statePersister.All()
}

// Get just wraps around the Get function of the internal persistence store.
func (stateCollection *StateCollection) Get(playerName string) (ReadonlyState, error) {
	return stateCollection.statePersister.Get(playerName)
}

// AvailableChatColors returns a deep copy of state persistence store's chat
// color slice.
func (stateCollection *StateCollection) AvailableChatColors() []string {
	numberOfColors := len(stateCollection.chatColorSlice)
	deepCopy := make([]string, numberOfColors)
	copy(deepCopy, stateCollection.chatColorSlice)

	return deepCopy
}

// Add ensures that the player definition has a chat color before calling
// the Add function of the internal persistence store.
func (stateCollection *StateCollection) Add(
	playerName string,
	chatColor string) error {
	if playerName == "" {
		return fmt.Errorf("Player must have a name")
	}

	if chatColor == "" {
		allPlayers, errorFromAll := stateCollection.statePersister.All()
		if errorFromAll != nil {
			return errorFromAll
		}

		playerCount := len(allPlayers)
		colorIndex := playerCount % stateCollection.numberOfColors
		chatColor = stateCollection.chatColorSlice[colorIndex]
	} else if !stateCollection.chatColorMap[chatColor] {
		return fmt.Errorf(
			"Chat color %v is not in list of valid colors %v",
			chatColor,
			stateCollection.chatColorSlice)
	}

	return stateCollection.statePersister.Add(playerName, chatColor)
}

// UpdateColor checks the validity of the color then calls the UpdateColor
// function of the internal persistence store.
func (stateCollection *StateCollection) UpdateColor(
	playerName string,
	chatColor string) error {
	if !stateCollection.chatColorMap[chatColor] {
		return fmt.Errorf(
			"Chat color %v is not in list of valid colors %v",
			chatColor,
			stateCollection.chatColorSlice)
	}

	return stateCollection.statePersister.UpdateColor(playerName, chatColor)
}

// Delete calls the Delete of the internal persistence store.
func (stateCollection *StateCollection) Delete(playerName string) error {
	return stateCollection.statePersister.Delete(playerName)
}

// Reset calls the Reset of the internal persistence store then adds the
// initial players again.
func (stateCollection *StateCollection) Reset() error {
	errorFromReset := stateCollection.statePersister.Reset()
	if errorFromReset != nil {
		return errorFromReset
	}

	return stateCollection.addInitialPlayers()
}

func (stateCollection *StateCollection) addInitialPlayers() error {
	for _, initialPlayerName := range stateCollection.initialPlayerNames {
		errorFromAdd := stateCollection.Add(initialPlayerName, "")

		if errorFromAdd != nil {
			return errorFromAdd
		}
	}

	return nil
}
