package game

import (
	"encoding/json"
	"net/http"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// GetAndPostHandler is a struct meant to encapsulate all the state co-ordinating all the games.
// It implements github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
type GetAndPostHandler struct {
	playerCollection player.Collection
	gameCollection   Collection
}

// NewGetAndPostHandler constructs a Handler object around the given game.Collection object.
func NewGetAndPostHandler(
	playerCollection player.Collection,
	gameCollection Collection) *GetAndPostHandler {
	return &GetAndPostHandler{playerCollection: playerCollection, gameCollection: gameCollection}
}

// HandleGet parses an HTTP GET request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (getAndPostHandler *GetAndPostHandler) HandleGet(
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "all-games-with-player":
		return getAndPostHandler.writeTurnSummariesForPlayer(relevantSegments[1:])
	case "game-as-seen-by-player":
		return getAndPostHandler.writeGameForPlayer(relevantSegments[1:])
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// HandlePost parses an HTTP POST request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (getAndPostHandler *GetAndPostHandler) HandlePost(
	httpBodyDecoder *json.Decoder,
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "create-new-game":
		return getAndPostHandler.handleNewGame(httpBodyDecoder, relevantSegments)
	case "player-action":
		return getAndPostHandler.handlePlayerAction(httpBodyDecoder, relevantSegments)
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// writeTurnSummariesForPlayer writes a JSON object into the HTTP response which has
// the list of turn summary objects as its "TurnSummaries" attribute.
func (getAndPostHandler *GetAndPostHandler) writeTurnSummariesForPlayer(
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine player", http.StatusBadRequest
	}

	playerIdentifier := relevantSegments[0]

	_, playerExists := getAndPostHandler.playerCollection.Get(playerIdentifier)

	if !playerExists {
		return "Player with identifier %v is not registered, cannot participate in games", http.StatusBadRequest
	}

	return TurnSummariesForFrontend(getAndPostHandler.gameCollection, playerIdentifier), http.StatusOK
}

// handleNewGame adds a new game to the map of game state objects.
func (getAndPostHandler *GetAndPostHandler) handleNewGame(
	httpBodyDecoder *json.Decoder,
	relevantSegments []string) (interface{}, int) {
	var newGame endpoint.GameDefinition

	parsingError := httpBodyDecoder.Decode(&newGame)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	if newGame.Name == "" {
		return "No name for new game parsed from JSON", http.StatusBadRequest
	}

	addError := getAndPostHandler.gameCollection.Add(newGame, getAndPostHandler.playerCollection)

	if addError != nil {
		return addError, http.StatusBadRequest
	}

	return "OK", http.StatusOK
}

// writeGameForPlayer writes a JSON representation of the current state of the game
// with the given name for the player with the given name.
func (getAndPostHandler *GetAndPostHandler) writeGameForPlayer(
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 2 {
		return "Not enough segments in URI to determine game name and player name", http.StatusBadRequest
	}

	gameIdentifier := relevantSegments[0]
	playerIdentifier := relevantSegments[1]

	gameState, isFound := getAndPostHandler.gameCollection.Get(gameIdentifier)

	if !isFound {
		errorMessage :=
			"Game " + gameIdentifier + " does not exist, cannot add chat from player " + playerIdentifier
		return errorMessage, http.StatusBadRequest
	}

	if !gameState.HasPlayerAsParticipant(playerIdentifier) {
		errorMessage :=
			"Player " + playerIdentifier + " is not a participant in game " + gameIdentifier
		return errorMessage, http.StatusBadRequest
	}

	return ForPlayer(gameState, playerIdentifier), http.StatusOK
}

// handlePlayerAction passes on the given player action to the relevant game.
func (getAndPostHandler *GetAndPostHandler) handlePlayerAction(
	httpBodyDecoder *json.Decoder,
	relevantSegments []string) (interface{}, int) {
	var playerAction endpoint.PlayerAction

	parsingError := httpBodyDecoder.Decode(&playerAction)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	gameState, isFound := getAndPostHandler.gameCollection.Get(playerAction.Game)

	if !isFound {
		errorMessage :=
			"Game " + playerAction.Game + " does not exist, cannot perform action from player " + playerAction.Player
		return errorMessage, http.StatusBadRequest
	}

	actingPlayer, isRegisteredPlayer := getAndPostHandler.playerCollection.Get(playerAction.Player)

	if !isRegisteredPlayer {
		errorMessage :=
			"Player " + playerAction.Player + " is not registered, should have no actions for game " + playerAction.Game
		return errorMessage, http.StatusBadRequest
	}

	if !gameState.HasPlayerAsParticipant(playerAction.Player) {
		errorMessage :=
			"Player " + playerAction.Player + " is not a participant in game " + playerAction.Game
		return errorMessage, http.StatusBadRequest
	}

	actionError := gameState.PerformAction(actingPlayer, playerAction)

	if actionError != nil {
		return actionError, http.StatusBadRequest
	}

	return "OK", http.StatusOK
}
