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
	case "record-chat-message":
		return gameHandler.handleRecordChatMessage(httpBodyDecoder, relevantSegments)
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// writeAvailableRulesets writes a JSON object into the HTTP response which has
// the list of available rulesets as its "Rulesets" attribute.
func (gameHandler *gameEndpointHandler) writeAvailableRulesets() (interface{}, int) {
	availableRulesetIdentifiers := game.ValidRulesetIdentifiers()

	selectableRulesets := make([]endpoint.SelectableRuleset, 0)

	for _, rulesetIdentifier := range availableRulesetIdentifiers {
		// There definitely will not be an error from RulesetFromIdentifier if we
		// iterate only over the valid identifiers.
		availableRuleset, _ := game.RulesetFromIdentifier(rulesetIdentifier)
		selectableRuleset :=
			endpoint.SelectableRuleset{
				Identifier:             rulesetIdentifier,
				Description:            availableRuleset.FrontendDescription(),
				MinimumNumberOfPlayers: availableRuleset.MinimumNumberOfPlayers(),
				MaximumNumberOfPlayers: availableRuleset.MaximumNumberOfPlayers(),
			}

		selectableRulesets = append(selectableRulesets, selectableRuleset)
	}

	endpointObject := endpoint.RulesetList{
		Rulesets: selectableRulesets,
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

	allGamesWithPlayer, viewError :=
		gameHandler.stateCollection.ViewAllWithPlayer(playerName)

	if viewError != nil {
		return viewError, http.StatusBadRequest
	}

	numberOfGamesWithPlayer := len(allGamesWithPlayer)

	turnSummaries := make([]endpoint.TurnSummary, numberOfGamesWithPlayer)

	for gameIndex := 0; gameIndex < numberOfGamesWithPlayer; gameIndex++ {
		gameView := allGamesWithPlayer[gameIndex]
		_, isPlayerTurn := gameView.CurrentTurnOrder()
		turnSummaries[gameIndex] = endpoint.TurnSummary{
			GameIdentifier: gameHandler.segmentTranslator.ToSegment(gameView.GameName()),
			GameName:       gameView.GameName(),
			IsPlayerTurn:   isPlayerTurn,
		}
	}

	endpointObject := endpoint.TurnSummaryList{
		TurnSummaries: turnSummaries,
	}

	return endpointObject, http.StatusOK
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
	gameName, gameIdentificationError :=
		gameHandler.segmentTranslator.FromSegment(gameIdentifier)

	if gameIdentificationError != nil {
		return gameIdentificationError, http.StatusBadRequest
	}

	playerIdentifier := relevantSegments[1]
	playerName, playerIdentificationError :=
		gameHandler.segmentTranslator.FromSegment(playerIdentifier)

	if playerIdentificationError != nil {
		return playerIdentificationError, http.StatusBadRequest
	}

	gameView, viewError :=
		gameHandler.stateCollection.ViewState(gameName, playerName)

	if viewError != nil {
		return viewError, http.StatusBadRequest
	}

	chatMessages := gameView.ChatLog().Sorted()
	numberOfMessages := len(chatMessages)
	chatLogForFrontend := make([]endpoint.ChatLogMessage, numberOfMessages)
	for messageIndex := 0; messageIndex < numberOfMessages; messageIndex++ {
		chatMessage := chatMessages[messageIndex]
		chatLogForFrontend[messageIndex] = endpoint.ChatLogMessage{
			TimestampInSeconds: chatMessage.CreationTime.Unix(),
			PlayerName:         chatMessage.PlayerName,
			ChatColor:          chatMessage.ChatColor,
			MessageText:        chatMessage.MessageText,
		}
	}

	endpointObject :=
		endpoint.GameView{
			ChatLog:                      chatLogForFrontend,
			ScoreSoFar:                   gameView.Score(),
			NumberOfReadyHints:           gameView.NumberOfReadyHints(),
			NumberOfSpentHints:           gameView.NumberOfSpentHints(),
			NumberOfMistakesStillAllowed: gameView.NumberOfMistakesStillAllowed(),
			NumberOfMistakesMade:         gameView.NumberOfMistakesMade(),
		}

	return endpointObject, http.StatusOK
}

// handleRecordChatMessage passes on the given chat message to the relevant game.
func (gameHandler *gameEndpointHandler) handleRecordChatMessage(
	httpBodyDecoder *json.Decoder,
	relevantSegments []string) (interface{}, int) {
	var playerChatMessage endpoint.PlayerChatMessage

	parsingError := httpBodyDecoder.Decode(&playerChatMessage)
	if parsingError != nil {
		return "Error parsing JSON: " + parsingError.Error(), http.StatusBadRequest
	}

	actionError :=
		gameHandler.stateCollection.RecordChatMessage(
			playerChatMessage.GameName,
			playerChatMessage.PlayerName,
			playerChatMessage.ChatMessage)

	if actionError != nil {
		return actionError, http.StatusBadRequest
	}

	return "OK", http.StatusOK
}
