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
	initialActionLog := []message.Readonly{
		message.NewReadonly("", "", ""),
		message.NewReadonly("", "", ""),
		message.NewReadonly("", "", ""),
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

			if readonlyState.Score() != 0 {
				unitTest.Fatalf(
					"Score() %v was not expected %v",
					readonlyState.Score(),
					0)
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

				for indexInHand := 0; indexInHand < numberOfCardsInHand; indexInHand++ {
					expectedInHand :=
						expectedNameWithHand.InitialHand[indexInHand]
					expectedVisible := expectedInHand.Readonly
					expectedInferred := expectedInHand.Inferred

					actualVisible, errorFromVisible :=
						readonlyState.VisibleCardInHand(playerName, indexInHand)

					if errorFromVisible != nil {
						unitTest.Fatalf(
							"VisibleCardInHand(%v, %v) %v produced error %v",
							playerName,
							indexInHand,
							actualVisible,
							errorFromVisible)
					}

					if actualVisible != expectedVisible {
						unitTest.Errorf(
							"VisibleCardInHand(%v, %v) %v was not expected %v",
							playerName,
							indexInHand,
							actualVisible,
							expectedVisible)
					}

					actualInferred, errorFromInferred :=
						readonlyState.InferredCardInHand(playerName, indexInHand)

					if errorFromInferred != nil {
						unitTest.Fatalf(
							"InferredCardInHand(%v, %v) %v produced error %v",
							playerName,
							indexInHand,
							actualInferred,
							errorFromInferred)
					}

					assertStringSlicesMatch(
						testIdentifier+"/inferred possible colors",
						unitTest,
						expectedInferred.PossibleColors(),
						actualInferred.PossibleColors())

					assertIntSlicesMatch(
						testIdentifier+"/inferred possible indices",
						unitTest,
						expectedInferred.PossibleIndices(),
						actualInferred.PossibleIndices())
				}
			}
		})
	}
}

func assertLogIsEmpty(
	testIdentifier string,
	unitTest *testing.T,
	actualLog []message.Readonly,
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
		if (actualLog[messageIndex].PlayerName() != "") ||
			(actualLog[messageIndex].TextColor() != "") ||
			(actualLog[messageIndex].MessageText() != "") {
			unitTest.Errorf(
				testIdentifier+"/log %+v had non-empty message",
				actualLog)
		}
	}
}
