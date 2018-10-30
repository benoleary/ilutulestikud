package persister_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/cloud"
	"github.com/benoleary/ilutulestikud/backend/player/persister"
)

const testProject = "Test-Project"

type mockLimitedIterator struct {
	ErrorToReturn error
}

func (mockIterator *mockLimitedIterator) DeserializeNext(
	deserializationDestination interface{}) error {
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

func TestReturnErrorFromInvalidProjectIdentifier(unitTest *testing.T) {
	invalidDatastoreClientProvider :=
		&mockClientProvider{
			ClientToReturn: nil,
			ErrorToReturn:  fmt.Errorf("Expected error"),
		}
	cloudDatastorePersister :=
		persister.NewInCloudDatastore(invalidDatastoreClientProvider)

	executionContext := context.Background()

	// We test that every kind of request generates an error.
	playerName := "Should Not Matter"
	playerColor := "Should not matter"
	errorFromAdd :=
		cloudDatastorePersister.Add(executionContext, playerName, playerColor)

	if errorFromAdd == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" invalid client provider, and got got nil error from"+
				" .Add(%v, %v,%v)",
			cloudDatastorePersister,
			executionContext,
			playerName,
			playerColor)
	}

	errorFromUpdateColor :=
		cloudDatastorePersister.UpdateColor(executionContext, playerName, playerColor)

	if errorFromUpdateColor == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" invalid client provider, and got got nil error from"+
				" .UpdateColor(%v, %v,%v)",
			cloudDatastorePersister,
			executionContext,
			playerName,
			playerColor)
	}

	unexpectedPlayer, errorFromGet :=
		cloudDatastorePersister.Get(executionContext, playerName)

	if errorFromGet == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" invalid client provider, and got %v from"+
				" .Get(%v, %v) instead of producing error",
			cloudDatastorePersister,
			unexpectedPlayer,
			executionContext,
			playerName)
	}

	errorFromDeleteRequest :=
		cloudDatastorePersister.Delete(executionContext, playerName)

	if errorFromDeleteRequest == nil {
		unitTest.Fatalf(
			"Successfully created Cloud Datastore persister %+v from"+
				" invalid client provider, and got got nil error from"+
				" .Delete(%v, %v)",
			cloudDatastorePersister,
			executionContext,
			playerName)
	}
}

func TestAllPropagatesAcquiralError(unitTest *testing.T) {
	invalidProjectIdentifier := ""
	invalidDatastoreClientProvider :=
		cloud.NewFixedProjectAndKeyDatastoreClientProvider(
			invalidProjectIdentifier,
			persister.CloudDatastoreKeyKind)
	cloudDatastorePersister :=
		persister.NewInCloudDatastore(invalidDatastoreClientProvider)

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

	mockProvider :=
		&mockClientProvider{
			ClientToReturn: mockClient,
			ErrorToReturn:  nil,
		}

	cloudDatastorePersister :=
		persister.NewInCloudDatastore(mockProvider)

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

	mockProvider :=
		&mockClientProvider{
			ClientToReturn: mockClient,
			ErrorToReturn:  nil,
		}

	cloudDatastorePersister :=
		persister.NewInCloudDatastore(mockProvider)

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

	mockProvider :=
		&mockClientProvider{
			ClientToReturn: mockClient,
			ErrorToReturn:  nil,
		}

	cloudDatastorePersister :=
		persister.NewInCloudDatastore(mockProvider)

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
