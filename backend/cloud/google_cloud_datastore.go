package cloud

import (
	"context"
	"fmt"

	"google.golang.org/api/iterator"
)

// IlutulestikudIdentifier is the string which identifies the project to the
// Google App Engine.
const IlutulestikudIdentifier = "ilutulestikud-191419"

// LimitedIterator defines wrappers around the subset of the functions of the
// datastore.Iterator struct used by the inCloudDatastorePersister struct.
type LimitedIterator interface {
	// DeserializeNext should deserialize the next element in the iterator
	// into the given interface.
	DeserializeNext(deserializationDestination interface{}) error

	// NextKey should simply move onto the next element, discarding whatever
	// the iterator was pointing at.
	NextKey() error
}

// LimitedClient defines the subset of the functions of the
// datastore.Client struct used by the inCloudDatastorePersister
// struct.
type LimitedClient interface {
	// AllOfKind should return an iterator to the set of all entities
	// of the kind known to the client.
	AllOfKind(executionContext context.Context) LimitedIterator

	// AllKeysMatching should return an iterator to the set of all keys
	// of the kind known to the client which match the given name.
	AllKeysMatching(
		executionContext context.Context,
		keyName string) LimitedIterator

	AllMatching(
		executionContext context.Context,
		filterExpression string,
		valueToMatch interface{}) LimitedIterator

	Get(
		executionContext context.Context,
		keyName string,
		deserializationDestination interface{}) error

	Put(
		executionContext context.Context,
		keyName string,
		deserializationSource interface{}) error

	Delete(
		executionContext context.Context,
		keyName string) error
}

// ClientProvider defines a factory interface which should provide
// implementations of the LimitedClient interface.
type ClientProvider interface {
	// NewClient should provide a new instance of an implementation
	// of the LimitedClient interface.
	NewClient(executionContext context.Context) (LimitedClient, error)
}

// DoesNameExist returns whether that name already in the datastore
// for the kind belonging to the client, along with an error if any
// problem was encountered.
func DoesNameExist(
	executionContext context.Context,
	limitedClient LimitedClient,
	nameToCheck string) (bool, error) {
	resultIterator :=
		limitedClient.AllKeysMatching(
			executionContext,
			nameToCheck)

	// If there is nothing already with the given name, the iterator
	// should immediately return an iterator.Done "error".
	errorFromInitialNext := resultIterator.NextKey()

	if errorFromInitialNext == iterator.Done {
		return false, nil
	}

	// Otherwise any error means that there was an actual error.
	if errorFromInitialNext != nil {
		errorFromInitialNextWithContext :=
			fmt.Errorf(
				"Trying to check for existing name %v produced error: %v",
				nameToCheck,
				errorFromInitialNext)
		return false, errorFromInitialNextWithContext
	}

	// If the first .Next(...) had no error, we check that the next
	// invocation gives iterator.Done - otherwise we report an error.
	errorFromSecondNext := resultIterator.NextKey()

	if errorFromSecondNext != iterator.Done {
		errorFromSecondNextWithContext :=
			fmt.Errorf(
				"Existing name %v was found but then checking for iterator.Done"+
					" produced error: %v",
				nameToCheck,
				errorFromSecondNext)

		return true, errorFromSecondNextWithContext
	}

	return true, nil
}
