package persister_test

import (
	"context"
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
)

// obviousTestInferred creates an inferred card with obvious test
// inferred knowledge, but the same length arrays as the ruleset
// gives.
func obviousTestInferred() card.Inferred {
	obviousTestColors := []string{}
	for _, rulesetColor := range testRuleset.ColorSuits() {
		obviousTestColors =
			append(obviousTestColors, "test replacement "+rulesetColor)
	}

	obviousTestIndices := []int{}
	for _, rulesetIndex := range testRuleset.DistinctPossibleIndices() {
		obviousTestIndices =
			append(obviousTestIndices, 100+rulesetIndex)
	}

	return card.Inferred{
		PossibleColors:  obviousTestColors,
		PossibleIndices: obviousTestIndices,
	}
}

var testReplacementInferred card.Inferred = obviousTestInferred()

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
			visibleHand, errorFromGet :=
				gameAndDescription.GameState.Read().VisibleHand(invalidPlayer)

			if errorFromGet == nil {
				unitTest.Fatalf(
					"VisibleHand(%v) %v did not produce expected error",
					invalidPlayer,
					visibleHand)
			}

			// There should have been no visible side-effects at all.
			assertGameStateAsExpectedLocallyAndRetrieved(
				testIdentifier,
				unitTest,
				gameAndDescription,
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
			inferredHand, errorFromGet :=
				gameAndDescription.GameState.Read().InferredHand(invalidPlayer)

			if errorFromGet == nil {
				unitTest.Fatalf(
					"InferredHand(%v) %v did not produce expected error",
					invalidPlayer,
					inferredHand)
			}

			// There should have been no visible side-effects at all.
			assertGameStateAsExpectedLocallyAndRetrieved(
				testIdentifier,
				unitTest,
				gameAndDescription,
				pristineState)
		})
	}
}

func TestRecordAndRetrieveSingleChatMessage(unitTest *testing.T) {
	testStartTime := time.Now()
	initialDeck := defaultTestRuleset.CopyOfFullCardset()

	// Default message.FromPlayer structs should have empty strings as expected.
	initialChatLog := make([]message.FromPlayer, logLengthForTest)

	for _, testPlayerWithHand := range threePlayersWithHands {
		gamesAndDescriptions :=
			prepareGameStates(
				unitTest,
				defaultTestRuleset,
				threePlayersWithHands,
				initialDeck,
				initialActionLogForDefaultThreePlayers)

		testPlayer := &mockPlayerState{
			testPlayerWithHand.PlayerName,
			defaultTestColor,
		}

		testMessage := "test message!"

		for _, gameAndDescription := range gamesAndDescriptions {
			testIdentifier :=
				"single chat message from " + testPlayer.Name() +
					"/" + gameAndDescription.PersisterDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

				errorFromChat :=
					gameAndDescription.GameState.RecordChatMessage(
						context.Background(),
						testPlayer,
						testMessage)

				if errorFromChat != nil {
					unitTest.Fatalf(
						"RecordChatMessage(%+v, %v) produced error %v",
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
				assertGameStateAsExpectedLocallyAndRetrieved(
					testIdentifier,
					unitTest,
					gameAndDescription,
					pristineState)
			})
		}
	}
}

func TestErrorFromActionsInvalidlyTakingCardFromHand(unitTest *testing.T) {
	initialDeck := []card.Defined{}

	actionMessage := "action message"
	testColor := "test color"
	numberOfHintsToAdd := 2
	numberOfMistakesToAdd := -1
	knowledgeOfNewCard := testReplacementInferred

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
						context.Background(),
						actionMessage,
						testPlayer,
						testCase.indexInHand,
						knowledgeOfNewCard,
						numberOfHintsToAdd,
						numberOfMistakesToAdd)

				if errorFromDiscardingCard == nil {
					unitTest.Fatalf(
						"EnactTurnByDiscardingAndReplacing(%v, %+v, %v, %+v, %v, %v)"+
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
				assertGameStateAsExpectedLocallyAndRetrieved(
					testIdentifier,
					unitTest,
					gameAndDescription,
					pristineState)

				errorFromPlayingCard :=
					gameAndDescription.GameState.EnactTurnByPlayingAndReplacing(
						context.Background(),
						actionMessage,
						testPlayer,
						testCase.indexInHand,
						knowledgeOfNewCard,
						numberOfHintsToAdd)

				if errorFromPlayingCard == nil {
					unitTest.Fatalf(
						"EnactTurnByPlayingAndReplacing(%v, %+v, %v, %+v, %v)"+
							" did not produce expected error",
						actionMessage,
						testPlayer,
						testCase.indexInHand,
						knowledgeOfNewCard,
						numberOfHintsToAdd)
				}

				// There should have been no visible side-effects apart from a change in the action log.
				pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
				assertGameStateAsExpectedLocallyAndRetrieved(
					testIdentifier,
					unitTest,
					gameAndDescription,
					pristineState)
			})
		}
	}
}

