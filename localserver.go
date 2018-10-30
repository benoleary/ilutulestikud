package main

import (
	"fmt"
	"net/http"

	"github.com/benoleary/ilutulestikud/backend/cloud"
	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/game"
	game_persister "github.com/benoleary/ilutulestikud/backend/game/persister"
	"github.com/benoleary/ilutulestikud/backend/player"
	player_persister "github.com/benoleary/ilutulestikud/backend/player/persister"
	"github.com/benoleary/ilutulestikud/backend/server"
	endpoint_parsing "github.com/benoleary/ilutulestikud/backend/server/endpoint/parsing"
)

// This main function just injects hard-coded dependencies.
func main() {
	fmt.Printf("Local server started.\n")
	contextProvider := &server.BackgroundContextProvider{}

	playerDatastoreClientProvider :=
		cloud.NewIlutulestikudDatastoreClientProvider(player_persister.CloudDatastoreKeyKind)
	playerPersister :=
		player_persister.NewInCloudDatastore(playerDatastoreClientProvider)
	playerCollection :=
		player.NewCollection(
			playerPersister,
			defaults.AvailableColors())

	gameDatastoreClientProvider :=
		cloud.NewIlutulestikudDatastoreClientProvider(game_persister.CloudDatastoreKeyKind)
	gamePersister :=
		game_persister.NewInCloudDatastore(gameDatastoreClientProvider)
	gameCollection :=
		game.NewCollection(
			gamePersister,
			8,
			playerCollection)

	// We could load the allowed origin from a file, but this app is very specific to a set of fixed addresses.
	serverState :=
		server.New(
			contextProvider,
			"http://localhost:4233",
			"Local version 2.0",
			&endpoint_parsing.Base32Translator{},
			playerCollection,
			gameCollection)
	http.HandleFunc("/backend/", serverState.HandleBackend)
	http.ListenAndServe(":8080", nil)
}
