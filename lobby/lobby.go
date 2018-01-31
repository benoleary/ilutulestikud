package lobby

import (
	"encoding/json"
	"github.com/benoleary/ilutulestikud/player"
	"net/http"
	"sync"
)

// State is a struct meant to encapsulate all the state required for the lobby concept to function.
type State struct {
	registeredPlayers map[string]*player.State
	mutualExclusion   sync.Mutex
}

// MakeEmpty constructs a State object with a non-nil (but empty) slice of player.State objects.
func CreateInitial() State {
	return State{defaultPlayers(), sync.Mutex{}}
}

// handleHttpRequest parses an HTTP request and responds with the appropriate function.
func (state *State) HandleHttpRequest(
	httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string) {
	switch httpRequest.Method {
	case http.MethodOptions:
		return
	case http.MethodGet:
		state.handleGetRequest(httpResponseWriter, httpRequest, relevantUriSegments)
	case http.MethodPost:
		state.handlePostRequest(httpResponseWriter, httpRequest, relevantUriSegments)
	default:
		http.Error(httpResponseWriter, "Method not GET or POST: "+httpRequest.Method, http.StatusBadRequest)
	}
}

// defaultPlayers returns a map of players created from default player names with colors
// according to the available chat colors, where the key is the player name.
func defaultPlayers() map[string]*player.State {
	initialColors := availableColors()
	numberOfColors := len(initialColors)
	initialNames := []string{"Mimi", "Aet", "Martin", "Markus", "Liisbet", "Madli", "Ben"}
	numberOfPlayers := len(initialNames)

	playerMap := make(map[string]*player.State, numberOfPlayers)

	for playerCount := 0; playerCount < numberOfPlayers; playerCount++ {
		playerName := initialNames[playerCount]

		// We cycle through all the colors again if there are more players than colors.
		playerColor := initialColors[playerCount%numberOfColors]

		playerMap[playerName] = player.CreateByNameAndColor(playerName, playerColor)
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

// handleGetRequest parses an HTTP GET request and responds with the appropriate function.
func (state *State) handleGetRequest(
	httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string) {
	if len(relevantUriSegments) < 1 {
		httpResponseWriter.WriteHeader(http.StatusBadRequest)
		httpResponseWriter.Write([]byte("Not enough segments in URI to determine what to do"))
		return
	}

	switch relevantUriSegments[0] {
	case "registered-players":
		state.writeRegisteredPlayerListJson(httpResponseWriter)
	case "available-colors":
		state.writeAvailableColorListJson(httpResponseWriter)
	default:
		http.NotFound(httpResponseWriter, httpRequest)
	}
}

// handlePostRequest parses an HTTP POST request and responds with the appropriate function.
func (state *State) handlePostRequest(
	httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string) {
	if len(relevantUriSegments) < 1 {
		httpResponseWriter.WriteHeader(http.StatusBadRequest)
		httpResponseWriter.Write([]byte("Not enough segments in URI to determine what to do"))
		return
	}

	switch relevantUriSegments[0] {
	case "new-player":
		state.handleNewPlayer(httpResponseWriter, httpRequest)
	case "update-player":
		state.handleUpdatePlayer(httpResponseWriter, httpRequest)
	case "reset-players":
		state.handleResetPlayers(httpResponseWriter, httpRequest)
	default:
		http.NotFound(httpResponseWriter, httpRequest)
	}
}

// writeRegisteredPlayerListJson writes a JSON object into the HTTP response which has
// the list of player objects as its "Players" attribute.
func (state *State) writeRegisteredPlayerListJson(httpResponseWriter http.ResponseWriter) {
	playerList := make([]player.State, 0, len(state.registeredPlayers))
	for _, registeredPlayer := range state.registeredPlayers {
		playerList = append(playerList, *registeredPlayer)
	}

	json.NewEncoder(httpResponseWriter).Encode(struct{ Players []player.State }{playerList})
}

// writeAvailableColorListJson writes a JSON object into the HTTP response which has
// the list of strings as its "Colors" attribute.
func (state *State) writeAvailableColorListJson(httpResponseWriter http.ResponseWriter) {
	json.NewEncoder(httpResponseWriter).Encode(struct{ Colors []string }{availableColors()})
}

// handleNewPlayer adds the player defined by the JSON of the request's body to the list
// of registered players, and returns the updated list as writeRegisteredPlayerNameListJson
// would.
func (state *State) handleNewPlayer(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Body == nil {
		http.Error(httpResponseWriter, "Empty request body", http.StatusBadRequest)
		return
	}

	var playerFromJson player.State
	parsingError := json.NewDecoder(httpRequest.Body).Decode(&playerFromJson)
	if parsingError != nil {
		http.Error(httpResponseWriter, "Error parsing JSON: "+parsingError.Error(), http.StatusBadRequest)
		return
	}

	_, playerExists := state.registeredPlayers[playerFromJson.Name]
	if playerExists {
		http.Error(httpResponseWriter, "Name "+playerFromJson.Name+" already registered", http.StatusBadRequest)
		return
	}

	playerColor := playerFromJson.Color
	if playerColor == "" {
		// The new player is assigned the next color in the list, cycling through all the colors
		// again if there are more players than colors. E.g. if there are already 5 players, the
		// 6th player gets the 6th color in the list. If there are only 4 colors in the list,
		// the new player would get the 2nd color. This does not account for players having
		// changed color, but it doesn't matter, as it is just a fun way of choosing an initial
		// color.
		colorList := availableColors()
		playerColor = colorList[len(state.registeredPlayers)%len(colorList)]
	}

	state.mutualExclusion.Lock()
	state.registeredPlayers[playerFromJson.Name] =
		player.CreateByNameAndColor(playerFromJson.Name, playerColor)
	state.mutualExclusion.Unlock()

	state.writeRegisteredPlayerListJson(httpResponseWriter)
}

// handleNewPlayer updates the player defined by the JSON of the request's body, taking the "Name"
// attribute as the key, and returns the updated list as writeRegisteredPlayerNameListJson
// would. Attributes which are present are updated, those which are missing remain unchanged.
func (state *State) handleUpdatePlayer(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Body == nil {
		http.Error(httpResponseWriter, "Empty request body", http.StatusBadRequest)
		return
	}

	var playerFromJson player.State
	parsingError := json.NewDecoder(httpRequest.Body).Decode(&playerFromJson)
	if parsingError != nil {
		http.Error(httpResponseWriter, "Error parsing JSON: "+parsingError.Error(), http.StatusBadRequest)
		return
	}

	existingPlayer, playerExists := state.registeredPlayers[playerFromJson.Name]
	if !playerExists {
		http.Error(httpResponseWriter, "Name "+playerFromJson.Name+" not found", http.StatusBadRequest)
		return
	}

	state.mutualExclusion.Lock()
	existingPlayer.UpdateNonEmptyStrings(&playerFromJson)
	state.mutualExclusion.Unlock()

	state.writeRegisteredPlayerListJson(httpResponseWriter)
}

// handleResetPlayers resets the player list to the initial list, and returns the updated list
// as writeRegisteredPlayerListJson would.
func (state *State) handleResetPlayers(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
	state.mutualExclusion.Lock()
	state.registeredPlayers = defaultPlayers()
	state.mutualExclusion.Unlock()

	state.writeRegisteredPlayerListJson(httpResponseWriter)
}
