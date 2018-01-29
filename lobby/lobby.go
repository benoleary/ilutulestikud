package lobby

import (
	"encoding/json"
	"github.com/benoleary/ilutulestikud/player"
	"net/http"
	"sync"
)

// State is a struct meant to encapsulate all the state required for the lobby page to function.
type State struct {
	registeredPlayers []player.State
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

func defaultPlayers() []player.State {
	initialColors := availableColors()
	return []player.State{
		player.CreateByNameAndColor("Mimi", initialColors[0]),
		player.CreateByNameAndColor("Aet", initialColors[1]),
		player.CreateByNameAndColor("Martin", initialColors[2]),
		player.CreateByNameAndColor("Markus", initialColors[3]),
		player.CreateByNameAndColor("Liisbet", initialColors[4]),
		player.CreateByNameAndColor("Madli", initialColors[5]),
		player.CreateByNameAndColor("Ben", initialColors[6])}
}

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

// playerNames returns an array of all the names of the registered players.
func (state *State) playerNames() []string {
	nameList := make([]string, 0, len(state.registeredPlayers))
	for _, registeredPlayer := range state.registeredPlayers {
		nameList = append(nameList, registeredPlayer.Name)
	}

	return nameList
}

// writeRegisteredPlayerListJson writes a JSON object into the HTTP response which has
// the list of player objects as its "Players" attribute.
func (state *State) writeRegisteredPlayerListJson(httpResponseWriter http.ResponseWriter) {
	json.NewEncoder(httpResponseWriter).Encode(struct{ Players []player.State }{state.registeredPlayers})
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

	existingNames := state.playerNames()
	for _, existingName := range existingNames {
		if existingName == playerFromJson.Name {
			http.Error(httpResponseWriter, "Name "+existingName+" already registered", http.StatusBadRequest)
			return
		}
	}

	playerColor := playerFromJson.Color
	if playerColor == "" {
		playerColor = "white"
	}

	state.mutualExclusion.Lock()
	state.registeredPlayers = append(
		state.registeredPlayers,
		player.CreateByNameAndColor(playerFromJson.Name, playerColor))
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

	for playerIndex := len(state.registeredPlayers) - 1; playerIndex >= 0; playerIndex-- {
		if state.registeredPlayers[playerIndex].Name == playerFromJson.Name {
			state.mutualExclusion.Lock()
			state.registeredPlayers[playerIndex].UpdateNonEmptyStrings(&playerFromJson)
			state.mutualExclusion.Unlock()

			state.writeRegisteredPlayerListJson(httpResponseWriter)
			return
		}
	}

	http.Error(httpResponseWriter, "Name "+playerFromJson.Name+" not found", http.StatusBadRequest)
}

// handleResetPlayers resets the player list to the initial list, and returns the updated list
// as writeRegisteredPlayerListJson would.
func (state *State) handleResetPlayers(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
	state.mutualExclusion.Lock()
	state.registeredPlayers = defaultPlayers()
	state.mutualExclusion.Unlock()

	state.writeRegisteredPlayerListJson(httpResponseWriter)
}
