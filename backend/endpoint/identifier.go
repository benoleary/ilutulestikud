package endpoint

// NameToIdentifier defines the interface for structs which should be able to encode a
// name as an identifier.
type NameToIdentifier interface {
	// Identifier should return the name encoded as an identifier for interaction between
	// frontend and backend.
	Identifier(name string) string

	// Name should return the identifier decoded to name for interaction between frontend
	// and backend.
	Name(identifier string) string
}