func TestValidDiscardOfCardWhenDeckNotYetEmpty(unitTest *testing.T) {
	expectedReplacementCard :=
		card.Defined{
			ColorSuit:     "a",
			SequenceIndex: 3,
		}

	numberOfHintsToAdd := -3
	numberOfMistakesToAdd := 2
	knowledgeOfNewCard := testReplacementInferred

	actionMessage := "action message"
	comparisonActionLog := make([]message.FromPlayer, 3)
	numberOfCopiedMessages :=
		copy(comparisonActionLog, initialActionLogForDefaultThreePlayers)
	if numberOfCopiedMessages != 3 {
		unitTest.Fatalf(
			"copy(%v, %v) returned %v",
			comparisonActionLog,
			initialActionLogForDefaultThreePlayers,
			numberOfCopiedMessages)
	}

	for _, testPlayerWithHand := range threePlayersWithHands {
		playerName := testPlayerWithHand.PlayerName

		testPlayer := &mockPlayerState{
			playerName,
			defaultTestColor,
		}

		for indexInHand := 0; indexInHand < len(testPlayerWithHand.InitialHand); indexInHand++ {
			// The initial deck has to be created afresh for each test case, or previous test
			// cases will interfere with subsequent tests.
			initialDeck :=
				[]card.Defined{
					expectedReplacementCard,
					card.Defined{
						ColorSuit:     "b",
						SequenceIndex: 2,
					},
					card.Defined{
						ColorSuit:     "c",
						SequenceIndex: 1,
					},
				}

			initialDeckSize := len(initialDeck)

			expectedDiscardedCard := testPlayerWithHand.InitialHand[indexInHand]

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
							context.Background(),
							actionMessage,
							testPlayer,
							indexInHand,
							knowledgeOfNewCard,
							numberOfHintsToAdd,
							numberOfMistakesToAdd)

					if errorFromDiscardingCard != nil {
						unitTest.Fatalf(
							"EnactTurnByDiscardingAndReplacing(%v, %+v, %v, %+v, %v, %v)"+
								" produced error %v ",
							actionMessage,
							testPlayer,
							indexInHand,
							knowledgeOfNewCard,
							numberOfHintsToAdd,
							numberOfMistakesToAdd,
							errorFromDiscardingCard)
					}

					// There should have been the following changes:
					pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
					pristineState.ActionLog[len(pristineState.ActionLog)-1] =
						message.NewFromPlayer(testPlayer.Name(), testPlayer.Color(), actionMessage)
					pristineState.DeckSize = initialDeckSize - 1
					pristineState.NumberOfReadyHints += numberOfHintsToAdd
					pristineState.NumberOfMistakesMade += numberOfMistakesToAdd
					pristineState.Turn += 1
					pristineState.VisibleCardInHand[playerName][indexInHand] = expectedReplacementCard
					pristineState.InferredCardInHand[playerName][indexInHand] = knowledgeOfNewCard
					pristineState.NumberOfDiscardedCards[expectedDiscardedCard.Defined] = 1
					assertGameStateAsExpectedLocallyAndRetrieved(
						testIdentifier,
						unitTest,
						gameAndDescription,
						pristineState)
				})
			}
		}
	}
}

