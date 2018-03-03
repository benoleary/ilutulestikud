package player_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// This just tests that the factory method does not cause any panics, and returns a non-nil pointer.
func TestNewState(unitTest *testing.T) {
	actualState := player.NewState("New Player", "Chat color")
	if actualState == nil {
		unitTest.Fatalf("New state was nil.")
	}
}

func TestState_UpdateNonEmptyStrings(unitTest *testing.T) {
	type testArguments struct {
		updaterReference endpoint.PlayerState
	}

	testCases := []struct {
		name      string
		state     *player.OriginalState
		arguments testArguments
	}{
		{
			name:      "OverwriteColor",
			state:     player.NewState("Player Name", "Original color"),
			arguments: testArguments{updaterReference: endpoint.PlayerState{Color: "Over-written color"}},
		},
	}
	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(t *testing.T) {
			testCase.state.UpdateNonEmptyStrings(testCase.arguments.updaterReference)
			if testCase.state.Color() != testCase.arguments.updaterReference.Color {
				unitTest.Fatalf("%v: color = %v, want %v", testCase.name, testCase.state.Color(), testCase.arguments.updaterReference.Color)
			}
		})
	}
}

func Test_ForBackend(unitTest *testing.T) {
	playerState := player.NewState("Player Name", "Chat Color")
	actualState := player.ForBackend(playerState)
	expectedState := endpoint.PlayerState{Name: "Player Name", Color: "Chat Color"}

	if actualState != expectedState {
		unitTest.Fatalf("player.State.ForBackend(): actual %v; expected %v", actualState, expectedState)
	}
}
