package persister_test

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/datastore"
	"github.com/benoleary/ilutulestikud/backend/cloud"
	"github.com/benoleary/ilutulestikud/backend/player/persister"
)

const testProject = "Test-Project"

type mockLimitedIterator struct {
	ErrorToReturn error
}

func (mockIterator *mockLimitedIterator) Next(
	deserializationDestination interface{}) (*datastore.Key, error) {
	return nil, mockIterator.ErrorToReturn
}

type mockLimitedClient struct {
	IteratorToReturn *mockLimitedIterator
	ErrorToReturn    error
}

func (mockClient *mockLimitedClient) KeyFor(
	nameForKey string) *datastore.Key {
	return nil
}

func (mockClient *mockLimitedClient) Run(
	executionContext context.Context,
	queryToRun *datastore.Query) cloud.LimitedIterator {
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
	playerName := "Should Not Matter"
	playerColor := "Should not matter"
	errorFromAdd :=
		cloudDatastorePersister.Add(executionContext, playerName, playerColor)

	if errorFromAdd == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" project identifier %v, and got got nil error from"+
				" .Add(%v, %v,%v)",
			cloudDatastorePersister,
			invalidProjectIdentifier,
			executionContext,
			playerName,
			playerColor)
	}

	errorFromUpdateColor :=
		cloudDatastorePersister.UpdateColor(executionContext, playerName, playerColor)

	if errorFromUpdateColor == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" project identifier %v, and got got nil error from"+
				" .UpdateColor(%v, %v,%v)",
			cloudDatastorePersister,
			invalidProjectIdentifier,
			executionContext,
			playerName,
			playerColor)
	}

	unexpectedPlayer, errorFromGet :=
		cloudDatastorePersister.Get(executionContext, playerName)

	if errorFromGet == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" project identifier %v, and got %v from"+
				" .Get(%v, %v) instead of producing error",
			cloudDatastorePersister,
			invalidProjectIdentifier,
			unexpectedPlayer,
			executionContext,
			playerName)
	}

	errorFromDeleteRequest :=
		cloudDatastorePersister.Delete(executionContext, playerName)

	if errorFromDeleteRequest == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" project identifier %v, and got got nil error from"+
				" .Delete(%v, %v)",
			cloudDatastorePersister,
			invalidProjectIdentifier,
			executionContext,
			playerName)
	}
}

func TestAllPropagatesAcquiralError(unitTest *testing.T) {
	cloudDatastorePersister :=
		persister.NewInCloudDatastore("")

	playerList, errorFromAll :=
		cloudDatastorePersister.All(nil)

	if errorFromAll == nil {
		unitTest.Fatalf(
			"All(nil) produced %v with nil error",
			playerList)
	}
}

func TestAllPropagatesIteratorError(unitTest *testing.T) {
	mockIterator :=
		&mockLimitedIterator{
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

	playerList, errorFromAll :=
		cloudDatastorePersister.All(context.Background())

	if errorFromAll == nil {
		unitTest.Fatalf(
			"All([background context]) produced %v with nil error",
			playerList)
	}
}

func TestGetInvalidPlayerProducesError(unitTest *testing.T) {
	invalidName := ""

	mockIterator :=
		&mockLimitedIterator{
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

	executionContext := context.Background()
	unexpectedPlayer, errorFromGet :=
		cloudDatastorePersister.Get(executionContext, invalidName)

	if errorFromGet == nil {
		unitTest.Fatalf(
			"Get(%v, %v) produced %+v instead of producing error",
			executionContext,
			invalidName,
			unexpectedPlayer)
	}
}

func TestUpdatePropagatesKeyCheckError(unitTest *testing.T) {
	mockIterator :=
		&mockLimitedIterator{
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

	executionContext := context.Background()
	playerName := "Should Not Matter"
	playerColor := "Should not matter"

	errorFromUpdateColor :=
		cloudDatastorePersister.UpdateColor(executionContext, playerName, playerColor)

	if errorFromUpdateColor == nil {
		unitTest.Fatalf(
			"UpdateColor([background context], %v, %v) produced nil error",
			playerName,
			playerColor)
	}
}
