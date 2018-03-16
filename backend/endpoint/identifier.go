package endpoint

import (
	"encoding/base32"
	"encoding/base64"
)

// NameToIdentifier defines the interface for structs which should be able to encode a
// name as an identifier which does not generate problematic characters for URIs
// (especially '/').
type NameToIdentifier interface {
	// Identifier should return the name encoded as an identifier for interaction between
	// frontend and backend.
	Identifier(name string) string
}

// Base64NameEncoder provides an easy way to encode names to base-64 representation of
// the bytes of the name's characters.
type Base64NameEncoder struct {
}

// Identifier encodes the name as a base-32 representation of the bytes of the name.
func (base64NameEncoder *Base64NameEncoder) Identifier(name string) string {
	return base64.StdEncoding.EncodeToString([]byte(name))
}

// Base32NameEncoder provides an easy way to encode names to base-32 representation of
// the bytes of the name's characters. (Base64 uses '/' which is incompatible with using
// '/' as a URI segment delimiter and using identifiers as URI segments.)
type Base32NameEncoder struct {
}

// Identifier encodes the name as a base-32 representation of the bytes of the name.
func (base32NameEncoder *Base32NameEncoder) Identifier(name string) string {
	return base32.StdEncoding.EncodeToString([]byte(name))
}
