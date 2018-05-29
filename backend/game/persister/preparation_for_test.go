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

func expectNoChanges(
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

func (expectedGame expectedState) assertAsExpected(
	unitTest *testing.T,
	actualGame game.ReadonlyState) {
	if (actualGame.Name() != expectedGame.Name) ||
		(actualGame.Ruleset() != expectedGame.Ruleset) ||
		(actualGame.CreationTime() != expectedGame.CreationTime) ||
		(actualGame.Turn() != expectedGame.Turn) ||
		(actualGame.Score() != expectedGame.Score) ||
		(actualGame.NumberOfReadyHints() != expectedGame.NumberOfReadyHints) ||
		(actualGame.NumberOfMistakesMade() != expectedGame.NumberOfMistakesMade) ||
		(actualGame.DeckSize() != expectedGame.DeckSize) ||
		(len(actualGame.PlayerNames()) != len(expectedGame.PlayerNames)) ||
		(len(actualGame.ChatLog()) != len(expectedGame.ChatLog)) ||
		(len(actualGame.ActionLog()) != len(expectedGame.ActionLog)) {
		unitTest.Fatalf(
			"actual %+v did not match expected %+v in easy comparisons",
			actualGame,
			expectedGame)
	}

	for chatIndex, chatMessage := range actualGame.ChatLog() {
		if chatMessage != expectedGame.ChatLog[chatIndex] {
			unitTest.Fatalf(
				"actual %+v did not match expected %+v in chat log",
				actualGame,
				expectedGame)
		}
	}

	for actionIndex, actionMessage := range actualGame.ActionLog() {
		if actionMessage != expectedGame.ActionLog[actionIndex] {
			unitTest.Fatalf(
				"actual %+v did not match expected %+v in action log",
				actualGame,
				expectedGame)
		}
	}

	for _, colorSuit := range actualGame.Ruleset().ColorSuits() {
		actualPlayedCards := actualGame.PlayedForColor(colorSuit)
		expectedPlayedCards := expectedGame.PlayedForColor[colorSuit]
		if len(actualPlayedCards) != len(expectedPlayedCards) {
			unitTest.Fatalf(
				"actual %+v did not match expected %+v in PlayedForColor",
				actualGame,
				expectedGame)
		}

		for cardIndex, actualPlayedCard := range actualPlayedCards {
			if actualPlayedCard != expectedPlayedCards[cardIndex] {
				unitTest.Fatalf(
					"actual %+v did not match expected %+v in PlayedForColor",
					actualGame,
					expectedGame)
			}
		}

		for _, sequenceIndex := range actualGame.Ruleset().DistinctPossibleIndices() {
			actualNumberOfDiscardedCopies :=
				actualGame.NumberOfDiscardedCards(colorSuit, sequenceIndex)
			discardedCard := card.NewReadonly(colorSuit, sequenceIndex)
			expectedNumberOfDiscardedCopies :=
				expectedGame.NumberOfDiscardedCards[discardedCard]

			if actualNumberOfDiscardedCopies != expectedNumberOfDiscardedCopies {
				unitTest.Fatalf(
					"actual %+v did not match expected %+v in NumberOfDiscardedCards",
					actualGame,
					expectedGame)
			}
		}
	}

	handSize :=
		actualGame.Ruleset().NumberOfCardsInPlayerHand(len(actualGame.PlayerNames()))

	for playerIndex, playerName := range actualGame.PlayerNames() {
		if playerName != expectedGame.PlayerNames[playerIndex] {
			unitTest.Fatalf(
				"actual %+v did not match expected %+v in player names",
				actualGame,
				expectedGame)
		}

		expectedVisibleHand := expectedGame.VisibleCardInHand[playerName]
		expectedInferredHand := expectedGame.InferredCardInHand[playerName]

		for handIndex := 0; handIndex < handSize; handIndex++ {
			visibleCard, errorFromVisible :=
				actualGame.VisibleCardInHand(playerName, handIndex)

			if errorFromVisible != nil {
				unitTest.Fatalf(
					"VisibleCardInHand(%+v, %+v) produced error %v",
					playerName,
					handIndex,
					errorFromVisible)
			}

			if visibleCard != expectedVisibleHand[handIndex] {
				unitTest.Fatalf(
					"actual %+v did not match expected %+v in visible hands",
					actualGame,
					expectedGame)
			}

			inferredCard, errorFromInferred :=
				actualGame.InferredCardInHand(playerName, handIndex)

			if errorFromInferred != nil {
				unitTest.Fatalf(
					"InferredCardInHand(%+v, %+v) produced error %v",
					playerName,
					handIndex,
					errorFromInferred)
			}

			expectedInferred := expectedInferredHand[handIndex]
			expectedColors := expectedInferred.PossibleColors()
			expectedIndices := expectedInferred.PossibleIndices()

			if (len(inferredCard.PossibleColors()) != len(expectedColors)) ||
				(len(inferredCard.PossibleIndices()) != len(expectedIndices)) {
				unitTest.Fatalf(
					"actual %+v did not match expected %+v in inferred hands",
					actualGame,
					expectedGame)
			}

			for colorIndex, actualColor := range inferredCard.PossibleColors() {
				if actualColor != expectedColors[colorIndex] {
					unitTest.Fatalf(
						"actual %+v did not match expected %+v in inferred hand colors",
						actualGame,
						expectedGame)
				}
			}

			for indexIndex, actualIndex := range inferredCard.PossibleIndices() {
				if actualIndex != expectedIndices[indexIndex] {
					unitTest.Fatalf(
						"actual %+v did not match expected %+v in inferred hand indices",
						actualGame,
						expectedGame)
				}
			}
		}
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
