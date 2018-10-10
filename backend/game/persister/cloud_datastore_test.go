package persister_test

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/datastore"
	"github.com/benoleary/ilutulestikud/backend/game/persister"
)

const testProject = "Test-Project"

type mockLimitedIterator struct {
	KeyToReturn   *datastore.Key
	ErrorToReturn error
}

func (mockIterator *mockLimitedIterator) Next(
	deserializationDestination interface{}) (*datastore.Key, error) {
	serializableState, isSerializableState :=
		deserializationDestination.(persister.SerializableState)

	if isSerializableState {
		// The only use case for this is in the test which gives back a
		// de-serialization error, which can conveniently cover the case
		// of being unable to de-serialize the ruleset. Hence we use -1,
		// which should not be a valid ruleset identifier.
		serializableState.RulesetIdentifier = -1
	}

	return mockIterator.KeyToReturn, mockIterator.ErrorToReturn
}

type mockLimitedClient struct {
	IteratorToReturn *mockLimitedIterator
	ErrorToReturn    error
}

func (mockClient *mockLimitedClient) Run(
	executionContext context.Context,
	queryToRun *datastore.Query) persister.LimitedIterator {
	return mockClient.IteratorToReturn
}

func (mockClient *mockLimitedClient) Get(
	executionContext context.Context,
	searchKey *datastore.Key,
	deserializationDestination interface{}) (err error) {
	return mockClient.ErrorToReturn
}

func (mockClient *mockLimitedClient) Put(
	executionContext context.Context,
	searchKey *datastore.Key,
	deserializationSource interface{}) (*datastore.Key, error) {
	return nil, mockClient.ErrorToReturn
}

func (mockClient *mockLimitedClient) Delete(
	executionContext context.Context,
	searchKey *datastore.Key) error {
	return mockClient.ErrorToReturn
}

func TestReturnErrorFromInvalidProjectIdentifier(unitTest *testing.T) {
	invalidProjectIdentifier := ""
	cloudDatastorePersister :=
		persister.NewInCloudDatastore(invalidProjectIdentifier)

	executionContext := context.Background()

	// We test that every kind of request generates an error.
	gameName := "Should not matter"
	unexpectedGame, errorFromReadAndWriteGameRequest :=
		cloudDatastorePersister.ReadAndWriteGame(executionContext, gameName)

	if errorFromReadAndWriteGameRequest == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" project identifier %v, and got %v from"+
				" .ReadAndWriteGame(%v, %v) instead of producing error",
			cloudDatastorePersister,
			invalidProjectIdentifier,
			unexpectedGame,
			executionContext,
			gameName)
	}

	playerName := "Should Not Matter"
	unexpectedGamesWithPlayer, errorFromReadAllWithPlayerRequest :=
		cloudDatastorePersister.ReadAllWithPlayer(executionContext, playerName)

	if errorFromReadAllWithPlayerRequest == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" project identifier %v, and got %v from"+
				" .ReadAllWithPlayer(%v, %v) instead of producing error",
			cloudDatastorePersister,
			invalidProjectIdentifier,
			unexpectedGamesWithPlayer,
			executionContext,
			playerName)
	}

	errorFromAddGameRequest :=
		cloudDatastorePersister.AddGame(executionContext, gameName, 0, nil, nil, nil, nil)

	if errorFromAddGameRequest == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" project identifier %v, and got got nil error from"+
				" .AddGame(%v, %v, 0, nil, nil, nil, nil)",
			cloudDatastorePersister,
			invalidProjectIdentifier,
			executionContext,
			gameName)
	}

	errorFromRemoveGameFromListForPlayerRequest :=
		cloudDatastorePersister.RemoveGameFromListForPlayer(
			executionContext,
			gameName,
			playerName)

	if errorFromRemoveGameFromListForPlayerRequest == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" project identifier %v, and got got nil error from"+
				" .RemoveGameFromListForPlayer(%v, %v, %v)",
			cloudDatastorePersister,
			invalidProjectIdentifier,
			executionContext,
			gameName,
			playerName)
	}

	errorFromDeleteRequest :=
		cloudDatastorePersister.Delete(executionContext, gameName)

	if errorFromDeleteRequest == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" project identifier %v, and got got nil error from"+
				" .Delete(%v, %v)",
			cloudDatastorePersister,
			invalidProjectIdentifier,
			executionContext,
			gameName)
	}
}

func TestReadAllWithPlayerPropagatesIteratorError(unitTest *testing.T) {
	mockIterator :=
		&mockLimitedIterator{
			KeyToReturn:   nil,
			ErrorToReturn: fmt.Errorf("Expected error"),
		}

	mockClient :=
		&mockLimitedClient{
			IteratorToReturn: mockIterator,
			ErrorToReturn:    nil,
		}

	cloudDatastorePersister :=
		persister.NewInCloudDatastoreWithGivenLimitedClient(
			testProject,
			mockClient)

	playerName := "Does Not Matter"
	gamesWithPlayer, errorFromReadAll :=
		cloudDatastorePersister.ReadAllWithPlayer(nil, playerName)

	if errorFromReadAll == nil {
		unitTest.Fatalf(
			"ReadAllWithPlayer(nil, %v) produced %v with nil error",
			playerName,
			gamesWithPlayer)
	}
}

func TestAddGamePropagatesIteratorError(unitTest *testing.T) {
	mockIterator :=
		&mockLimitedIterator{
			KeyToReturn:   nil,
			ErrorToReturn: fmt.Errorf("Expected error"),
		}

	mockClient :=
		&mockLimitedClient{
			IteratorToReturn: mockIterator,
			ErrorToReturn:    nil,
		}

	cloudDatastorePersister :=
		persister.NewInCloudDatastoreWithGivenLimitedClient(
			testProject,
			mockClient)

	gameName := "does not matter"
	errorFromAddGame :=
		cloudDatastorePersister.AddGame(nil, gameName, 0, nil, nil, nil, nil)

	if errorFromAddGame == nil {
		unitTest.Fatalf(
			"AddGame(nil, %v, 0, nil, nil, nil, nil) produced nil error",
			gameName)
	}
}

func TestReadAllWithPlayerPropagatesDeserializationError(unitTest *testing.T) {
	mockIterator :=
		&mockLimitedIterator{
			KeyToReturn:   nil,
			ErrorToReturn: nil,
		}

	mockClient :=
		&mockLimitedClient{
			IteratorToReturn: mockIterator,
			ErrorToReturn:    nil,
		}

	cloudDatastorePersister :=
		persister.NewInCloudDatastoreWithGivenLimitedClient(
			testProject,
			mockClient)

	playerName := "Does Not Matter"
	gamesWithPlayer, errorFromReadAllWithPlayer :=
		cloudDatastorePersister.ReadAllWithPlayer(nil, playerName)

	if errorFromReadAllWithPlayer == nil {
		unitTest.Fatalf(
			"ReadAllWithPlayer(nil, %v) produced %v with nil error",
			playerName,
			gamesWithPlayer)
	}
}
