package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/player"
)

type httpGetAndPostHandler interface {
	HandleGet(relevantSegments []string) (interface{}, int)

	HandlePost(httpBodyDecoder *json.Decoder, relevantSegments []string) (interface{}, int)
}

type playerCollection interface {
	// Add should add a new player to the collection, defined by the given arguments.
	Add(playerName string, chatColor string) error

	// UpdateColor should update the given player with the given chat color.
	UpdateColor(playerName string, chatColor string) error

	// Get should return a read-only state for the identified player.
	Get(playerIdentifier string) (player.ReadonlyState, error)

	// Reset should reset the players to the initial set.
	Reset()

	// All should return a slice of all the players in the collection. The order is not
	// mandated, and may even change with repeated calls to the same unchanged collection
	// (analogously to the entry set of a standard Golang map, for example), though of
	// course an implementation may order the slice consistently.
	All() []player.ReadonlyState

	// AvailableChatColors should return the chat colors available to the collection.
	AvailableChatColors() []string
}

type gameCollection interface {
	// ViewState should return a view around the read-only game state corresponding
	// to the given name as seen by the given player. If the game does not exist or
	// the player is not a participant, it should return an error.
	ViewState(gameName string, playerName string) (*game.PlayerView, error)

	// ViewAllWithPlayer should return a slice of read-only views on all the games in the
	// collection which have the given player as a participant. It should return an
	// error if there is a problem wrapping any of the read-only game states in a view.
	// The order is not mandated, and may even change with repeated calls to the same
	// unchanged collection (analogously to the entry set of a standard Golang map, for
	// example), though of course an implementation may order the slice consistently.
	ViewAllWithPlayer(playerName string) ([]*game.PlayerView, error)

	// RecordChatMessage should find the given game and record the given chat message
	// from the given player, or return an error.
	RecordChatMessage(
		gameName string,
		playerName string,
		chatMessage string) error

	// AddNew should add a new game to the collection based on the given arguments.
	AddNew(
		gameName string,
		gameRuleset game.Ruleset,
		playerNames []string) error
}

// State contains all the state to allow the backend to function.
type State struct {
	accessControlAllowedOrigin string
	playerHandler              *playerEndpointHandler
	gameHandler                *gameEndpointHandler
}

// New creates a new State object and returns a pointer to it, assuming that the
// given handlers are consistent.
func New(
	accessControlAllowedOrigin string,
	segmentTranslator EndpointSegmentTranslator,
	playerStateCollection playerCollection,
	gameStateCollection gameCollection) *State {
	return &State{
		accessControlAllowedOrigin: accessControlAllowedOrigin,
		playerHandler: &playerEndpointHandler{
			stateCollection:   playerStateCollection,
			segmentTranslator: segmentTranslator,
		},
		gameHandler: &gameEndpointHandler{
			stateCollection:   gameStateCollection,
			segmentTranslator: segmentTranslator,
		},
	}
}

// HandleBackend calls functions according to the second segment of the URI, assuming
// that the first segment is "backend".
func (state *State) HandleBackend(
	httpResponseWriter http.ResponseWriter,
	httpRequest *http.Request) {
	// If an allowed origin for access control has been set, we set all the headers to allow it.
	if state.accessControlAllowedOrigin != "" {
		httpResponseWriter.Header().Set("Access-Control-Allow-Origin", state.accessControlAllowedOrigin)
		httpResponseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		httpResponseWriter.Header().Set(
			"Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}

	// There should always be an initial "/", but unless it is present, with at least one character
	// following it as the first URI segment, we do not process the request.
	if (len(httpRequest.URL.Path) < 2) || (httpRequest.URL.Path[0] != '/') {
		http.NotFound(httpResponseWriter, httpRequest)
		return
	}

	pathSegments := parsePathSegments(httpRequest)

	// There is no default if there is no URI segment or not enough segments.
	if (pathSegments == nil) || (len(pathSegments) < 2) {
		http.NotFound(httpResponseWriter, httpRequest)
		return
	}

	// We choose the interface which will handle the GET or POST based on the
	// first segment of the URI after "backend".
	var requestHandler httpGetAndPostHandler
	switch pathSegments[1] {
	case "player":
		requestHandler = state.playerHandler
	case "game":
		requestHandler = state.gameHandler
	default:
		http.NotFound(httpResponseWriter, httpRequest)
		return
	}

	var objectForBody interface{}
	var httpStatus int

	switch httpRequest.Method {
	case http.MethodOptions:
		return
	case http.MethodGet:
		objectForBody, httpStatus = requestHandler.HandleGet(pathSegments[2:])
	case http.MethodPost:
		{
			if httpRequest.Body == nil {
				http.Error(httpResponseWriter, "Empty request body", http.StatusBadRequest)
				return
			}

			objectForBody, httpStatus =
				requestHandler.HandlePost(json.NewDecoder(httpRequest.Body), pathSegments[2:])
		}
	default:
		http.Error(httpResponseWriter, "Method not GET or POST: "+httpRequest.Method, http.StatusBadRequest)
	}

	// If the status is OK, writing the header with OK won't make any difference.
	httpResponseWriter.WriteHeader(httpStatus)
	json.NewEncoder(httpResponseWriter).Encode(objectForBody)
}

// parsePathSegments returns the segments of the URI path as a slice of a string array.
func parsePathSegments(httpRequest *http.Request) []string {
	// The initial character is '/' so we skip it to avoid an empty string as the first element.
	return strings.Split(httpRequest.URL.Path[1:], "/")
}
