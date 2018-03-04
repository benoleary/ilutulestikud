package main

import (
	"fmt"
	"net/http"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/player"
	"github.com/benoleary/ilutulestikud/backend/server"
)

func main() {
	fmt.Printf("Local server started.\n")

	// This main function just injects hard-coded dependencies.
	playerFactory := &player.ThreadsafeFactory{}
	initialPlayers := player.DefaultPlayers()
	playerHandler := player.NewGetAndPostHandler(playerFactory, initialPlayers)
	gameHandler := game.NewGetAndPostHandler(playerHandler)

	// We could load the allowed origin from a file, but this app is very specific to a set of fixed addresses.
	serverState := server.New("http://localhost:4233", playerHandler, gameHandler)
	http.HandleFunc("/backend/", serverState.HandleBackend)
	http.ListenAndServe(":8080", nil)
}
