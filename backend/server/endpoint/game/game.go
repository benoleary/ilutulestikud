package game

// This package is exported as game and yet also imports a different package as game.
// This is not a problem as imported package names are local to the file.
import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/server/endpoint/parsing"
)

// Handler is a struct meant to encapsulate all the state co-ordinating
// interaction with all the games through the endpoints.
// It implements the
// github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler interface.
type Handler struct {
	stateCollection   StateCollection
	segmentTranslator parsing.SegmentTranslator
}

// New returns a pointer to a new Handler.
func New(
	collectionOfStates StateCollection,
	translatorForSegments parsing.SegmentTranslator) *Handler {
	return &Handler{
		stateCollection:   collectionOfStates,
		segmentTranslator: translatorForSegments,
	}
}

// HandleGet parses an HTTP GET request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (handler *Handler) HandleGet(
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "available-rulesets":
		return handler.writeAvailableRulesets()
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
func (handler *Handler) HandlePost(
	httpBodyDecoder *json.Decoder,
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "create-new-game":
		return handler.handleNewGame(httpBodyDecoder, relevantSegments)
	case "record-chat-message":
		return handler.handleRecordChatMessage(httpBodyDecoder, relevantSegments)
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// writeAvailableRulesets writes a JSON object into the HTTP response which has
// the list of available rulesets as its "Rulesets" attribute.
func (handler *Handler) writeAvailableRulesets() (interface{}, int) {
	availableRulesetIdentifiers := game.ValidRulesetIdentifiers()

	selectableRulesets := make([]parsing.SelectableRuleset, 0)

	for _, rulesetIdentifier := range availableRulesetIdentifiers {
		// There definitely will not be an error from RulesetFromIdentifier if we
		// iterate only over the valid identifiers.
		availableRuleset, _ := game.RulesetFromIdentifier(rulesetIdentifier)
		selectableRuleset :=
			parsing.SelectableRuleset{
				Identifier:             rulesetIdentifier,
				Description:            availableRuleset.FrontendDescription(),
				MinimumNumberOfPlayers: availableRuleset.MinimumNumberOfPlayers(),
				MaximumNumberOfPlayers: availableRuleset.MaximumNumberOfPlayers(),
			}

		selectableRulesets = append(selectableRulesets, selectableRuleset)
	}

	endpointObject := parsing.RulesetList{
		Rulesets: selectableRulesets,
	}

	return endpointObject, http.StatusOK
}

// writeTurnSummariesForPlayer writes a JSON object into the HTTP response which has
// the list of turn summary objects as its "TurnSummaries" attribute.
func (handler *Handler) writeTurnSummariesForPlayer(
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine player", http.StatusBadRequest
	}

	playerIdentifier := relevantSegments[0]

	playerName, errorFromIdentification :=
		handler.segmentTranslator.FromSegment(playerIdentifier)

	if errorFromIdentification != nil {
		return errorFromIdentification, http.StatusBadRequest
	}

	allGamesWithPlayer, errorFromView :=
		handler.stateCollection.ViewAllWithPlayer(playerName)

	if errorFromView != nil {
		return errorFromView, http.StatusBadRequest
	}

	numberOfGamesWithPlayer := len(allGamesWithPlayer)

	turnSummaries := make([]parsing.TurnSummary, numberOfGamesWithPlayer)

	for gameIndex := 0; gameIndex < numberOfGamesWithPlayer; gameIndex++ {
		gameView := allGamesWithPlayer[gameIndex]
		_, isPlayerTurn := gameView.CurrentTurnOrder()
		turnSummaries[gameIndex] = parsing.TurnSummary{
			GameIdentifier: handler.segmentTranslator.ToSegment(gameView.GameName()),
			GameName:       gameView.GameName(),
			IsPlayerTurn:   isPlayerTurn,
		}
	}

	endpointObject := parsing.TurnSummaryList{
		TurnSummaries: turnSummaries,
	}

	return endpointObject, http.StatusOK
}

