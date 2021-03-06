package persister_test

import (
	"context"
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game/message"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
)

func assertPlayersMatchNames(
	testIdentifier string,
	unitTest *testing.T,
	expectedPlayerNames map[string]bool,
	actualPlayerNames []string) {
	if len(actualPlayerNames) != len(expectedPlayerNames) {
		unitTest.Fatalf(
			testIdentifier+"/expected players %v, actual list %v",
			expectedPlayerNames,
			actualPlayerNames)
	}

	for _, actualName := range actualPlayerNames {
		if !expectedPlayerNames[actualName] {
			unitTest.Fatalf(
				testIdentifier+"/expected players %v, actual list %v",
				expectedPlayerNames,
				actualPlayerNames)
		}
	}
}

func assertStringSlicesMatch(
	testIdentifier string,
	unitTest *testing.T,
	expectedSlice []string,
	actualSlice []string) {
	numberOfExpected := len(expectedSlice)
	if len(actualSlice) != numberOfExpected {
		unitTest.Fatalf(
			testIdentifier+"/actual %v did not match expected %v",
			actualSlice,
			expectedSlice)
	}

	for sliceIndex := 0; sliceIndex < numberOfExpected; sliceIndex++ {
		expectedString := expectedSlice[sliceIndex]
		actualString := actualSlice[sliceIndex]
		if actualString != expectedString {
			unitTest.Fatalf(
				testIdentifier+"/actual %v did not match expected %v",
				actualSlice,
				expectedSlice)
		}
	}
}

func assertIntSlicesMatch(
	testIdentifier string,
	unitTest *testing.T,
	expectedSlice []int,
	actualSlice []int) {
	numberOfExpected := len(expectedSlice)
	if len(expectedSlice) != numberOfExpected {
		unitTest.Fatalf(
			testIdentifier+"/actual %v did not match expected %v",
			actualSlice,
			expectedSlice)
	}

	for sliceIndex := 0; sliceIndex < numberOfExpected; sliceIndex++ {
		expectedString := expectedSlice[sliceIndex]
		actualString := actualSlice[sliceIndex]
		if actualString != expectedString {
			unitTest.Fatalf(
				testIdentifier+"/actual %v did not match expected %v",
				actualSlice,
				expectedSlice)
		}
	}
}

func assertGameStateAsExpectedLocallyAndRetrieved(
	testIdentifier string,
	unitTest *testing.T,
	actualGameAndPersister gameAndDescription,
	expectedGame expectedState) {
	// We check that the local copy of the game state is as expected.
	actualLocal := actualGameAndPersister.GameState.Read()
	assertGameStateAsExpected(
		testIdentifier+"/local state",
		unitTest,
		actualLocal,
		expectedGame)

	// We check that the version of the game state retrieved from the
	// persister is as expected.
	actualRetrieved, errorFromRetrieval :=
		actualGameAndPersister.GamePersister.ReadAndWriteGame(
			context.Background(),
			actualLocal.Name())

	if errorFromRetrieval != nil {
		unitTest.Fatalf(
			"%v/Unable to retrieve actual game state: %v",
			testIdentifier,
			errorFromRetrieval)
	}

	assertGameStateAsExpected(
		testIdentifier+"/local state",
		unitTest,
		actualRetrieved.Read(),
		expectedGame)
}

