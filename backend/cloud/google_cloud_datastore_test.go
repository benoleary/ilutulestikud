package cloud_test

import (
	"context"
	"fmt"
	"testing"

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

func (mockIterator *mockLimitedIterator) nextError() error {
	errorToReturn := mockIterator.ErrorsToReturn[0]
	if len(mockIterator.ErrorsToReturn) > 1 {
		mockIterator.ErrorsToReturn = mockIterator.ErrorsToReturn[1:]
	}

	return errorToReturn
}

func (mockIterator *mockLimitedIterator) DeserializeNext(
	deserializationDestination interface{}) error {
	return mockIterator.nextError()
}

func (mockIterator *mockLimitedIterator) NextKey() error {
	return mockIterator.nextError()
}

type mockLimitedClient struct {
	IteratorToReturn *mockLimitedIterator
	ErrorToReturn    error
}

// AllOfKind returns an iterator to the set of all entities of the
// kind known to the client.
func (mockClient *mockLimitedClient) AllOfKind(
	executionContext context.Context) cloud.LimitedIterator {
	return mockClient.IteratorToReturn
}

// AllKeysMatching returns an iterator to the set of all keys of the
// kind known to the client which match the given name.
func (mockClient *mockLimitedClient) AllKeysMatching(
	executionContext context.Context,
	keyName string) cloud.LimitedIterator {
	return mockClient.IteratorToReturn
}

// AllMatching returns an iterator to the set of entities which are
// selected by the given filter, according the rules of the Google
// Cloud Datastore (which can have surprising results when searching
// through entities which have arrays).
func (mockClient *mockLimitedClient) AllMatching(
	executionContext context.Context,
	filterExpression string,
	valueToMatch interface{}) cloud.LimitedIterator {
	return mockClient.IteratorToReturn
}

func (mockClient *mockLimitedClient) Get(
	executionContext context.Context,
	keyName string,
	deserializationDestination interface{}) error {
	return mockClient.ErrorToReturn
}

func (mockClient *mockLimitedClient) Put(
	executionContext context.Context,
	keyName string,
	deserializationSource interface{}) error {
	return mockClient.ErrorToReturn
}

func (mockClient *mockLimitedClient) Delete(
	executionContext context.Context,
	keyName string) error {
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

	isAlreadyInDatastore, errorFromCheck :=
		cloud.DoesNameExist(
			context.Background(),
			mockClient,
			invalidName)

	if isAlreadyInDatastore || (errorFromCheck != nil) {
		unitTest.Fatalf(
			"DoesNameExist([background context], %+v, %v) produced %v, %+v",
			mockClient,
			invalidName,
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

	isAlreadyInDatastore, errorFromCheck :=
		cloud.DoesNameExist(
			context.Background(),
			mockClient,
			invalidName)

	if errorFromCheck == nil {
		unitTest.Fatalf(
			"DoesNameExist([background context], %+v, %v) produced %v, %+v",
			mockClient,
			invalidName,
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

	isAlreadyInDatastore, errorFromCheck :=
		cloud.DoesNameExist(
			context.Background(),
			mockClient,
			invalidName)

	if errorFromCheck == nil {
		unitTest.Fatalf(
			"DoesNameExist([background context], %+v, %v) produced %v, %+v",
			mockClient,
			invalidName,
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

	isAlreadyInDatastore, errorFromCheck :=
		cloud.DoesNameExist(
			context.Background(),
			mockClient,
			testValidName)

	if !isAlreadyInDatastore || (errorFromCheck != nil) {
		unitTest.Fatalf(
			"DoesNameExist([background context], %+v, %v) produced %v, %+v",
			mockClient,
			invalidName,
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

	errorFromPut :=
		wrappedClient.Put(
			context.Background(),
			testValidName,
			&objectToPersist)

	if errorFromPut != nil {
		unitTest.Fatalf(
			"Put([background context], %+v, [pointer to %+v]) produced error %+v",
			testValidName,
			objectToPersist,
			errorFromPut)
	}

	var retrievedObject testObject
	errorFromGet :=
		wrappedClient.Get(
			context.Background(),
			testValidName,
			&retrievedObject)

	if errorFromGet != nil {
		unitTest.Fatalf(
			"Get([background context], %+v, [pointer to %+v]) produced error %+v",
			testValidName,
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
			testValidName)

	if errorFromDelete != nil {
		unitTest.Fatalf(
			"Delete([background context], %+v) produced error %+v",
			testValidName,
			errorFromDelete)
	}

	// We run a query that just fetches all the test objects
	// - and there should be none now.
	resultIterator := wrappedClient.AllOfKind(context.Background())

	if resultIterator == nil {
		unitTest.Fatalf("AllOfKind([background context]) produced nil iterator")
	}

	errorFromNext := resultIterator.NextKey()
	if errorFromNext != iterator.Done {
		unitTest.Fatalf(
			"NextKey() produced error %v",
			errorFromNext)
	}
}

func createClient(unitTest *testing.T) cloud.LimitedClient {
	clientProvider :=
		cloud.NewIlutulestikudDatastoreClientProvider(testKind)

	cloudDatastoreClient, errorFromCloudDatastore :=
		clientProvider.NewClient(
			context.Background())
	if errorFromCloudDatastore != nil {
		unitTest.Fatalf(
			"Error when trying to create client: %v",
			errorFromCloudDatastore)
	}

	for _, testName := range testNames {
		errorFromInitialDelete :=
			cloudDatastoreClient.Delete(
				context.Background(),
				testName)
		unitTest.Logf(
			"Error from delete of %v when wrapping client: %v",
			testName,
			errorFromInitialDelete)
	}

	return cloudDatastoreClient
}
