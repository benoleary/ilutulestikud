package player

import (
	"encoding/json"
	"net/http"
	"sync"
)

// Handler is a struct meant to encapsulate all the state co-ordinating all the players.
// It implements github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
type Handler struct {
	registeredPlayers map[string]*State
	mutualExclusion   sync.Mutex
}

// NewHandler constructs a Handler object with a non-nil, non-empty slice of State objects,
// returning a pointer to the newly-created object.
func NewHandler() *Handler {
	return &Handler{defaultPlayers(), sync.Mutex{}}
}

// HandleGet parses an HTTP GET request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (handler *Handler) HandleGet(
	httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantSegments []string) {
	if len(relevantSegments) < 1 {
		httpResponseWriter.WriteHeader(http.StatusBadRequest)
		httpResponseWriter.Write([]byte("Not enough segments in URI to determine what to do"))
		return
	}

	switch relevantSegments[0] {
	case "registered-players":
		handler.writeRegisteredPlayers(httpResponseWriter)
	case "available-colors":
		handler.writeAvailableColors(httpResponseWriter)
	default:
		http.NotFound(httpResponseWriter, httpRequest)
	}
}

// HandlePost parses an HTTP POST request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (handler *Handler) HandlePost(
	httpResponseWriter http.ResponseWriter,
	httpRequest *http.Request,
	relevantSegments []string) {
	if len(relevantSegments) < 1 {
		httpResponseWriter.WriteHeader(http.StatusBadRequest)
		httpResponseWriter.Write([]byte("Not enough segments in URI to determine what to do"))
		return
	}

	switch relevantSegments[0] {
	case "new-player":
		handler.handleNewPlayer(httpResponseWriter, httpRequest)
	case "update-player":
		handler.handleUpdatePlayer(httpResponseWriter, httpRequest)
	case "reset-players":
		handler.handleResetPlayers(httpResponseWriter, httpRequest)
	default:
		http.NotFound(httpResponseWriter, httpRequest)
	}
}

// GetPlayerByName returns a pointer to the player state which has the given name.
func (handler *Handler) GetPlayerByName(playerName string) *State {
	return handler.registeredPlayers[playerName]
}

// defaultPlayers returns a map of players created from default player names with colors
// according to the available chat colors, where the key is the player name.
func defaultPlayers() map[string]*State {
	initialColors := availableColors()
	numberOfColors := len(initialColors)
	initialNames := []string{"Mimi", "Aet", "Martin", "Markus", "Liisbet", "Madli", "Ben"}
	numberOfPlayers := len(initialNames)

	playerMap := make(map[string]*State, numberOfPlayers)

	for playerCount := 0; playerCount < numberOfPlayers; playerCount++ {
		playerName := initialNames[playerCount]

		// We cycle through all the colors again if there are more players than colors.
		playerColor := initialColors[playerCount%numberOfColors]

		playerMap[playerName] = NewState(playerName, playerColor)
	}

	return playerMap
}

// availableColors returns a list of colors which can be selected as chat colors for players.
func availableColors() []string {
	return []string{
		"pink",
		"red",
		"orange",
		"yellow",
		"green",
		"blue",
		"purple",
		"white"}
}

// writeRegisteredPlayers writes a JSON object into the HTTP response which has
// the list of player objects as its "Players" attribute. The order of the players is not
// consistent with repeated calls even if the map of players does not change, since the
// Go compiler actually does randomize the iteration order of the map entries by design.
func (handler *Handler) writeRegisteredPlayers(httpResponseWriter http.ResponseWriter) {
	playerList := make([]State, 0, len(handler.registeredPlayers))
	for _, registeredPlayer := range handler.registeredPlayers {
		playerList = append(playerList, *registeredPlayer)
	}

	json.NewEncoder(httpResponseWriter).Encode(struct{ Players []State }{playerList})
}

// writeAvailableColors writes a JSON object into the HTTP response which has
// the list of strings as its "Colors" attribute.
func (handler *Handler) writeAvailableColors(httpResponseWriter http.ResponseWriter) {
	json.NewEncoder(httpResponseWriter).Encode(struct{ Colors []string }{availableColors()})
}

// handleNewPlayer adds the player defined by the JSON of the request's body to the list
// of registered players, and returns the updated list as writeRegisteredPlayerNameListJson
// would.
func (handler *Handler) handleNewPlayer(
	httpResponseWriter http.ResponseWriter,
	httpRequest *http.Request) {
	if httpRequest.Body == nil {
		http.Error(httpResponseWriter, "Empty request body", http.StatusBadRequest)
		return
	}

	var newPlayer State
	parsingError := json.NewDecoder(httpRequest.Body).Decode(&newPlayer)
	if parsingError != nil {
		http.Error(httpResponseWriter, "Error parsing JSON: "+parsingError.Error(), http.StatusBadRequest)
		return
	}

	_, playerExists := handler.registeredPlayers[newPlayer.Name]
	if playerExists {
		http.Error(httpResponseWriter, "Name "+newPlayer.Name+" already registered", http.StatusBadRequest)
		return
	}

	playerColor := newPlayer.Color
	if playerColor == "" {
		// The new player is assigned the next color in the list, cycling through all the colors
		// again if there are more players than colors. E.g. if there are already 5 players, the
		// 6th player gets the 6th color in the list. If there are only 4 colors in the list,
		// the new player would get the 2nd color. This does not account for players having
		// changed color, but it doesn't matter, as it is just a fun way of choosing an initial
		// color.
		colorList := availableColors()
		playerColor = colorList[len(handler.registeredPlayers)%len(colorList)]
	}

	handler.mutualExclusion.Lock()
	handler.registeredPlayers[newPlayer.Name] =
		NewState(newPlayer.Name, playerColor)
	handler.mutualExclusion.Unlock()

	handler.writeRegisteredPlayers(httpResponseWriter)
}

// handleUpdatePlayer updates the player defined by the JSON of the request's body, taking the "Name"
// attribute as the key, and returns the updated list as writeRegisteredPlayers would. Attributes
// which are present are updated, those which are missing remain unchanged.
func (handler *Handler) handleUpdatePlayer(
	httpResponseWriter http.ResponseWriter,
	httpRequest *http.Request) {
	if httpRequest.Body == nil {
		http.Error(httpResponseWriter, "Empty request body", http.StatusBadRequest)
		return
	}

	var newPlayer State
	parsingError := json.NewDecoder(httpRequest.Body).Decode(&newPlayer)
	if parsingError != nil {
		http.Error(httpResponseWriter, "Error parsing JSON: "+parsingError.Error(), http.StatusBadRequest)
		return
	}

	existingPlayer, playerExists := handler.registeredPlayers[newPlayer.Name]
	if !playerExists {
		http.Error(httpResponseWriter, "Name "+newPlayer.Name+" not found", http.StatusBadRequest)
		return
	}

	handler.mutualExclusion.Lock()
	existingPlayer.UpdateNonEmptyStrings(&newPlayer)
	handler.mutualExclusion.Unlock()

	handler.writeRegisteredPlayers(httpResponseWriter)
}

// handleResetPlayers resets the player list to the initial list, and returns the updated list
// as writeRegisteredPlayers would.
func (handler *Handler) handleResetPlayers(
	httpResponseWriter http.ResponseWriter,
	httpRequest *http.Request) {
	handler.mutualExclusion.Lock()
	handler.registeredPlayers = defaultPlayers()
	handler.mutualExclusion.Unlock()

	handler.writeRegisteredPlayers(httpResponseWriter)
}
