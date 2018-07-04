package persister_test

import (
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
	"github.com/benoleary/ilutulestikud/backend/game/persister"
)

const testGameName = "test game"
const logLengthForTest = 8

var defaultTestRuleset game.Ruleset = game.NewStandardWithoutRainbow()
var colorsForTest = defaultTestRuleset.ColorSuits()
var threeColors = []string{colorsForTest[0], colorsForTest[1], colorsForTest[2]}
var fourColors = []string{colorsForTest[0], colorsForTest[1], colorsForTest[2], colorsForTest[3]}
var indicesForTest = defaultTestRuleset.DistinctPossibleIndices()
var threeIndices = []int{indicesForTest[0], indicesForTest[1], indicesForTest[2]}
var fourIndices = []int{indicesForTest[0], indicesForTest[1], indicesForTest[2], indicesForTest[3]}

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
				Readonly: card.NewReadonly(
					colorsForTest[0],
					indicesForTest[0]),
				Inferred: card.NewInferred(
					threeColors,
					threeIndices),
			},
			card.InHand{
				Readonly: card.NewReadonly(
					colorsForTest[0],
					indicesForTest[0]),
				Inferred: card.NewInferred(
					threeColors,
					threeIndices),
			},
			card.InHand{
				Readonly: card.NewReadonly(
					colorsForTest[0],
					indicesForTest[1]),
				Inferred: card.NewInferred(
					threeColors,
					threeIndices),
			},
			card.InHand{
				Readonly: card.NewReadonly(
					colorsForTest[0],
					indicesForTest[0]),
				Inferred: card.NewInferred(
					threeColors,
					threeIndices),
			},
			card.InHand{
				Readonly: card.NewReadonly(
					colorsForTest[0],
					indicesForTest[1]),
				Inferred: card.NewInferred(
					threeColors,
					threeIndices),
			},
		},
	},
	game.PlayerNameWithHand{
		PlayerName: defaultTestPlayers[1],
		InitialHand: []card.InHand{
			card.InHand{
				Readonly: card.NewReadonly(
					colorsForTest[0],
					indicesForTest[0]),
				Inferred: card.NewInferred(
					threeColors,
					fourIndices),
			},
			card.InHand{
				Readonly: card.NewReadonly(
					colorsForTest[1],
					indicesForTest[0]),
				Inferred: card.NewInferred(
					fourColors,
					threeIndices),
			},
			card.InHand{
				Readonly: card.NewReadonly(
					colorsForTest[1],
					indicesForTest[1]),
				Inferred: card.NewInferred(
					threeColors,
					threeIndices),
			},
			card.InHand{
				Readonly: card.NewReadonly(
					colorsForTest[2],
					indicesForTest[1]),
				Inferred: card.NewInferred(
					threeColors,
					threeIndices),
			},
			card.InHand{
				Readonly: card.NewReadonly(
					colorsForTest[2],
					indicesForTest[2]),
				Inferred: card.NewInferred(
					threeColors,
					threeIndices),
			},
		},
	},
	game.PlayerNameWithHand{
		PlayerName: defaultTestPlayers[2],
		InitialHand: []card.InHand{
			card.InHand{
				Readonly: card.NewReadonly(
					colorsForTest[2],
					indicesForTest[2]),
				Inferred: card.NewInferred(
					threeColors,
					fourIndices),
			},
			card.InHand{
				Readonly: card.NewReadonly(
					colorsForTest[1],
					indicesForTest[2]),
				Inferred: card.NewInferred(
					threeColors,
					threeIndices),
			},
			card.InHand{
				Readonly: card.NewReadonly(
					colorsForTest[0],
					indicesForTest[2]),
				Inferred: card.NewInferred(
					threeColors,
					threeIndices),
			},
			card.InHand{
				Readonly: card.NewReadonly(
					colorsForTest[3],
					indicesForTest[2]),
				Inferred: card.NewInferred(
					fourColors,
					threeIndices),
			},
			card.InHand{
				Readonly: card.NewReadonly(
					colorsForTest[3],
					indicesForTest[0]),
				Inferred: card.NewInferred(
					fourColors,
					threeIndices),
			},
		},
	},
}

var initialActionMessagesForDefaultThreePlayers = []string{
	"initial player one action",
	"initial player two action",
	"initial player three action",
}

