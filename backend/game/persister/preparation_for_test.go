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

	handSize := pristineRuleset.NumberOfCardsInPlayerHand(len(pristineState.PlayerNames()))
	visibleHands := make(map[string][]card.Readonly, 0)
	inferredHands := make(map[string][]card.Inferred, 0)

	for _, playerName := range pristineState.PlayerNames() {
		visibleHand := make([]card.Readonly, 0)
		inferredHand := make([]card.Inferred, 0)

		for handIndex := 0; handIndex < handSize; handIndex++ {
			visibleCard, errorFromVisible :=
				pristineState.VisibleCardInHand(playerName, handIndex)

			if errorFromVisible != nil {
				unitTest.Fatalf(
					"VisibleCardInHand(%+v, %+v) produced error %v",
					playerName,
					handIndex,
					errorFromVisible)
			}

			visibleHand = append(visibleHand, visibleCard)

			inferredCard, errorFromInferred :=
				pristineState.InferredCardInHand(playerName, handIndex)

			if errorFromInferred != nil {
				unitTest.Fatalf(
					"InferredCardInHand(%+v, %+v) produced error %v",
					playerName,
					handIndex,
					errorFromInferred)
			}

			inferredHand = append(inferredHand, inferredCard)
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
		Score:                  pristineState.Score(),
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
