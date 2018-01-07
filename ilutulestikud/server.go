package main

import (
	"fmt"
	"github.com/benoleary/ilutulestikud/lobby"
	"github.com/benoleary/ilutulestikud/parseuri"
	"net/http"
	"os"
)

type State struct {
	lobbyState lobby.State
}

// rootHandler calls functions according to the second segment of the URI, assuming that the first
// segment is "backend".
func (state *State) handleBackend(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
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

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("This program needs exactly one argument, which is the directory to serve for the Angular code")
		os.Exit(1)
	}

	serverState := State{lobby.CreateEmpty()}

	angularDirectory := os.Args[1]
	httpFileServer := http.FileServer(http.Dir(angularDirectory))
	http.Handle("/client/", http.StripPrefix("/client", httpFileServer))
	http.HandleFunc("/backend/", serverState.handleBackend)
	http.ListenAndServe(":8080", nil)
}