func TestValidDiscardOfCardWhichEmptiesDeck(unitTest *testing.T) {
	expectedReplacementCard :=
		card.Defined{
			ColorSuit:     "a",
			SequenceIndex: 3,
		}

	numberOfHintsToAdd := -3
	numberOfMistakesToAdd := 2
	knowledgeOfNewCard := testReplacementInferred

	actionMessage := "action message"
	comparisonActionLog := make([]message.FromPlayer, 3)
	numberOfCopiedMessages :=
		copy(comparisonActionLog, initialActionLogForDefaultThreePlayers)
	if numberOfCopiedMessages != 3 {
		unitTest.Fatalf(
			"copy(%v, %v) returned %v",
			comparisonActionLog,
			initialActionLogForDefaultThreePlayers,
			numberOfCopiedMessages)
	}

	for _, testPlayerWithHand := range threePlayersWithHands {
		playerName := testPlayerWithHand.PlayerName

		testPlayer := &mockPlayerState{
			playerName,
			defaultTestColor,
		}

		for indexInHand := 0; indexInHand < len(testPlayerWithHand.InitialHand); indexInHand++ {
			// The initial deck has to be created afresh for each test case, or previous test
			// cases will interfere with subsequent tests.
			initialDeck :=
				[]card.Defined{
					expectedReplacementCard,
				}

			initialDeckSize := len(initialDeck)

			expectedDiscardedCard := testPlayerWithHand.InitialHand[indexInHand]

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
							context.Background(),
							actionMessage,
							testPlayer,
							indexInHand,
							knowledgeOfNewCard,
							numberOfHintsToAdd,
							numberOfMistakesToAdd)

					if errorFromDiscardingCard != nil {
						unitTest.Fatalf(
							"EnactTurnByDiscardingAndReplacing(%v, %+v, %v, %+v, %v, %v)"+
								" produced error %v ",
							actionMessage,
							testPlayer,
							indexInHand,
							knowledgeOfNewCard,
							numberOfHintsToAdd,
							numberOfMistakesToAdd,
							errorFromDiscardingCard)
					}

					// There should have been the following changes:
					pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
					pristineState.ActionLog[len(pristineState.ActionLog)-1] =
						message.NewFromPlayer(testPlayer.Name(), testPlayer.Color(), actionMessage)
					pristineState.DeckSize = initialDeckSize - 1
					pristineState.NumberOfReadyHints += numberOfHintsToAdd
					pristineState.NumberOfMistakesMade += numberOfMistakesToAdd
					pristineState.Turn += 1
					pristineState.VisibleCardInHand[playerName][indexInHand] = expectedReplacementCard
					pristineState.InferredCardInHand[playerName][indexInHand] = knowledgeOfNewCard
					pristineState.NumberOfDiscardedCards[expectedDiscardedCard.Defined] = 1
					assertGameStateAsExpectedLocallyAndRetrieved(
						testIdentifier,
						unitTest,
						gameAndDescription,
						pristineState)
				})
			}
		}
	}
}

func TestValidDiscardOfCardWhenDeckAlreadyEmpty(unitTest *testing.T) {
	// In this case, the initial deck does not have to be created afresh
	// for each test case, because it starts empty and stays empty.
	initialDeck := []card.Defined{}

	initialDeckSize := len(initialDeck)
	numberOfHintsToAdd := -3
	numberOfMistakesToAdd := 2
	knowledgeOfNewCard := testReplacementInferred

	actionMessage := "action message"
	comparisonActionLog := make([]message.FromPlayer, 3)
	numberOfCopiedMessages :=
		copy(comparisonActionLog, initialActionLogForDefaultThreePlayers)
	if numberOfCopiedMessages != 3 {
		unitTest.Fatalf(
			"copy(%v, %v) returned %v",
			comparisonActionLog,
			initialActionLogForDefaultThreePlayers,
			numberOfCopiedMessages)
	}

	for _, testPlayerWithHand := range threePlayersWithHands {
		playerName := testPlayerWithHand.PlayerName

		testPlayer := &mockPlayerState{
			playerName,
			defaultTestColor,
		}

		for indexInHand := 0; indexInHand < len(testPlayerWithHand.InitialHand); indexInHand++ {
			expectedDiscardedCard := testPlayerWithHand.InitialHand[indexInHand]

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
							context.Background(),
							actionMessage,
							testPlayer,
							indexInHand,
							knowledgeOfNewCard,
							numberOfHintsToAdd,
							numberOfMistakesToAdd)

					if errorFromDiscardingCard != nil {
						unitTest.Fatalf(
							"EnactTurnByDiscardingAndReplacing(%v, %+v, %v, %+v, %v, %v)"+
								" produced error %v ",
							actionMessage,
							testPlayer,
							indexInHand,
							knowledgeOfNewCard,
							numberOfHintsToAdd,
							numberOfMistakesToAdd,
							errorFromDiscardingCard)
					}

					// There should have been the following changes:
					pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
					pristineState.ActionLog[len(pristineState.ActionLog)-1] =
						message.NewFromPlayer(testPlayer.Name(), testPlayer.Color(), actionMessage)
					pristineState.DeckSize = initialDeckSize
					pristineState.NumberOfReadyHints += numberOfHintsToAdd
					pristineState.NumberOfMistakesMade += numberOfMistakesToAdd
					pristineState.Turn += 1
					pristineState.TurnsTakenWithEmptyDeck += 1
					pristineVisibleHand := pristineState.VisibleCardInHand[playerName]
					pristineState.VisibleCardInHand[playerName] =
						append(pristineVisibleHand[:indexInHand], pristineVisibleHand[indexInHand+1:]...)
					pristineInferredHand := pristineState.InferredCardInHand[playerName]
					pristineState.InferredCardInHand[playerName] =
						append(pristineInferredHand[:indexInHand], pristineInferredHand[indexInHand+1:]...)
					pristineState.NumberOfDiscardedCards[expectedDiscardedCard.Defined] = 1
					assertGameStateAsExpectedLocallyAndRetrieved(
						testIdentifier,
						unitTest,
						gameAndDescription,
						pristineState)
				})
			}
		}
	}
}

