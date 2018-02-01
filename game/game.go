package game

import (
	"encoding/json"
	"github.com/benoleary/ilutulestikud/lobby"
	"github.com/benoleary/ilutulestikud/player"
	"net/http"
	"sync"
)

// SingleState is a struct meant to encapsulate all the state required for a single game to function.
type SingleState struct {
	gameName             string
	participatingPlayers []*player.State
	mutualExclusion      sync.Mutex
}

// CollectionState is a struct meant to encapsulate all the state co-odrinating all the games.
// It implements github.com/benoleary/ilutulestikud/http.GetAndPostHandler.
type CollectionState struct {
	lobbyState      *lobby.State
	activeGames     map[string]*SingleState
	mutualExclusion sync.Mutex
}

// CreateCollectionState constructs a CollectionState object with a pointer to the lobby which
// contains the players.
func CreateCollectionState(lobbyState *lobby.State) CollectionState {
	return CollectionState{lobbyState, make(map[string]*SingleState, 0), sync.Mutex{}}
}

// CreateSingleState constructs a SingleState object with a non-nil, non-empty slice of player.State objects.
func createSingleState(gameName string, lobbyState *lobby.State, playerNames []string) *SingleState {
	numberOfPlayers := len(playerNames)
	playerStates := make([]*player.State, numberOfPlayers)
	for playerIndex := 0; playerIndex < numberOfPlayers; playerIndex++ {
		playerStates[playerIndex] = lobbyState.GetPlayerByName(playerNames[playerIndex])
	}

	return &SingleState{gameName, playerStates, sync.Mutex{}}
}

// HandleGetRequest parses an HTTP GET request and responds with the appropriate function.
// This implements github.com/benoleary/ilutulestikud/http.GetHandler.
func (collectionState *CollectionState) HandleGet(
	httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string) {
	if len(relevantUriSegments) < 1 {
		httpResponseWriter.WriteHeader(http.StatusBadRequest)
		httpResponseWriter.Write([]byte("Not enough segments in URI to determine what to do"))
		return
	}

	switch relevantUriSegments[0] {
	case "all-games-with-player":
		collectionState.writeGamesWithPlayerListJson(httpResponseWriter, relevantUriSegments[1:])
	default:
		http.NotFound(httpResponseWriter, httpRequest)
	}
}

// HandlePost parses an HTTP POST request and responds with the appropriate function.
// This implements github.com/benoleary/ilutulestikud/http.PostHandler.
func (collectionState *CollectionState) HandlePost(
	httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string) {
	if len(relevantUriSegments) < 1 {
		httpResponseWriter.WriteHeader(http.StatusBadRequest)
		httpResponseWriter.Write([]byte("Not enough segments in URI to determine what to do"))
		return
	}

	switch relevantUriSegments[0] {
	case "create-new-game":
		collectionState.handleNewGame(httpResponseWriter, httpRequest)
	default:
		http.NotFound(httpResponseWriter, httpRequest)
	}
}

// writeRegisteredPlayerListJson writes a JSON object into the HTTP response which has
// the list of player objects as its "Players" attribute.
func (collectionState *CollectionState) writeGamesWithPlayerListJson(
	httpResponseWriter http.ResponseWriter, relevantUriSegments []string) {
	if len(relevantUriSegments) < 1 {
		httpResponseWriter.WriteHeader(http.StatusBadRequest)
		httpResponseWriter.Write([]byte("Not enough segments in URI to determine player name"))
		return
	}

	playerName := relevantUriSegments[0]

	gameList := make([]string, 0)
	for _, activeGame := range collectionState.activeGames {
		for _, participatingPlayer := range activeGame.participatingPlayers {
			if participatingPlayer.Name == playerName {
				gameList = append(gameList, activeGame.gameName)
				break
			}
		}
	}

	json.NewEncoder(httpResponseWriter).Encode(struct{ Games []string }{gameList})
}

// handleNewGame adds a new game to the map of SingleState objects.
func (collectionState *CollectionState) handleNewGame(
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

	_, gameExists := collectionState.activeGames[newGameFromJson.Name]
	if gameExists {
		http.Error(httpResponseWriter, "Name "+newGameFromJson.Name+" already exists", http.StatusBadRequest)
		return
	}

	collectionState.mutualExclusion.Lock()
	collectionState.activeGames[newGameFromJson.Name] =
		createSingleState(
			newGameFromJson.Name,
			collectionState.lobbyState,
			newGameFromJson.Players)
	collectionState.mutualExclusion.Unlock()

	httpResponseWriter.WriteHeader(http.StatusOK)
}
