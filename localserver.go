package main

import (
	"fmt"
	"net/http"

	"github.com/benoleary/ilutulestikud/backend/endpoint"

	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/player"
	"github.com/benoleary/ilutulestikud/backend/server"
)

func main() {
	fmt.Printf("Local server started.\n")

	// This main function just injects hard-coded dependencies.
	playerPersister := player.NewInMemoryPersister()
	playerCollection :=
		player.NewCollection(
			playerPersister,
			defaults.InitialPlayerNames(),
			defaults.AvailableColors())
	gameCollection := game.NewInMemoryCollection(&endpoint.Base32NameEncoder{})

	// We could load the allowed origin from a file, but this app is very specific to a set of fixed addresses.
	serverState :=
		server.New(
			"http://localhost:4233",
			&server.Base32Translator{},
			playerCollection,
			gameCollection)
	http.HandleFunc("/backend/", serverState.HandleBackend)
	http.ListenAndServe(":8080", nil)
}