func TestValidPlayOfCardWhenDeckNotYetEmpty(unitTest *testing.T) {
	expectedReplacementCard :=
		card.Defined{
			ColorSuit:     "a",
			SequenceIndex: 3,
		}

	numberOfHintsToAdd := -2
	knowledgeOfNewCard := testReplacementInferred

	actionMessage := "action message"
	comparisonActionLog := make([]message.FromPlayer, 3)
	numberOfCopiedMessages :=
		copy(comparisonActionLog, initialActionLogForDefaultThreePlayers)
	if numberOfCopiedMessages != 3 {
		unitTest.Fatalf(
			"copy(%v, %v) returned %v",
			comparisonActionLog,
			initialActionLogForDefaultThreePlayers,
			numberOfCopiedMessages)
	}

	for _, testPlayerWithHand := range threePlayersWithHands {
		playerName := testPlayerWithHand.PlayerName

		testPlayer := &mockPlayerState{
			playerName,
			defaultTestColor,
		}

		for indexInHand := 0; indexInHand < len(testPlayerWithHand.InitialHand); indexInHand++ {
			// The initial deck has to be created afresh for each test case, or previous test
			// cases will interfere with subsequent tests.
			initialDeck :=
				[]card.Defined{
					expectedReplacementCard,
					card.Defined{
						ColorSuit:     "b",
						SequenceIndex: 2,
					},
					card.Defined{
						ColorSuit:     "c",
						SequenceIndex: 1,
					},
				}

			initialDeckSize := len(initialDeck)

			expectedPlayedCard := testPlayerWithHand.InitialHand[indexInHand]

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
							context.Background(),
							actionMessage,
							testPlayer,
							indexInHand,
							knowledgeOfNewCard,
							numberOfHintsToAdd)

					if errorFromPlayingCard != nil {
						unitTest.Fatalf(
							"EnactTurnByPlayingAndReplacing(%v, %+v, %v, %+v, %v)"+
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
					pristineState.ActionLog[len(pristineState.ActionLog)-1] =
						message.NewFromPlayer(testPlayer.Name(), testPlayer.Color(), actionMessage)
					pristineState.DeckSize = initialDeckSize - 1
					pristineState.NumberOfReadyHints += numberOfHintsToAdd
					pristineState.Turn += 1
					pristineState.VisibleCardInHand[playerName][indexInHand] = expectedReplacementCard
					pristineState.InferredCardInHand[playerName][indexInHand] = knowledgeOfNewCard
					pristineState.PlayedForColor[expectedPlayedCard.ColorSuit] =
						[]card.Defined{expectedPlayedCard.Defined}
					assertGameStateAsExpectedLocallyAndRetrieved(
						testIdentifier,
						unitTest,
						gameAndDescription,
						pristineState)
				})
			}
		}
	}
}

func TestValidPlayOfCardWhichEmptiesDeck(unitTest *testing.T) {
	expectedReplacementCard :=
		card.Defined{
			ColorSuit:     "a",
			SequenceIndex: 3,
		}

	numberOfHintsToAdd := -2
	knowledgeOfNewCard := testReplacementInferred

	actionMessage := "action message"
	comparisonActionLog := make([]message.FromPlayer, 3)
	numberOfCopiedMessages :=
		copy(comparisonActionLog, initialActionLogForDefaultThreePlayers)
	if numberOfCopiedMessages != 3 {
		unitTest.Fatalf(
			"copy(%v, %v) returned %v",
			comparisonActionLog,
			initialActionLogForDefaultThreePlayers,
			numberOfCopiedMessages)
	}

	for _, testPlayerWithHand := range threePlayersWithHands {
		playerName := testPlayerWithHand.PlayerName

		testPlayer := &mockPlayerState{
			playerName,
			defaultTestColor,
		}

		for indexInHand := 0; indexInHand < len(testPlayerWithHand.InitialHand); indexInHand++ {
			// The initial deck has to be created afresh for each test case, or previous test
			// cases will interfere with subsequent tests.
			initialDeck :=
				[]card.Defined{
					expectedReplacementCard,
				}

			initialDeckSize := len(initialDeck)

			expectedPlayedCard := testPlayerWithHand.InitialHand[indexInHand]

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
							context.Background(),
							actionMessage,
							testPlayer,
							indexInHand,
							knowledgeOfNewCard,
							numberOfHintsToAdd)

					if errorFromPlayingCard != nil {
						unitTest.Fatalf(
							"EnactTurnByPlayingAndReplacing(%v, %+v, %v, %+v, %v)"+
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
					pristineState.ActionLog[len(pristineState.ActionLog)-1] =
						message.NewFromPlayer(testPlayer.Name(), testPlayer.Color(), actionMessage)
					pristineState.DeckSize = initialDeckSize - 1
					pristineState.NumberOfReadyHints += numberOfHintsToAdd
					pristineState.Turn += 1
					pristineState.VisibleCardInHand[playerName][indexInHand] = expectedReplacementCard
					pristineState.InferredCardInHand[playerName][indexInHand] = knowledgeOfNewCard
					pristineState.PlayedForColor[expectedPlayedCard.ColorSuit] =
						[]card.Defined{expectedPlayedCard.Defined}
					assertGameStateAsExpectedLocallyAndRetrieved(
						testIdentifier,
						unitTest,
						gameAndDescription,
						pristineState)
				})
			}
		}
	}
}

