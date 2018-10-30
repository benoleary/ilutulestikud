package cloud

import (
	"context"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

// WrappingLimitedIterator wraps an iterator in order to make it more portable.
type WrappingLimitedIterator struct {
	wrappedInterface *datastore.Iterator
}

// DeserializeNext deserializes the next element in the iterator
// into the given interface.
func (wrappingIterator *WrappingLimitedIterator) DeserializeNext(
	deserializationDestination interface{}) error {
	_, errorFromDeserialization :=
		wrappingIterator.wrappedInterface.Next(deserializationDestination)

	return errorFromDeserialization
}

// NextKey simply moves onto the next element, discarding whatever
// the iterator was pointing at.
func (wrappingIterator *WrappingLimitedIterator) NextKey() error {
	var keyToBeIgnored datastore.Key
	_, errorFromDeserialization :=
		wrappingIterator.wrappedInterface.Next(&keyToBeIgnored)

	return errorFromDeserialization
}

// IsDone returns true if the given error matches the error used
// to denote that the iterator is done.
func (wrappingIterator *WrappingLimitedIterator) IsDone(
	errorFromLastNext error) bool {
	return errorFromLastNext == iterator.Done
}

// WrappingLimitedClient wraps a Google Cloud Datastore client in order
// to implement an interface which uses a subset of the possible functions
// from a datastore.Client, in order to make it easier to abstract.
type WrappingLimitedClient struct {
	wrappedInterface *datastore.Client
	keyKind          string
}

// AllOfKind returns an iterator to the set of all entities of the
// kind known to the client.
func (wrappingClient *WrappingLimitedClient) AllOfKind(
	executionContext context.Context) LimitedIterator {
	queryForAllOfKind := datastore.NewQuery(wrappingClient.keyKind)

	resultIterator := wrappingClient.wrappedInterface.Run(
		executionContext,
		queryForAllOfKind)

	return &WrappingLimitedIterator{wrappedInterface: resultIterator}
}

// AllKeysMatching returns an iterator to the set of all keys of the
// kind known to the client which match the given name.
func (wrappingClient *WrappingLimitedClient) AllKeysMatching(
	executionContext context.Context,
	keyName string) LimitedIterator {
	keyForName := wrappingClient.keyForKind(keyName)
	queryForNameAlreadyExists :=
		datastore.NewQuery("").Filter("__key__ =", keyForName).KeysOnly()

	resultIterator := wrappingClient.wrappedInterface.Run(
		executionContext,
		queryForNameAlreadyExists)

	return &WrappingLimitedIterator{wrappedInterface: resultIterator}
}

// AllMatching returns an iterator to the set of entities which are
// selected by the given filter, according the rules of the Google
// Cloud Datastore (which can have surprising results when searching
// through entities which have arrays).
func (wrappingClient *WrappingLimitedClient) AllMatching(
	executionContext context.Context,
	filterExpression string,
	valueToMatch interface{}) LimitedIterator {
	queryOnMatchingValue :=
		datastore.NewQuery(wrappingClient.keyKind).Filter(
			filterExpression,
			valueToMatch)

	resultIterator :=
		wrappingClient.wrappedInterface.Run(
			executionContext,
			queryOnMatchingValue)

	return &WrappingLimitedIterator{wrappedInterface: resultIterator}
}

// Get implements a wrapper for the Google Cloud Datastore client's Get
// function.
func (wrappingClient *WrappingLimitedClient) Get(
	executionContext context.Context,
	nameForKey string,
	deserializationDestination interface{}) error {
	return wrappingClient.wrappedInterface.Get(
		executionContext,
		wrappingClient.keyForKind(nameForKey),
		deserializationDestination)
}

// Put implements a wrapper for the Google Cloud Datastore client's Put
// function.
func (wrappingClient *WrappingLimitedClient) Put(
	executionContext context.Context,
	nameForKey string,
	deserializationSource interface{}) error {
	_, errorFromPut := wrappingClient.wrappedInterface.Put(
		executionContext,
		wrappingClient.keyForKind(nameForKey),
		deserializationSource)

	return errorFromPut
}

// Delete implements a wrapper for the Google Cloud Datastore client's Delete
// (taking a string and making a key out of it).
func (wrappingClient *WrappingLimitedClient) Delete(
	executionContext context.Context,
	nameForKey string) error {
	return wrappingClient.wrappedInterface.Delete(
		executionContext,
		wrappingClient.keyForKind(nameForKey))
}

// keyForKind makes a basic key for the given name, for the client's kind.
func (wrappingClient *WrappingLimitedClient) keyForKind(
	nameForKey string) *datastore.Key {
	return datastore.NameKey(wrappingClient.keyKind, nameForKey, nil)
}

// FixedProjectAndKeyDatastoreClientProvider creates new datastore.Client objects.
type FixedProjectAndKeyDatastoreClientProvider struct {
	projectIdentifier string
	keyKind           string
}

// NewFixedProjectAndKeyDatastoreClientProvider returns a ClientProvider which
// creates clients for the given project in the Google Cloud Datastore.
func NewFixedProjectAndKeyDatastoreClientProvider(
	projectIdentifier string,
	keyKind string) DatastoreClientProvider {
	return &FixedProjectAndKeyDatastoreClientProvider{
		projectIdentifier: projectIdentifier,
		keyKind:           keyKind,
	}
}

// NewIlutulestikudDatastoreClientProvider returns a ClientProvider which
// creates clients for the Ilutulestikud project in the Google Cloud Datastore.
func NewIlutulestikudDatastoreClientProvider(keyKind string) DatastoreClientProvider {
	return NewFixedProjectAndKeyDatastoreClientProvider(
		IlutulestikudIdentifier,
		keyKind)
}

// NewClient wraps a WrappingLimitedClient around a new datastore.Client object.
func (clientProvider *FixedProjectAndKeyDatastoreClientProvider) NewClient(
	executionContext context.Context) (LimitedClient, error) {
	cloudDatastoreClient, errorFromCloudDatastore :=
		datastore.NewClient(executionContext, clientProvider.projectIdentifier)

	if errorFromCloudDatastore != nil {
		return nil, errorFromCloudDatastore
	}

	wrappedClient := &WrappingLimitedClient{
		wrappedInterface: cloudDatastoreClient,
		keyKind:          clientProvider.keyKind,
	}

	return wrappedClient, nil
}
