package server_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/server"
)

func TestDecodedEncodingIsInvariant(unitTest *testing.T) {
	testCases := []struct {
		testName            string
		translatorReference server.EndpointSegmentTranslator
	}{
		{
			testName:            "Base32",
			translatorReference: &server.Base32Translator{},
		},
		{
			testName:            "Base64",
			translatorReference: &server.Base64Translator{},
		},
		{
			testName:            "No-operation",
			translatorReference: &server.NoOperationTranslator{},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			originalString := "test/string/with/odd/characters?#"
			encodedString := testCase.translatorReference.ToSegment(originalString)
			decodedString, decodingError := testCase.translatorReference.FromSegment(encodedString)

			if decodingError != nil {
				unitTest.Fatalf(
					testCase.testName+"/decoding %v produced an error: %v",
					encodedString,
					decodingError)
			}

			if decodedString != originalString {
				unitTest.Fatalf(
					testCase.testName+"/encoding to %v then decoding to %v did not match original %v",
					encodedString,
					decodedString,
					originalString)
			}
		})
	}
}

func TestDecodingStringWithInvalidCharacterProducesError(unitTest *testing.T) {
	// Only the base-32 and base-64 translators get tested as the no-operation string
	// does not have any disallowed characters.
	testCases := []struct {
		testName            string
		translatorReference server.EndpointSegmentTranslator
	}{
		{
			testName:            "Base32",
			translatorReference: &server.Base32Translator{},
		},
		{
			testName:            "Base64",
			translatorReference: &server.Base64Translator{},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			invalidEncodedString := "\""
			_, decodingError := testCase.translatorReference.FromSegment(invalidEncodedString)

			if decodingError == nil {
				unitTest.Fatalf(
					testCase.testName+"/decoding invalid %v produced no error",
					invalidEncodedString,
					decodingError)
			}
		})
	}
}