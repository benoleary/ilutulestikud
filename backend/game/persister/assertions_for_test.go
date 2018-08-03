package persister_test

import (
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

func assertGameStateAsExpected(
	testIdentifier string,
	unitTest *testing.T,
	actualGame game.ReadonlyState,
	expectedGame expectedState) {
	if (actualGame.Name() != expectedGame.Name) ||
		(actualGame.Ruleset() != expectedGame.Ruleset) ||
		(actualGame.CreationTime() != expectedGame.CreationTime) ||
		(actualGame.Turn() != expectedGame.Turn) ||
		(actualGame.TurnsTakenWithEmptyDeck() != expectedGame.TurnsTakenWithEmptyDeck) ||
		(actualGame.NumberOfReadyHints() != expectedGame.NumberOfReadyHints) ||
		(actualGame.NumberOfMistakesMade() != expectedGame.NumberOfMistakesMade) ||
		(actualGame.DeckSize() != expectedGame.DeckSize) ||
		(len(actualGame.PlayerNames()) != len(expectedGame.PlayerNames)) ||
		(len(actualGame.ChatLog()) != len(expectedGame.ChatLog)) ||
		(len(actualGame.ActionLog()) != len(expectedGame.ActionLog)) {
		unitTest.Fatalf(
			testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin easy comparisons",
			actualGame,
			expectedGame)
	}

	toleranceInNanosecondsForLogMessages := 100 * 1000

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
					testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin visible hands\n"+
						"actual:\n%+v\n\nexpected:\n%v\n\n",
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
					testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin inferred hands\n"+
						"actual:\n%+v\n\nexpected:\n%v\n\n",
					actualGame,
					expectedGame,
					inferredHand,
					expectedInferredHand)
			}

			for colorIndex, actualColor := range inferredCard.PossibleColors {
				if actualColor != expectedColors[colorIndex] {
					unitTest.Fatalf(
						"actual\n  %+v\ndid not match expected\n  %+v\nin inferred hand colors\n"+
							"actual:\n%+v\n\nexpected:\n%v\n\n",
						actualGame,
						expectedGame,
						inferredCard.PossibleColors,
						expectedColors)
				}
			}

			for indexIndex, actualIndex := range inferredCard.PossibleIndices {
				if actualIndex != expectedIndices[indexIndex] {
					unitTest.Fatalf(
						testIdentifier+"/actual\n  %+v\ndid not match expected\n  %+v\nin inferred hand indices\n"+
							"actual:\n%+v\n\nexpected:\n%v\n\n",
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
