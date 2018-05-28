package persister_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
	"github.com/benoleary/ilutulestikud/backend/game/persister"
)

const testGameName = "test game"
const logLengthForTest = 8

var defaultTestRuleset game.Ruleset = game.NewStandardWithoutRainbow()

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

var defaultTestPlayers []string = []string{
	"Player One",
	"Player Two",
	"Player Three",
}

func playerNameSet(namesWithHands []game.PlayerNameWithHand) map[string]bool {
	playerNameMap := make(map[string]bool, 0)
	for _, nameWithHand := range namesWithHands {
		playerNameMap[nameWithHand.PlayerName] = true
	}

	return playerNameMap
}

var threePlayersWithHands = []game.PlayerNameWithHand{
	game.PlayerNameWithHand{
		PlayerName: defaultTestPlayers[0],
		InitialHand: []card.InHand{
			card.InHand{
				Readonly: card.NewReadonly("a", 1),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c"},
					[]int{1, 2, 3}),
			},
			card.InHand{
				Readonly: card.NewReadonly("a", 1),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c"},
					[]int{1, 2, 3}),
			},
			card.InHand{
				Readonly: card.NewReadonly("a", 2),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c"},
					[]int{1, 2, 3}),
			},
			card.InHand{
				Readonly: card.NewReadonly("a", 1),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c"},
					[]int{1, 2, 3}),
			},
			card.InHand{
				Readonly: card.NewReadonly("a", 2),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c"},
					[]int{1, 2, 3}),
			},
		},
	},
	game.PlayerNameWithHand{
		PlayerName: defaultTestPlayers[1],
		InitialHand: []card.InHand{
			card.InHand{
				Readonly: card.NewReadonly("a", 1),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c"},
					[]int{1, 2, 3, 4}),
			},
			card.InHand{
				Readonly: card.NewReadonly("b", 1),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c", "d"},
					[]int{1, 2, 3}),
			},
			card.InHand{
				Readonly: card.NewReadonly("b", 2),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c"},
					[]int{1, 2, 3}),
			},
			card.InHand{
				Readonly: card.NewReadonly("c", 2),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c"},
					[]int{1, 2, 3}),
			},
			card.InHand{
				Readonly: card.NewReadonly("c", 3),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c"},
					[]int{1, 2, 3}),
			},
		},
	},
	game.PlayerNameWithHand{
		PlayerName: defaultTestPlayers[2],
		InitialHand: []card.InHand{
			card.InHand{
				Readonly: card.NewReadonly("c", 3),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c"},
					[]int{1, 2, 3, 4}),
			},
			card.InHand{
				Readonly: card.NewReadonly("b", 3),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c"},
					[]int{1, 2, 3}),
			},
			card.InHand{
				Readonly: card.NewReadonly("a", 3),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c"},
					[]int{1, 2, 3}),
			},
			card.InHand{
				Readonly: card.NewReadonly("d", 3),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c", "d"},
					[]int{1, 2, 3}),
			},
			card.InHand{
				Readonly: card.NewReadonly("d", 1),
				Inferred: card.NewInferred(
					[]string{"a", "b", "c", "d"},
					[]int{1, 2, 3}),
			},
		},
	},
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
	playersInTurnOrderWithInitialHands []game.PlayerNameWithHand,
	initialDeck []card.Readonly,
	initialActionLog []message.Readonly) []gameAndDescription {
	statePersisters := preparePersisters()

	numberOfPersisters := len(statePersisters)

	gamesAndDescriptions := make([]gameAndDescription, numberOfPersisters)

	for persisterIndex := 0; persisterIndex < numberOfPersisters; persisterIndex++ {
		statePersister := statePersisters[persisterIndex]

		errorFromAdd :=
			statePersister.GamePersister.AddGame(
				testGameName,
				logLengthForTest,
				initialActionLog,
				gameRuleset,
				playersInTurnOrderWithInitialHands,
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