var defaultTestColor = "default test color"

var initialActionLogForDefaultThreePlayers = []message.Readonly{
	message.NewReadonly(
		threePlayersWithHands[0].PlayerName,
		defaultTestColor,
		initialActionMessagesForDefaultThreePlayers[0]),
	message.NewReadonly(
		threePlayersWithHands[1].PlayerName,
		defaultTestColor,
		initialActionMessagesForDefaultThreePlayers[1]),
	message.NewReadonly(
		threePlayersWithHands[2].PlayerName,
		defaultTestColor,
		initialActionMessagesForDefaultThreePlayers[2]),
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

type expectedState struct {
	Name                   string
	Ruleset                game.Ruleset
	PlayerNames            []string
	CreationTime           time.Time
	ChatLog                []message.Readonly
	ActionLog              []message.Readonly
	Turn                   int
	Score                  int
	NumberOfReadyHints     int
	NumberOfMistakesMade   int
	DeckSize               int
	PlayedForColor         map[string][]card.Readonly
	NumberOfDiscardedCards map[card.Readonly]int
	VisibleCardInHand      map[string][]card.Readonly
	InferredCardInHand     map[string][]card.Inferred
}

func prepareExpected(
	unitTest *testing.T,
	pristineState game.ReadonlyState) expectedState {
	if pristineState == nil {
		unitTest.Fatalf("nil game.ReadonlyState")
	}

	pristineRuleset := pristineState.Ruleset()

	playedCards := make(map[string][]card.Readonly, 0)
	discardedCards := make(map[card.Readonly]int, 0)
	for _, colorSuit := range pristineRuleset.ColorSuits() {
		playedCards[colorSuit] = pristineState.PlayedForColor(colorSuit)

		for _, sequenceIndex := range pristineRuleset.DistinctPossibleIndices() {
			numberOfDiscardedCopies := pristineState.NumberOfDiscardedCards(colorSuit, sequenceIndex)
			if numberOfDiscardedCopies != 0 {
				discardedCard := card.NewReadonly(colorSuit, sequenceIndex)
				discardedCards[discardedCard] = numberOfDiscardedCopies
			}
		}
	}

	visibleHands := make(map[string][]card.Readonly, 0)
	inferredHands := make(map[string][]card.Inferred, 0)

	for _, playerName := range pristineState.PlayerNames() {
		visibleHand, errorFromVisible :=
			pristineState.VisibleHand(playerName)
		if errorFromVisible != nil {
			unitTest.Fatalf(
				"VisibleHand(%v) produced error %v",
				playerName,
				errorFromVisible)
		}

		inferredHand, errorFromInferred :=
			pristineState.InferredHand(playerName)
		if errorFromInferred != nil {
			unitTest.Fatalf(
				"InferredHand(%v) produced error %v",
				playerName,
				errorFromInferred)
		}

		visibleHands[playerName] = visibleHand
		inferredHands[playerName] = inferredHand
	}

	return expectedState{
		Name:                   pristineState.Name(),
		Ruleset:                pristineRuleset,
		PlayerNames:            pristineState.PlayerNames(),
		CreationTime:           pristineState.CreationTime(),
		ChatLog:                copyLog(unitTest, pristineState.ChatLog()),
		ActionLog:              copyLog(unitTest, pristineState.ActionLog()),
		Turn:                   pristineState.Turn(),
		NumberOfReadyHints:     pristineState.NumberOfReadyHints(),
		NumberOfMistakesMade:   pristineState.NumberOfMistakesMade(),
		DeckSize:               pristineState.DeckSize(),
		PlayedForColor:         playedCards,
		NumberOfDiscardedCards: discardedCards,
		VisibleCardInHand:      visibleHands,
		InferredCardInHand:     inferredHands,
	}
}

func copyLog(
	unitTest *testing.T,
	sourceLog []message.Readonly) []message.Readonly {
	numberOfMessages := len(sourceLog)
	copiedLog := make([]message.Readonly, numberOfMessages)
	numberOfMessagesCopied := copy(copiedLog, sourceLog)

	if numberOfMessagesCopied != numberOfMessages {
		unitTest.Fatalf(
			"copied %v elements instead of %v elements from slice %+v",
			numberOfMessagesCopied,
			numberOfMessages,
			sourceLog)
	}

	return copiedLog
}
