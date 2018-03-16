package assertgame

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game"
)

// StateIsCorrect fails the unit test if the given game.State does not
// have the right characteristics.
func StateIsCorrect(
	unitTest *testing.T,
	expectedGameName string,
	expectedPlayers []string,
	actualGame game.State,
	testIdentifier string) {
	if actualGame.Name() != expectedGameName {
		unitTest.Fatalf(
			testIdentifier+": game %v was found but had name %v.",
			expectedGameName,
			actualGame.Name())
	}

	actualPlayers := actualGame.Players()
	playerSlicesMatch := (len(actualPlayers) == len(expectedPlayers))

	if playerSlicesMatch {
		for playerIndex := 0; playerIndex < len(actualPlayers); playerIndex++ {
			playerSlicesMatch =
				(actualPlayers[playerIndex].Identifier() == expectedPlayers[playerIndex])
			if !playerSlicesMatch {
				break
			}
		}
	}

	if !playerSlicesMatch {
		unitTest.Fatalf(
			testIdentifier+": game %v was found but had players %v instead of expected %v.",
			expectedGameName,
			actualPlayers,
			expectedPlayers)
	}
}
