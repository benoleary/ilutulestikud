package player

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/benoleary/ilutulestikud/backendjson"
	"github.com/benoleary/ilutulestikud/chat"
)

// GetAndPostHandler is a struct meant to encapsulate all the state co-ordinating all the players.
// It implements github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
type GetAndPostHandler struct {
	registeredPlayers map[string]*OriginalState
	mutualExclusion   sync.Mutex
}

// NewGetAndPostHandler constructs a GetAndPostHandler object with a non-nil, non-empty slice
// of State objects, returning a pointer to the newly-created object.
func NewGetAndPostHandler() *GetAndPostHandler {
	return &GetAndPostHandler{
		registeredPlayers: defaultPlayers(),
		mutualExclusion:   sync.Mutex{}}
}

// HandleGet parses an HTTP GET request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (handler *GetAndPostHandler) HandleGet(relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "registered-players":
		return handler.writeRegisteredPlayers()
	case "available-colors":
		return handler.writeAvailableColors()
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// HandlePost parses an HTTP POST request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (handler *GetAndPostHandler) HandlePost(httpBodyDecoder *json.Decoder, relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "new-player":
		return handler.handleNewPlayer(httpBodyDecoder)
	case "update-player":
		return handler.handleUpdatePlayer(httpBodyDecoder)
	case "reset-players":
		return handler.handleResetPlayers()
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// GetPlayerByName returns a pointer to the player state which has the given name, with
// a bool that is true if the player was found, as a map normally returns.
func (handler *GetAndPostHandler) GetPlayerByName(playerName string) (*OriginalState, bool) {
	foundPlayer, isFound := handler.registeredPlayers[playerName]
	return foundPlayer, isFound
}

// defaultPlayers returns a map of players created from default player names with colors
// according to the available chat colors, where the key is the player name.
func defaultPlayers() map[string]*OriginalState {
	initialNames := []string{"Mimi", "Aet", "Martin", "Markus", "Liisbet", "Madli", "Ben"}
	numberOfPlayers := len(initialNames)

	playerMap := make(map[string]*OriginalState, numberOfPlayers)

	for playerCount := 0; playerCount < numberOfPlayers; playerCount++ {
		playerName := initialNames[playerCount]

		playerMap[playerName] = NewState(playerName, chat.DefaultColor(playerCount))
	}

	return playerMap
}

// writeRegisteredPlayers writes a JSON object into the HTTP response which has
// the list of player objects as its "Players" attribute. The order of the players is not
// consistent with repeated calls even if the map of players does not change, since the
// Go compiler actually does randomize the iteration order of the map entries by design.
func (handler *GetAndPostHandler) writeRegisteredPlayers() (interface{}, int) {
	playerList := make([]backendjson.PlayerState, 0, len(handler.registeredPlayers))
	for _, registeredPlayer := range handler.registeredPlayers {
		playerList = append(playerList, ForBackend(registeredPlayer))
	}

	return backendjson.PlayerStateList{Players: playerList}, http.StatusOK
}

// writeAvailableColors writes a JSON object into the HTTP response which has
// the list of strings as its "Colors" attribute.
func (handler *GetAndPostHandler) writeAvailableColors() (interface{}, int) {
	return backendjson.ChatColorList{Colors: chat.AvailableColors()}, http.StatusOK
}

// handleNewPlayer adds the player defined by the JSON of the request's body to the list
// of registered players, and returns the updated list as writeRegisteredPlayerNameListJson
// would.
func (handler *GetAndPostHandler) handleNewPlayer(httpBodyDecoder *json.Decoder) (interface{}, int) {
	var newPlayer backendjson.PlayerState
	parsingError := httpBodyDecoder.Decode(&newPlayer)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	if newPlayer.Name == "" {
		return "No name for new player parsed from JSON", http.StatusBadRequest
	}

	_, playerExists := handler.registeredPlayers[newPlayer.Name]
	if playerExists {
		return "Name " + newPlayer.Name + " already registered", http.StatusBadRequest
	}

	playerColor := newPlayer.Color
	if playerColor == "" {
		// The new player is assigned the next color in the list, cycling through all the colors
		// again if there are more players than colors. E.g. if there are already 5 players, the
		// 6th player gets the 6th color in the list. If there are only 4 colors in the list,
		// the new player would get the 2nd color. This does not account for players having
		// changed color, but it doesn't matter, as it is just a fun way of choosing an initial
		// color.
		playerColor = chat.DefaultColor(len(handler.registeredPlayers))
	}

	handler.mutualExclusion.Lock()
	handler.registeredPlayers[newPlayer.Name] =
		NewState(newPlayer.Name, playerColor)
	handler.mutualExclusion.Unlock()

	return handler.writeRegisteredPlayers()
}

// handleUpdatePlayer updates the player defined by the JSON of the request's body, taking the "Name"
// attribute as the key, and returns the updated list as writeRegisteredPlayers would. Attributes
// which are present are updated, those which are missing remain unchanged.
func (handler *GetAndPostHandler) handleUpdatePlayer(httpBodyDecoder *json.Decoder) (interface{}, int) {
	var newPlayer backendjson.PlayerState
	parsingError := httpBodyDecoder.Decode(&newPlayer)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	existingPlayer, playerExists := handler.registeredPlayers[newPlayer.Name]
	if !playerExists {
		return "Name " + newPlayer.Name + " not found", http.StatusBadRequest
	}

	handler.mutualExclusion.Lock()
	existingPlayer.UpdateNonEmptyStrings(newPlayer)
	handler.mutualExclusion.Unlock()

	return handler.writeRegisteredPlayers()
}

// handleResetPlayers resets the player list to the initial list, and returns the updated list
// as writeRegisteredPlayers would.
func (handler *GetAndPostHandler) handleResetPlayers() (interface{}, int) {
	handler.mutualExclusion.Lock()
	handler.registeredPlayers = defaultPlayers()
	handler.mutualExclusion.Unlock()

	return handler.writeRegisteredPlayers()
}
