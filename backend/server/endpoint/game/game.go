package game

// This package is exported as game and yet also imports a different package as game.
// This is not a problem as imported package names are local to the file.
import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
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
		return handler.handleNewGame(httpBodyDecoder)
	case "record-chat-message":
		return handler.handleRecordChatMessage(httpBodyDecoder)
	case "take-turn-by-discarding":
		return handler.handleTakeTurnByDiscarding(httpBodyDecoder)
	case "take-turn-by-attempting-to-play":
		return handler.handleTakeTurnByAttemptingToPlay(httpBodyDecoder)
	case "take-turn-by-hinting-color":
		return handler.handleTakeTurnByHintingColor(httpBodyDecoder)
	case "take-turn-by-hinting-number":
		return handler.handleTakeTurnByHintingNumber(httpBodyDecoder)
	case "leave-game":
		return handler.handleLeaveGame(httpBodyDecoder)
	case "delete-game":
		return handler.handleDeleteGame(httpBodyDecoder)
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
		_, playerTurnIndex := gameView.CurrentTurnOrder()
		turnSummaries[gameIndex] = parsing.TurnSummary{
			GameIdentifier: handler.segmentTranslator.ToSegment(gameView.GameName()),
			GameName:       gameView.GameName(),
			IsPlayerTurn:   playerTurnIndex == 0,
		}
	}

	endpointObject := parsing.TurnSummaryList{
		TurnSummaries: turnSummaries,
	}

	return endpointObject, http.StatusOK
}

// handleNewGame adds a new game to the map of game state objects.
func (handler *Handler) handleNewGame(
	httpBodyDecoder *json.Decoder) (interface{}, int) {
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
	gameName, playerName, errorFromParsing := handler.parseGameAndPlayer(relevantSegments)
	if errorFromParsing != nil {
		return errorFromParsing, http.StatusBadRequest
	}

	gameView, errorFromView := handler.stateCollection.ViewState(gameName, playerName)
	if errorFromView != nil {
		return errorFromView, http.StatusInternalServerError
	}

	handsBeforeThisPlayer, handsAfterThisPlayer, isViewingPlayerTurn, errorFromVisibleHands :=
		handler.visibleHandsBeforeAndAfter(gameView)
	if errorFromVisibleHands != nil {
		return errorFromVisibleHands, http.StatusInternalServerError
	}

	handOfThisPlayer, errorFromInferredHand :=
		handler.playerKnowledgeOfHand(
			gameView,
			playerName)
	if errorFromInferredHand != nil {
		return errorFromInferredHand, http.StatusInternalServerError
	}

	gameIsFinished := gameView.GameIsFinished()

	endpointObject :=
		parsing.GameView{
			ChatLog:                            handler.logForFrontend(gameView.ChatLog()),
			ActionLog:                          handler.logForFrontend(gameView.ActionLog()),
			GameIsFinished:                     gameView.GameIsFinished(),
			ScoreSoFar:                         gameView.Score(),
			NumberOfReadyHints:                 gameView.NumberOfReadyHints(),
			MaximumNumberOfHints:               gameView.MaximumNumberOfHints(),
			HintColorSuits:                     gameView.ColorsAvailableAsHint(),
			HintSequenceIndices:                gameView.IndicesAvailableAsHint(),
			NumberOfMistakesMade:               gameView.NumberOfMistakesMade(),
			NumberOfMistakesIndicatingGameOver: gameView.NumberOfMistakesIndicatingGameOver(),
			NumberOfCardsLeftInDeck:            gameView.DeckSize(),
			PlayedCards:                        playedCards(gameView.PlayedCards()),
			DiscardedCards:                     cardsForFrontend(gameView.DiscardedCards()),
			HandsBeforeThisPlayer:              handsBeforeThisPlayer,
			HandOfThisPlayer:                   handOfThisPlayer,
			HandsAfterThisPlayer:               handsAfterThisPlayer,
			ThisPlayerCanTakeTurn:              !gameIsFinished && isViewingPlayerTurn,
		}

	return endpointObject, http.StatusOK
}

