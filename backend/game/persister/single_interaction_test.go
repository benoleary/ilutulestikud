package persister_test

import (
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
)

func TestErrorFromInvalidPlayerVisibleHand(unitTest *testing.T) {
	initialDeck := defaultTestRuleset.CopyOfFullCardset()

	// A nil initial action log should not be a problem for this test.
	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			nil)

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"visible hand for invalid player/" + gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

			invalidPlayer := "Invalid Player"
			soughtIndex := 0
			visibleCard, errorFromGet :=
				gameAndDescription.GameState.Read().VisibleCardInHand(invalidPlayer, soughtIndex)

			if errorFromGet == nil {
				unitTest.Fatalf(
					"VisibleCardInHand(%v, %v) %v did not produce expected error",
					invalidPlayer,
					soughtIndex,
					visibleCard)
			}

			// There should have been no visible side-effects at all.
			assertGameStateAsExpected(
				testIdentifier,
				unitTest,
				gameAndDescription.GameState.Read(),
				pristineState)
		})
	}
}

func TestErrorFromInvalidPlayerInferredHand(unitTest *testing.T) {
	initialDeck := defaultTestRuleset.CopyOfFullCardset()

	// A nil initial action log should not be a problem for this test.
	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			nil)

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"inferred hand for invalid player/" + gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

			invalidPlayer := "Invalid Player"
			soughtIndex := 0
			inferredCard, errorFromGet :=
				gameAndDescription.GameState.Read().InferredCardInHand(invalidPlayer, soughtIndex)

			if errorFromGet == nil {
				unitTest.Fatalf(
					"InferredCardInHand(%v, %v) %v did not produce expected error",
					invalidPlayer,
					soughtIndex,
					inferredCard)
			}

			// There should have been no visible side-effects at all.
			assertGameStateAsExpected(
				testIdentifier,
				unitTest,
				gameAndDescription.GameState.Read(),
				pristineState)
		})
	}
}

func TestRecordAndRetrieveSingleMessages(unitTest *testing.T) {
	testStartTime := time.Now()
	initialDeck := defaultTestRuleset.CopyOfFullCardset()

	testColor := "test color"

	initialActionMessages :=
		[]string{
			"initial player one action",
			"initial player two action",
			"initial player three action",
		}

	initialActionLog :=
		[]message.Readonly{
			message.NewReadonly(
				threePlayersWithHands[0].PlayerName,
				testColor,
				initialActionMessages[0]),
			message.NewReadonly(
				threePlayersWithHands[1].PlayerName,
				testColor,
				initialActionMessages[1]),
			message.NewReadonly(
				threePlayersWithHands[2].PlayerName,
				testColor,
				initialActionMessages[2]),
		}

	comparisonActionLog := make([]message.Readonly, 3)
	numberOfCopiedMessages := copy(comparisonActionLog, initialActionLog)
	if numberOfCopiedMessages != 3 {
		unitTest.Fatalf(
			"copy(%v, %v) returned %v",
			comparisonActionLog,
			initialActionLog,
			numberOfCopiedMessages)
	}

	// Default message.Readonly structs should have empty strings as expected.
	initialChatLog := make([]message.Readonly, logLengthForTest)

	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			initialActionLog)

	testPlayer := &mockPlayerState{
		threePlayersWithHands[0].PlayerName,
		testColor,
	}

	testMessage := "test message!"

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"single chat message/" + gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

			errorFromChat :=
				gameAndDescription.GameState.RecordChatMessage(testPlayer, testMessage)

			if errorFromChat != nil {
				unitTest.Fatalf(
					"RecordChatMessage(%v, %v) produced error %v",
					testPlayer,
					testMessage,
					errorFromChat)
			}

			assertLogWithSingleMessageIsCorrect(
				testIdentifier+"/chat log",
				unitTest,
				gameAndDescription.GameState.Read().ChatLog(),
				logLengthForTest,
				testPlayer.Name(),
				testPlayer.Color(),
				testMessage,
				initialChatLog,
				testStartTime,
				time.Now())

			// There should have been no other changes.
			pristineState.ChatLog = gameAndDescription.GameState.Read().ChatLog()
			assertGameStateAsExpected(
				testIdentifier,
				unitTest,
				gameAndDescription.GameState.Read(),
				pristineState)

			// We check the initial action log before recording an action message
			// as well as after. In this case though, the "last" message is the
			// original last message.
			assertLogWithSingleMessageIsCorrect(
				testIdentifier+"/action log before recording",
				unitTest,
				gameAndDescription.GameState.Read().ActionLog(),
				len(comparisonActionLog),
				comparisonActionLog[2].PlayerName(),
				comparisonActionLog[2].TextColor(),
				comparisonActionLog[2].MessageText(),
				comparisonActionLog,
				comparisonActionLog[2].CreationTime(),
				time.Now())

			errorFromAction :=
				gameAndDescription.GameState.RecordActionMessage(testPlayer, testMessage)

			if errorFromAction != nil {
				unitTest.Fatalf(
					"RecordActionMessage(%v, %v) produced error %v",
					testPlayer,
					testMessage,
					errorFromAction)
			}

			// For comparison after recording, we take a slice which misses out on the old
			// first message.
			assertLogWithSingleMessageIsCorrect(
				testIdentifier+"/action log after recording",
				unitTest,
				gameAndDescription.GameState.Read().ActionLog(),
				len(comparisonActionLog),
				testPlayer.Name(),
				testPlayer.Color(),
				testMessage,
				comparisonActionLog[1:],
				testStartTime,
				time.Now())

			// There should have been no other changes.
			pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
			assertGameStateAsExpected(
				testIdentifier,
				unitTest,
				gameAndDescription.GameState.Read(),
				pristineState)
		})
	}
}

