package persister_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/persister"
	"github.com/benoleary/ilutulestikud/backend/player"
)

const testGameName = "test game"

var defaultTestRuleset game.Ruleset = &game.StandardWithoutRainbowRuleset{}

type mockPlayerState struct {
	mockName  string
	mockColor string
}

func (mockState *mockPlayerState) Name() string {
	return mockState.mockName
}

func (mockState *mockPlayerState) Color() string {
	return mockState.mockColor
}

var defaultTestPlayers []player.ReadonlyState = []player.ReadonlyState{
	&mockPlayerState{
		mockName:  "Player One",
		mockColor: "color one",
	},
	&mockPlayerState{
		mockName:  "Player Two",
		mockColor: "color two",
	},
	&mockPlayerState{
		mockName:  "Player Three",
		mockColor: "color three",
	},
}

func playerNameSet(playerStates []player.ReadonlyState) map[string]bool {
	playerNames := make(map[string]bool, 0)
	for _, playerState := range playerStates {
		playerNames[playerState.Name()] = true
	}

	return playerNames
}

type persisterAndDescription struct {
	GamePersister        game.StatePersister
	PersisterDescription string
}

func preparePersisters() []persisterAndDescription {
	return []persisterAndDescription{
		persisterAndDescription{
			GamePersister:        persister.NewInMemory(),
			PersisterDescription: "in-memory persister",
		},
	}
}

type gameAndDescription struct {
	GameState            game.ReadAndWriteState
	PersisterDescription string
}

func prepareGameStates(
	unitTest *testing.T,
	gameRuleset game.Ruleset,
	playerStates []player.ReadonlyState,
	initialDeck []card.Readonly) []gameAndDescription {
	statePersisters := preparePersisters()

	numberOfPersisters := len(statePersisters)

	gamesAndDescriptions := make([]gameAndDescription, numberOfPersisters)

	for persisterIndex := 0; persisterIndex < numberOfPersisters; persisterIndex++ {
		statePersister := statePersisters[persisterIndex]
		errorFromAdd :=
			statePersister.GamePersister.AddGame(
				testGameName,
				gameRuleset,
				playerStates,
				initialDeck)

		if errorFromAdd != nil {
			unitTest.Fatalf("Error when adding game: %v", errorFromAdd)
		}

		gameState, errorFromGet :=
			statePersister.GamePersister.ReadAndWriteGame(testGameName)

		if errorFromGet != nil {
			unitTest.Fatalf("Error when getting game: %v", errorFromGet)
		}

		gamesAndDescriptions[persisterIndex] =
			gameAndDescription{
				GameState:            gameState,
				PersisterDescription: statePersister.PersisterDescription,
			}
	}

	return gamesAndDescriptions
}
