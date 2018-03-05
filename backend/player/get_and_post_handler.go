package player

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// GetAndPostHandler is a struct meant to encapsulate all the state co-ordinating all the players.
// It implements github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
type GetAndPostHandler struct {
	playerCollection Collection
}

// NewGetAndPostHandler constructs a GetAndPostHandler object with a non-nil, non-empty slice
// of State objects, returning a pointer to the newly-created object.
func NewGetAndPostHandler(stateCollection Collection) *GetAndPostHandler {
	return &GetAndPostHandler{playerCollection: stateCollection}
}

// HandleGet parses an HTTP GET request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (getAndPostHandler *GetAndPostHandler) HandleGet(relevantSegments []string) (interface{}, int) {
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
func (getAndPostHandler *GetAndPostHandler) HandlePost(httpBodyDecoder *json.Decoder, relevantSegments []string) (interface{}, int) {
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
	return ForEndpoint(getAndPostHandler.playerCollection), http.StatusOK
}

// writeAvailableColors writes a JSON object into the HTTP response which has
// the list of strings as its "Colors" attribute.
func (getAndPostHandler *GetAndPostHandler) writeAvailableColors() (interface{}, int) {
	return AvailableChatColorsForEndpoint(getAndPostHandler.playerCollection), http.StatusOK
}

// handleNewPlayer adds the player defined by the JSON of the request's body to the list
// of registered players, and returns the updated list as writeRegisteredPlayerNameListJson
// would.
func (getAndPostHandler *GetAndPostHandler) handleNewPlayer(httpBodyDecoder *json.Decoder) (interface{}, int) {
	var endpointPlayer endpoint.PlayerState
	parsingError := httpBodyDecoder.Decode(&endpointPlayer)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	if endpointPlayer.Name == "" {
		return "No name for new player parsed from JSON", http.StatusBadRequest
	}

	if strings.Contains(endpointPlayer.Name, "/") {
		return "Player names cannot contain the slash character '/', sorry", http.StatusBadRequest
	}

	addError := getAndPostHandler.playerCollection.Add(endpointPlayer)

	if addError != nil {
		return addError, http.StatusBadRequest
	}

	return getAndPostHandler.writeRegisteredPlayers()
}

// handleUpdatePlayer updates the player defined by the JSON of the request's body, taking the "Name"
// attribute as the key, and returns the updated list as writeRegisteredPlayers would. Attributes
// which are present are updated, those which are missing remain unchanged.
func (getAndPostHandler *GetAndPostHandler) handleUpdatePlayer(httpBodyDecoder *json.Decoder) (interface{}, int) {
	var newPlayer endpoint.PlayerState
	parsingError := httpBodyDecoder.Decode(&newPlayer)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	existingPlayer, playerExists := getAndPostHandler.playerCollection.Get(newPlayer.Identifier)

	// This is for backwards-compatibility with a frontend which still uses the player
	// name as the identifier.
	if !playerExists {
		playerIdentifier := base64.StdEncoding.EncodeToString([]byte(newPlayer.Name))
		playerFromEncodedName, encodedNameFound := getAndPostHandler.playerCollection.Get(playerIdentifier)
		playerExists = encodedNameFound
		existingPlayer = playerFromEncodedName
	}

	if !playerExists {
		return "Name " + newPlayer.Name + " not found", http.StatusBadRequest
	}

	existingPlayer.UpdateFromPresentAttributes(newPlayer)

	return getAndPostHandler.writeRegisteredPlayers()
}

// handleResetPlayers resets the player list to the initial list, and returns the updated list
// as writeRegisteredPlayers would.
func (getAndPostHandler *GetAndPostHandler) handleResetPlayers() (interface{}, int) {
	getAndPostHandler.playerCollection.Reset()

	return getAndPostHandler.writeRegisteredPlayers()
}