func TestErrorFromDrawingFromEmptyDeck(unitTest *testing.T) {
	initialDeck := []card.Readonly{}

	// A nil initial action log should not be a problem for this test.
	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			nil)

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"drawing from empty deck/" + gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())
			drawnCard, errorFromDraw :=
				gameAndDescription.GameState.DrawCard()

			if errorFromDraw == nil {
				unitTest.Fatalf(
					"DrawCard() %v did not produce expected error",
					drawnCard)
			}

			// There should have been no visible side-effects at all.
			assertGameStateAsExpected(
				testIdentifier,
				unitTest,
				gameAndDescription.GameState.Read(),
				pristineState)
		})
	}
}

func TestDrawingFromValidDeck(unitTest *testing.T) {
	expectedCard := card.NewReadonly("a", 3)
	initialDeck :=
		[]card.Readonly{
			expectedCard,
			card.NewReadonly("b", 2),
			card.NewReadonly("c", 1),
		}

	initialDeckSize := len(initialDeck)

	// A nil initial action log should not be a problem for this test.
	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			nil)

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"drawing from valid deck/" + gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

			if gameAndDescription.GameState.Read().DeckSize() != initialDeckSize {
				unitTest.Fatalf(
					"initial DeckSize() %v did not match expected %v",
					gameAndDescription.GameState.Read().DeckSize(),
					initialDeckSize)
			}

			drawnCard, errorFromDraw :=
				gameAndDescription.GameState.DrawCard()

			if errorFromDraw != nil {
				unitTest.Fatalf(
					"DrawCard() produced error %v",
					errorFromDraw)
			}

			if drawnCard != expectedCard {
				unitTest.Fatalf(
					"DrawCard() %v did not match expected %v",
					drawnCard,
					expectedCard)
			}

			if gameAndDescription.GameState.Read().DeckSize() != (initialDeckSize - 1) {
				unitTest.Fatalf(
					"after drawing, DeckSize() %v did not match expected %v",
					gameAndDescription.GameState.Read().DeckSize(),
					initialDeckSize-1)
			}

			// There should have been no other changes.
			pristineState.DeckSize = initialDeckSize - 1
			assertGameStateAsExpected(
				testIdentifier,
				unitTest,
				gameAndDescription.GameState.Read(),
				pristineState)
		})
	}
}

func TestErrorFromInvalidReplacementInHand(unitTest *testing.T) {
	// Using nil for the initial action log and for the initial deck
	// should not be problematic for this test.
	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			nil,
			nil)

	cardToInsert := card.InHand{
		Readonly: card.NewReadonly("replacement color", 5),
		Inferred: card.NewInferred([]string{"no idea", "not a clue"}, []int{1, 2, 3}),
	}

	handSize :=
		defaultTestRuleset.NumberOfCardsInPlayerHand(len(threePlayersWithHands))

	testCases := []struct {
		testName    string
		playerName  string
		indexInHand int
	}{
		{
			testName:    "player with no hand",
			playerName:  "Invalid Player",
			indexInHand: 0,
		},
		{
			testName:    "negative index",
			playerName:  defaultTestPlayers[0],
			indexInHand: -1,
		},
		{
			testName:    "too large index",
			playerName:  defaultTestPlayers[0],
			indexInHand: handSize,
		},
	}

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"invalid replacement of card in hand/" + gameAndDescription.PersisterDescription

		for _, testCase := range testCases {
			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

				cardFromOriginalHand, errorFromReplacement :=
					gameAndDescription.GameState.ReplaceCardInHand(
						testCase.playerName,
						testCase.indexInHand,
						cardToInsert)

				if errorFromReplacement == nil {
					unitTest.Fatalf(
						"ReplaceCardInHand(%v, %v, %+v) %+v did not produce expected error",
						testCase.playerName,
						testCase.indexInHand,
						cardToInsert,
						cardFromOriginalHand)
				}

				// There should have been no visible side-effects at all.
				assertGameStateAsExpected(
					testIdentifier,
					unitTest,
					gameAndDescription.GameState.Read(),
					pristineState)
			})
		}
	}
}

