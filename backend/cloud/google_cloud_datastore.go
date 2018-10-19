package cloud

import (
	"context"
	"fmt"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

// IlutulestikudIdentifier is the string which identifies the project to the
// Google App Engine.
const IlutulestikudIdentifier = "ilutulestikud-191419"

// LimitedIterator defines the subset of the functions of the
// datastore.Iterator struct used by the inCloudDatastorePersister
// struct.
type LimitedIterator interface {
	Next(deserializationDestination interface{}) (*datastore.Key, error)
}

// LimitedClient defines the subset of the functions of the
// datastore.Client struct used by the inCloudDatastorePersister
// struct.
type LimitedClient interface {
	KeyFor(nameForKey string) *datastore.Key

	Run(
		executionContext context.Context,
		queryToRun *datastore.Query) LimitedIterator

	Get(
		executionContext context.Context,
		keyForEntity *datastore.Key,
		deserializationDestination interface{}) (err error)

	Put(
		executionContext context.Context,
		keyForEntity *datastore.Key,
		deserializationSource interface{}) (*datastore.Key, error)

	Delete(
		executionContext context.Context,
		keyForEntity *datastore.Key) error
}

// KeyWithIfNameExists returns a key for the given name along with
// whether that name already in the datastore for the kind belonging
// to the client, along with an error if any problem was encountered.
func KeyWithIfNameExists(
	executionContext context.Context,
	limitedClient LimitedClient,
	nameToCheck string) (*datastore.Key, bool, error) {
	keyForName := limitedClient.KeyFor(nameToCheck)

	queryForNameAlreadyExists :=
		datastore.NewQuery("").Filter("__key__ =", keyForName).KeysOnly()

	resultIterator :=
		limitedClient.Run(
			executionContext,
			queryForNameAlreadyExists)

	// If there is nothing already with the given name, the iterator
	// should immediately return an iterator.Done "error".
	var keyOfExistingName datastore.Key
	_, errorFromInitialNext := resultIterator.Next(&keyOfExistingName)

	if errorFromInitialNext == iterator.Done {
		return keyForName, false, nil
	}

	// Otherwise any error means that there was an actual error.
	if errorFromInitialNext != nil {
		errorFromInitialNextWithContext :=
			fmt.Errorf(
				"Trying to check for existing name %v produced error: %v",
				nameToCheck,
				errorFromInitialNext)
		return nil, false, errorFromInitialNextWithContext
	}

	// If the first .Next(...) had no error, we check that the next
	// invocation gives iterator.Done - otherwise we report an error.
	_, errorFromSecondNext := resultIterator.Next(&keyOfExistingName)

	if errorFromSecondNext != iterator.Done {
		errorFromSecondNextWithContext :=
			fmt.Errorf(
				"Existing name %v was found but then checking for iterator.Done"+
					" produced error: %v",
				nameToCheck,
				errorFromSecondNext)

		return nil, true, errorFromSecondNextWithContext
	}

	return keyForName, true, nil
}

// WrappingLimitedClient wraps a Google Cloud Datastore client in order
// to implement an interface which uses a subset of the possible functions
// from a datastore.Client, in order to make it easier to abstract.
type WrappingLimitedClient struct {
	wrappedInterface *datastore.Client
	keyKind          string
}

// WrapDatastoreClient returns a pointer to a WrappingLimitedClient
// wrapped around the given datastore client.
func WrapDatastoreClient(
	wrappedInterface *datastore.Client,
	keyKind string) *WrappingLimitedClient {
	return &WrappingLimitedClient{
		wrappedInterface: wrappedInterface,
		keyKind:          keyKind,
	}
}

// KeyFor creates a key out of the given entity name with the stored key
// kind, without specifying an ancestor.
func (wrappingClient *WrappingLimitedClient) KeyFor(
	nameForKey string) *datastore.Key {
	return datastore.NameKey(wrappingClient.keyKind, nameForKey, nil)
}

// Run implements a wrapper for the Google Cloud Datastore client's Run
// function.
func (wrappingClient *WrappingLimitedClient) Run(
	executionContext context.Context,
	queryToRun *datastore.Query) LimitedIterator {
	return wrappingClient.wrappedInterface.Run(
		executionContext,
		queryToRun)
}

// Get implements a wrapper for the Google Cloud Datastore client's Get
// function.
func (wrappingClient *WrappingLimitedClient) Get(
	executionContext context.Context,
	keyForEntity *datastore.Key,
	deserializationDestination interface{}) (err error) {
	return wrappingClient.wrappedInterface.Get(
		executionContext,
		keyForEntity,
		deserializationDestination)
}

// Put implements a wrapper for the Google Cloud Datastore client's Put
// function.
func (wrappingClient *WrappingLimitedClient) Put(
	executionContext context.Context,
	keyForEntity *datastore.Key,
	deserializationSource interface{}) (*datastore.Key, error) {
	return wrappingClient.wrappedInterface.Put(
		executionContext,
		keyForEntity,
		deserializationSource)
}

// Delete implements a wrapper for the Google Cloud Datastore client's Delete
// (taking a string and making a key out of it).
func (wrappingClient *WrappingLimitedClient) Delete(
	executionContext context.Context,
	keyForEntity *datastore.Key) error {
	return wrappingClient.wrappedInterface.Delete(
		executionContext,
		keyForEntity)
}
