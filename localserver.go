package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/game"
	game_persister "github.com/benoleary/ilutulestikud/backend/game/persister"
	"github.com/benoleary/ilutulestikud/backend/player"
	player_persister "github.com/benoleary/ilutulestikud/backend/player/persister"
	"github.com/benoleary/ilutulestikud/backend/server"
	endpoint_parsing "github.com/benoleary/ilutulestikud/backend/server/endpoint/parsing"
)

func main() {
	fmt.Printf("Local server started.\n")

	// This main function just injects hard-coded dependencies.
	playerPersister := player_persister.NewInMemory()
	playerCollection, errorCreatingPlayerCollection :=
		player.NewCollection(
			playerPersister,
			defaults.InitialPlayerNames,
			defaults.AvailableColors)

	if errorCreatingPlayerCollection != nil {
		log.Fatalf(
			"Error when creating player collection: %v",
			errorCreatingPlayerCollection)
	}

	gamePersister := game_persister.NewInMemory()
	gameCollection :=
		game.NewCollection(
			gamePersister,
			playerCollection)

	// We could load the allowed origin from a file, but this app is very specific to a set of fixed addresses.
	serverState :=
		server.New(
			"http://localhost:4233",
			&endpoint_parsing.Base32Translator{},
			playerCollection,
			gameCollection)
	http.HandleFunc("/backend/", serverState.HandleBackend)
	http.ListenAndServe(":8080", nil)
}
