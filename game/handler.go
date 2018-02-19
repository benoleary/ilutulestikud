package game

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"

	"github.com/benoleary/ilutulestikud/player"
)

// Handler is a struct meant to encapsulate all the state co-ordinating all the games.
// It implements github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
type Handler struct {
	playerHandler   *player.Handler
	gameStates      map[string]*State
	mutualExclusion sync.Mutex
}

// NewHandler constructs a Handler object with a pointer to the player.Handler which
// handles the players, returning a pointer to the newly-created object.
func NewHandler(playerHandler *player.Handler) *Handler {
	return &Handler{playerHandler, make(map[string]*State, 0), sync.Mutex{}}
}

// HandleGet parses an HTTP GET request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (handler *Handler) HandleGet(
	httpResponseWriter http.ResponseWriter,
	httpRequest *http.Request,
	relevantSegments []string) {
	if len(relevantSegments) < 1 {
		httpResponseWriter.WriteHeader(http.StatusBadRequest)
		httpResponseWriter.Write([]byte("Not enough segments in URI to determine what to do"))
		return
	}

	switch relevantSegments[0] {
	case "all-games-with-player":
		handler.writeTurnSummariesForPlayer(httpResponseWriter, relevantSegments[1:])
	case "game-as-seen-by-player":
		handler.writeGameForPlayer(httpResponseWriter, relevantSegments[1:])
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
	case "create-new-game":
		handler.handleNewGame(httpResponseWriter, httpRequest)
	case "send-chat-message":
		handler.handleNewChatMessage(httpResponseWriter, httpRequest)
	default:
		http.NotFound(httpResponseWriter, httpRequest)
	}
}

// turnSummary contains the information to determine what games involve a player and whose turn it is.
// All the fields need to be public so that the JSON encoder can see them to serialize them.
type turnSummary struct {
	GameName                   string
	CreationTimestampInSeconds int64
	TurnNumber                 int
	PlayersInNextTurnOrder     []string
	IsPlayerTurn               bool
}

// writeTurnSummariesForPlayer writes a JSON object into the HTTP response which has
// the list of turn summary objects as its "TurnSummaries" attribute.
func (handler *Handler) writeTurnSummariesForPlayer(
	httpResponseWriter http.ResponseWriter, relevantSegments []string) {
	if len(relevantSegments) < 1 {
		httpResponseWriter.WriteHeader(http.StatusBadRequest)
		httpResponseWriter.Write([]byte("Not enough segments in URI to determine player name"))
		return
	}

	playerName := relevantSegments[0]

	gameList := make([]*State, 0)
	for _, gameState := range handler.gameStates {
		if gameState.HasPlayerAsParticipant(playerName) {
			gameList = append(gameList, gameState)
		}
	}

	sort.Sort(byCreationTime(gameList))

	numberOfGamesWithPlayer := len(gameList)
	turnSummaries := make([]turnSummary, numberOfGamesWithPlayer)
	for gameIndex := 0; gameIndex < numberOfGamesWithPlayer; gameIndex++ {
		nameOfGame := gameList[gameIndex].gameName
		gameTurn := gameList[gameIndex].turnNumber

		gameParticipants := gameList[gameIndex].participatingPlayers
		numberOfParticipants := len(gameParticipants)

		playerNamesInTurnOrder := make([]string, numberOfParticipants)

		for playerIndex := 0; playerIndex < numberOfParticipants; playerIndex++ {
			// Game turns begin with 1 rather than 0, so this sets the player names in order,
			// wrapping index back to 0 when at the end of the list.
			// E.g. turn 3, 5 players: playerNamesInTurnOrder will start with
			// gameParticipants[2], then [3], then [4], then [0], then [1].
			playerNamesInTurnOrder[playerIndex] =
				gameParticipants[(playerIndex+gameTurn-1)%numberOfParticipants].Name
		}

		turnSummaries[gameIndex] = turnSummary{
			nameOfGame,
			gameList[gameIndex].creationTime.Unix(),
			gameTurn,
			playerNamesInTurnOrder,
			playerName == playerNamesInTurnOrder[0]}
	}

	json.NewEncoder(httpResponseWriter).Encode(struct{ TurnSummaries []turnSummary }{turnSummaries[:]})
}

