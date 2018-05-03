package game_test

import (
	"bytes"
	"encoding/json"
	"testing"
)

// This file is duplicated as both the player and game packages make use of
// it for their tests, but it should not be exported as part of non-test code,
// yet it is impossible to import test-only packages.

func DecoderAroundInterface(
	unitTest *testing.T,
	testIdentifier string,
	jsonBody interface{}) *json.Decoder {
	bytesBuffer := new(bytes.Buffer)
	errorFromEncoding := json.NewEncoder(bytesBuffer).Encode(jsonBody)

	if errorFromEncoding != nil {
		unitTest.Fatalf(
			testIdentifier+"/encoding %v as JSON generated an error: %v",
			jsonBody,
			errorFromEncoding)
	}

	return json.NewDecoder(bytes.NewReader(bytesBuffer.Bytes()))
}
