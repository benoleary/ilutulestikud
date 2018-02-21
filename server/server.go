package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/benoleary/ilutulestikud/game"
	"github.com/benoleary/ilutulestikud/player"
)

type httpGetAndPostHandler interface {
	HandleGet(relevantSegments []string) (interface{}, int)

	HandlePost(httpBodyDecoder *json.Decoder, relevantSegments []string) (interface{}, int)
}

// State contains all the state to allow the backend to function.
type State struct {
	accessControlAllowedOrigin string
	playerHandler              *player.Handler
	gameHandler                *game.Handler
}

// New creates a new State object and returns a pointer to it.
func New(accessControlAllowedOrigin string) *State {
	playerHandler := player.NewHandler()
	return &State{
		accessControlAllowedOrigin: accessControlAllowedOrigin,
		playerHandler:              playerHandler,
		gameHandler:                game.NewHandler(playerHandler)}
}

// HandleBackend calls functions according to the second segment of the URI, assuming that the first
// segment is "backend".
func (state *State) HandleBackend(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
	// If an allowed origin for access control has been set, we set all the headers to allow it.
	if state.accessControlAllowedOrigin != "" {
		httpResponseWriter.Header().Set("Access-Control-Allow-Origin", state.accessControlAllowedOrigin)
		httpResponseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		httpResponseWriter.Header().Set(
			"Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}

	if httpRequest.URL.Path == "" || httpRequest.URL.Path == "/" {
		http.Redirect(httpResponseWriter, httpRequest, "/client", http.StatusFound)
		return
	}

	pathSegments := parsePathSegments(httpRequest)

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

			objectForBody, httpStatus = requestHandler.HandlePost(json.NewDecoder(httpRequest.Body), pathSegments[2:])
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
