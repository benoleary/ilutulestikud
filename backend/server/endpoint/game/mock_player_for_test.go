package game_test

import (
	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/player"
)

var colorsAvailableInTest []string = defaults.AvailableColors()

type mockPlayerState struct {
	name  string
	color string
}

// Name returns the private name field.
func (playerState *mockPlayerState) Name() string {
	return playerState.name
}

// Color returns the private color field.
func (playerState *mockPlayerState) Color() string {
	return playerState.color
}

var testPlayerStates []player.ReadonlyState = []player.ReadonlyState{
	&mockPlayerState{
		name:  "Player One",
		color: colorsAvailableInTest[0],
	},
	// Player Two has the same color as Player One
	&mockPlayerState{
		name:  "Player Two",
		color: colorsAvailableInTest[0],
	},
	&mockPlayerState{
		name:  "Player Three",
		color: colorsAvailableInTest[1],
	},
}
