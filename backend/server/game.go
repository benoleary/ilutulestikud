package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/game"
)

// ByCreationTime implements sort interface for []game.ReadonlyState based on the return
// from its CreationTime().
type ByCreationTime []game.ReadonlyState

// gameEndpointHandler is a struct meant to encapsulate all the state co-ordinating
// interaction with all the games through the endpoints.
type gameEndpointHandler struct {
	stateCollection   gameCollection
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
	endpointObject := endpoint.RulesetList{
		Rulesets: game.AvailableRulesets(),
	}

	return endpointObject, http.StatusOK
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

	return gameHandler.stateCollection.TurnSummariesForFrontend(playerName), http.StatusOK
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
		gameHandler.stateCollection.AddNew(
			gameDefinition)

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

	gameState, isFound := gameHandler.stateCollection.ReadState(gameIdentifier)

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

	endpointObject :=
		endpoint.GameView{
			ChatLog:                      gameState.ChatLog().ForFrontend(),
			ScoreSoFar:                   gameState.Score(),
			NumberOfReadyHints:           gameState.NumberOfReadyHints(),
			NumberOfSpentHints:           MaximumNumberOfHints - gameState.NumberOfReadyHints(),
			NumberOfMistakesStillAllowed: MaximumNumberOfMistakesAllowed - gameState.NumberOfMistakesMade(),
			NumberOfMistakesMade:         gameState.NumberOfMistakesMade(),
		}

	return endpointObject, http.StatusOK
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
		gameHandler.stateCollection.PerformAction(
			playerAction)

	if actionError != nil {
		return actionError, http.StatusBadRequest
	}

	return "OK", http.StatusOK
}

// TurnSummariesForFrontend writes the turn summary information for each game which has
// the given player into the relevant JSON object for the frontend.
func (gameCollection *StateCollection) TurnSummariesForFrontend(playerName string) endpoint.TurnSummaryList {
	gameList := gameCollection.statePersister.readAllWithPlayer(playerName)

	sort.Sort(ByCreationTime(gameList))

	numberOfGamesWithPlayer := len(gameList)

	turnSummaries := make([]endpoint.TurnSummary, numberOfGamesWithPlayer)
	for gameIndex := 0; gameIndex < numberOfGamesWithPlayer; gameIndex++ {
		nameOfGame := gameList[gameIndex].Name()
		gameTurn := gameList[gameIndex].Turn()

		gameParticipants := gameList[gameIndex].Players()
		numberOfParticipants := len(gameParticipants)

		playerNamesInTurnOrder := make([]string, numberOfParticipants)

		turnsUntilPlayer := 0
		for playerIndex := 0; playerIndex < numberOfParticipants; playerIndex++ {
			// Game turns begin with 1 rather than 0, so this sets the player names in order,
			// wrapping index back to 0 when at the end of the list.
			// E.g. turn 3, 5 players: playerNamesInTurnOrder will start with
			// gameParticipants[2], then [3], then [4], then [0], then [1].
			playerInTurnOrder :=
				gameParticipants[(playerIndex+gameTurn-1)%numberOfParticipants]
			playerNamesInTurnOrder[playerIndex] =
				playerInTurnOrder.Name()

			if playerName == playerInTurnOrder.Name() {
				turnsUntilPlayer = playerIndex
			}
		}

		turnSummaries[gameIndex] = endpoint.TurnSummary{
			GameIdentifier:             gameList[gameIndex].Identifier(),
			GameName:                   nameOfGame,
			RulesetDescription:         gameList[gameIndex].Ruleset().FrontendDescription(),
			CreationTimestampInSeconds: gameList[gameIndex].CreationTime().Unix(),
			TurnNumber:                 gameTurn,
			PlayerNamesInNextTurnOrder: playerNamesInTurnOrder,
			IsPlayerTurn:               turnsUntilPlayer == 0,
		}
	}

	return endpoint.TurnSummaryList{TurnSummaries: turnSummaries}
}

// Len implements part of the sort interface for ByCreationTime.
func (byCreationTime ByCreationTime) Len() int {
	return len(byCreationTime)
}

// Swap implements part of the sort interface for ByCreationTime.
func (byCreationTime ByCreationTime) Swap(firstIndex int, secondIndex int) {
	byCreationTime[firstIndex], byCreationTime[secondIndex] =
		byCreationTime[secondIndex], byCreationTime[firstIndex]
}

// Less implements part of the sort interface for ByCreationTime.
func (byCreationTime ByCreationTime) Less(firstIndex int, secondIndex int) bool {
	return byCreationTime[firstIndex].CreationTime().Before(
		byCreationTime[secondIndex].CreationTime())
}