// handleRecordChatMessage passes on the given chat message to the relevant game.
func (handler *Handler) handleRecordChatMessage(
	httpBodyDecoder *json.Decoder) (interface{}, int) {
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

	errorFromRecordChatMessage :=
		actionExecutor.RecordChatMessage(playerChatMessage.ChatMessage)

	if errorFromRecordChatMessage != nil {
		return errorFromRecordChatMessage, http.StatusInternalServerError
	}

	return "OK", http.StatusOK
}

// handleTakeTurnByDiscarding passes on the given parameters of taking a turn by
// discarding a card to the relevant game.
func (handler *Handler) handleTakeTurnByDiscarding(
	httpBodyDecoder *json.Decoder) (interface{}, int) {
	var playerCardIndication parsing.PlayerCardIndication

	errorFromParse := httpBodyDecoder.Decode(&playerCardIndication)
	if errorFromParse != nil {
		return "Error parsing JSON: " + errorFromParse.Error(), http.StatusBadRequest
	}

	actionExecutor, errorFromExecutor :=
		handler.stateCollection.ExecuteAction(
			playerCardIndication.GameName,
			playerCardIndication.PlayerName)

	if errorFromExecutor != nil {
		return errorFromExecutor, http.StatusBadRequest
	}

	errorFromTakeTurnByDiscarding :=
		actionExecutor.TakeTurnByDiscarding(playerCardIndication.CardIndex)

	if errorFromTakeTurnByDiscarding != nil {
		return errorFromTakeTurnByDiscarding, http.StatusBadRequest
	}

	return "OK", http.StatusOK
}

// handleTakeTurnByAttemptingToPlay passes on the given parameters of taking a turn
// by attempting to play a card to the relevant game.
func (handler *Handler) handleTakeTurnByAttemptingToPlay(
	httpBodyDecoder *json.Decoder) (interface{}, int) {
	var playerCardIndication parsing.PlayerCardIndication

	errorFromParse := httpBodyDecoder.Decode(&playerCardIndication)
	if errorFromParse != nil {
		return "Error parsing JSON: " + errorFromParse.Error(), http.StatusBadRequest
	}

	actionExecutor, errorFromExecutor :=
		handler.stateCollection.ExecuteAction(
			playerCardIndication.GameName,
			playerCardIndication.PlayerName)

	if errorFromExecutor != nil {
		return errorFromExecutor, http.StatusBadRequest
	}

	errorFromTakeTurnByPlaying :=
		actionExecutor.TakeTurnByPlaying(playerCardIndication.CardIndex)

	if errorFromTakeTurnByPlaying != nil {
		return errorFromTakeTurnByPlaying, http.StatusBadRequest
	}

	return "OK", http.StatusOK
}

// handleTakeTurnByHintingColor passes on the given parameters of taking a turn
// by hinting about color to the relevant game.
func (handler *Handler) handleTakeTurnByHintingColor(
	httpBodyDecoder *json.Decoder) (interface{}, int) {
	var playerColorHint parsing.PlayerColorHint

	errorFromParse := httpBodyDecoder.Decode(&playerColorHint)
	if errorFromParse != nil {
		return "Error parsing JSON: " + errorFromParse.Error(), http.StatusBadRequest
	}

	actionExecutor, errorFromExecutor :=
		handler.stateCollection.ExecuteAction(
			playerColorHint.GameName,
			playerColorHint.PlayerName)

	if errorFromExecutor != nil {
		return errorFromExecutor, http.StatusBadRequest
	}

	errorFromTakeTurnByHinting :=
		actionExecutor.TakeTurnByHintingColor(
			playerColorHint.ReceiverName,
			playerColorHint.HintedColor)

	if errorFromTakeTurnByHinting != nil {
		return errorFromTakeTurnByHinting, http.StatusBadRequest
	}

	return "OK", http.StatusOK
}

