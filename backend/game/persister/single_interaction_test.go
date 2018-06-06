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

	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			initialActionLogForDefaultThreePlayers)

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

func TestRecordAndRetrieveSingleChatMessage(unitTest *testing.T) {
	testStartTime := time.Now()
	initialDeck := defaultTestRuleset.CopyOfFullCardset()

	// Default message.Readonly structs should have empty strings as expected.
	initialChatLog := make([]message.Readonly, logLengthForTest)

	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			initialActionLogForDefaultThreePlayers)

	testPlayer := &mockPlayerState{
		threePlayersWithHands[0].PlayerName,
		defaultTestColor,
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
		})
	}
}

func TestErrorFromActionsInvalidlyTakingCardFromHand(unitTest *testing.T) {
	initialDeck := []card.Readonly{}

	actionMessage := "action message"
	testColor := "test color"
	numberOfHintsToAdd := 2
	numberOfMistakesToAdd := -1
	knowledgeOfNewCard :=
		card.NewInferred(
			[]string{"no idea", "not a clue"},
			[]int{1, 2, 3})

	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			initialActionLogForDefaultThreePlayers)

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
		for _, testCase := range testCases {
			testIdentifier :=
				"invalid take-card-from-hand action/" +
					gameAndDescription.PersisterDescription +
					"/" + testCase.testName

			testPlayer := &mockPlayerState{
				testCase.playerName,
				testColor,
			}

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

				errorFromDiscardingCard :=
					gameAndDescription.GameState.EnactTurnByDiscardingAndReplacing(
						actionMessage,
						testPlayer,
						testCase.indexInHand,
						knowledgeOfNewCard,
						numberOfHintsToAdd,
						numberOfMistakesToAdd)

				if errorFromDiscardingCard == nil {
					unitTest.Fatalf(
						"EnactTurnByDiscardingAndReplacing(%v, %v, %v, %v, %v, %v)"+
							" did not produce expected error",
						actionMessage,
						testPlayer,
						testCase.indexInHand,
						knowledgeOfNewCard,
						numberOfHintsToAdd,
						numberOfMistakesToAdd)
				}

				// There should have been no visible side-effects apart from a change in the action log.
				pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
				assertGameStateAsExpected(
					testIdentifier,
					unitTest,
					gameAndDescription.GameState.Read(),
					pristineState)

				errorFromPlayingCard :=
					gameAndDescription.GameState.EnactTurnByPlayingAndReplacing(
						actionMessage,
						testPlayer,
						testCase.indexInHand,
						knowledgeOfNewCard,
						numberOfHintsToAdd)

				if errorFromPlayingCard == nil {
					unitTest.Fatalf(
						"EnactTurnByPlayingAndReplacing(%v, %v, %v, %v, %v)"+
							" did not produce expected error",
						actionMessage,
						testPlayer,
						testCase.indexInHand,
						knowledgeOfNewCard,
						numberOfHintsToAdd)
				}

				// There should have been no visible side-effects apart from a change in the action log.
				pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
				assertGameStateAsExpected(
					testIdentifier,
					unitTest,
					gameAndDescription.GameState.Read(),
					pristineState)
			})
		}
	}
}

func TestValidDiscardOfCardWhenDeckAlreadyEmpty(unitTest *testing.T) {
	initialDeck := []card.Readonly{}

	initialDeckSize := len(initialDeck)
	numberOfHintsToAdd := -3
	numberOfMistakesToAdd := 2
	knowledgeOfNewCard :=
		card.NewInferred(
			[]string{"no idea", "not a clue"},
			[]int{1, 2, 3})

	actionMessage := "action message"
	comparisonActionLog := make([]message.Readonly, 3)
	numberOfCopiedMessages :=
		copy(comparisonActionLog, initialActionLogForDefaultThreePlayers)
	if numberOfCopiedMessages != 3 {
		unitTest.Fatalf(
			"copy(%v, %v) returned %v",
			comparisonActionLog,
			initialActionLogForDefaultThreePlayers,
			numberOfCopiedMessages)
	}

	playerName := threePlayersWithHands[0].PlayerName

	testPlayer := &mockPlayerState{
		playerName,
		defaultTestColor,
	}

	indexInHand := 1
	expectedDiscardedCard := threePlayersWithHands[0].InitialHand[indexInHand]

	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			initialActionLogForDefaultThreePlayers)

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"valid play of card when deck already empty/" +
				gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

			if gameAndDescription.GameState.Read().DeckSize() != initialDeckSize {
				unitTest.Fatalf(
					"initial DeckSize() %v did not match expected %v",
					gameAndDescription.GameState.Read().DeckSize(),
					initialDeckSize)
			}

			errorFromDiscardingCard :=
				gameAndDescription.GameState.EnactTurnByDiscardingAndReplacing(
					actionMessage,
					testPlayer,
					indexInHand,
					knowledgeOfNewCard,
					numberOfHintsToAdd,
					numberOfMistakesToAdd)

			if errorFromDiscardingCard != nil {
				unitTest.Fatalf(
					"EnactTurnByDiscardingAndReplacing(%v, %v, %v, %v, %v, %v)"+
						" produced error %v ",
					actionMessage,
					testPlayer,
					indexInHand,
					knowledgeOfNewCard,
					numberOfHintsToAdd,
					numberOfMistakesToAdd,
					errorFromDiscardingCard)
			}

			// There should have been no other changes.
			pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
			pristineState.DeckSize = initialDeckSize
			pristineState.NumberOfReadyHints += numberOfHintsToAdd
			pristineState.NumberOfMistakesMade += numberOfMistakesToAdd
			pristineState.Turn += 1
			pristineVisibleHand := pristineState.VisibleCardInHand[playerName]
			pristineState.VisibleCardInHand[playerName] =
				append(pristineVisibleHand[:indexInHand], pristineVisibleHand[indexInHand+1:]...)
			pristineInferredHand := pristineState.InferredCardInHand[playerName]
			pristineState.InferredCardInHand[playerName] =
				append(pristineInferredHand[:indexInHand], pristineInferredHand[indexInHand+1:]...)
			pristineState.NumberOfDiscardedCards[expectedDiscardedCard.Readonly] = 1
			assertGameStateAsExpected(
				testIdentifier,
				unitTest,
				gameAndDescription.GameState.Read(),
				pristineState)
		})
	}
}