func assertGameStateAsExpected(
	testIdentifier string,
	unitTest *testing.T,
	actualGame game.ReadonlyState,
	expectedGame expectedState) {
	if actualGame.Name() != expectedGame.Name {
		unitTest.Fatalf(
			testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin name - expected %v, actual %v",
			actualGame,
			expectedGame,
			expectedGame.Name,
			actualGame.Name())
	}

	if actualGame.Ruleset().BackendIdentifier() != expectedGame.Ruleset.BackendIdentifier() {
		unitTest.Fatalf(
			testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin ruleset backend identifier - expected %v, actual %v",
			actualGame,
			expectedGame,
			expectedGame.Ruleset.BackendIdentifier(),
			actualGame.Ruleset().BackendIdentifier())
	}

	if actualGame.CreationTime() != expectedGame.CreationTime {
		unitTest.Fatalf(
			testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin creation time - expected %v, actual %v",
			actualGame,
			expectedGame,
			expectedGame.CreationTime,
			actualGame.CreationTime())
	}

	if actualGame.Turn() != expectedGame.Turn {
		unitTest.Fatalf(
			testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin turn - expected %v, actual %v",
			actualGame,
			expectedGame,
			expectedGame.Turn,
			actualGame.Turn())
	}

	if actualGame.TurnsTakenWithEmptyDeck() != expectedGame.TurnsTakenWithEmptyDeck {
		unitTest.Fatalf(
			testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin turns taken with empty deck - expected %v, actual %v",
			actualGame,
			expectedGame,
			expectedGame.TurnsTakenWithEmptyDeck,
			actualGame.TurnsTakenWithEmptyDeck())
	}

	if actualGame.NumberOfReadyHints() != expectedGame.NumberOfReadyHints {
		unitTest.Fatalf(
			testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin number of ready hints - expected %v, actual %v",
			actualGame,
			expectedGame,
			expectedGame.NumberOfReadyHints,
			actualGame.NumberOfReadyHints())
	}

	if actualGame.NumberOfMistakesMade() != expectedGame.NumberOfMistakesMade {
		unitTest.Fatalf(
			testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin number of mistakes made - expected %v, actual %v",
			actualGame,
			expectedGame,
			expectedGame.NumberOfMistakesMade,
			actualGame.NumberOfMistakesMade())
	}

	if actualGame.DeckSize() != expectedGame.DeckSize {
		unitTest.Fatalf(
			testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin deck size - expected %v, actual %v",
			actualGame,
			expectedGame,
			expectedGame.DeckSize,
			actualGame.DeckSize())
	}

	if len(actualGame.PlayerNames()) != len(expectedGame.PlayerNames) {
		unitTest.Fatalf(
			testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin number of player names - expected %v, actual %v",
			actualGame,
			expectedGame,
			expectedGame.PlayerNames,
			actualGame.PlayerNames())
	}

	if len(actualGame.ChatLog()) != len(expectedGame.ChatLog) {
		unitTest.Fatalf(
			testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin length of chat log - expected %v, actual %v",
			actualGame,
			expectedGame,
			expectedGame.ChatLog,
			actualGame.ChatLog())
	}

	if len(actualGame.ActionLog()) != len(expectedGame.ActionLog) {
		unitTest.Fatalf(
			testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin length of action log - expected %v, actual %v",
			actualGame,
			expectedGame,
			expectedGame.ActionLog,
			actualGame.ActionLog())
	}

	// We allow up to half a second of variance in the timestamps.
	toleranceInNanosecondsForLogMessages := 500 * 1000 * 1000

	for indexInLog, actualMessage := range actualGame.ChatLog() {
		expectedMessage := expectedGame.ChatLog[indexInLog]
		if !doMessagesMatchwithinTimeTolerance(
			actualMessage,
			expectedMessage,
			toleranceInNanosecondsForLogMessages) {
			unitTest.Fatalf(
				testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin chat log\n"+
					"actual:\n%+v\n\nexpected:\n%+v\n\n",
				actualGame,
				expectedGame,
				actualGame.ChatLog(),
				expectedGame.ChatLog)
		}
	}

	for indexInLog, actualMessage := range actualGame.ActionLog() {
		expectedMessage := expectedGame.ActionLog[indexInLog]
		if !doMessagesMatchwithinTimeTolerance(
			actualMessage,
			expectedMessage,
			toleranceInNanosecondsForLogMessages) {
			unitTest.Fatalf(
				testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin action log\n"+
					"actual:\n%+v\n\nexpected:\n%+v\n\n",
				actualGame,
				expectedGame,
				actualGame.ActionLog(),
				expectedGame.ActionLog)
		}
	}

	for _, colorSuit := range actualGame.Ruleset().ColorSuits() {
		actualPlayedCards := actualGame.PlayedForColor(colorSuit)
		expectedPlayedCards := expectedGame.PlayedForColor[colorSuit]
		if len(actualPlayedCards) != len(expectedPlayedCards) {
			unitTest.Fatalf(
				testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin PlayedForColor\n"+
					"actual:\n%+v\n\nexpected:\n%+v\n\n",
				actualGame,
				expectedGame,
				actualPlayedCards,
				expectedPlayedCards)
		}

		for cardIndex, actualPlayedCard := range actualPlayedCards {
			if actualPlayedCard != expectedPlayedCards[cardIndex] {
				unitTest.Fatalf(
					testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin PlayedForColor\n"+
						"actual:\n%+v\n\nexpected:\n%+v\n\n",
					actualGame,
					expectedGame,
					actualPlayedCards,
					expectedPlayedCards)
			}
		}

		for _, sequenceIndex := range actualGame.Ruleset().DistinctPossibleIndices() {
			actualNumberOfDiscardedCopies :=
				actualGame.NumberOfDiscardedCards(colorSuit, sequenceIndex)
			discardedCard :=
				card.Defined{
					ColorSuit:     colorSuit,
					SequenceIndex: sequenceIndex,
				}
			expectedNumberOfDiscardedCopies :=
				expectedGame.NumberOfDiscardedCards[discardedCard]

			if actualNumberOfDiscardedCopies != expectedNumberOfDiscardedCopies {
				unitTest.Fatalf(
					testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin NumberOfDiscardedCards",
					actualGame,
					expectedGame)
			}
		}
	}

	for playerIndex, playerName := range actualGame.PlayerNames() {
		if playerName != expectedGame.PlayerNames[playerIndex] {
			unitTest.Fatalf(
				testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin player names\n"+
					"actual:\n%+v\n\nexpected:\n%+v\n\n",
				actualGame,
				expectedGame,
				actualGame.PlayerNames(),
				expectedGame.PlayerNames)
		}

		expectedVisibleHand := expectedGame.VisibleCardInHand[playerName]
		expectedInferredHand := expectedGame.InferredCardInHand[playerName]

		// It could be that the hand size is less than the ruleset decrees, if we're on the last turn.
		handSize := len(expectedVisibleHand)

		visibleHand, errorFromVisible :=
			actualGame.VisibleHand(playerName)
		if errorFromVisible != nil {
			unitTest.Fatalf(
				"VisibleHand(%v) produced error %v",
				playerName,
				errorFromVisible)
		}

		inferredHand, errorFromInferred :=
			actualGame.InferredHand(playerName)
		if errorFromInferred != nil {
			unitTest.Fatalf(
				"InferredHand(%v) produced error %v",
				playerName,
				errorFromInferred)
		}

		for indexInHand := 0; indexInHand < handSize; indexInHand++ {
			if visibleHand[indexInHand] != expectedVisibleHand[indexInHand] {
				unitTest.Fatalf(
					testIdentifier+"/player %v %v/actual\n  %+v\ndid not match expected\n  %+v\nin visible hands\n"+
						"actual:\n%+v\n\nexpected:\n%v\n\n",
					playerIndex,
					playerName,
					actualGame,
					expectedGame,
					visibleHand,
					expectedVisibleHand)
			}

			inferredCard := inferredHand[indexInHand]

			expectedInferred := expectedInferredHand[indexInHand]
			expectedColors := expectedInferred.PossibleColors
			expectedIndices := expectedInferred.PossibleIndices

			if (len(inferredCard.PossibleColors) != len(expectedColors)) ||
				(len(inferredCard.PossibleIndices) != len(expectedIndices)) {
				unitTest.Fatalf(
					testIdentifier+"/player %v %v/actual\n  %+v\ndid not match expected\n  %+v\nin inferred hands\n"+
						"actual:\n%+v\n\nexpected:\n%v\n\n",
					playerIndex,
					playerName,
					actualGame,
					expectedGame,
					inferredHand,
					expectedInferredHand)
			}

			for colorIndex, actualColor := range inferredCard.PossibleColors {
				if actualColor != expectedColors[colorIndex] {
					unitTest.Fatalf(
						testIdentifier+"/player %v %v/actual\n  %+v\ndid not match expected\n  %+v\n"+
							"in inferred hand colors\nactual:\n%+v\n\nexpected:\n%v\n\n",
						playerIndex,
						playerName,
						actualGame,
						expectedGame,
						inferredCard.PossibleColors,
						expectedColors)
				}
			}

			for indexIndex, actualIndex := range inferredCard.PossibleIndices {
				if actualIndex != expectedIndices[indexIndex] {
					unitTest.Fatalf(
						testIdentifier+"/player %v %v/actual\n  %+v\ndid not match expected\n  %+v\n"+
							"in inferred hand indices\nactual:\n%+v\n\nexpected:\n%v\n\n",
						playerIndex,
						playerName,
						actualGame,
						expectedGame,
						inferredCard.PossibleIndices,
						expectedIndices)
				}
			}
		}
	}
}

func doMessagesMatchwithinTimeTolerance(
	actualMessage message.FromPlayer,
	expectedMessage message.FromPlayer,
	toleranceInNanoseconds int) bool {
	if (actualMessage.PlayerName != expectedMessage.PlayerName) ||
		(actualMessage.TextColor != expectedMessage.TextColor) ||
		(actualMessage.MessageText != expectedMessage.MessageText) {
		return false
	}

	timeDifference := actualMessage.CreationTime.Sub(expectedMessage.CreationTime)
	if timeDifference < 0 {
		timeDifference = -timeDifference
	}

	return timeDifference < time.Duration(toleranceInNanoseconds)
}