func TestValidPlayOfCardWhenDeckAlreadyEmpty(unitTest *testing.T) {
	// In this case, the initial deck does not have to be created afresh
	// for each test case, because it starts empty and stays empty.
	initialDeck := []card.Defined{}

	initialDeckSize := len(initialDeck)
	numberOfHintsToAdd := -2
	knowledgeOfNewCard := testReplacementInferred

	actionMessage := "action message"
	comparisonActionLog := make([]message.FromPlayer, 3)
	numberOfCopiedMessages :=
		copy(comparisonActionLog, initialActionLogForDefaultThreePlayers)
	if numberOfCopiedMessages != 3 {
		unitTest.Fatalf(
			"copy(%v, %v) returned %v",
			comparisonActionLog,
			initialActionLogForDefaultThreePlayers,
			numberOfCopiedMessages)
	}

	for _, testPlayerWithHand := range threePlayersWithHands {
		playerName := testPlayerWithHand.PlayerName

		testPlayer := &mockPlayerState{
			playerName,
			defaultTestColor,
		}

		for indexInHand := 0; indexInHand < len(testPlayerWithHand.InitialHand); indexInHand++ {
			expectedPlayedCard := testPlayerWithHand.InitialHand[indexInHand]

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
							context.Background(),
							actionMessage,
							testPlayer,
							indexInHand,
							knowledgeOfNewCard,
							numberOfHintsToAdd)

					if errorFromPlayingCard != nil {
						unitTest.Fatalf(
							"EnactTurnByPlayingAndReplacing(%v, %+v, %v, %+v, %v)"+
								" produced error %v ",
							actionMessage,
							testPlayer,
							indexInHand,
							knowledgeOfNewCard,
							numberOfHintsToAdd,
							errorFromPlayingCard)
					}

					// There should have been the following changes:
					pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
					pristineState.ActionLog[len(pristineState.ActionLog)-1] =
						message.NewFromPlayer(testPlayer.Name(), testPlayer.Color(), actionMessage)
					pristineState.DeckSize = initialDeckSize
					pristineState.NumberOfReadyHints += numberOfHintsToAdd
					pristineState.Turn += 1
					pristineState.TurnsTakenWithEmptyDeck += 1
					pristineVisibleHand := pristineState.VisibleCardInHand[playerName]
					pristineState.VisibleCardInHand[playerName] =
						append(pristineVisibleHand[:indexInHand], pristineVisibleHand[indexInHand+1:]...)
					pristineInferredHand := pristineState.InferredCardInHand[playerName]
					pristineState.InferredCardInHand[playerName] =
						append(pristineInferredHand[:indexInHand], pristineInferredHand[indexInHand+1:]...)
					pristineState.PlayedForColor[expectedPlayedCard.ColorSuit] =
						[]card.Defined{expectedPlayedCard.Defined}
					assertGameStateAsExpectedLocallyAndRetrieved(
						testIdentifier,
						unitTest,
						gameAndDescription,
						pristineState)
				})
			}
		}
	}
}

