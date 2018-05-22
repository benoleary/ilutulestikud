package persister_test

import (
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/log"

	"github.com/benoleary/ilutulestikud/backend/game"
)

func TestSetUpInitialMetadataCorrectly(unitTest *testing.T) {
	testStartTime := time.Now()
	emptyMessage := log.Message{}

	threePlayersWithHands :=
		[]game.PlayerNameWithHand{
			game.PlayerNameWithHand{
				PlayerName: defaultTestPlayers[0],
				InitialHand: []card.Inferred{
					card.NewInferred(
						card.NewReadonly("a", 1),
						[]string{"a", "b", "c"},
						[]int{1, 2, 3}),
					card.NewInferred(
						card.NewReadonly("a", 1),
						[]string{"a", "b", "c"},
						[]int{1, 2, 3}),
					card.NewInferred(
						card.NewReadonly("a", 2),
						[]string{"a", "b", "c"},
						[]int{1, 2, 3}),
					card.NewInferred(
						card.NewReadonly("a", 1),
						[]string{"a", "b", "c"},
						[]int{1, 2, 3}),
					card.NewInferred(
						card.NewReadonly("a", 2),
						[]string{"a", "b", "c"},
						[]int{1, 2, 3}),
				},
			},
			game.PlayerNameWithHand{
				PlayerName: defaultTestPlayers[1],
				InitialHand: []card.Inferred{
					card.NewInferred(
						card.NewReadonly("a", 1),
						[]string{"a", "b", "c"},
						[]int{1, 2, 3, 4}),
					card.NewInferred(
						card.NewReadonly("b", 1),
						[]string{"a", "b", "c", "d"},
						[]int{1, 2, 3}),
					card.NewInferred(
						card.NewReadonly("b", 2),
						[]string{"a", "b", "c"},
						[]int{1, 2, 3}),
					card.NewInferred(
						card.NewReadonly("c", 2),
						[]string{"a", "b", "c"},
						[]int{1, 2, 3}),
					card.NewInferred(
						card.NewReadonly("c", 3),
						[]string{"a", "b", "c"},
						[]int{1, 2, 3}),
				},
			},
			game.PlayerNameWithHand{
				PlayerName: defaultTestPlayers[2],
				InitialHand: []card.Inferred{
					card.NewInferred(
						card.NewReadonly("c", 3),
						[]string{"a", "b", "c"},
						[]int{1, 2, 3, 4}),
					card.NewInferred(
						card.NewReadonly("b", 3),
						[]string{"a", "b", "c"},
						[]int{1, 2, 3}),
					card.NewInferred(
						card.NewReadonly("a", 3),
						[]string{"a", "b", "c"},
						[]int{1, 2, 3}),
					card.NewInferred(
						card.NewReadonly("d", 3),
						[]string{"a", "b", "c", "d"},
						[]int{1, 2, 3}),
					card.NewInferred(
						card.NewReadonly("d", 1),
						[]string{"a", "b", "c", "d"},
						[]int{1, 2, 3}),
				},
			},
		}

	numberOfParticipants := len(threePlayersWithHands)
	initialDeck := defaultTestRuleset.CopyOfFullCardset()

	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck)

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

			chatLog := readonlyState.ChatLog()
			if chatLog == nil {
				unitTest.Fatalf("ChatLog() was nil")
			}

			logMessages := chatLog.SortedCopyOfMessages()

			if len(logMessages) != logLengthForTest {
				unitTest.Fatalf(
					"ChatLog() had wrong number of messages %v, expected %v",
					logMessages,
					logLengthForTest)
			}

			for messageIndex := 0; messageIndex < logLengthForTest; messageIndex++ {
				if logMessages[messageIndex] != emptyMessage {
					unitTest.Errorf(
						"ChatLog() %v had non-empty message",
						logMessages)
				}
			}

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
					expectedInferred :=
						expectedNameWithHand.InitialHand[indexInHand]
					expectedVisible := expectedInferred.UnderlyingCard()

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

					if actualInferred.UnderlyingCard() != expectedVisible {
						unitTest.Errorf(
							"InferredCardInHand(%v, %v) %v was not expected %v",
							playerName,
							indexInHand,
							actualInferred,
							expectedInferred)
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
