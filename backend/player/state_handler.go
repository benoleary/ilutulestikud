package player

import (
	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// StateHandler wraps around a player.StateCollection to encapsulate logic acting on the
// functions of the interface.
type StateHandler struct {
	PlayerStates StateCollection
}

// RegisteredPlayersForEndpoint writes relevant parts of the handler's collection's players
// into the JSON object for the frontend as a list of player objects as its "Players"
// attribute. he order of the players may not be consistent with repeated calls, as the
// order of All is not guaranteed to be consistent.
func (stateHandler StateHandler) RegisteredPlayersForEndpoint() endpoint.PlayerList {
	playerStates := stateHandler.PlayerStates.All()
	playerList := make([]endpoint.PlayerState, 0, len(playerStates))
	for _, playerState := range playerStates {
		playerList = append(playerList, endpoint.PlayerState{
			Identifier: playerState.Identifier(),
			Name:       playerState.Name(),
			Color:      playerState.Color(),
		})
	}

	return endpoint.PlayerList{Players: playerList}
}

// AvailableChatColorsForEndpoint writes the chat colors available to the handler's
// collection into the JSON object for the frontend.
func (stateHandler StateHandler) AvailableChatColorsForEndpoint() endpoint.ChatColorList {
	return endpoint.ChatColorList{
		Colors: stateHandler.PlayerStates.AvailableChatColors(),
	}
}
