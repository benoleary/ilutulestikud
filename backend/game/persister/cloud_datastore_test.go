package persister_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/cloud"
	"github.com/benoleary/ilutulestikud/backend/game/persister"
)

const testProject = "Test-Project"

type mockLimitedIterator struct {
	ErrorToReturn error
}

func (mockIterator *mockLimitedIterator) DeserializeNext(
	deserializationDestination interface{}) error {
	serializableState, isSerializableState :=
		deserializationDestination.(persister.SerializableState)

	if isSerializableState {
		// The only use case for this is in the test which gives back a
		// de-serialization error, which can conveniently cover the case
		// of being unable to de-serialize the ruleset. Hence we use -1,
		// which should not be a valid ruleset identifier.
		serializableState.RulesetIdentifier = -1
	}

	return mockIterator.ErrorToReturn
}

func (mockIterator *mockLimitedIterator) NextKey() error {
	return mockIterator.ErrorToReturn
}

type mockLimitedClient struct {
	IteratorToReturn *mockLimitedIterator
	ErrorToReturn    error
}

func (mockClient *mockLimitedClient) AllOfKind(
	executionContext context.Context) cloud.LimitedIterator {
	return mockClient.IteratorToReturn
}

func (mockClient *mockLimitedClient) AllKeysMatching(
	executionContext context.Context,
	keyName string) cloud.LimitedIterator {
	return mockClient.IteratorToReturn
}

func (mockClient *mockLimitedClient) AllMatching(
	executionContext context.Context,
	filterExpression string,
	valueToMatch interface{}) cloud.LimitedIterator {
	return mockClient.IteratorToReturn
}

func (mockClient *mockLimitedClient) Get(
	executionContext context.Context,
	nameForKey string,
	deserializationDestination interface{}) error {
	return mockClient.ErrorToReturn
}

func (mockClient *mockLimitedClient) Put(
	executionContext context.Context,
	nameForKey string,
	deserializationSource interface{}) error {
	return mockClient.ErrorToReturn
}

func (mockClient *mockLimitedClient) Delete(
	executionContext context.Context,
	nameForKey string) error {
	return mockClient.ErrorToReturn
}

type mockClientProvider struct {
	ClientToReturn *mockLimitedClient
	ErrorToReturn  error
}

func (mockProvider *mockClientProvider) NewClient(
	executionContext context.Context) (cloud.LimitedClient, error) {
	return mockProvider.ClientToReturn, mockProvider.ErrorToReturn
}

func TestPropagateErrorFromClientProvider(unitTest *testing.T) {
	invalidDatastoreClientProvider :=
		&mockClientProvider{
			ClientToReturn: nil,
			ErrorToReturn:  fmt.Errorf("Expected error"),
		}
	cloudDatastorePersister :=
		persister.NewInCloudDatastore(invalidDatastoreClientProvider)

	executionContext := context.Background()

	// We test that every kind of request generates an error.
	gameName := "Should not matter"
	unexpectedGame, errorFromReadAndWriteGameRequest :=
		cloudDatastorePersister.ReadAndWriteGame(executionContext, gameName)

	if errorFromReadAndWriteGameRequest == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" invalid client provider, and got %v from"+
				" .ReadAndWriteGame(%v, %v) instead of producing error",
			cloudDatastorePersister,
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
				" invalid client provider, and got %v from"+
				" .ReadAllWithPlayer(%v, %v) instead of producing error",
			cloudDatastorePersister,
			unexpectedGamesWithPlayer,
			executionContext,
			playerName)
	}

	errorFromAddGameRequest :=
		cloudDatastorePersister.AddGame(executionContext, gameName, 0, nil, nil, nil, nil)

	if errorFromAddGameRequest == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" invalid client provider, and got got nil error from"+
				" .AddGame(%v, %v, 0, nil, nil, nil, nil)",
			cloudDatastorePersister,
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
				" invalid client provider, and got got nil error from"+
				" .RemoveGameFromListForPlayer(%v, %v, %v)",
			cloudDatastorePersister,
			executionContext,
			gameName,
			playerName)
	}

	errorFromDeleteRequest :=
		cloudDatastorePersister.Delete(executionContext, gameName)

	if errorFromDeleteRequest == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" invalid client provider, and got got nil error from"+
				" .Delete(%v, %v)",
			cloudDatastorePersister,
			executionContext,
			gameName)
	}
}

func TestReadAllWithPlayerPropagatesIteratorError(unitTest *testing.T) {
	mockIterator :=
		&mockLimitedIterator{
			ErrorToReturn: fmt.Errorf("Expected error"),
		}

	mockClient :=
		&mockLimitedClient{
			IteratorToReturn: mockIterator,
			ErrorToReturn:    nil,
		}

	mockProvider :=
		&mockClientProvider{
			ClientToReturn: mockClient,
			ErrorToReturn:  nil,
		}

	cloudDatastorePersister :=
		persister.NewInCloudDatastore(mockProvider)

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
			ErrorToReturn: fmt.Errorf("Expected error"),
		}

	mockClient :=
		&mockLimitedClient{
			IteratorToReturn: mockIterator,
			ErrorToReturn:    nil,
		}

	mockProvider :=
		&mockClientProvider{
			ClientToReturn: mockClient,
			ErrorToReturn:  nil,
		}

	cloudDatastorePersister :=
		persister.NewInCloudDatastore(mockProvider)

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
			ErrorToReturn: nil,
		}

	mockClient :=
		&mockLimitedClient{
			IteratorToReturn: mockIterator,
			ErrorToReturn:    nil,
		}

	mockProvider :=
		&mockClientProvider{
			ClientToReturn: mockClient,
			ErrorToReturn:  nil,
		}

	cloudDatastorePersister :=
		persister.NewInCloudDatastore(mockProvider)

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
