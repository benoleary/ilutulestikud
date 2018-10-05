package persister_test

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/datastore"
	"github.com/benoleary/ilutulestikud/backend/game/persister"
)

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

func TestConstructorDoesNotCausePanic(unitTest *testing.T) {
	cloudDatastorePersister := persister.NewInCloudDatastore(nil)

	if cloudDatastorePersister == nil {
		unitTest.Fatalf("Created nil persister")
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
		persister.NewInCloudDatastoreAroundLimitedClient(mockClient)

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
		persister.NewInCloudDatastoreAroundLimitedClient(mockClient)

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
		persister.NewInCloudDatastoreAroundLimitedClient(mockClient)

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
