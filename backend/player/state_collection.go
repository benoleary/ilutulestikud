package player

import (
	"context"
	"fmt"
)

// StateCollection wraps around a player.StatePersister to encapsulate logic acting on
// the functions of the interface. It also has the responsibility of maintaining the
// list of chat colors which are available to players, and providing default colors if
// player definitions do not contain specific colors.
type StateCollection struct {
	statePersister StatePersister
	chatColorSlice []string
	chatColorMap   map[string]bool
	numberOfColors int
}

// NewCollection creates a new StateCollection around the given StatePersister and list
// of chat colors, giving default colors to the initial players. It returns nil and an
// error if given no chat colors.
func NewCollection(
	statePersister StatePersister,
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
			statePersister: statePersister,
			chatColorSlice: uniqueColors,
			chatColorMap:   colorMap,
			numberOfColors: len(uniqueColors),
		}

	return newCollection, nil
}

// All just wraps around the All function of the internal persistence store.
func (stateCollection *StateCollection) All(
	executionContext context.Context) ([]ReadonlyState, error) {
	return stateCollection.statePersister.All(executionContext)
}

// Get just wraps around the Get function of the internal persistence store.
func (stateCollection *StateCollection) Get(
	executionContext context.Context,
	playerName string) (ReadonlyState, error) {
	return stateCollection.statePersister.Get(executionContext, playerName)
}

// AvailableChatColors returns a deep copy of state persistence store's chat
// color slice, and ignores the context.
func (stateCollection *StateCollection) AvailableChatColors(
	executionContext context.Context) []string {
	numberOfColors := len(stateCollection.chatColorSlice)
	deepCopy := make([]string, numberOfColors)
	copy(deepCopy, stateCollection.chatColorSlice)

	return deepCopy
}

// Add ensures that the player definition has a chat color before calling
// the Add function of the internal persistence store.
func (stateCollection *StateCollection) Add(
	executionContext context.Context,
	playerName string,
	chatColor string) error {
	if playerName == "" {
		return fmt.Errorf("Player must have a name")
	}

	if chatColor == "" {
		allPlayers, errorFromAll :=
			stateCollection.statePersister.All(executionContext)
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

	return stateCollection.statePersister.Add(
		executionContext,
		playerName,
		chatColor)
}

// UpdateColor checks the validity of the color then calls the UpdateColor
// function of the internal persistence store.
func (stateCollection *StateCollection) UpdateColor(
	executionContext context.Context,
	playerName string,
	chatColor string) error {
	if !stateCollection.chatColorMap[chatColor] {
		return fmt.Errorf(
			"Chat color %v is not in list of valid colors %v",
			chatColor,
			stateCollection.chatColorSlice)
	}

	return stateCollection.statePersister.UpdateColor(
		executionContext,
		playerName,
		chatColor)
}

// Delete calls the Delete of the internal persistence store.
func (stateCollection *StateCollection) Delete(
	executionContext context.Context,
	playerName string) error {
	return stateCollection.statePersister.Delete(
		executionContext,
		playerName)
}