func TestValidDiscardOfCardWhenDeckNotYetEmpty(unitTest *testing.T) {
	expectedReplacementCard := card.NewReadonly("a", 3)
	initialDeck :=
		[]card.Readonly{
			expectedReplacementCard,
			card.NewReadonly("b", 2),
			card.NewReadonly("c", 1),
		}

	initialDeckSize := len(initialDeck)
	numberOfHintsToAdd := -3
	numberOfMistakesToAdd := 2
	knowledgeOfNewCard :=
		card.NewInferred(
			[]string{"no idea", "not a clue"},
			[]int{1, 2, 3})

	actionMessage := "action message"
	comparisonActionLog := make([]message.Readonly, 3)
	numberOfCopiedMessages :=
		copy(comparisonActionLog, initialActionLogForDefaultThreePlayers)
	if numberOfCopiedMessages != 3 {
		unitTest.Fatalf(
			"copy(%v, %v) returned %v",
			comparisonActionLog,
			initialActionLogForDefaultThreePlayers,
			numberOfCopiedMessages)
	}

	playerName := threePlayersWithHands[0].PlayerName

	testPlayer := &mockPlayerState{
		playerName,
		defaultTestColor,
	}

	indexInHand := 1
	expectedDiscardedCard := threePlayersWithHands[0].InitialHand[indexInHand]

	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			initialActionLogForDefaultThreePlayers)

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"valid play of card when deck not yet empty/" +
				gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

			if gameAndDescription.GameState.Read().DeckSize() != initialDeckSize {
				unitTest.Fatalf(
					"initial DeckSize() %v did not match expected %v",
					gameAndDescription.GameState.Read().DeckSize(),
					initialDeckSize)
			}

			errorFromDiscardingCard :=
				gameAndDescription.GameState.EnactTurnByDiscardingAndReplacing(
					actionMessage,
					testPlayer,
					indexInHand,
					knowledgeOfNewCard,
					numberOfHintsToAdd,
					numberOfMistakesToAdd)

			if errorFromDiscardingCard != nil {
				unitTest.Fatalf(
					"EnactTurnByDiscardingAndReplacing(%v, %v, %v, %v, %v, %v)"+
						" produced error %v ",
					actionMessage,
					testPlayer,
					indexInHand,
					knowledgeOfNewCard,
					numberOfHintsToAdd,
					numberOfMistakesToAdd,
					errorFromDiscardingCard)
			}

			// There should have been no other changes.
			pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
			pristineState.DeckSize = initialDeckSize - 1
			pristineState.NumberOfReadyHints += numberOfHintsToAdd
			pristineState.NumberOfMistakesMade += numberOfMistakesToAdd
			pristineState.Turn += 1
			pristineState.VisibleCardInHand[playerName][indexInHand] = expectedReplacementCard
			pristineState.InferredCardInHand[playerName][indexInHand] = knowledgeOfNewCard
			pristineState.NumberOfDiscardedCards[expectedDiscardedCard.Readonly] = 1
			assertGameStateAsExpected(
				testIdentifier,
				unitTest,
				gameAndDescription.GameState.Read(),
				pristineState)
		})
	}
}
func TestValidPlayOfCardWhenDeckAlreadyEmpty(unitTest *testing.T) {
	initialDeck := []card.Readonly{}

	initialDeckSize := len(initialDeck)
	numberOfHintsToAdd := -2
	knowledgeOfNewCard :=
		card.NewInferred(
			[]string{"no idea", "not a clue"},
			[]int{1, 2, 3})

	actionMessage := "action message"
	comparisonActionLog := make([]message.Readonly, 3)
	numberOfCopiedMessages :=
		copy(comparisonActionLog, initialActionLogForDefaultThreePlayers)
	if numberOfCopiedMessages != 3 {
		unitTest.Fatalf(
			"copy(%v, %v) returned %v",
			comparisonActionLog,
			initialActionLogForDefaultThreePlayers,
			numberOfCopiedMessages)
	}

	playerName := threePlayersWithHands[0].PlayerName

	testPlayer := &mockPlayerState{
		playerName,
		defaultTestColor,
	}

	indexInHand := 1
	expectedPlayedCard := threePlayersWithHands[0].InitialHand[indexInHand]

	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			initialActionLogForDefaultThreePlayers)

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"valid play of card when deck already empty/" +
				gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

			if gameAndDescription.GameState.Read().DeckSize() != initialDeckSize {
				unitTest.Fatalf(
					"initial DeckSize() %v did not match expected %v",
					gameAndDescription.GameState.Read().DeckSize(),
					initialDeckSize)
			}

			errorFromPlayingCard :=
				gameAndDescription.GameState.EnactTurnByPlayingAndReplacing(
					actionMessage,
					testPlayer,
					indexInHand,
					knowledgeOfNewCard,
					numberOfHintsToAdd)

			if errorFromPlayingCard != nil {
				unitTest.Fatalf(
					"EnactTurnByPlayingAndReplacing(%v, %v, %v, %v, %v)"+
						" produced error %v ",
					actionMessage,
					testPlayer,
					indexInHand,
					knowledgeOfNewCard,
					numberOfHintsToAdd,
					errorFromPlayingCard)
			}

			// There should have been no other changes.
			pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
			pristineState.DeckSize = initialDeckSize
			pristineState.NumberOfReadyHints += numberOfHintsToAdd
			pristineState.Turn += 1
			pristineVisibleHand := pristineState.VisibleCardInHand[playerName]
			pristineState.VisibleCardInHand[playerName] =
				append(pristineVisibleHand[:indexInHand], pristineVisibleHand[indexInHand+1:]...)
			pristineInferredHand := pristineState.InferredCardInHand[playerName]
			pristineState.InferredCardInHand[playerName] =
				append(pristineInferredHand[:indexInHand], pristineInferredHand[indexInHand+1:]...)
			pristineState.PlayedForColor[expectedPlayedCard.ColorSuit()] =
				[]card.Readonly{expectedPlayedCard.Readonly}
			assertGameStateAsExpected(
				testIdentifier,
				unitTest,
				gameAndDescription.GameState.Read(),
				pristineState)
		})
	}
}

