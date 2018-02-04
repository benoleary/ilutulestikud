package game

import (
	"encoding/json"
	"github.com/benoleary/ilutulestikud/player"
	"net/http"
	"sync"
)

// state is a struct meant to encapsulate all the state required for a single game to function.
type state struct {
	gameName             string
	participatingPlayers []*player.State
	mutualExclusion      sync.Mutex
}

// createState constructs a state object with a non-nil, non-empty slice of player.State objects.
func createState(gameName string, playerHandler *player.Handler, playerNames []string) *state {
	numberOfPlayers := len(playerNames)
	playerStates := make([]*player.State, numberOfPlayers)
	for playerIndex := 0; playerIndex < numberOfPlayers; playerIndex++ {
		playerStates[playerIndex] = playerHandler.GetPlayerByName(playerNames[playerIndex])
	}

	return &state{gameName, playerStates, sync.Mutex{}}
}

// Handler is a struct meant to encapsulate all the state co-ordinating all the games.
// It implements github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
type Handler struct {
	playerHandler   *player.Handler
	gameStates      map[string]*state
	mutualExclusion sync.Mutex
}

// CreateHandler constructs a Handler object with a pointer to the player.Handler which
// handles the players.
func CreateHandler(playerHandler *player.Handler) Handler {
	return Handler{playerHandler, make(map[string]*state, 0), sync.Mutex{}}
}

// HandleGetRequest parses an HTTP GET request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (handler *Handler) HandleGet(
	httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string) {
	if len(relevantUriSegments) < 1 {
		httpResponseWriter.WriteHeader(http.StatusBadRequest)
		httpResponseWriter.Write([]byte("Not enough segments in URI to determine what to do"))
		return
	}

	switch relevantUriSegments[0] {
	case "all-games-with-player":
		handler.writeGamesWithPlayerListJson(httpResponseWriter, relevantUriSegments[1:])
	default:
		http.NotFound(httpResponseWriter, httpRequest)
	}
}

// HandlePost parses an HTTP POST request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (handler *Handler) HandlePost(
	httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string) {
	if len(relevantUriSegments) < 1 {
		httpResponseWriter.WriteHeader(http.StatusBadRequest)
		httpResponseWriter.Write([]byte("Not enough segments in URI to determine what to do"))
		return
	}

	switch relevantUriSegments[0] {
	case "create-new-game":
		handler.handleNewGame(httpResponseWriter, httpRequest)
	default:
		http.NotFound(httpResponseWriter, httpRequest)
	}
}

// writeRegisteredPlayerListJson writes a JSON object into the HTTP response which has
// the list of player objects as its "Players" attribute.
func (handler *Handler) writeGamesWithPlayerListJson(
	httpResponseWriter http.ResponseWriter, relevantUriSegments []string) {
	if len(relevantUriSegments) < 1 {
		httpResponseWriter.WriteHeader(http.StatusBadRequest)
		httpResponseWriter.Write([]byte("Not enough segments in URI to determine player name"))
		return
	}

	playerName := relevantUriSegments[0]

	gameList := make([]string, 0)
	for _, gameState := range handler.gameStates {
		for _, participatingPlayer := range gameState.participatingPlayers {
			if participatingPlayer.Name == playerName {
				gameList = append(gameList, gameState.gameName)
				break
			}
		}
	}

	json.NewEncoder(httpResponseWriter).Encode(struct{ Games []string }{gameList})
}

// handleNewGame adds a new game to the map of game state objects.
func (handler *Handler) handleNewGame(
	httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Body == nil {
		http.Error(httpResponseWriter, "Empty request body", http.StatusBadRequest)
		return
	}

	var newGameFromJson struct {
		Name    string
		Players []string
	}

	parsingError := json.NewDecoder(httpRequest.Body).Decode(&newGameFromJson)
	if parsingError != nil {
		http.Error(httpResponseWriter, "Error parsing JSON: "+parsingError.Error(), http.StatusBadRequest)
		return
	}

	_, gameExists := handler.gameStates[newGameFromJson.Name]
	if gameExists {
		http.Error(httpResponseWriter, "Name "+newGameFromJson.Name+" already exists", http.StatusBadRequest)
		return
	}

	handler.mutualExclusion.Lock()
	handler.gameStates[newGameFromJson.Name] =
		createState(
			newGameFromJson.Name,
			handler.playerHandler,
			newGameFromJson.Players)
	handler.mutualExclusion.Unlock()

	httpResponseWriter.WriteHeader(http.StatusOK)
}