func TestErrorFromHintToInvalidPlayer(unitTest *testing.T) {
	initialDeck := []card.Defined{}

	actionMessage := "action message"

	actingPlayerName := threePlayersWithHands[0].PlayerName

	actingPlayer := &mockPlayerState{
		actingPlayerName,
		defaultTestColor,
	}

	receivingPlayerName := "Not A. Participant"

	numberOfHintsToSubtract := 3
	handSize := len(threePlayersWithHands[0].InitialHand)

	// It is not important to make valid inferred cards for this test.
	updatedInferredHand := make([]card.Inferred, handSize)

	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			initialActionLogForDefaultThreePlayers)

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"invalid take-card-from-hand action/" +
				gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

			errorFromHint :=
				gameAndDescription.GameState.EnactTurnByUpdatingHandWithHint(
					context.Background(),
					actionMessage,
					actingPlayer,
					receivingPlayerName,
					updatedInferredHand,
					numberOfHintsToSubtract)

			if errorFromHint == nil {
				unitTest.Fatalf(
					"EnactTurnByUpdatingHandWithHint(%v, %+v, %v, %+v, %v)"+
						" did not produce expected error",
					actionMessage,
					actingPlayer,
					receivingPlayerName,
					updatedInferredHand,
					numberOfHintsToSubtract)
			}

			// There should have been no visible side-effects apart from a change in the action log.
			pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
			assertGameStateAsExpectedLocallyAndRetrieved(
				testIdentifier,
				unitTest,
				gameAndDescription,
				pristineState)
		})
	}
}

func TestErrorFromHintWithTooSmallInferredHand(unitTest *testing.T) {
	initialDeck := []card.Defined{}

	actionMessage := "action message"

	actingPlayerName := threePlayersWithHands[0].PlayerName

	actingPlayer := &mockPlayerState{
		actingPlayerName,
		defaultTestColor,
	}

	receivingPlayerWithHand := threePlayersWithHands[1]
	receivingPlayerName := receivingPlayerWithHand.PlayerName

	numberOfHintsToSubtract := 4
	tooSmallHandSize := len(receivingPlayerWithHand.InitialHand) - 1

	// It is not important to make valid inferred cards for this test.
	updatedInferredHand := make([]card.Inferred, tooSmallHandSize)

	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			initialActionLogForDefaultThreePlayers)

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"invalid take-card-from-hand action/" +
				gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

			errorFromHint :=
				gameAndDescription.GameState.EnactTurnByUpdatingHandWithHint(
					context.Background(),
					actionMessage,
					actingPlayer,
					receivingPlayerName,
					updatedInferredHand,
					numberOfHintsToSubtract)

			if errorFromHint == nil {
				unitTest.Fatalf(
					"EnactTurnByUpdatingHandWithHint(%v, %+v, %v, %+v, %v)"+
						" did not produce expected error",
					actionMessage,
					actingPlayer,
					receivingPlayerName,
					updatedInferredHand,
					numberOfHintsToSubtract)
			}

			// There should have been no visible side-effects apart from a change in the action log.
			pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
			assertGameStateAsExpectedLocallyAndRetrieved(
				testIdentifier,
				unitTest,
				gameAndDescription,
				pristineState)
		})
	}
}

func TestErrorFromHintWithTooLargeInferredHand(unitTest *testing.T) {
	initialDeck := []card.Defined{}

	actionMessage := "action message"

	actingPlayerName := threePlayersWithHands[0].PlayerName

	actingPlayer := &mockPlayerState{
		actingPlayerName,
		defaultTestColor,
	}

	receivingPlayerWithHand := threePlayersWithHands[1]
	receivingPlayerName := receivingPlayerWithHand.PlayerName

	numberOfHintsToSubtract := 5
	tooSmallHandSize := len(receivingPlayerWithHand.InitialHand) + 1

	// It is not important to make valid inferred cards for this test.
	updatedInferredHand := make([]card.Inferred, tooSmallHandSize)

	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			initialActionLogForDefaultThreePlayers)

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"invalid take-card-from-hand action/" +
				gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

			errorFromHint :=
				gameAndDescription.GameState.EnactTurnByUpdatingHandWithHint(
					context.Background(),
					actionMessage,
					actingPlayer,
					receivingPlayerName,
					updatedInferredHand,
					numberOfHintsToSubtract)

			if errorFromHint == nil {
				unitTest.Fatalf(
					"EnactTurnByUpdatingHandWithHint(%v, %+v, %v, %+v, %v)"+
						" did not produce expected error",
					actionMessage,
					actingPlayer,
					receivingPlayerName,
					updatedInferredHand,
					numberOfHintsToSubtract)
			}

			// There should have been no visible side-effects apart from a change in the action log.
			pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
			assertGameStateAsExpectedLocallyAndRetrieved(
				testIdentifier,
				unitTest,
				gameAndDescription,
				pristineState)
		})
	}
}