// handleNewGame adds a new game to the map of game state objects.
func (handler *Handler) handleNewGame(
	httpBodyDecoder *json.Decoder,
	relevantSegments []string) (interface{}, int) {
	var gameDefinition parsing.GameDefinition

	errorFromParse := httpBodyDecoder.Decode(&gameDefinition)
	if errorFromParse != nil {
		return "Error parsing JSON: " + errorFromParse.Error(), http.StatusBadRequest
	}

	gameRuleset, unknownRulesetError :=
		game.RulesetFromIdentifier(gameDefinition.RulesetIdentifier)
	if unknownRulesetError != nil {
		return unknownRulesetError, http.StatusBadRequest
	}

	errorFromAdd :=
		handler.stateCollection.AddNew(
			gameDefinition.GameName,
			gameRuleset,
			gameDefinition.PlayerNames)

	if errorFromAdd != nil {
		return errorFromAdd, http.StatusBadRequest
	}

	gameIdentifier := handler.segmentTranslator.ToSegment(gameDefinition.GameName)

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
func (handler *Handler) writeGameForPlayer(
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 2 {
		return "Not enough segments in URI to determine game name and player name", http.StatusBadRequest
	}

	gameIdentifier := relevantSegments[0]
	gameName, gameerrorFromIdentification :=
		handler.segmentTranslator.FromSegment(gameIdentifier)

	if gameerrorFromIdentification != nil {
		return gameerrorFromIdentification, http.StatusBadRequest
	}

	playerIdentifier := relevantSegments[1]
	playerName, playererrorFromIdentification :=
		handler.segmentTranslator.FromSegment(playerIdentifier)

	if playererrorFromIdentification != nil {
		return playererrorFromIdentification, http.StatusBadRequest
	}

	gameView, errorFromView :=
		handler.stateCollection.ViewState(gameName, playerName)

	if errorFromView != nil {
		return errorFromView, http.StatusBadRequest
	}

	chatMessages := gameView.SortedChatLog()
	numberOfChatMessages := len(chatMessages)
	chatLogForFrontend := make([]parsing.LogMessage, numberOfChatMessages)
	for messageIndex := 0; messageIndex < numberOfChatMessages; messageIndex++ {
		chatMessage := chatMessages[messageIndex]
		chatLogForFrontend[messageIndex] = parsing.LogMessage{
			TimestampInSeconds: chatMessage.CreationTime.Unix(),
			PlayerName:         chatMessage.PlayerName,
			TextColor:          chatMessage.TextColor,
			MessageText:        chatMessage.MessageText,
		}
	}

	actionMessages := gameView.SortedChatLog()
	numberOfActionMessages := len(actionMessages)
	actionLogForFrontend := make([]parsing.LogMessage, numberOfActionMessages)
	for messageIndex := 0; messageIndex < numberOfActionMessages; messageIndex++ {
		actionMessage := actionMessages[messageIndex]
		actionLogForFrontend[messageIndex] = parsing.LogMessage{
			TimestampInSeconds: actionMessage.CreationTime.Unix(),
			PlayerName:         actionMessage.PlayerName,
			TextColor:          actionMessage.TextColor,
			MessageText:        actionMessage.MessageText,
		}
	}

	endpointObject :=
		parsing.GameView{
			ChatLog:                      chatLogForFrontend,
			ActionLog:                    actionLogForFrontend,
			ScoreSoFar:                   gameView.Score(),
			NumberOfReadyHints:           gameView.NumberOfReadyHints(),
			NumberOfSpentHints:           gameView.NumberOfSpentHints(),
			NumberOfMistakesStillAllowed: gameView.NumberOfMistakesStillAllowed(),
			NumberOfMistakesMade:         gameView.NumberOfMistakesMade(),
		}

	return endpointObject, http.StatusOK
}

// handleRecordChatMessage passes on the given chat message to the relevant game.
func (handler *Handler) handleRecordChatMessage(
	httpBodyDecoder *json.Decoder,
	relevantSegments []string) (interface{}, int) {
	var playerChatMessage parsing.PlayerChatMessage

	errorFromParse := httpBodyDecoder.Decode(&playerChatMessage)
	if errorFromParse != nil {
		return "Error parsing JSON: " + errorFromParse.Error(), http.StatusBadRequest
	}

	actionExecutor, errorFromExecutor :=
		handler.stateCollection.ExecuteAction(
			playerChatMessage.GameName,
			playerChatMessage.PlayerName)

	if errorFromExecutor != nil {
		return errorFromExecutor, http.StatusBadRequest
	}

	actionExecutor.RecordChatMessage(playerChatMessage.ChatMessage)

	return "OK", http.StatusOK
}
