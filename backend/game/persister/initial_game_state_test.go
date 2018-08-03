package persister_test

import (
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game/message"
)

func TestSetUpInitialMetadataCorrectly(unitTest *testing.T) {
	testStartTime := time.Now()

	numberOfParticipants := len(threePlayersWithHands)
	initialDeck := defaultTestRuleset.CopyOfFullCardset()

	// For this test, it is most convenient to check that both
	// logs have empty messages.
	initialActionLog := []message.FromPlayer{
		message.NewFromPlayer("", "", ""),
		message.NewFromPlayer("", "", ""),
		message.NewFromPlayer("", "", ""),
	}

	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			initialActionLog)

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"Initial metadata/" + gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			readonlyState := gameAndDescription.GameState.Read()

			if readonlyState == nil {
				unitTest.Fatalf("Read() returned nil")
			}

			if readonlyState.Name() != testGameName {
				unitTest.Fatalf(
					"Name() %v was not expected %v",
					readonlyState.Name(),
					testGameName)
			}

			if readonlyState.Ruleset() != defaultTestRuleset {
				unitTest.Fatalf(
					"Ruleset() %v was not expected %v",
					readonlyState.Ruleset(),
					defaultTestRuleset)
			}

			assertPlayersMatchNames(
				testIdentifier,
				unitTest,
				playerNameSet(threePlayersWithHands),
				readonlyState.PlayerNames())

			comparisonTime := time.Now()

			if !readonlyState.CreationTime().After(testStartTime) ||
				!readonlyState.CreationTime().Before(comparisonTime) {
				unitTest.Fatalf(
					"CreationTime() %v was not in expected range %v to %v",
					readonlyState.CreationTime(),
					testStartTime,
					comparisonTime)
			}

			assertLogIsEmpty(
				testIdentifier+"/ChatLog()",
				unitTest,
				readonlyState.ChatLog(),
				logLengthForTest)

			assertLogIsEmpty(
				testIdentifier+"/ActionLog()",
				unitTest,
				readonlyState.ActionLog(),
				numberOfParticipants)

			if readonlyState.Turn() != 1 {
				unitTest.Fatalf(
					"Turn() %v was not expected %v",
					readonlyState.Turn(),
					1)
			}

			if readonlyState.NumberOfReadyHints() != defaultTestRuleset.MaximumNumberOfHints() {
				unitTest.Fatalf(
					"NumberOfReadyHints() %v was not expected %v",
					readonlyState.NumberOfReadyHints(),
					defaultTestRuleset.MaximumNumberOfHints())
			}

			if readonlyState.NumberOfMistakesMade() != 0 {
				unitTest.Fatalf(
					"NumberOfMistakesMade() %v was not expected %v",
					readonlyState.NumberOfMistakesMade(),
					0)
			}

			if readonlyState.DeckSize() != len(initialDeck) {
				unitTest.Fatalf(
					"DeckSize() %v was not expected %v",
					readonlyState.DeckSize(),
					len(initialDeck))
			}

			for _, colorSuit := range defaultTestRuleset.ColorSuits() {
				playedCards := readonlyState.PlayedForColor(colorSuit)
				if len(playedCards) != 0 {
					unitTest.Fatalf(
						"PlayedForColor(%v) was %v rather than expected empty list",
						colorSuit,
						playedCards)
				}

				for _, sequenceIndex := range defaultTestRuleset.DistinctPossibleIndices() {
					numberOfDiscardedCards := readonlyState.NumberOfDiscardedCards(colorSuit, sequenceIndex)
					if numberOfDiscardedCards != 0 {
						unitTest.Fatalf(
							"NumberOfDiscardedCards(%v, %v) %v was not expected %v",
							colorSuit,
							sequenceIndex,
							numberOfDiscardedCards,
							0)
					}
				}
			}

			numberOfCardsInHand :=
				defaultTestRuleset.NumberOfCardsInPlayerHand(numberOfParticipants)

			for playerIndex, playerWithHand := range threePlayersWithHands {
				expectedNameWithHand := threePlayersWithHands[playerIndex]

				if len(expectedNameWithHand.InitialHand) != numberOfCardsInHand {
					unitTest.Fatalf(
						"expected hand %v not set up correctly, requires %v cards per hand",
						expectedNameWithHand,
						numberOfCardsInHand)
				}

				playerName := playerWithHand.PlayerName

				actualVisibleHand, errorFromVisible :=
					readonlyState.VisibleHand(playerName)

				if errorFromVisible != nil {
					unitTest.Fatalf(
						"VisibleHand(%v) %v produced error %v",
						playerName,
						actualVisibleHand,
						errorFromVisible)
				}

				actualInferredHand, errorFromInferred :=
					readonlyState.InferredHand(playerName)

				if errorFromInferred != nil {
					unitTest.Fatalf(
						"InferredHand(%v) %v produced error %v",
						playerName,
						actualInferredHand,
						errorFromInferred)
				}

				for indexInHand := 0; indexInHand < numberOfCardsInHand; indexInHand++ {
					expectedInHand :=
						expectedNameWithHand.InitialHand[indexInHand]
					expectedVisible := expectedInHand.Defined
					expectedInferred := expectedInHand.Inferred
					if actualVisibleHand[indexInHand] != expectedVisible {
						unitTest.Errorf(
							"VisibleHand(%v) %v at index %v did not match expected %v",
							playerName,
							actualVisibleHand[indexInHand],
							indexInHand,
							expectedVisible)
					}

					assertStringSlicesMatch(
						testIdentifier+"/inferred possible colors",
						unitTest,
						expectedInferred.PossibleColors,
						actualInferredHand[indexInHand].PossibleColors)

					assertIntSlicesMatch(
						testIdentifier+"/inferred possible indices",
						unitTest,
						expectedInferred.PossibleIndices,
						actualInferredHand[indexInHand].PossibleIndices)
				}
			}
		})
	}
}

func assertLogIsEmpty(
	testIdentifier string,
	unitTest *testing.T,
	actualLog []message.FromPlayer,
	expectedLogLength int) {
	if actualLog == nil {
		unitTest.Fatalf("log was nil")
	}

	if len(actualLog) != expectedLogLength {
		unitTest.Fatalf(
			testIdentifier+"/log %+v had wrong number of messages, expected %v",
			actualLog,
			expectedLogLength)
	}

	for messageIndex := 0; messageIndex < expectedLogLength; messageIndex++ {
		if (actualLog[messageIndex].PlayerName != "") ||
			(actualLog[messageIndex].TextColor != "") ||
			(actualLog[messageIndex].MessageText != "") {
			unitTest.Errorf(
				testIdentifier+"/log %+v had non-empty message",
				actualLog)
		}
	}
}
