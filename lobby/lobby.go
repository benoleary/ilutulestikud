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
	return []player.State{
		player.CreateByNameAndColor("Mimi", "red"),
		player.CreateByNameAndColor("Aet", "white"),
		player.CreateByNameAndColor("Martin", "green"),
		player.CreateByNameAndColor("Markus", "blue"),
		player.CreateByNameAndColor("Liisbet", "yellow"),
		player.CreateByNameAndColor("Madli", "orange"),
		player.CreateByNameAndColor("Ben", "purple")}
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

// handleNewPlayer adds the player defined by the JSON of the request's body to the list
// of registered players, and returns the updated list as writeRegisteredPlayerNameListJson
// would.
func (state *State) handleNewPlayer(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Body == nil {
		http.Error(httpResponseWriter, "Empty request body", http.StatusBadRequest)
		return
	}

	var jsonObject struct {
		Name  string
		Color string
	}
	parsingError := json.NewDecoder(httpRequest.Body).Decode(&jsonObject)
	if parsingError != nil {
		http.Error(httpResponseWriter, "Error parsing JSON: "+parsingError.Error(), http.StatusBadRequest)
		return
	}

	existingNames := state.playerNames()
	for _, existingName := range existingNames {
		if existingName == jsonObject.Name {
			http.Error(httpResponseWriter, "Name "+existingName+" already registered", http.StatusBadRequest)
			return
		}
	}

	playerColor := jsonObject.Color
	if playerColor == "" {
		playerColor = "white"
	}

	state.mutualExclusion.Lock()
	state.registeredPlayers = append(
		state.registeredPlayers,
		player.CreateByNameAndColor(jsonObject.Name, playerColor))
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