func TestValidHintWhenDeckAlreadyEmpty(unitTest *testing.T) {
	// In this case, the initial deck does not have to be created afresh
	// for each test case, because it starts empty and stays empty.
	initialDeck := []card.Defined{}

	initialDeckSize := len(initialDeck)

	actionMessage := "action message"
	comparisonActionLog := make([]message.FromPlayer, 3)
	numberOfCopiedMessages :=
		copy(comparisonActionLog, initialActionLogForDefaultThreePlayers)
	if numberOfCopiedMessages != 3 {
		unitTest.Fatalf(
			"copy(%v, %v) returned %v",
			comparisonActionLog,
			initialActionLogForDefaultThreePlayers,
			numberOfCopiedMessages)
	}

	numberOfHintsToSubtract := 2

	for actingPlayerIndex, actingPlayerWithHand := range threePlayersWithHands {
		actingPlayerName := actingPlayerWithHand.PlayerName

		actingPlayer := &mockPlayerState{
			actingPlayerName,
			defaultTestColor,
		}

		for receivingPlayerIndex, receivingPlayerWithHand := range threePlayersWithHands {
			if receivingPlayerIndex == actingPlayerIndex {
				continue
			}

			receivingPlayerName := receivingPlayerWithHand.PlayerName
			handSize := len(receivingPlayerWithHand.InitialHand)

			// It is not important to make valid inferred cards for this test.
			updatedInferredHand := make([]card.Inferred, handSize)
			testInferredColors := []string{"a test color", "another test color"}
			for indexInHand := 0; indexInHand < handSize; indexInHand++ {
				updatedInferredHand[indexInHand] =
					card.Inferred{
						PossibleColors:  testInferredColors,
						PossibleIndices: []int{indexInHand},
					}
			}

			gamesAndDescriptions :=
				prepareGameStates(
					unitTest,
					defaultTestRuleset,
					threePlayersWithHands,
					initialDeck,
					initialActionLogForDefaultThreePlayers)

			for _, gameAndDescription := range gamesAndDescriptions {
				testIdentifier :=
					"valid hint when deck already empty/" +
						gameAndDescription.PersisterDescription

				unitTest.Run(testIdentifier, func(unitTest *testing.T) {
					pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

					if gameAndDescription.GameState.Read().DeckSize() != initialDeckSize {
						unitTest.Fatalf(
							"initial DeckSize() %v did not match expected %v",
							gameAndDescription.GameState.Read().DeckSize(),
							initialDeckSize)
					}

					errorFromHint :=
						gameAndDescription.GameState.EnactTurnByUpdatingHandWithHint(
							context.Background(),
							actionMessage,
							actingPlayer,
							receivingPlayerName,
							updatedInferredHand,
							numberOfHintsToSubtract)

					if errorFromHint != nil {
						unitTest.Fatalf(
							"EnactTurnByUpdatingHandWithHint(%v, %+v, %v, %+v, %v)"+
								" produced error %v ",
							actionMessage,
							actingPlayer,
							receivingPlayerName,
							updatedInferredHand,
							numberOfHintsToSubtract,
							errorFromHint)
					}

					// There should have been the following changes:
					pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
					pristineState.ActionLog[len(pristineState.ActionLog)-1] =
						message.NewFromPlayer(actingPlayer.Name(), actingPlayer.Color(), actionMessage)
					pristineState.DeckSize = initialDeckSize
					pristineState.NumberOfReadyHints -= numberOfHintsToSubtract
					pristineState.Turn += 1
					pristineState.TurnsTakenWithEmptyDeck += 1
					pristineState.InferredCardInHand[receivingPlayerName] = updatedInferredHand
					assertGameStateAsExpectedLocallyAndRetrieved(
						testIdentifier,
						unitTest,
						gameAndDescription,
						pristineState)
				})
			}
		}
	}
}

