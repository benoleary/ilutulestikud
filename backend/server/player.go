package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// playerEndpointHandler is a struct meant to encapsulate all the state making the
// player states available to the endpoints.
// It implements github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
type playerEndpointHandler struct {
	stateCollection   playerCollection
	segmentTranslator EndpointSegmentTranslator
}

// HandleGet parses an HTTP GET request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (playerHandler *playerEndpointHandler) HandleGet(
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "registered-players":
		return playerHandler.writeRegisteredPlayers()
	case "available-colors":
		return playerHandler.writeAvailableColors()
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// HandlePost parses an HTTP POST request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (playerHandler *playerEndpointHandler) HandlePost(
	httpBodyDecoder *json.Decoder,
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "new-player":
		return playerHandler.handleNewPlayer(httpBodyDecoder)
	case "update-player":
		return playerHandler.handleUpdatePlayer(httpBodyDecoder)
	case "reset-players":
		return playerHandler.handleResetPlayers()
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// writeRegisteredPlayers writes a JSON object into the HTTP response which has
// the list of player objects as its "Players" attribute. The order of the players
// may not consistent with repeated calls as ForEndpoint does not guarantee it.
func (playerHandler *playerEndpointHandler) writeRegisteredPlayers() (interface{}, int) {
	playerStates := playerHandler.stateCollection.All()
	playerList := make([]endpoint.PlayerState, 0, len(playerStates))
	for _, playerState := range playerStates {
		playerName := playerState.Name()
		playerList = append(playerList, endpoint.PlayerState{
			Identifier: playerHandler.segmentTranslator.ToSegment(playerName),
			Name:       playerName,
			Color:      playerState.Color(),
		})
	}

	endpointObject := endpoint.PlayerList{
		Players: playerList,
	}

	return endpointObject, http.StatusOK
}

// writeAvailableColors writes a JSON object into the HTTP response which has
// the list of strings as its "Colors" attribute.
func (playerHandler *playerEndpointHandler) writeAvailableColors() (interface{}, int) {
	endpointObject := endpoint.ChatColorList{
		Colors: playerHandler.stateCollection.AvailableChatColors(),
	}

	return endpointObject, http.StatusOK
}

// handleNewPlayer adds the player defined by the JSON of the request's body to the list
// of registered players, and returns the updated list as writeRegisteredPlayerNameListJson
// would.
func (playerHandler *playerEndpointHandler) handleNewPlayer(
	httpBodyDecoder *json.Decoder) (interface{}, int) {
	var endpointPlayer endpoint.PlayerState
	parsingError := httpBodyDecoder.Decode(&endpointPlayer)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	addError := playerHandler.stateCollection.Add(endpointPlayer.Name, endpointPlayer.Color)

	if addError != nil {
		return addError, http.StatusBadRequest
	}

	playerIdentifier := playerHandler.segmentTranslator.ToSegment(endpointPlayer.Name)

	if strings.Contains(playerIdentifier, "/") {
		errorMessage := fmt.Sprintf(
			"Server set up with encoding which cannot convert %v to identifier with '/' in it",
			endpointPlayer.Name)
		return errorMessage, http.StatusBadRequest
	}

	return playerHandler.writeRegisteredPlayers()
}

// handleUpdatePlayer updates the player defined by the JSON of the request's body, taking
// the "Name" attribute as the key, and returns the updated list as writeRegisteredPlayers
// would. Attributes which are present are updated, those which are missing remain unchanged.
func (playerHandler *playerEndpointHandler) handleUpdatePlayer(
	httpBodyDecoder *json.Decoder) (interface{}, int) {
	var playerUpdate endpoint.PlayerState
	parsingError := httpBodyDecoder.Decode(&playerUpdate)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	updateError :=
		playerHandler.stateCollection.UpdateColor(playerUpdate.Name, playerUpdate.Color)

	if updateError != nil {
		return updateError, http.StatusBadRequest
	}

	return playerHandler.writeRegisteredPlayers()
}

// handleResetPlayers resets the player list to the initial list, and returns the updated list
// as writeRegisteredPlayers would.
func (playerHandler *playerEndpointHandler) handleResetPlayers() (interface{}, int) {
	playerHandler.stateCollection.Reset()

	return playerHandler.writeRegisteredPlayers()
}
