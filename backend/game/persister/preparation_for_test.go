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
				Defined: card.Defined{
					ColorSuit:     colorsForTest[0],
					SequenceIndex: indicesForTest[0],
				},
				Inferred: card.Inferred{
					PossibleColors:  threeColors,
					PossibleIndices: threeIndices,
				},
			},
			card.InHand{
				Defined: card.Defined{
					ColorSuit:     colorsForTest[0],
					SequenceIndex: indicesForTest[0],
				},
				Inferred: card.Inferred{
					PossibleColors:  threeColors,
					PossibleIndices: threeIndices,
				},
			},
			card.InHand{
				Defined: card.Defined{
					ColorSuit:     colorsForTest[0],
					SequenceIndex: indicesForTest[1],
				},
				Inferred: card.Inferred{
					PossibleColors:  threeColors,
					PossibleIndices: threeIndices,
				},
			},
			card.InHand{
				Defined: card.Defined{
					ColorSuit:     colorsForTest[0],
					SequenceIndex: indicesForTest[0],
				},
				Inferred: card.Inferred{
					PossibleColors:  threeColors,
					PossibleIndices: threeIndices,
				},
			},
			card.InHand{
				Defined: card.Defined{
					ColorSuit:     colorsForTest[0],
					SequenceIndex: indicesForTest[1],
				},
				Inferred: card.Inferred{
					PossibleColors:  threeColors,
					PossibleIndices: threeIndices,
				},
			},
		},
	},
	game.PlayerNameWithHand{
		PlayerName: defaultTestPlayers[1],
		InitialHand: []card.InHand{
			card.InHand{
				Defined: card.Defined{
					ColorSuit:     colorsForTest[0],
					SequenceIndex: indicesForTest[0],
				},
				Inferred: card.Inferred{
					PossibleColors:  threeColors,
					PossibleIndices: fourIndices,
				},
			},
			card.InHand{
				Defined: card.Defined{
					ColorSuit:     colorsForTest[1],
					SequenceIndex: indicesForTest[0],
				},
				Inferred: card.Inferred{
					PossibleColors:  fourColors,
					PossibleIndices: threeIndices,
				},
			},
			card.InHand{
				Defined: card.Defined{
					ColorSuit:     colorsForTest[1],
					SequenceIndex: indicesForTest[1],
				},
				Inferred: card.Inferred{
					PossibleColors:  threeColors,
					PossibleIndices: threeIndices,
				},
			},
			card.InHand{
				Defined: card.Defined{
					ColorSuit:     colorsForTest[2],
					SequenceIndex: indicesForTest[1],
				},
				Inferred: card.Inferred{
					PossibleColors:  threeColors,
					PossibleIndices: threeIndices,
				},
			},
			card.InHand{
				Defined: card.Defined{
					ColorSuit:     colorsForTest[2],
					SequenceIndex: indicesForTest[2],
				},
				Inferred: card.Inferred{
					PossibleColors:  threeColors,
					PossibleIndices: threeIndices,
				},
			},
		},
	},
	game.PlayerNameWithHand{
		PlayerName: defaultTestPlayers[2],
		InitialHand: []card.InHand{
			card.InHand{
				Defined: card.Defined{
					ColorSuit:     colorsForTest[2],
					SequenceIndex: indicesForTest[2],
				},
				Inferred: card.Inferred{
					PossibleColors:  threeColors,
					PossibleIndices: fourIndices,
				},
			},
			card.InHand{
				Defined: card.Defined{
					ColorSuit:     colorsForTest[1],
					SequenceIndex: indicesForTest[2],
				},
				Inferred: card.Inferred{
					PossibleColors:  threeColors,
					PossibleIndices: threeIndices,
				},
			},
			card.InHand{
				Defined: card.Defined{
					ColorSuit:     colorsForTest[0],
					SequenceIndex: indicesForTest[2],
				},
				Inferred: card.Inferred{
					PossibleColors:  threeColors,
					PossibleIndices: threeIndices,
				},
			},
			card.InHand{
				Defined: card.Defined{
					ColorSuit:     colorsForTest[3],
					SequenceIndex: indicesForTest[2],
				},
				Inferred: card.Inferred{
					PossibleColors:  fourColors,
					PossibleIndices: threeIndices,
				},
			},
			card.InHand{
				Defined: card.Defined{
					ColorSuit:     colorsForTest[3],
					SequenceIndex: indicesForTest[0],
				},
				Inferred: card.Inferred{
					PossibleColors:  fourColors,
					PossibleIndices: threeIndices,
				},
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

var initialActionLogForDefaultThreePlayers = []message.FromPlayer{
	message.NewFromPlayer(
		threePlayersWithHands[0].PlayerName,
		defaultTestColor,
		initialActionMessagesForDefaultThreePlayers[0]),
	message.NewFromPlayer(
		threePlayersWithHands[1].PlayerName,
		defaultTestColor,
		initialActionMessagesForDefaultThreePlayers[1]),
	message.NewFromPlayer(
		threePlayersWithHands[2].PlayerName,
		defaultTestColor,
		initialActionMessagesForDefaultThreePlayers[2]),
}

type persisterAndDescription struct {
	GamePersister        game.StatePersister
	PersisterDescription string
}

func preparePersisters(unitTest *testing.T) []persisterAndDescription {
	//	inCloudDatastore, errorFromNewInCloudDatastore :=
	//		persister.NewInCloudDatastore(context.Background())
	//
	//	if errorFromNewInCloudDatastore != nil {
	//		unitTest.Fatalf(
	//			"Error when creating persister for Cloud Datastore: %v",
	//			errorFromNewInCloudDatastore)
	//	}

	return []persisterAndDescription{
		persisterAndDescription{
			GamePersister:        persister.NewInMemory(),
			PersisterDescription: "in-memory persister",
		},
		//		persisterAndDescription{
		//			GamePersister:        inCloudDatastore,
		//			PersisterDescription: "in-Cloud-Datastore persister",
		//		},
	}
}

type gameAndDescription struct {
	persisterAndDescription
	GameState game.ReadAndWriteState
}

func prepareGameStates(
	unitTest *testing.T,
	gameRuleset game.Ruleset,
	playersInTurnOrderWithInitialHands []game.PlayerNameWithHand,
	initialDeck []card.Defined,
	initialActionLog []message.FromPlayer) []gameAndDescription {
	statePersisters := preparePersisters(unitTest)

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
				persisterAndDescription: statePersister,
				GameState:               gameState,
			}
	}

	return gamesAndDescriptions
}

type expectedState struct {
	Name                    string
	Ruleset                 game.Ruleset
	PlayerNames             []string
	CreationTime            time.Time
	ChatLog                 []message.FromPlayer
	ActionLog               []message.FromPlayer
	Turn                    int
	TurnsTakenWithEmptyDeck int
	Score                   int
	NumberOfReadyHints      int
	NumberOfMistakesMade    int
	DeckSize                int
	PlayedForColor          map[string][]card.Defined
	NumberOfDiscardedCards  map[card.Defined]int
	VisibleCardInHand       map[string][]card.Defined
	InferredCardInHand      map[string][]card.Inferred
}

func prepareExpected(
	unitTest *testing.T,
	pristineState game.ReadonlyState) expectedState {
	if pristineState == nil {
		unitTest.Fatalf("nil game.ReadonlyState")
	}

	pristineRuleset := pristineState.Ruleset()

	playedCards := make(map[string][]card.Defined, 0)
	discardedCards := make(map[card.Defined]int, 0)
	for _, colorSuit := range pristineRuleset.ColorSuits() {
		playedCards[colorSuit] = pristineState.PlayedForColor(colorSuit)

		for _, sequenceIndex := range pristineRuleset.DistinctPossibleIndices() {
			numberOfDiscardedCopies := pristineState.NumberOfDiscardedCards(colorSuit, sequenceIndex)
			if numberOfDiscardedCopies != 0 {
				discardedCard :=
					card.Defined{
						ColorSuit:     colorSuit,
						SequenceIndex: sequenceIndex,
					}
				discardedCards[discardedCard] = numberOfDiscardedCopies
			}
		}
	}

	visibleHands := make(map[string][]card.Defined, 0)
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
		Name:         pristineState.Name(),
		Ruleset:      pristineRuleset,
		PlayerNames:  pristineState.PlayerNames(),
		CreationTime: pristineState.CreationTime(),
		ChatLog:      copyLog(unitTest, pristineState.ChatLog()),
		ActionLog:    copyLog(unitTest, pristineState.ActionLog()),
		Turn:         pristineState.Turn(),
		TurnsTakenWithEmptyDeck: pristineState.TurnsTakenWithEmptyDeck(),
		NumberOfReadyHints:      pristineState.NumberOfReadyHints(),
		NumberOfMistakesMade:    pristineState.NumberOfMistakesMade(),
		DeckSize:                pristineState.DeckSize(),
		PlayedForColor:          playedCards,
		NumberOfDiscardedCards:  discardedCards,
		VisibleCardInHand:       visibleHands,
		InferredCardInHand:      inferredHands,
	}
}

func copyLog(
	unitTest *testing.T,
	sourceLog []message.FromPlayer) []message.FromPlayer {
	numberOfMessages := len(sourceLog)
	copiedLog := make([]message.FromPlayer, numberOfMessages)
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
