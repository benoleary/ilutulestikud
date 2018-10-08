package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"

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
	setupContext := context.Background()
	contextProvider := &server.BackgroundContextProvider{}

	postgresqlUsername := os.Getenv("POSTGRESQL_USERNAME")
	postgresqlPassword := os.Getenv("POSTGRESQL_PASSWORD")
	postgresqlPlayerdb := os.Getenv("POSTGRESQL_PLAYERDB")
	postgresqlLocation := os.Getenv("POSTGRESQL_LOCATION")
	connectionString :=
		fmt.Sprintf(
			"user=%v password=%v dbname=%v %v",
			postgresqlUsername,
			postgresqlPassword,
			postgresqlPlayerdb,
			postgresqlLocation)

	playerPersister := player_persister.NewInPostgresql(connectionString)

	playerCollection, errorCreatingPlayerCollection :=
		player.NewCollection(
			playerPersister,
			defaults.AvailableColors())

	datastoreClient, errorFromCloudDatastore :=
		datastore.NewClient(
			setupContext,
			game_persister.IlutulestikudIdentifier)

	if errorFromCloudDatastore != nil {
		log.Fatalf(
			"Error when creating client for Cloud Datastore: %v",
			errorFromCloudDatastore)
	}

	if errorCreatingPlayerCollection != nil {
		log.Fatalf(
			"Error when creating player collection: %v",
			errorCreatingPlayerCollection)
	}

	gamePersister := game_persister.NewInCloudDatastore(datastoreClient)
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
			&endpoint_parsing.Base32Translator{},
			playerCollection,
			gameCollection)
	http.HandleFunc("/backend/", serverState.HandleBackend)
	http.ListenAndServe(":8080", nil)
}