// handleTakeTurnByHintingNumber passes on the given parameters of taking a turn
// by hinting about index to the relevant game.
func (handler *Handler) handleTakeTurnByHintingNumber(
	httpBodyDecoder *json.Decoder) (interface{}, int) {
	var playerIndexHint parsing.PlayerIndexHint

	errorFromParse := httpBodyDecoder.Decode(&playerIndexHint)
	if errorFromParse != nil {
		return "Error parsing JSON: " + errorFromParse.Error(), http.StatusBadRequest
	}

	actionExecutor, errorFromExecutor :=
		handler.stateCollection.ExecuteAction(
			playerIndexHint.GameName,
			playerIndexHint.PlayerName)

	if errorFromExecutor != nil {
		return errorFromExecutor, http.StatusBadRequest
	}

	errorFromTakeTurnByHinting :=
		actionExecutor.TakeTurnByHintingIndex(
			playerIndexHint.ReceiverName,
			playerIndexHint.HintedNumber)

	if errorFromTakeTurnByHinting != nil {
		return errorFromTakeTurnByHinting, http.StatusBadRequest
	}

	return "OK", http.StatusOK
}

// handleLeaveGame passes on the given game name and player name to the collection so that
// the game can be removed from the list of games which is given for the player.
func (handler *Handler) handleLeaveGame(httpBodyDecoder *json.Decoder) (interface{}, int) {
	var leavingInformation parsing.PlayerInGameIndication

	errorFromParse := httpBodyDecoder.Decode(&leavingInformation)
	if errorFromParse != nil {
		return "Error parsing JSON: " + errorFromParse.Error(), http.StatusBadRequest
	}

	errorFromLeaving :=
		handler.stateCollection.RemoveGameFromListForPlayer(
			leavingInformation.GameName,
			leavingInformation.PlayerName)
	if errorFromLeaving != nil {
		return errorFromLeaving, http.StatusInternalServerError
	}

	return "OK", http.StatusOK
}

// handleDeleteGame passes on the given game name to the collection so that the game can
// be deleted.
func (handler *Handler) handleDeleteGame(httpBodyDecoder *json.Decoder) (interface{}, int) {
	var gameToDelete parsing.GameDefinition

	errorFromParse := httpBodyDecoder.Decode(&gameToDelete)
	if errorFromParse != nil {
		return "Error parsing JSON: " + errorFromParse.Error(), http.StatusBadRequest
	}

	errorFromDeletion := handler.stateCollection.Delete(gameToDelete.GameName)
	if errorFromDeletion != nil {
		return errorFromDeletion, http.StatusInternalServerError
	}

	return "OK", http.StatusOK
}

func (handler *Handler) parseGameAndPlayer(
	relevantSegments []string) (string, string, error) {
	if len(relevantSegments) < 2 {
		errorBecauseNotEnoughSegments :=
			fmt.Errorf("Not enough segments in URI to determine game name and player name")
		return "error", "error", errorBecauseNotEnoughSegments
	}

	gameIdentifier := relevantSegments[0]
	gameName, errorFromGameIdentification :=
		handler.segmentTranslator.FromSegment(gameIdentifier)

	if errorFromGameIdentification != nil {
		return "error", "error", errorFromGameIdentification
	}

	playerIdentifier := relevantSegments[1]
	playerName, errorFromPlayerIdentification :=
		handler.segmentTranslator.FromSegment(playerIdentifier)

	if errorFromPlayerIdentification != nil {
		return "error", "error", errorFromPlayerIdentification
	}

	return gameName, playerName, nil
}

func (handler *Handler) logForFrontend(
	backendMessages []message.Readonly) []parsing.LogMessage {
	numberOfMessages := len(backendMessages)
	logForFrontend := make([]parsing.LogMessage, numberOfMessages)
	for messageIndex := 0; messageIndex < numberOfMessages; messageIndex++ {
		backendMessage := backendMessages[messageIndex]
		logForFrontend[messageIndex] = parsing.LogMessage{
			TimestampInSeconds: backendMessage.CreationTime().Unix(),
			PlayerName:         backendMessage.PlayerName(),
			TextColor:          backendMessage.TextColor(),
			MessageText:        backendMessage.MessageText(),
		}
	}

	return logForFrontend
}

