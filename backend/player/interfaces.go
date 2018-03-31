package player

// ReadonlyState defines the interface for structs which should encapsulate the state
// of a player which can be read but not written.
type ReadonlyState interface {
	// Name should return the name of the player as known to the games.
	Name() string

	// Color should return the color that the player uses for chat messages.
	Color() string
}

// StatePersister defines the interface for structs which should be able to create
// objects implementing the ReadOnly interface out of instances of endpoint.PlayerState,
// which normally represents de-serialized JSON from the frontend, and for tracking the
// objects by their identifier, which is the player name encoded in a way compatible with
// the identifier being a segment of a URI with '/' as a delimiter.
type StatePersister interface {
	// add should add an element to the collection which is a new object implementing
	// the ReadonlyState interface with information given by the arguments. If there was
	// no problem, the returned error should be nil. It should return an error if the
	// player already exists.
	add(playerName string, chatColor string) error

	// updateColor should update the given player to have the given chat color.
	// This should be thread-safe. It should return an error if there was a problem,
	// including if the player is not registered.
	updateColor(playerName string, chatColor string) error

	// get should return the ReadOnly corresponding to the given player identifier if it
	// exists already along with an error which of course should be nil if there was no
	// problem. If the player does not exist, a non-nil error should be returned along with
	// a nil ReadOnly.
	get(playerIdentifier string) (ReadonlyState, error)

	// all should return a slice of all the State instances in the collection. The order
	// is not mandated, and may even change with repeated calls to the same unchanged
	// Collection (analogously to the entry set of a standard Golang map, for example),
	// though of course an implementation may order the slice consistently.
	all() []ReadonlyState

	// reset should remove all players.
	reset()
}