func TestReplaceCardInHand(unitTest *testing.T) {
	// Using nil for the initial action log and for the initial deck
	// should not be problematic for this test.
	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			nil,
			nil)

	playerForTest := defaultTestPlayers[1]
	handSize :=
		defaultTestRuleset.NumberOfCardsInPlayerHand(len(threePlayersWithHands))

	cardsToInsert := []card.InHand{
		card.InHand{
			Readonly: card.NewReadonly("first replacement color", 5),
			Inferred: card.NewInferred([]string{"no idea", "not a clue"}, []int{5, 15}),
		},
		card.InHand{
			Readonly: card.NewReadonly("second replacement color", 4),
			Inferred: card.NewInferred([]string{"no idea", "not a clue"}, []int{4, 14}),
		},
		card.InHand{
			Readonly: card.NewReadonly("third replacement color", 3),
			Inferred: card.NewInferred([]string{"no idea", "not a clue"}, []int{3, 13}),
		},
		card.InHand{
			Readonly: card.NewReadonly("fourth replacement color", 2),
			Inferred: card.NewInferred([]string{"no idea", "not a clue"}, []int{2, 12}),
		},
		card.InHand{
			Readonly: card.NewReadonly("fifth replacement color", 1),
			Inferred: card.NewInferred([]string{"no idea", "not a clue"}, []int{1, 10}),
		},
	}

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"valid replacement of card in hand/" + gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

			for indexInHand := 0; indexInHand < handSize; indexInHand++ {
				cardToInsert := cardsToInsert[indexInHand]

				expectedCardFromHand :=
					pristineState.VisibleCardInHand[playerForTest][indexInHand]
				cardFromOriginalHand, errorFromReplacement :=
					gameAndDescription.GameState.ReplaceCardInHand(
						playerForTest,
						indexInHand,
						cardToInsert)

				if errorFromReplacement != nil {
					unitTest.Fatalf(
						"ReplaceCardInHand(%v, %v, %+v) produced error %v",
						playerForTest,
						indexInHand,
						cardToInsert,
						cardFromOriginalHand)
				}

				if cardFromOriginalHand != expectedCardFromHand {
					unitTest.Fatalf(
						"ReplaceCardInHand(%v, %v, %+v) %+v did not return expected %+v",
						playerForTest,
						indexInHand,
						cardToInsert,
						cardFromOriginalHand,
						expectedCardFromHand)
				}

				// There should have been no other changes.
				pristineState.VisibleCardInHand[playerForTest][indexInHand] = cardToInsert.Readonly
				pristineState.InferredCardInHand[playerForTest][indexInHand] = cardToInsert.Inferred
				assertGameStateAsExpected(
					testIdentifier,
					unitTest,
					gameAndDescription.GameState.Read(),
					pristineState)
			}
		})
	}
}

