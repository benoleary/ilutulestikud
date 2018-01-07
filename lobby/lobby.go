package lobby

import (
	"encoding/json"
	"github.com/benoleary/ilutulestikud/player"
	"net/http"
	"sync"
)

// Lobby is a struct meant to encapsulate all the state required for the lobby page to function.
type State struct {
	registeredPlayers []player.State
	mutualExclusion   sync.Mutex
}

// MakeEmpty constructs a State object with a non-nil (but empty) slice of player.State objects.
func CreateEmpty() State {
	return State{make([]player.State, 0, 8), sync.Mutex{}}
}

// handleHttpRequest parses an HTTP request and responds with the appropriate function.
func (state *State) HandleHttpRequest(
	httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string) {
	// While developing locally, we allow any request from the local Angular server
	/*
		httpResponseWriter.Header().Set("Access-Control-Allow-Origin", "http://localhost:4233")
		httpResponseWriter.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		httpResponseWriter.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	*/

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

// handleGetRequest parses an HTTP GET request and responds with the appropriate function.
func (state *State) handleGetRequest(
	httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string) {
	if len(relevantUriSegments) < 1 {
		httpResponseWriter.WriteHeader(http.StatusBadRequest)
		httpResponseWriter.Write([]byte("Not enough segments in URI to determine what to do"))
		return
	}

	switch relevantUriSegments[0] {
	case "registered-player-names":
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
	default:
		http.NotFound(httpResponseWriter, httpRequest)
	}
}

// playerNames returns an array of all the names of the registered players.
func (state *State) playerNames() []string {
	nameList := make([]string, 0, len(state.registeredPlayers))
	for _, registeredPlayer := range state.registeredPlayers {
		nameList = append(nameList, registeredPlayer.Name())
	}

	return nameList
}

// writeRegisteredPlayerListJson writes a JSON object into the HTTP response which has
// the list of player names as its "Names" attribute.
func (state *State) writeRegisteredPlayerListJson(httpResponseWriter http.ResponseWriter) {
	json.NewEncoder(httpResponseWriter).Encode(struct{ Names []string }{state.playerNames()})
}

// writeRegisteredPlayerListJson writes a JSON object into the HTTP response which has
// the list of player names as its "Names" attribute.
func (state *State) handleNewPlayer(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Body == nil {
		http.Error(httpResponseWriter, "Empty request body", http.StatusBadRequest)
		return
	}

	var jsonObject struct{ Name string }
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

	state.mutualExclusion.Lock()
	state.registeredPlayers = append(state.registeredPlayers, player.CreateByNameOnly(jsonObject.Name))
	state.mutualExclusion.Unlock()

	state.writeRegisteredPlayerListJson(httpResponseWriter)
}