func TestValidPlayOfCardWhenDeckNotYetEmpty(unitTest *testing.T) {
	expectedReplacementCard := card.NewReadonly("a", 3)
	initialDeck :=
		[]card.Readonly{
			expectedReplacementCard,
			card.NewReadonly("b", 2),
			card.NewReadonly("c", 1),
		}

	initialDeckSize := len(initialDeck)
	numberOfHintsToAdd := -2
	knowledgeOfNewCard :=
		card.NewInferred(
			[]string{"no idea", "not a clue"},
			[]int{1, 2, 3})

	actionMessage := "action message"
	comparisonActionLog := make([]message.Readonly, 3)
	numberOfCopiedMessages :=
		copy(comparisonActionLog, initialActionLogForDefaultThreePlayers)
	if numberOfCopiedMessages != 3 {
		unitTest.Fatalf(
			"copy(%v, %v) returned %v",
			comparisonActionLog,
			initialActionLogForDefaultThreePlayers,
			numberOfCopiedMessages)
	}

	playerName := threePlayersWithHands[0].PlayerName

	testPlayer := &mockPlayerState{
		playerName,
		defaultTestColor,
	}

	indexInHand := 1
	expectedPlayedCard := threePlayersWithHands[0].InitialHand[indexInHand]

	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			initialActionLogForDefaultThreePlayers)

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"valid play of card when deck not yet empty/" +
				gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

			if gameAndDescription.GameState.Read().DeckSize() != initialDeckSize {
				unitTest.Fatalf(
					"initial DeckSize() %v did not match expected %v",
					gameAndDescription.GameState.Read().DeckSize(),
					initialDeckSize)
			}

			errorFromPlayingCard :=
				gameAndDescription.GameState.EnactTurnByPlayingAndReplacing(
					actionMessage,
					testPlayer,
					indexInHand,
					knowledgeOfNewCard,
					numberOfHintsToAdd)

			if errorFromPlayingCard != nil {
				unitTest.Fatalf(
					"EnactTurnByPlayingAndReplacing(%v, %v, %v, %v, %v)"+
						" produced error %v ",
					actionMessage,
					testPlayer,
					indexInHand,
					knowledgeOfNewCard,
					numberOfHintsToAdd,
					errorFromPlayingCard)
			}

			// There should have been no other changes.
			pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
			pristineState.DeckSize = initialDeckSize - 1
			pristineState.NumberOfReadyHints += numberOfHintsToAdd
			pristineState.Turn += 1
			pristineState.VisibleCardInHand[playerName][indexInHand] = expectedReplacementCard
			pristineState.InferredCardInHand[playerName][indexInHand] = knowledgeOfNewCard
			pristineState.PlayedForColor[expectedPlayedCard.ColorSuit()] =
				[]card.Readonly{expectedPlayedCard.Readonly}
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
