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
// objects implementing the ReadOnly interface out of player names with colors.
type StatePersister interface {
	// All should return a slice of all the State instances in the persistence store.
	// The order is not mandated, and may even change with repeated calls to the same
	// unchanged persistence store (analogously to the entry set of a standard Golang
	// map, for example), though of course an implementation may order the slice
	// consistently.
	All() []ReadonlyState

	// Get should return the read-only state corresponding to the given player name if it
	// exists already along with an error which of course should be nil if there was no
	// problem. If the player does not exist, a non-nil error should be returned along with
	// nil for the read-only state.
	Get(playerName string) (ReadonlyState, error)

	// Add should add an element to the persistence store which is a new object
	// implementing the ReadonlyState interface with information given by the arguments. If
	// there was no problem, the returned error should be nil. It should return an error if
	// the player already exists.
	Add(playerName string, chatColor string) error

	// UpdateColor should update the given player to have the given chat color.
	// This should be thread-safe. It should return an error if there was a problem,
	// including if the player is not registered.
	UpdateColor(playerName string, chatColor string) error

	// Delete should delete the given player from the persistence store.
	Delete(playerName string) error

	// Reset should remove all players.
	Reset()
}
