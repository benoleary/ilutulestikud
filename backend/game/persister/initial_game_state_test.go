package persister_test

import (
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game/chat"
)

func TestInitialMetadataAreCorrect(unitTest *testing.T) {
	testStartTime := time.Now()
	emptyMessage := chat.Message{}
	numberOfParticipants := len(defaultTestPlayers)
	initialDeck := defaultTestRuleset.CopyOfFullCardset()

	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			defaultTestPlayers,
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
				playerNameSet(defaultTestPlayers),
				readonlyState.Players())

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

			logMessages := chatLog.Sorted()

			if len(logMessages) != chat.LogSize {
				unitTest.Fatalf(
					"ChatLog() had wrong number of messages %v, expected %v",
					logMessages,
					chat.LogSize)
			}

			for messageIndex := 0; messageIndex < chat.LogSize; messageIndex++ {
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
				playedCard, hasAnyBeenPlayed := readonlyState.LastPlayedForColor(colorSuit)
				if hasAnyBeenPlayed {
					unitTest.Fatalf(
						"LastPlayedForColor(%v) was %v rather than expected error placeholder",
						colorSuit,
						playedCard)
				}

				for _, sequenceIndex := range defaultTestRuleset.SequenceIndices() {
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

			for _, playerState := range defaultTestPlayers {
				playerName := playerState.Name()

				for indexInHand := 0; indexInHand < numberOfCardsInHand; indexInHand++ {
					visibleCard, errorFromVisible :=
						readonlyState.VisibleCardInHand(playerName, indexInHand)

					if errorFromVisible != nil {
						unitTest.Fatalf(
							"VisibleCardInHand(%v, %v) %v produced error %v",
							playerName,
							indexInHand,
							visibleCard,
							errorFromVisible)
					}

					if visibleCard == visibleCard {
						unitTest.Errorf(
							"Need to work out how to check hands")
					}

					inferredCard, errorFromInferred :=
						readonlyState.InferredCardInHand(playerName, indexInHand)

					if errorFromInferred != nil {
						unitTest.Fatalf(
							"InferredCardInHand(%v, %v) %v produced error %v",
							playerName,
							indexInHand,
							visibleCard,
							errorFromInferred)
					}

					if len(inferredCard.PossibleColors()) != -1 {
						unitTest.Errorf(
							"Need to work out how to check hands")
					}
				}
			}
		})
	}
}
