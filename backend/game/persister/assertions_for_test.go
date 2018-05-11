package persister_test

import (
	"testing"
)

func assertPlayersMatchNames(
	testIdentifier string,
	unitTest *testing.T,
	expectedPlayerNames map[string]bool,
	actualPlayerNames []string) {
	if len(actualPlayerNames) != len(expectedPlayerNames) {
		unitTest.Fatalf(
			testIdentifier+"/expected players %v, actual list %v",
			expectedPlayerNames,
			actualPlayerNames)
	}

	for _, actualName := range actualPlayerNames {
		if !expectedPlayerNames[actualName] {
			unitTest.Fatalf(
				testIdentifier+"/expected players %v, actual list %v",
				expectedPlayerNames,
				actualPlayerNames)
		}
	}
}

func assertStringSlicesMatch(
	testIdentifier string,
	unitTest *testing.T,
	expectedSlice []string,
	actualSlice []string) {
	numberOfExpected := len(expectedSlice)
	if len(expectedSlice) != numberOfExpected {
		unitTest.Fatalf(
			testIdentifier+"/actual %v did not match expected %v",
			actualSlice,
			expectedSlice)
	}

	for sliceIndex := 0; sliceIndex < numberOfExpected; sliceIndex++ {
		expectedString := expectedSlice[sliceIndex]
		actualString := actualSlice[sliceIndex]
		if actualString != expectedString {
			unitTest.Fatalf(
				testIdentifier+"/actual %v did not match expected %v",
				actualSlice,
				expectedSlice)
		}
	}
}

func assertIntSlicesMatch(
	testIdentifier string,
	unitTest *testing.T,
	expectedSlice []int,
	actualSlice []int) {
	numberOfExpected := len(expectedSlice)
	if len(expectedSlice) != numberOfExpected {
		unitTest.Fatalf(
			testIdentifier+"/actual %v did not match expected %v",
			actualSlice,
			expectedSlice)
	}

	for sliceIndex := 0; sliceIndex < numberOfExpected; sliceIndex++ {
		expectedString := expectedSlice[sliceIndex]
		actualString := actualSlice[sliceIndex]
		if actualString != expectedString {
			unitTest.Fatalf(
				testIdentifier+"/actual %v did not match expected %v",
				actualSlice,
				expectedSlice)
		}
	}
}
