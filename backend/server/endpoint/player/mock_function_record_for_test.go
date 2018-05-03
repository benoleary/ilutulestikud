package player_test

import (
	"testing"
)

// This file is duplicated as both the player and game packages make use of
// it for their tests, but it should not be exported as part of non-test code,
// yet it is impossible to import test-only packages.

// functionNameAndArgument is for the mock collections to use to record what was
// asked of them.
type functionNameAndArgument struct {
	FunctionName     string
	FunctionArgument interface{}
}

type stringPair struct {
	first  string
	second string
}

type stringTriple struct {
	first  string
	second string
	third  string
}

func assertNoFunctionWasCalled(
	unitTest *testing.T,
	actualRecords []functionNameAndArgument,
	testIdentifier string) {
	if len(actualRecords) != 0 {
		unitTest.Fatalf(
			testIdentifier+": unexpectedly called mock collection methods %v",
			actualRecords)
	}
}

func assertFunctionRecordIsCorrect(
	unitTest *testing.T,
	actualRecord functionNameAndArgument,
	expectedRecord functionNameAndArgument,
	testIdentifier string) {
	if actualRecord != expectedRecord {
		unitTest.Fatalf(
			testIdentifier+"/function record mismatch: actual = %v, expected = %v",
			actualRecord,
			expectedRecord)
	}
}

func assertFunctionRecordsAreCorrect(
	unitTest *testing.T,
	actualRecords []functionNameAndArgument,
	expectedRecords []functionNameAndArgument,
	testIdentifier string) {
	expectedNumberOfRecords := len(expectedRecords)

	if len(actualRecords) != expectedNumberOfRecords {
		unitTest.Fatalf(
			testIdentifier+"/function record list length mismatch: actual = %v, expected = %v",
			actualRecords,
			expectedRecords)
	}

	for recordIndex := 0; recordIndex < expectedNumberOfRecords; recordIndex++ {
		actualRecord := actualRecords[recordIndex]
		expectedRecord := expectedRecords[recordIndex]
		if actualRecord != expectedRecord {
			unitTest.Fatalf(
				testIdentifier+
					"/function record[%v] mismatch: actual = %v, expected = %v (list: actual = %v, expected = %v)",
				recordIndex,
				actualRecord,
				expectedRecord,
				actualRecords,
				expectedRecords)
		}
	}
}
