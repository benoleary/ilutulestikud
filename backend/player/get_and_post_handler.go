package player

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// GetAndPostHandler is a struct meant to encapsulate all the state co-ordinating all the
// players.
// It implements github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
type GetAndPostHandler struct {
	stateHandler *StateHandler
}

// NewGetAndPostHandler constructs a GetAndPostHandler object with a non-nil, non-empty slice
// of State objects, returning a pointer to the newly-created object.
func NewGetAndPostHandler(playerHandler *StateHandler) *GetAndPostHandler {
	return &GetAndPostHandler{
		stateHandler: playerHandler,
	}
}

// HandleGet parses an HTTP GET request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (getAndPostHandler *GetAndPostHandler) HandleGet(
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "registered-players":
		return getAndPostHandler.writeRegisteredPlayers()
	case "available-colors":
		return getAndPostHandler.writeAvailableColors()
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// HandlePost parses an HTTP POST request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (getAndPostHandler *GetAndPostHandler) HandlePost(
	httpBodyDecoder *json.Decoder,
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "new-player":
		return getAndPostHandler.handleNewPlayer(httpBodyDecoder)
	case "update-player":
		return getAndPostHandler.handleUpdatePlayer(httpBodyDecoder)
	case "reset-players":
		return getAndPostHandler.handleResetPlayers()
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// writeRegisteredPlayers writes a JSON object into the HTTP response which has
// the list of player objects as its "Players" attribute. The order of the players
// may not consistent with repeated calls as ForEndpoint does not guarantee it.
func (getAndPostHandler *GetAndPostHandler) writeRegisteredPlayers() (interface{}, int) {
	return getAndPostHandler.stateHandler.RegisteredPlayersForEndpoint(), http.StatusOK
}

// writeAvailableColors writes a JSON object into the HTTP response which has
// the list of strings as its "Colors" attribute.
func (getAndPostHandler *GetAndPostHandler) writeAvailableColors() (interface{}, int) {
	return getAndPostHandler.stateHandler.AvailableChatColorsForEndpoint(), http.StatusOK
}

// handleNewPlayer adds the player defined by the JSON of the request's body to the list
// of registered players, and returns the updated list as writeRegisteredPlayerNameListJson
// would.
func (getAndPostHandler *GetAndPostHandler) handleNewPlayer(
	httpBodyDecoder *json.Decoder) (interface{}, int) {
	var endpointPlayer endpoint.PlayerState
	parsingError := httpBodyDecoder.Decode(&endpointPlayer)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	if endpointPlayer.Name == "" {
		return "No name for new player parsed from JSON", http.StatusBadRequest
	}

	playerIdentifier, addError := getAndPostHandler.stateHandler.Add(endpointPlayer)

	if addError != nil {
		return addError, http.StatusBadRequest
	}

	if strings.Contains(playerIdentifier, "/") {
		errorMessage := fmt.Sprintf(
			"Server set up with encoding which cannot convert %v to identifier with '/' in it",
			endpointPlayer.Name)
		return errorMessage, http.StatusBadRequest
	}

	return getAndPostHandler.writeRegisteredPlayers()
}

// handleUpdatePlayer updates the player defined by the JSON of the request's body, taking
// the "Name" attribute as the key, and returns the updated list as writeRegisteredPlayers
// would. Attributes which are present are updated, those which are missing remain unchanged.
func (getAndPostHandler *GetAndPostHandler) handleUpdatePlayer(
	httpBodyDecoder *json.Decoder) (interface{}, int) {
	var playerUpdate endpoint.PlayerState
	parsingError := httpBodyDecoder.Decode(&playerUpdate)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	updateError :=
		getAndPostHandler.stateHandler.UpdateFromPresentAttributes(playerUpdate)

	if updateError != nil {
		return updateError, http.StatusBadRequest
	}

	return getAndPostHandler.writeRegisteredPlayers()
}

// handleResetPlayers resets the player list to the initial list, and returns the updated list
// as writeRegisteredPlayers would.
func (getAndPostHandler *GetAndPostHandler) handleResetPlayers() (interface{}, int) {
	getAndPostHandler.stateHandler.Reset()

	return getAndPostHandler.writeRegisteredPlayers()
}
