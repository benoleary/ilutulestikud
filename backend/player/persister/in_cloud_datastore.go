package persister

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/benoleary/ilutulestikud/backend/cloud"
	"github.com/benoleary/ilutulestikud/backend/player"
	"google.golang.org/api/iterator"
)

const keyKind = "Player"

// inCloudDatastorePersister stores game states by creating
// inCloudDatastoreStates and saving them as game.ReadAndWriteStates
// in Google Cloud Datastore.
type inCloudDatastorePersister struct {
	randomNumberGenerator *rand.Rand
	projectIdentifier     string
	datastoreClient       cloud.LimitedClient
}

// NewInCloudDatastore creates a game state persister.
func NewInCloudDatastore(projectIdentifier string) player.StatePersister {
	return NewInCloudDatastoreWithGivenLimitedClient(projectIdentifier, nil)
}

// NewInCloudDatastoreWithGivenLimitedClient creates a game state
// persister using a given LimitedClient implementation.
func NewInCloudDatastoreWithGivenLimitedClient(
	projectIdentifier string,
	datastoreClient cloud.LimitedClient) player.StatePersister {
	return &inCloudDatastorePersister{
		randomNumberGenerator: rand.New(rand.NewSource(time.Now().Unix())),
		projectIdentifier:     projectIdentifier,
		datastoreClient:       datastoreClient,
	}
}

// Add inserts the given name and color as a row in the database.
func (playerPersister *inCloudDatastorePersister) Add(
	executionContext context.Context,
	playerName string,
	chatColor string) error {
	return playerPersister.insertOrOverwrite(
		executionContext,
		playerName,
		chatColor,
		false)
}

// UpdateColor updates the given player to have the given chat color. It
// relies on the PostgreSQL driver to ensure thread safety.
func (playerPersister *inCloudDatastorePersister) UpdateColor(
	executionContext context.Context,
	playerName string,
	chatColor string) error {
	return playerPersister.insertOrOverwrite(
		executionContext,
		playerName,
		chatColor,
		true)
}

// Get returns the ReadOnly corresponding to the given player name if it exists.
func (playerPersister *inCloudDatastorePersister) Get(
	executionContext context.Context,
	playerName string) (player.ReadonlyState, error) {
	initializedClient, errorFromAcquiral :=
		playerPersister.acquireClientIfValidName(executionContext, playerName)

	if errorFromAcquiral != nil {
		return nil, errorFromAcquiral
	}

	playerKey := initializedClient.KeyFor(playerName)
	serializableState := player.ReadAndWriteState{}

	errorFromGet :=
		initializedClient.Get(
			executionContext,
			playerKey,
			&serializableState)

	return &serializableState, errorFromGet
}

// All returns a slice of all the players in the collection as ReadonlyState
// instances, ordered as given by the database.
func (playerPersister *inCloudDatastorePersister) All(
	executionContext context.Context) ([]player.ReadonlyState, error) {
	initializedClient, errorFromAcquiral :=
		playerPersister.acquireClient(executionContext)

	if errorFromAcquiral != nil {
		return nil, errorFromAcquiral
	}

	// We do not want to filter anything from the query for entities
	// of the player type.
	queryOnKind := datastore.NewQuery(keyKind)

	resultIterator :=
		initializedClient.Run(executionContext, queryOnKind)

	playerStates := []player.ReadonlyState{}

	for {
		var retrievedPlayer player.ReadAndWriteState
		_, errorFromNext := resultIterator.Next(&retrievedPlayer)

		if errorFromNext == iterator.Done {
			break
		}

		if errorFromNext != nil {
			return nil, errorFromNext
		}

		playerStates = append(playerStates, &retrievedPlayer)
	}

	return playerStates, nil
}

// Delete deletes the given game from the collection. It returns an error
// if the Cloud Datastore API returns an error.
func (playerPersister *inCloudDatastorePersister) Delete(
	executionContext context.Context,
	playerName string) error {
	initializedClient, errorFromAcquiral :=
		playerPersister.acquireClient(executionContext)

	if errorFromAcquiral != nil {
		return errorFromAcquiral
	}

	return initializedClient.Delete(
		executionContext,
		initializedClient.KeyFor(playerName))
}

// acquireClient returns the connection to the Cloud Datastore,
// initializing it if it has not already been initialized.
func (playerPersister *inCloudDatastorePersister) acquireClient(
	executionContext context.Context) (cloud.LimitedClient, error) {
	if playerPersister.datastoreClient == nil {
		cloudDatastoreClient, errorFromCloudDatastore :=
			datastore.NewClient(
				executionContext,
				playerPersister.projectIdentifier)
		if errorFromCloudDatastore != nil {
			return nil, errorFromCloudDatastore
		}

		playerPersister.datastoreClient =
			cloud.WrapDatastoreClient(cloudDatastoreClient, keyKind)
	}

	return playerPersister.datastoreClient, nil
}

func (playerPersister *inCloudDatastorePersister) acquireClientIfValidName(
	executionContext context.Context,
	playerName string) (cloud.LimitedClient, error) {
	if playerName == "" {
		return nil, fmt.Errorf("Player must have a name")
	}

	return playerPersister.acquireClient(executionContext)
}

func (playerPersister *inCloudDatastorePersister) insertOrOverwrite(
	executionContext context.Context,
	playerName string,
	chatColor string,
	isUpdate bool) error {
	initializedClient, errorFromAcquiral :=
		playerPersister.acquireClientIfValidName(executionContext, playerName)

	if errorFromAcquiral != nil {
		return errorFromAcquiral
	}

	playerKey, isAlreadyInDatastore, errorFromCheck :=
		cloud.KeyWithIfNameExists(
			executionContext,
			initializedClient,
			playerName)

	if errorFromCheck != nil {
		return errorFromCheck
	}

	if isAlreadyInDatastore && !isUpdate {
		return fmt.Errorf("Player with name %v already exists", playerName)
	}

	if !isAlreadyInDatastore && isUpdate {
		return fmt.Errorf("Player with name %v does not exist", playerName)
	}

	serializableState :=
		player.ReadAndWriteState{
			PlayerName: playerName,
			ChatColor:  chatColor,
		}

	_, errorFromPut :=
		initializedClient.Put(
			executionContext,
			playerKey,
			&serializableState)

	return errorFromPut
}