func (handler *Handler) visibleHandsBeforeAndAfter(
	gameView game.ViewForPlayer) ([]parsing.VisibleHand, []parsing.VisibleHand, bool, error) {
	playersInTurnOrder, playerIndexInTurnOrder := gameView.CurrentTurnOrder()
	numberOfPlayers := len(playersInTurnOrder)

	allVisibleHands := make([]parsing.VisibleHand, numberOfPlayers-1)

	for playerIndex := 0; playerIndex < numberOfPlayers; playerIndex++ {
		if playerIndex != playerIndexInTurnOrder {
			playerWithVisibleHand := playersInTurnOrder[playerIndex]
			visibleHandFromView, playerChatColor, errorFromVisibleHand :=
				gameView.VisibleHand(playerWithVisibleHand)
			if errorFromVisibleHand != nil {
				return nil, nil, false, errorFromVisibleHand
			}

			numberOfCardsInHand := len(visibleHandFromView)
			handCards := make([]parsing.VisibleCard, numberOfCardsInHand)
			for cardIndex := 0; cardIndex < numberOfCardsInHand; cardIndex++ {
				visibleCard := visibleHandFromView[cardIndex]
				handCards[cardIndex] = parsing.VisibleCard{
					ColorSuit:     visibleCard.ColorSuit(),
					SequenceIndex: visibleCard.SequenceIndex(),
				}
			}

			knowledgeOfOwnHand, errorFromInferredHand :=
				handler.playerKnowledgeOfHand(
					gameView,
					playersInTurnOrder[playerIndex])
			if errorFromInferredHand != nil {
				return nil, nil, false, errorFromInferredHand
			}

			visibleHandForFrontend :=
				parsing.VisibleHand{
					PlayerName:         playerWithVisibleHand,
					PlayerColor:        playerChatColor,
					HandCards:          handCards,
					KnowledgeOfOwnHand: knowledgeOfOwnHand,
				}

			if playerIndex < playerIndexInTurnOrder {
				allVisibleHands[playerIndex] = visibleHandForFrontend
			} else {
				allVisibleHands[playerIndex-1] = visibleHandForFrontend
			}
		}
	}

	handsBeforeViewingPlayer := allVisibleHands[:playerIndexInTurnOrder]
	handsAfterViewingPlayer := allVisibleHands[playerIndexInTurnOrder:]

	return handsBeforeViewingPlayer, handsAfterViewingPlayer, playerIndexInTurnOrder == 0, nil
}

func (handler *Handler) playerKnowledgeOfHand(
	gameView game.ViewForPlayer,
	holdingPlayer string) ([]parsing.CardFromBehind, error) {
	inferredHandFromView, errorFromInferredHand :=
		gameView.KnowledgeOfOwnHand(holdingPlayer)
	if errorFromInferredHand != nil {
		return nil, errorFromInferredHand
	}

	numberOfCardsInHand := len(inferredHandFromView)
	handFromBehind := make([]parsing.CardFromBehind, numberOfCardsInHand)
	for cardIndex := 0; cardIndex < numberOfCardsInHand; cardIndex++ {
		inferredCard := inferredHandFromView[cardIndex]
		handFromBehind[cardIndex] =
			parsing.CardFromBehind{
				PossibleColorSuits:      inferredCard.PossibleColors(),
				PossibleSequenceIndices: inferredCard.PossibleIndices(),
			}
	}

	return handFromBehind, nil
}

func playedCards(playedPilesFromView [][]card.Readonly) [][]parsing.VisibleCard {
	numberOfPiles := len(playedPilesFromView)

	playedPilesForFrontend := make([][]parsing.VisibleCard, numberOfPiles)

	for pileIndex := 0; pileIndex < numberOfPiles; pileIndex++ {
		playedPilesForFrontend[pileIndex] =
			cardsForFrontend(playedPilesFromView[pileIndex])
	}

	return playedPilesForFrontend
}

func cardsForFrontend(backendCards []card.Readonly) []parsing.VisibleCard {
	numberOfCards := len(backendCards)

	forFrontend := make([]parsing.VisibleCard, numberOfCards)

	for cardIndex := 0; cardIndex < numberOfCards; cardIndex++ {
		backendCard := backendCards[cardIndex]

		forFrontend[cardIndex] = parsing.VisibleCard{
			ColorSuit:     backendCard.ColorSuit(),
			SequenceIndex: backendCard.SequenceIndex(),
		}
	}

	return forFrontend
}
