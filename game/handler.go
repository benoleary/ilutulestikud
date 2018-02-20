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
func (handler *Handler) HandleGet(relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "all-games-with-player":
		return handler.writeTurnSummariesForPlayer(relevantSegments[1:])
	case "game-as-seen-by-player":
		return handler.writeGameForPlayer(relevantSegments[1:])
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// HandlePost parses an HTTP POST request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (handler *Handler) HandlePost(httpBodyDecoder *json.Decoder, relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "create-new-game":
		return handler.handleNewGame(httpBodyDecoder, relevantSegments)
	case "send-chat-message":
		return handler.handleNewChatMessage(httpBodyDecoder, relevantSegments)
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// turnSummary contains the information to determine what games involve a player and whose turn it is.
// All the fields need to be public so that the JSON encoder can see them to serialize them.
// The creation timestamp is int64 because that is what time.Unix() returns.
type turnSummary struct {
	GameName                   string
	CreationTimestampInSeconds int64
	TurnNumber                 int
	PlayersInNextTurnOrder     []string
	IsPlayerTurn               bool
}

// writeTurnSummariesForPlayer writes a JSON object into the HTTP response which has
// the list of turn summary objects as its "TurnSummaries" attribute.
func (handler *Handler) writeTurnSummariesForPlayer(relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine player name", http.StatusBadRequest
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

	return struct{ TurnSummaries []turnSummary }{turnSummaries[:]}, http.StatusOK
}

// handleNewGame adds a new game to the map of game state objects.
func (handler *Handler) handleNewGame(httpBodyDecoder *json.Decoder, relevantSegments []string) (interface{}, int) {
	var newGame struct {
		Name    string
		Players []string
	}

	parsingError := httpBodyDecoder.Decode(&newGame)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	_, gameExists := handler.gameStates[newGame.Name]
	if gameExists {
		return "Name " + newGame.Name + " already exists", http.StatusBadRequest
	}

	handler.mutualExclusion.Lock()
	handler.gameStates[newGame.Name] =
		NewState(
			newGame.Name,
			handler.playerHandler,
			newGame.Players)
	handler.mutualExclusion.Unlock()

	return "OK", http.StatusOK
}

// writeGameForPlayer writes a JSON representation of the current state of the game
// with the given name for the player with the given name.
func (handler *Handler) writeGameForPlayer(relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 2 {
		return "Not enough segments in URI to determine game name and player name", http.StatusBadRequest
	}

	gameName := relevantSegments[0]
	playerName := relevantSegments[1]

	gameState, isFound := handler.gameStates[gameName]
	if !isFound {
		return " game " + gameName + " does not exist, cannot add chat from player " + playerName, http.StatusBadRequest
	}

	if !gameState.HasPlayerAsParticipant(playerName) {
		return "Player " + playerName + " is not a participant in game " + gameName, http.StatusBadRequest
	}

	return struct{ Knowledge PlayerKnowledge }{gameState.ForPlayer(playerName)}, http.StatusOK
}

// handleNewChatMessage adds the given chat message to the relevant game state,
// as coming from the given player.
func (handler *Handler) handleNewChatMessage(httpBodyDecoder *json.Decoder, relevantSegments []string) (interface{}, int) {
	var chatMessage struct {
		Player  string
		Game    string
		Message string
	}

	parsingError := httpBodyDecoder.Decode(&chatMessage)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	chattingPlayer, playerFound := handler.playerHandler.GetPlayerByName(chatMessage.Player)
	if !playerFound {
		return "Player " + chatMessage.Player + " is not registered, and is not a participant in game " + chatMessage.Game, http.StatusBadRequest
	}

	gameState, isFound := handler.gameStates[chatMessage.Game]
	if !isFound {
		return "Game " + chatMessage.Game + " does not exist, cannot add chat from player " + chatMessage.Player, http.StatusBadRequest
	}

	if !gameState.HasPlayerAsParticipant(chatMessage.Player) {
		return "Player " + chatMessage.Player + " is not a participant in game " + chatMessage.Game, http.StatusBadRequest
	}

	gameState.RecordPlayerChatMessage(chattingPlayer, chatMessage.Message)

	return "OK", http.StatusOK
}