func TestValidHintWhenDeckNotYetEmpty(unitTest *testing.T) {
	// In this case, the initial deck does not have to be created afresh
	// for each test case, because giving a hint does not involve drawing
	// from the deck.
	initialDeck :=
		[]card.Defined{
			card.Defined{
				ColorSuit:     "a",
				SequenceIndex: 3,
			},
			card.Defined{
				ColorSuit:     "b",
				SequenceIndex: 2,
			},
			card.Defined{
				ColorSuit:     "c",
				SequenceIndex: 1,
			},
		}

	initialDeckSize := len(initialDeck)

	actionMessage := "action message"
	comparisonActionLog := make([]message.FromPlayer, 3)
	numberOfCopiedMessages :=
		copy(comparisonActionLog, initialActionLogForDefaultThreePlayers)
	if numberOfCopiedMessages != 3 {
		unitTest.Fatalf(
			"copy(%v, %v) returned %v",
			comparisonActionLog,
			initialActionLogForDefaultThreePlayers,
			numberOfCopiedMessages)
	}

	numberOfHintsToSubtract := 3

	for actingPlayerIndex, actingPlayerWithHand := range threePlayersWithHands {
		actingPlayerName := actingPlayerWithHand.PlayerName

		actingPlayer := &mockPlayerState{
			actingPlayerName,
			defaultTestColor,
		}

		for receivingPlayerIndex, receivingPlayerWithHand := range threePlayersWithHands {
			if receivingPlayerIndex == actingPlayerIndex {
				continue
			}

			receivingPlayerName := receivingPlayerWithHand.PlayerName
			handSize := len(receivingPlayerWithHand.InitialHand)

			// It is not important to make valid inferred cards for this test.
			updatedInferredHand := make([]card.Inferred, handSize)
			testInferredColors := []string{"a test color", "another test color"}
			for indexInHand := 0; indexInHand < handSize; indexInHand++ {
				updatedInferredHand[indexInHand] =
					card.Inferred{
						PossibleColors:  testInferredColors,
						PossibleIndices: []int{indexInHand},
					}
			}

			gamesAndDescriptions :=
				prepareGameStates(
					unitTest,
					defaultTestRuleset,
					threePlayersWithHands,
					initialDeck,
					initialActionLogForDefaultThreePlayers)

			for _, gameAndDescription := range gamesAndDescriptions {
				testIdentifier :=
					"valid hint when deck not yet empty/" +
						gameAndDescription.PersisterDescription

				unitTest.Run(testIdentifier, func(unitTest *testing.T) {
					pristineState := prepareExpected(unitTest, gameAndDescription.GameState.Read())

					if gameAndDescription.GameState.Read().DeckSize() != initialDeckSize {
						unitTest.Fatalf(
							"initial DeckSize() %v did not match expected %v",
							gameAndDescription.GameState.Read().DeckSize(),
							initialDeckSize)
					}

					errorFromHint :=
						gameAndDescription.GameState.EnactTurnByUpdatingHandWithHint(
							context.Background(),
							actionMessage,
							actingPlayer,
							receivingPlayerName,
							updatedInferredHand,
							numberOfHintsToSubtract)

					if errorFromHint != nil {
						unitTest.Fatalf(
							"EnactTurnByUpdatingHandWithHint(%v, %+v, %v, %+v, %v)"+
								" produced error %v ",
							actionMessage,
							actingPlayer,
							receivingPlayerName,
							updatedInferredHand,
							numberOfHintsToSubtract,
							errorFromHint)
					}

					// There should have been the following changes:
					pristineState.ActionLog = gameAndDescription.GameState.Read().ActionLog()
					pristineState.ActionLog[len(pristineState.ActionLog)-1] =
						message.NewFromPlayer(actingPlayer.Name(), actingPlayer.Color(), actionMessage)
					pristineState.DeckSize = initialDeckSize
					pristineState.NumberOfReadyHints -= numberOfHintsToSubtract
					pristineState.Turn += 1
					pristineState.InferredCardInHand[receivingPlayerName] = updatedInferredHand
					assertGameStateAsExpectedLocallyAndRetrieved(
						testIdentifier,
						unitTest,
						gameAndDescription,
						pristineState)
				})
			}
		}
	}
}

func assertLogWithSingleMessageIsCorrect(
	testIdentifier string,
	unitTest *testing.T,
	logMessages []message.FromPlayer,
	expectedLogLength int,
	expectedPlayerName string,
	expectedTextColor string,
	expectedSingleMessage string,
	expectedInitialMessages []message.FromPlayer,
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
	if (firstMessage.PlayerName != expectedPlayerName) ||
		(firstMessage.TextColor != expectedTextColor) ||
		(firstMessage.MessageText != expectedSingleMessage) {
		unitTest.Fatalf(
			testIdentifier+
				"/first message %+v was not as expected: player name %v, text color %v, message %v",
			firstMessage,
			expectedPlayerName,
			expectedTextColor,
			expectedSingleMessage)
	}

	recordingTime := firstMessage.CreationTime

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
		if (actualMessage.PlayerName != expectedMessage.PlayerName) ||
			(actualMessage.TextColor != expectedMessage.TextColor) ||
			(actualMessage.MessageText != expectedMessage.MessageText) {
			unitTest.Errorf(
				testIdentifier+
					"/log\n %+v\n did not have expected other messages\n %+v\n",
				logMessages,
				expectedInitialMessages)
		}
	}
}
