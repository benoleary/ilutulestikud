package server

import (
	"fmt"
	"github.com/benoleary/ilutulestikud/lobby"
	"github.com/benoleary/ilutulestikud/parseuri"
	"net/http"
)

type State struct {
	accessControlAllowedOrigin string
	lobbyState                 lobby.State
}

func CreateNew(accessControlAllowedOrigin string) State {
	return State{accessControlAllowedOrigin, lobby.CreateEmpty()}
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
	fmt.Printf("handleBackend: pathSegments = %v\n\n", pathSegments)
	switch pathSegments[1] {
	case "lobby":
		state.handleLobby(httpResponseWriter, httpRequest, pathSegments[2:])
	case "game":
		state.handleGame(httpResponseWriter, httpRequest, pathSegments[2:])
	default:
		http.NotFound(httpResponseWriter, httpRequest)
	}
}

// handleLobby delegates responsibility for handling the HTTP request to the state's lobby state object.
func (state *State) handleLobby(httpResponseWriter http.ResponseWriter, httpRequest *http.Request, uriSegments []string) {
	state.lobbyState.HandleHttpRequest(httpResponseWriter, httpRequest, uriSegments)
}

// handleGame is currently a placeholder.
func (state *State) handleGame(httpResponseWriter http.ResponseWriter, httpRequest *http.Request, uriSegments []string) {
	returnHtml := `<h1>You tried to do something for a game. Well done!</h1>`
	fmt.Fprintf(httpResponseWriter, returnHtml)
}
