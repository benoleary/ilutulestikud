package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/benoleary/ilutulestikud/backend/server/endpoint/game"
	"github.com/benoleary/ilutulestikud/backend/server/endpoint/parsing"
	"github.com/benoleary/ilutulestikud/backend/server/endpoint/player"
)

// State contains all the state to allow the backend to function.
type State struct {
	contextProvider            ContextProvider
	accessControlAllowedOrigin string
	backendVersion             string
	playerHandler              httpGetAndPostHandler
	gameHandler                httpGetAndPostHandler
}

// New creates a new State object with handlers built around the given
// state collections.
func New(
	contextProvider ContextProvider,
	accessControlAllowedOrigin string,
	backendVersion string,
	segmentTranslator parsing.SegmentTranslator,
	playerStateCollection player.StateCollection,
	gameStateCollection game.StateCollection) *State {
	return NewWithGivenHandlers(
		contextProvider,
		accessControlAllowedOrigin,
		backendVersion,
		segmentTranslator,
		player.New(playerStateCollection, segmentTranslator),
		game.New(gameStateCollection, segmentTranslator))
}

// NewWithGivenHandlers creates a new State object and returns a pointer to it,
// assuming that the given handlers are consistent.
func NewWithGivenHandlers(
	contextProvider ContextProvider,
	accessControlAllowedOrigin string,
	backendVersion string,
	segmentTranslator parsing.SegmentTranslator,
	handlerForPlayer httpGetAndPostHandler,
	handlerForGame httpGetAndPostHandler) *State {
	return &State{
		contextProvider:            contextProvider,
		accessControlAllowedOrigin: accessControlAllowedOrigin,
		backendVersion:             backendVersion,
		playerHandler:              handlerForPlayer,
		gameHandler:                handlerForGame,
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
	case "version":
		{
			versionForBody := &parsing.VersionForBody{Version: state.backendVersion}
			json.NewEncoder(httpResponseWriter).Encode(versionForBody)
			return
		}
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
		objectForBody, httpStatus =
			requestHandler.HandleGet(
				state.contextProvider.FromRequest(httpRequest),
				pathSegments[2:])
	case http.MethodPost:
		{
			if httpRequest.Body == nil {
				http.Error(httpResponseWriter, "Empty request body", http.StatusBadRequest)
				return
			}

			objectForBody, httpStatus =
				requestHandler.HandlePost(
					state.contextProvider.FromRequest(httpRequest),
					json.NewDecoder(httpRequest.Body),
					pathSegments[2:])
		}
	default:
		http.Error(httpResponseWriter, "Method not GET or POST: "+httpRequest.Method, http.StatusBadRequest)
	}

	// If the status is OK, writing the header with OK won't make any difference.
	httpResponseWriter.WriteHeader(httpStatus)

	errorMessageForBody, isError := objectForBody.(error)

	// Since errors go out just as strings, we wrap the .Error() string in an object recognized
	// by the frontend.
	if isError {
		objectForBody = &parsing.ErrorForBody{
			Error: errorMessageForBody.Error(),
		}
	}

	json.NewEncoder(httpResponseWriter).Encode(objectForBody)
}

// parsePathSegments returns the segments of the URI path as a slice of a string array.
func parsePathSegments(httpRequest *http.Request) []string {
	// The initial character is '/' so we skip it to avoid an empty string as the first element.
	return strings.Split(httpRequest.URL.Path[1:], "/")
}
