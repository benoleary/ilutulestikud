package player_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backendjson"
	"github.com/benoleary/ilutulestikud/player"
)

// This just tests that the factory method does not cause any panics, and returns a non-nil pointer.
func TestNewState(unitTest *testing.T) {
	type testArguments struct {
		nameForNewPlayer  string
		colorForNewPlayer string
	}

	testCases := []struct {
		name      string
		arguments testArguments
	}{
		{
			name: "ConstructorDoesNotReturnNil",
			arguments: testArguments{
				nameForNewPlayer:  "New Player",
				colorForNewPlayer: "Chat color",
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			actualState := player.NewState(testCase.arguments.nameForNewPlayer, testCase.arguments.colorForNewPlayer)
			if actualState == nil {
				unitTest.Errorf("%v: new state was nil.", testCase.name)
			}
		})
	}
}

func TestState_UpdateNonEmptyStrings(unitTest *testing.T) {
	type testArguments struct {
		updaterReference backendjson.PlayerState
	}

	testCases := []struct {
		name      string
		state     *player.State
		arguments testArguments
	}{
		{
			name:      "OverwriteColor",
			state:     player.NewState("Player Name", "Original color"),
			arguments: testArguments{updaterReference: backendjson.PlayerState{Color: "Over-written color"}},
		},
	}
	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(t *testing.T) {
			testCase.state.UpdateNonEmptyStrings(testCase.arguments.updaterReference)
			if testCase.state.Color != testCase.arguments.updaterReference.Color {
				unitTest.Fatalf("%v: color = %v, want %v", testCase.name, testCase.state.Color, testCase.arguments.updaterReference.Color)
			}
		})
	}
}

func TestState_ForBackend(unitTest *testing.T) {
	testCases := []struct {
		name     string
		state    *player.State
		expected backendjson.PlayerState
	}{
		{
			name:     "OnlyCase",
			state:    player.NewState("Player Name", "Chat Color"),
			expected: backendjson.PlayerState{Name: "Player Name", Color: "Chat Color"},
		},
	}
	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			actualState := testCase.state.ForBackend()

			if actualState != testCase.expected {
				unitTest.Fatalf("%v: state.ForBackend() = %v, want %v", testCase.name, actualState, testCase.expected)
			}
		})
	}
}
