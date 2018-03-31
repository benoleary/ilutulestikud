package server

import (
	"encoding/base32"
	"encoding/base64"
)

// EndpointSegmentTranslator defines the interface for structs which should be
// able to encode a name as an identifier to be a segment in a URI, in a way so
// that it does not generate problematic characters for URIs (especially '/').
type EndpointSegmentTranslator interface {
	// ToSegment should return the name encoded as an identifier for interaction between
	// frontend and backend.
	ToSegment(nameToEncode string) string

	// FromSegment should return the name decoded from an identifier for interaction
	// between frontend and backend.
	FromSegment(segmentToDecode string) (string, error)
}

// Base32Translator provides an easy way to encode names to base-32 representation of
// the bytes of the name's characters. (Base64 uses '/' which is incompatible with using
// '/' as a URI segment delimiter and using identifiers as URI segments.)
type Base32Translator struct {
}

// ToSegment encodes the name as a base-32 representation of the bytes of the name.
func (segmentTranslator *Base32Translator) ToSegment(nameToEncode string) string {
	return base32.StdEncoding.EncodeToString([]byte(nameToEncode))
}

// FromSegment decodes the name from a base-32 representation of the bytes of the name.
func (segmentTranslator *Base32Translator) FromSegment(segmentToDecode string) (string, error) {
	return decodeBytesIfNoError(base32.StdEncoding.DecodeString(segmentToDecode))
}

// Base64Translator provides an easy way to encode names to base-64 representation of
// the bytes of the name's characters. It mainly exists for tests demonstrating that
// the handler classes forbid identifiers with '/' in them, and base-64
type Base64Translator struct {
}

// ToSegment encodes the name as a base-64 representation of the bytes of the name.
func (segmentTranslator *Base64Translator) ToSegment(nameToEncode string) string {
	return base64.StdEncoding.EncodeToString([]byte(nameToEncode))
}

// FromSegment decodes the name from a base-64 representation of the bytes of the name.
func (segmentTranslator *Base64Translator) FromSegment(segmentToDecode string) (string, error) {
	return decodeBytesIfNoError(base64.StdEncoding.DecodeString(segmentToDecode))
}

func decodeBytesIfNoError(decodedBytes []byte, decodingError error) (string, error) {
	if decodingError != nil {
		return "error", decodingError
	}

	return string(decodedBytes), nil
}
