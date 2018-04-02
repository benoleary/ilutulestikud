package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/game"
)

// gameEndpointHandler is a struct meant to encapsulate all the state co-ordinating
// interaction with all the games through the endpoints.
type gameEndpointHandler struct {
	gameCollection    game.StateCollection
	segmentTranslator EndpointSegmentTranslator
}

// HandleGet parses an HTTP GET request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (gameHandler *gameEndpointHandler) HandleGet(
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "available-rulesets":
		return gameHandler.writeAvailableRulesets()
	case "all-games-with-player":
		return gameHandler.writeTurnSummariesForPlayer(relevantSegments[1:])
	case "game-as-seen-by-player":
		return gameHandler.writeGameForPlayer(relevantSegments[1:])
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// HandlePost parses an HTTP POST request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (gameHandler *gameEndpointHandler) HandlePost(
	httpBodyDecoder *json.Decoder,
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "create-new-game":
		return gameHandler.handleNewGame(httpBodyDecoder, relevantSegments)
	case "player-action":
		return gameHandler.handlePlayerAction(httpBodyDecoder, relevantSegments)
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// writeAvailableRulesets writes a JSON object into the HTTP response which has
// the list of available rulesets as its "Rulesets" attribute.
func (gameHandler *gameEndpointHandler) writeAvailableRulesets() (interface{}, int) {
	return endpoint.RulesetList{Rulesets: game.AvailableRulesets()}, http.StatusOK
}

// writeTurnSummariesForPlayer writes a JSON object into the HTTP response which has
// the list of turn summary objects as its "TurnSummaries" attribute.
func (gameHandler *gameEndpointHandler) writeTurnSummariesForPlayer(
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine player", http.StatusBadRequest
	}

	playerIdentifier := relevantSegments[0]

	playerName, identificationError :=
		gameHandler.segmentTranslator.FromSegment(playerIdentifier)

	if identificationError != nil {
		return identificationError, http.StatusBadRequest
	}

	return game.TurnSummariesForFrontend(gameHandler.gameCollection, playerName), http.StatusOK
}

// handleNewGame adds a new game to the map of game state objects.
func (gameHandler *gameEndpointHandler) handleNewGame(
	httpBodyDecoder *json.Decoder,
	relevantSegments []string) (interface{}, int) {
	var gameDefinition endpoint.GameDefinition

	parsingError := httpBodyDecoder.Decode(&gameDefinition)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	addError :=
		game.AddNew(
			gameDefinition,
			gameHandler.gameCollection)

	if addError != nil {
		return addError, http.StatusBadRequest
	}

	gameIdentifier := gameHandler.segmentTranslator.ToSegment(gameDefinition.GameName)

	if strings.Contains(gameIdentifier, "/") {
		errorMessage := fmt.Sprintf(
			"Server set up with encoding which cannot convert %v to identifier with '/' in it",
			gameDefinition.GameName)
		return errorMessage, http.StatusBadRequest
	}

	return "OK", http.StatusOK
}

// writeGameForPlayer writes a JSON representation of the current state of the game
// with the given name for the player with the given name.
func (gameHandler *gameEndpointHandler) writeGameForPlayer(
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 2 {
		return "Not enough segments in URI to determine game name and player name", http.StatusBadRequest
	}

	gameIdentifier := relevantSegments[0]
	playerIdentifier := relevantSegments[1]

	gameState, isFound := game.ReadState(gameHandler.gameCollection, gameIdentifier)

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

	return game.ForPlayer(gameState, playerIdentifier), http.StatusOK
}

// handlePlayerAction passes on the given player action to the relevant game.
func (gameHandler *gameEndpointHandler) handlePlayerAction(
	httpBodyDecoder *json.Decoder,
	relevantSegments []string) (interface{}, int) {

	var playerAction endpoint.PlayerAction

	parsingError := httpBodyDecoder.Decode(&playerAction)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	actionError :=
		game.PerformAction(
			gameHandler.gameCollection,
			playerAction)

	if actionError != nil {
		return actionError, http.StatusBadRequest
	}

	return "OK", http.StatusOK
}
