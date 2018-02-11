package game

import (
	"encoding/json"
	"github.com/benoleary/ilutulestikud/player"
	"net/http"
	"sort"
	"sync"
	"time"
)

// state is a struct meant to encapsulate all the state required for a single game to function.
type state struct {
	gameName             string
	creationTime         time.Time
	participatingPlayers []*player.State
	turnNumber           int
	mutualExclusion      sync.Mutex
}

// createState constructs a state object with a non-nil, non-empty slice of player.State objects.
func createState(gameName string, playerHandler *player.Handler, playerNames []string) *state {
	numberOfPlayers := len(playerNames)
	playerStates := make([]*player.State, numberOfPlayers)
	for playerIndex := 0; playerIndex < numberOfPlayers; playerIndex++ {
		playerStates[playerIndex] = playerHandler.GetPlayerByName(playerNames[playerIndex])
	}

	return &state{gameName, time.Now(), playerStates, 1, sync.Mutex{}}
}

// hasPlayerAsParticipant returns true if the given player name matches
// the name of any of the game's participating players.
func (gameState *state) hasPlayerAsParticipant(playerName string) bool {
	for _, participatingPlayer := range gameState.participatingPlayers {
		if participatingPlayer.Name == playerName {
			return true
		}
	}

	return false
}

// byCreationTime implements sort.Interface for []*state based on the creationTime field.
type byCreationTime []*state

// Len implements part of the sort.Interface for byCreationTime.
func (statePointerArray byCreationTime) Len() int {
	return len(statePointerArray)
}

// Swap implements part of the sort.Interface for byCreationTime.
func (statePointerArray byCreationTime) Swap(firstIndex int, secondIndex int) {
	statePointerArray[firstIndex], statePointerArray[secondIndex] =
		statePointerArray[secondIndex], statePointerArray[firstIndex]
}

// Less implements part of the sort.Interface for byCreationTime.
func (statePointerArray byCreationTime) Less(firstIndex int, secondIndex int) bool {
	return statePointerArray[firstIndex].creationTime.Before(
		statePointerArray[secondIndex].creationTime)
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

// turnSummary contains the information to determine what games involve a player and whose turn it is.
// All the fields need to be public so that the JSON encoder can see them to serialize them.
type turnSummary struct {
	GameName               string
	TurnNumber             int
	PlayersInNextTurnOrder []string
	IsPlayerTurn           bool
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

	gameList := make([]*state, 0)
	for _, gameState := range handler.gameStates {
		if gameState.hasPlayerAsParticipant(playerName) {
			gameList = append(gameList, gameState)
		}
	}

	sort.Sort(byCreationTime(gameList))

	numberOfGamesWithPlayer := len(gameList)
	namesOfGamesWithPlayer := make([]string, numberOfGamesWithPlayer)
	turnSummaries := make([]turnSummary, numberOfGamesWithPlayer)
	for gameIndex := 0; gameIndex < numberOfGamesWithPlayer; gameIndex++ {
		nameOfGame := gameList[gameIndex].gameName
		gameTurn := gameList[gameIndex].turnNumber
		namesOfGamesWithPlayer[gameIndex] = nameOfGame

		gameParticipants := gameList[gameIndex].participatingPlayers
		numberOfParticipants := len(gameParticipants)

		playerNamesInTurnOrder := make([]string, numberOfParticipants)

		for playerIndex := 0; playerIndex < numberOfParticipants; playerIndex++ {
			// Game turns begin with 1 rather than 0, so this sets the player names in order,
			// wrapping index back to 0 when at the end of the list.
			// E.g turn 3, 5 players: playerNamesInTurnOrder will start with
			// gameParticipants[2], then [3], then [4], then [0], then [1].
			playerNamesInTurnOrder[playerIndex] =
				gameParticipants[(playerIndex+gameTurn-1)%numberOfParticipants].Name
		}

		turnSummaries[gameIndex] = turnSummary{
			nameOfGame,
			gameTurn,
			playerNamesInTurnOrder,
			playerName == playerNamesInTurnOrder[0]}
	}

	json.NewEncoder(httpResponseWriter).Encode(struct {
		Games         []string
		TurnSummaries []turnSummary
	}{namesOfGamesWithPlayer, turnSummaries[:]})
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
