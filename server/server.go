package server

import (
	"github.com/benoleary/ilutulestikud/game"
	"github.com/benoleary/ilutulestikud/parseuri"
	"github.com/benoleary/ilutulestikud/player"
	"net/http"
)

type httpGetAndPostHandler interface {
	HandleGet(
		httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string)

	HandlePost(
		httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string)
}

type State struct {
	accessControlAllowedOrigin string
	playerHandler              player.Handler
	gameHandler                game.Handler
}

func CreateNew(accessControlAllowedOrigin string) State {
	playerHandler := player.CreateHandler()
	return State{accessControlAllowedOrigin, playerHandler, game.CreateHandler(&playerHandler)}
}

// rootHandler calls functions according to the second segment of the URI, assuming that the first
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

	pathSegments := parseuri.PathSegments(httpRequest)

	// We choose the interface which will handle the GET or POST based on the
	// first segment of the URI after "backend".
	var requestHandler httpGetAndPostHandler
	switch pathSegments[1] {
	// Deprecated: "lobby" exists for backwards compatibility with the front-end.
	case "lobby":
		requestHandler = &state.playerHandler
	case "player":
		requestHandler = &state.playerHandler
	case "game":
		requestHandler = &state.gameHandler
	default:
		http.NotFound(httpResponseWriter, httpRequest)
	}

	switch httpRequest.Method {
	case http.MethodOptions:
		return
	case http.MethodGet:
		requestHandler.HandleGet(httpResponseWriter, httpRequest, pathSegments[2:])
	case http.MethodPost:
		requestHandler.HandlePost(httpResponseWriter, httpRequest, pathSegments[2:])
	default:
		http.Error(httpResponseWriter, "Method not GET or POST: "+httpRequest.Method, http.StatusBadRequest)
	}
}
