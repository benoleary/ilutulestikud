package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/player"
	"github.com/benoleary/ilutulestikud/backend/server"
)

func main() {
	fmt.Printf("Local server started.\n")

	// This main function just injects hard-coded dependencies.
	playerPersister := player.NewInMemoryPersister()
	playerCollection, errorCreatingPlayerCollection :=
		player.NewCollection(
			playerPersister,
			defaults.InitialPlayerNames(),
			defaults.AvailableColors())

	if errorCreatingPlayerCollection != nil {
		log.Fatalf(
			"Error when creating player collection: %v",
			errorCreatingPlayerCollection)
	}

	gamePersister := game.NewInMemoryPersister()
	gameCollection :=
		game.NewCollection(
			gamePersister,
			playerCollection)

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
