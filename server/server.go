package server

import (
	"github.com/benoleary/ilutulestikud/game"
	"github.com/benoleary/ilutulestikud/httphandler"
	"github.com/benoleary/ilutulestikud/lobby"
	"github.com/benoleary/ilutulestikud/parseuri"
	"net/http"
)

type State struct {
	accessControlAllowedOrigin string
	lobbyState                 lobby.State
	activeGameCollection       game.CollectionState
}

func CreateNew(accessControlAllowedOrigin string) State {
	lobbyState := lobby.CreateInitial()
	return State{accessControlAllowedOrigin, lobbyState, game.CreateCollectionState(&lobbyState)}
}

// rootHandler calls functions according to the second segment of the URI, assuming that the first
// segment is "backend".
func (state *State) HandleBackend(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
	// If an allowed origin for access control has been set, we set all the headers to allow it.
	if state.accessControlAllowedOrigin != "" {
		httpResponseWriter.Header().Set("Access-Control-Allow-Origin", state.accessControlAllowedOrigin)
		httpResponseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		httpResponseWriter.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}

	if httpRequest.URL.Path == "" || httpRequest.URL.Path == "/" {
		http.Redirect(httpResponseWriter, httpRequest, "/client", http.StatusFound)
		return
	}

	pathSegments := parseuri.PathSegments(httpRequest)

	// We choose the interface which will handle the GET or POST based on the
	// first segment of the URI after "backend".
	var requestHandler httphandler.GetAndPostHandler
	switch pathSegments[1] {
	case "lobby":
		requestHandler = &state.lobbyState
	case "game":
		requestHandler = &state.activeGameCollection
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