// handleNewGame adds a new game to the map of game state objects.
func (handler *Handler) handleNewGame(
	httpResponseWriter http.ResponseWriter,
	httpRequest *http.Request) {
	if httpRequest.Body == nil {
		http.Error(httpResponseWriter, "Empty request body", http.StatusBadRequest)
		return
	}

	var newGame struct {
		Name    string
		Players []string
	}

	parsingError := json.NewDecoder(httpRequest.Body).Decode(&newGame)
	if parsingError != nil {
		http.Error(httpResponseWriter, "Error parsing JSON: "+parsingError.Error(), http.StatusBadRequest)
		return
	}

	_, gameExists := handler.gameStates[newGame.Name]
	if gameExists {
		http.Error(httpResponseWriter, "Name "+newGame.Name+" already exists", http.StatusBadRequest)
		return
	}

	handler.mutualExclusion.Lock()
	handler.gameStates[newGame.Name] =
		NewState(
			newGame.Name,
			handler.playerHandler,
			newGame.Players)
	handler.mutualExclusion.Unlock()

	httpResponseWriter.WriteHeader(http.StatusOK)
}

// gameForPlayer returns the State pointer for the game with the given name, unless no player
// with the given name is a participant, in which case an error message is written in the
// provided *http.ResponseWriter ifnot nil, and nil and false are returned.
func (handler *Handler) gameWithParticipant(
	gameName string,
	playerName string,
	httpResponseWriter http.ResponseWriter) (*State, bool) {
	gameState := handler.gameStates[gameName]
	if !gameState.HasPlayerAsParticipant(playerName) {
		if httpResponseWriter != nil {
			httpResponseWriter.WriteHeader(http.StatusBadRequest)
			httpResponseWriter.Write([]byte(
				"Player " + playerName + " is not a participant in game " + gameName))
		}

		return nil, false
	}

	return gameState, true
}

// writeGameForPlayer writes a JSON representation of the current state of the game
// with the given name for the player with the given name.
func (handler *Handler) writeGameForPlayer(
	httpResponseWriter http.ResponseWriter,
	relevantSegments []string) {
	if len(relevantSegments) < 2 {
		httpResponseWriter.WriteHeader(http.StatusBadRequest)
		httpResponseWriter.Write([]byte("Not enough segments in URI to determine game name and player name"))
		return
	}

	gameName := relevantSegments[0]
	playerName := relevantSegments[1]

	gameState, validParticipant := handler.gameWithParticipant(
		gameName, playerName, httpResponseWriter)
	if !validParticipant {
		return
	}

	json.NewEncoder(httpResponseWriter).Encode(struct{ Knowledge PlayerKnowledge }{gameState.ForPlayer(playerName)})
}

// handleNewChatMessage adds the given chat message to the relevant game state,
// as coming from the given player.
func (handler *Handler) handleNewChatMessage(
	httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Body == nil {
		http.Error(httpResponseWriter, "Empty request body", http.StatusBadRequest)
		return
	}

	var chatMessage struct {
		Player  string
		Game    string
		Message string
	}

	parsingError := json.NewDecoder(httpRequest.Body).Decode(&chatMessage)
	if parsingError != nil {
		http.Error(httpResponseWriter, "Error parsing JSON: "+parsingError.Error(), http.StatusBadRequest)
		return
	}

	gameState, validParticipant := handler.gameWithParticipant(
		chatMessage.Game, chatMessage.Player, httpResponseWriter)
	if !validParticipant {
		return
	}

	gameState.RecordPlayerChatMessage(chatMessage.Player, chatMessage.Message)

	httpResponseWriter.WriteHeader(http.StatusOK)
}