func TestAddCardsToTwoSequences(unitTest *testing.T) {
	// Using nil for the initial action log and for the initial deck
	// should not be problematic for this test.
	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			nil,
			nil)

	colorsInTest := defaultTestRuleset.ColorSuits()
	firstColor := colorsInTest[0]
	secondColor := colorsInTest[1]

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"add cards to sequences/" + gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

			firstCardOfFirstColor := card.NewReadonly(firstColor, 1)

			errorFromFirstPlay :=
				gameAndDescription.GameState.AddCardToPlayedSequence(
					firstCardOfFirstColor)

			if errorFromFirstPlay != nil {
				unitTest.Fatalf(
					"AddCardToPlayedSequence(%+v) produced error %v",
					firstCardOfFirstColor,
					errorFromFirstPlay)
			}

			// There should have been no other changes.
			pristineState.PlayedForColor[firstColor] =
				[]card.Readonly{
					firstCardOfFirstColor,
				}
			assertGameStateAsExpected(
				testIdentifier,
				unitTest,
				gameAndDescription.GameState.Read(),
				pristineState)

			secondCardOfFirstColor := card.NewReadonly(firstColor, 2)

			errorFromSecondPlay :=
				gameAndDescription.GameState.AddCardToPlayedSequence(
					secondCardOfFirstColor)

			if errorFromSecondPlay != nil {
				unitTest.Fatalf(
					"AddCardToPlayedSequence(%+v) produced error %v",
					secondCardOfFirstColor,
					errorFromSecondPlay)
			}

			// There should have been no other changes.
			pristineState.PlayedForColor[firstColor] =
				[]card.Readonly{
					firstCardOfFirstColor,
					secondCardOfFirstColor,
				}
			assertGameStateAsExpected(
				testIdentifier,
				unitTest,
				gameAndDescription.GameState.Read(),
				pristineState)

			firstCardOfSecondColor := card.NewReadonly(secondColor, 1)

			errorFromThirdPlay :=
				gameAndDescription.GameState.AddCardToPlayedSequence(
					firstCardOfSecondColor)

			if errorFromThirdPlay != nil {
				unitTest.Fatalf(
					"AddCardToPlayedSequence(%+v) produced error %v",
					firstCardOfSecondColor,
					errorFromThirdPlay)
			}

			// There should have been no other changes.
			pristineState.PlayedForColor[secondColor] =
				[]card.Readonly{
					firstCardOfSecondColor,
				}
			assertGameStateAsExpected(
				testIdentifier,
				unitTest,
				gameAndDescription.GameState.Read(),
				pristineState)

			thirdCardOfFirstColor := card.NewReadonly(firstColor, 3)

			errorFromFourthPlay :=
				gameAndDescription.GameState.AddCardToPlayedSequence(
					thirdCardOfFirstColor)

			if errorFromFourthPlay != nil {
				unitTest.Fatalf(
					"AddCardToPlayedSequence(%+v) produced error %v",
					secondCardOfFirstColor,
					errorFromFourthPlay)
			}

			// There should have been no other changes.
			pristineState.PlayedForColor[firstColor] =
				[]card.Readonly{
					firstCardOfFirstColor,
					secondCardOfFirstColor,
					thirdCardOfFirstColor,
				}
			assertGameStateAsExpected(
				testIdentifier,
				unitTest,
				gameAndDescription.GameState.Read(),
				pristineState)
		})
	}
}

func assertLogWithSingleMessageIsCorrect(
	testIdentifier string,
	unitTest *testing.T,
	logMessages []message.Readonly,
	expectedLogLength int,
	expectedPlayerName string,
	expectedTextColor string,
	expectedSingleMessage string,
	expectedInitialMessages []message.Readonly,
	earliestTimeForMessage time.Time,
	latestTimeForMessage time.Time) {

	if len(logMessages) != expectedLogLength {
		unitTest.Fatalf(
			testIdentifier+"/wrong number of messages %v, expected %v",
			logMessages,
			expectedLogLength)
	}

	// The first message starts at the end of the log, since there
	// have been no other messages.
	firstMessage := logMessages[expectedLogLength-1]
	if (firstMessage.PlayerName() != expectedPlayerName) ||
		(firstMessage.TextColor() != expectedTextColor) ||
		(firstMessage.MessageText() != expectedSingleMessage) {
		unitTest.Fatalf(
			testIdentifier+
				"/first message %+v was not as expected: player name %v, text color %v, message %v",
			firstMessage,
			expectedPlayerName,
			expectedTextColor,
			expectedSingleMessage)
	}

	recordingTime := firstMessage.CreationTime()

	if (recordingTime.Before(earliestTimeForMessage)) ||
		(recordingTime.After(latestTimeForMessage)) {
		unitTest.Fatalf(
			testIdentifier+
				"/first message %v was not between %v and %v",
			firstMessage,
			earliestTimeForMessage,
			latestTimeForMessage)
	}

	for messageIndex := 0; messageIndex < expectedLogLength-1; messageIndex++ {
		expectedMessage := expectedInitialMessages[messageIndex]
		actualMessage := logMessages[messageIndex]
		if (actualMessage.PlayerName() != expectedMessage.PlayerName()) ||
			(actualMessage.TextColor() != expectedMessage.TextColor()) ||
			(actualMessage.MessageText() != expectedMessage.MessageText()) {
			unitTest.Errorf(
				testIdentifier+
					"/log\n %+v\n did not have expected other messages\n %+v\n",
				logMessages,
				expectedInitialMessages)
		}
	}
}
