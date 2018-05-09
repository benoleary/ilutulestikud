package persister_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/player"
)

func assertPlayersMatchNames(
	testIdentifier string,
	unitTest *testing.T,
	expectedPlayerNames map[string]bool,
	actualPlayerStates []player.ReadonlyState) {
	if len(actualPlayerStates) != len(expectedPlayerNames) {
		unitTest.Fatalf(
			testIdentifier+"/expected players %v, actual list %v",
			expectedPlayerNames,
			actualPlayerStates)
	}

	for _, actualPlayer := range actualPlayerStates {
		if !expectedPlayerNames[actualPlayer.Name()] {
			unitTest.Fatalf(
				testIdentifier+"/expected players %v, actual list %v",
				expectedPlayerNames,
				actualPlayerStates)
		}
	}
}
