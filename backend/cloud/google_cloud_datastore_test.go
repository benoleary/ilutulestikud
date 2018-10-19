package cloud_test

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/datastore"
	"github.com/benoleary/ilutulestikud/backend/cloud"
	"google.golang.org/api/iterator"
)

const testKind = "integration_test"
const invalidName = "Invalid Name"
const testValidName = "Test Valid Name"

var testNames = []string{
	invalidName,
	testValidName,
}

type testObject struct {
	Name    string
	Comment string
}

// We have a mock in order to test the logic of KeyWithIfNameExists
type mockLimitedIterator struct {
	ErrorsToReturn []error
}

func (mockIterator *mockLimitedIterator) Next(
	deserializationDestination interface{}) (*datastore.Key, error) {
	errorToReturn := mockIterator.ErrorsToReturn[0]
	if len(mockIterator.ErrorsToReturn) > 1 {
		mockIterator.ErrorsToReturn = mockIterator.ErrorsToReturn[1:]
	}

	return nil, errorToReturn
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

func TestInvalidNameDoesNotGetKey(unitTest *testing.T) {
	mockIterator :=
		mockLimitedIterator{
			ErrorsToReturn: []error{
				iterator.Done,
				fmt.Errorf("Unexpected error"),
			},
		}

	mockClient :=
		&mockLimitedClient{
			IteratorToReturn: &mockIterator,
			ErrorToReturn:    nil,
		}

	actualKey, isAlreadyInDatastore, errorFromCheck :=
		cloud.KeyWithIfNameExists(
			context.Background(),
			mockClient,
			invalidName)

	if isAlreadyInDatastore || (errorFromCheck != nil) {
		unitTest.Fatalf(
			"KeyWithIfNameExists([background context], %+v, %v) produced %+v, %v, %+v",
			mockClient,
			invalidName,
			actualKey,
			isAlreadyInDatastore,
			errorFromCheck)
	}
}

func TestKeyWithIfNameExistsPropagatesIteratorError(unitTest *testing.T) {
	mockIterator :=
		mockLimitedIterator{
			ErrorsToReturn: []error{
				fmt.Errorf("Expected error"),
				fmt.Errorf("Unexpected error"),
			},
		}

	mockClient :=
		&mockLimitedClient{
			IteratorToReturn: &mockIterator,
			ErrorToReturn:    nil,
		}

	actualKey, isAlreadyInDatastore, errorFromCheck :=
		cloud.KeyWithIfNameExists(
			context.Background(),
			mockClient,
			invalidName)

	if errorFromCheck == nil {
		unitTest.Fatalf(
			"KeyWithIfNameExists([background context], %+v, %v) produced %+v, %v, %+v",
			mockClient,
			invalidName,
			actualKey,
			isAlreadyInDatastore,
			errorFromCheck)
	}
}

func TestKeyWithIfNameExistsGivesErrorIfIteratorNotDoneAfterValidName(unitTest *testing.T) {
	mockIterator :=
		mockLimitedIterator{
			ErrorsToReturn: []error{
				nil,
				fmt.Errorf("Expected error"),
				fmt.Errorf("Unexpected error"),
			},
		}

	mockClient :=
		&mockLimitedClient{
			IteratorToReturn: &mockIterator,
			ErrorToReturn:    nil,
		}

	actualKey, isAlreadyInDatastore, errorFromCheck :=
		cloud.KeyWithIfNameExists(
			context.Background(),
			mockClient,
			invalidName)

	if errorFromCheck == nil {
		unitTest.Fatalf(
			"KeyWithIfNameExists([background context], %+v, %v) produced %+v, %v, %+v",
			mockClient,
			invalidName,
			actualKey,
			isAlreadyInDatastore,
			errorFromCheck)
	}
}

func TestValidNameDoesGetKey(unitTest *testing.T) {
	mockIterator :=
		mockLimitedIterator{
			ErrorsToReturn: []error{
				nil,
				iterator.Done,
				fmt.Errorf("Unexpected error"),
			},
		}

	mockClient :=
		&mockLimitedClient{
			IteratorToReturn: &mockIterator,
			ErrorToReturn:    nil,
		}

	actualKey, isAlreadyInDatastore, errorFromCheck :=
		cloud.KeyWithIfNameExists(
			context.Background(),
			mockClient,
			testValidName)

	if !isAlreadyInDatastore || (errorFromCheck != nil) {
		unitTest.Fatalf(
			"KeyWithIfNameExists([background context], %+v, %v) produced %+v, %v, %+v",
			mockClient,
			invalidName,
			actualKey,
			isAlreadyInDatastore,
			errorFromCheck)
	}
}

func TestCreateThenGetThenDeleteThenRun(unitTest *testing.T) {
	wrappedClient := createClient(unitTest)

	objectToPersist := testObject{
		Name:    "Test Name",
		Comment: "Test comment",
	}

	expectedKey := wrappedClient.KeyFor(testValidName)

	keyFromPut, errorFromPut :=
		wrappedClient.Put(
			context.Background(),
			expectedKey,
			&objectToPersist)

	if errorFromPut != nil {
		unitTest.Fatalf(
			"Put([background context], %+v, [pointer to %+v]) produced error %+v",
			expectedKey,
			objectToPersist,
			errorFromPut)
	}

	if keyFromPut != expectedKey {
		unitTest.Fatalf(
			"Put([background context], %+v, [pointer to %+v]) produced key %+v"+
				" which does not match expected key %+v",
			expectedKey,
			objectToPersist,
			keyFromPut,
			expectedKey)
	}

	var retrievedObject testObject
	errorFromGet :=
		wrappedClient.Get(
			context.Background(),
			expectedKey,
			&retrievedObject)

	if errorFromGet != nil {
		unitTest.Fatalf(
			"Get([background context], %+v, [pointer to %+v]) produced error %+v",
			expectedKey,
			retrievedObject,
			errorFromGet)
	}

	if retrievedObject != objectToPersist {
		unitTest.Fatalf(
			"Retrieved %+v instead of expected %+v",
			retrievedObject,
			objectToPersist)
	}

	errorFromDelete :=
		wrappedClient.Delete(
			context.Background(),
			expectedKey)

	if errorFromDelete != nil {
		unitTest.Fatalf(
			"Delete([background context], %+v) produced error %+v",
			expectedKey,
			errorFromDelete)
	}

	// We run a query that just fetches all the test objects
	// - and there should be none now.
	queryOnName := datastore.NewQuery(testKind)

	resultIterator :=
		wrappedClient.Run(context.Background(), queryOnName)

	if resultIterator == nil {
		unitTest.Fatalf(
			"Run([background context], %+v) produced nil iterator",
			queryOnName)
	}

	var keyOfExistingObject datastore.Key
	_, errorFromNext := resultIterator.Next(&keyOfExistingObject)
	if errorFromNext != iterator.Done {
		unitTest.Fatalf(
			"Next([pointer to %+v]) produced error %v",
			keyOfExistingObject,
			errorFromNext)
	}
}

func createClient(unitTest *testing.T) cloud.LimitedClient {
	cloudDatastoreClient, errorFromCloudDatastore :=
		datastore.NewClient(
			context.Background(),
			cloud.IlutulestikudIdentifier)
	if errorFromCloudDatastore != nil {
		unitTest.Fatalf(
			"Error when trying to create client: %v",
			errorFromCloudDatastore)
	}

	wrappedClient := cloud.WrapDatastoreClient(cloudDatastoreClient, testKind)

	for _, testName := range testNames {
		errorFromInitialDelete :=
			wrappedClient.Delete(
				context.Background(),
				wrappedClient.KeyFor(testName))
		unitTest.Logf(
			"Error from delete of %v when wrapping client: %v",
			testName,
			errorFromInitialDelete)
	}

	return wrappedClient
}
