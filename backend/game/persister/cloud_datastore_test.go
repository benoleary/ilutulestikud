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
