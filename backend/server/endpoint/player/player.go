package player

// This package is exported as player and yet also imports a different package as player.
// This is not a problem as imported package names are local to the file.
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/benoleary/ilutulestikud/backend/server/endpoint/parsing"
)

// Handler is a struct meant to encapsulate all the state making the player states
// available to the endpoints.
// It implements github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
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
	requestContext context.Context,
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "registered-players":
		return handler.writeRegisteredPlayers(requestContext)
	case "available-colors":
		return handler.writeAvailableColors(requestContext)
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// HandlePost parses an HTTP POST request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (handler *Handler) HandlePost(
	requestContext context.Context,
	httpBodyDecoder *json.Decoder,
	relevantSegments []string) (interface{}, int) {
	if len(relevantSegments) < 1 {
		return "Not enough segments in URI to determine what to do", http.StatusBadRequest
	}

	switch relevantSegments[0] {
	case "new-player":
		return handler.handleNewPlayer(requestContext, httpBodyDecoder)
	case "update-player":
		return handler.handleUpdatePlayer(requestContext, httpBodyDecoder)
	case "delete-player":
		return handler.handleDeletePlayer(requestContext, httpBodyDecoder)
	default:
		return "URI segment " + relevantSegments[0] + " not valid", http.StatusNotFound
	}
}

// writeRegisteredPlayers writes a JSON object into the HTTP response which has
// the list of player objects as its "Players" attribute. The order of the players
// may not consistent with repeated calls as ForEndpoint does not guarantee it.
func (handler *Handler) writeRegisteredPlayers(
	requestContext context.Context) (interface{}, int) {
	playerStates, errorFromAll := handler.stateCollection.All(requestContext)
	if errorFromAll != nil {
		return errorFromAll, http.StatusInternalServerError
	}

	playerList := make([]parsing.PlayerState, 0, len(playerStates))
	for _, playerState := range playerStates {
		playerName := playerState.Name()
		playerList = append(playerList, parsing.PlayerState{
			Identifier: handler.segmentTranslator.ToSegment(playerName),
			Name:       playerName,
			Color:      playerState.Color(),
		})
	}

	endpointObject := parsing.PlayerList{
		Players: playerList,
	}

	return endpointObject, http.StatusOK
}

// writeAvailableColors writes a JSON object into the HTTP response which has
// the list of strings as its "Colors" attribute.
func (handler *Handler) writeAvailableColors(
	requestContext context.Context) (interface{}, int) {
	endpointObject := parsing.ChatColorList{
		Colors: handler.stateCollection.AvailableChatColors(requestContext),
	}

	return endpointObject, http.StatusOK
}

// handleNewPlayer adds the player defined by the JSON of the request's body to the list
// of registered players, and returns the updated list as writeRegisteredPlayerNameListJson
// would.
func (handler *Handler) handleNewPlayer(
	requestContext context.Context,
	httpBodyDecoder *json.Decoder) (interface{}, int) {
	var endpointPlayer parsing.PlayerState
	errorFromParse := httpBodyDecoder.Decode(&endpointPlayer)
	if errorFromParse != nil {
		return "Error parsing JSON: " + errorFromParse.Error(), http.StatusBadRequest
	}

	errorFromAdd :=
		handler.stateCollection.Add(requestContext, endpointPlayer.Name, endpointPlayer.Color)

	if errorFromAdd != nil {
		return errorFromAdd, http.StatusBadRequest
	}

	playerIdentifier := handler.segmentTranslator.ToSegment(endpointPlayer.Name)

	if strings.Contains(playerIdentifier, "/") {
		errorMessage := fmt.Sprintf(
			"Server set up with encoding which cannot convert %v to identifier with '/' in it",
			endpointPlayer.Name)
		return errorMessage, http.StatusBadRequest
	}

	return handler.writeRegisteredPlayers(requestContext)
}

// handleUpdatePlayer updates the player defined by the JSON of the request's body, taking
// the "Name" attribute as the key, and returns the updated list as writeRegisteredPlayers
// would. Attributes which are present are updated, those which are missing remain unchanged.
func (handler *Handler) handleUpdatePlayer(
	requestContext context.Context,
	httpBodyDecoder *json.Decoder) (interface{}, int) {
	var playerUpdate parsing.PlayerState
	errorFromParse := httpBodyDecoder.Decode(&playerUpdate)
	if errorFromParse != nil {
		return "Error parsing JSON: " + errorFromParse.Error(), http.StatusBadRequest
	}

	updateError :=
		handler.stateCollection.UpdateColor(
			requestContext,
			playerUpdate.Name,
			playerUpdate.Color)

	if updateError != nil {
		return updateError, http.StatusBadRequest
	}

	return handler.writeRegisteredPlayers(requestContext)
}

// handleUpdatePlayer updates the player defined by the JSON of the request's body, taking
// the "Name" attribute as the key, and returns the updated list as writeRegisteredPlayers
// would. Attributes which are present are updated, those which are missing remain unchanged.
func (handler *Handler) handleDeletePlayer(
	requestContext context.Context,
	httpBodyDecoder *json.Decoder) (interface{}, int) {
	var playerToDelete parsing.PlayerState
	errorFromParse := httpBodyDecoder.Decode(&playerToDelete)
	if errorFromParse != nil {
		return "Error parsing JSON: " + errorFromParse.Error(), http.StatusBadRequest
	}

	deleteError :=
		handler.stateCollection.Delete(requestContext, playerToDelete.Name)
	if deleteError != nil {
		return deleteError, http.StatusInternalServerError
	}

	return handler.writeRegisteredPlayers(requestContext)
}
