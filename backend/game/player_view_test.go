package game_test

import (
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
)

func TestWrapperFunctions(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[0]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForName = gameName
	mockReadAndWriteState.ReturnForRuleset = testRuleset

	testTurn := 3
	mockReadAndWriteState.ReturnForTurn = testTurn
	testPlayersInCurrentTurnOrder :=
		[]string{
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
		}

	// This means that the viewing player should appear as the third player in the
	// list of next players.
	expectedTurnIndex := 2

	testReadyHints := 5
	testMaximumHints := testRuleset.MaximumNumberOfHints()
	mockReadAndWriteState.ReturnForNumberOfReadyHints = testReadyHints

	testMistakesMade := 5
	testMistakesForGameOver := testRuleset.NumberOfMistakesIndicatingGameOver()
	mockReadAndWriteState.ReturnForNumberOfMistakesMade = testMistakesMade

	testDeckSize := 11
	mockReadAndWriteState.ReturnForDeckSize = testDeckSize

	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder

	testChatLog :=
		[]message.Readonly{
			message.NewReadonly(testPlayersInOriginalOrder[1], "a color", "Several words"),
			message.NewReadonly(testPlayersInOriginalOrder[1], "a color", "More words"),
		}
	mockReadAndWriteState.ReturnForChatLog = testChatLog

	testActionLog :=
		[]message.Readonly{
			message.NewReadonly(testPlayersInOriginalOrder[1], "a color", "An action"),
			message.NewReadonly(testPlayersInOriginalOrder[2], "different color", "Different action"),
			message.NewReadonly(testPlayersInOriginalOrder[3], "another color", "Another action"),
		}
	mockReadAndWriteState.ReturnForActionLog = testActionLog

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	viewForPlayer, errorFromViewState :=
		gameCollection.ViewState(
			gameName,
			playerName)

	if errorFromViewState != nil {
		unitTest.Fatalf(
			"ViewState(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromViewState)
	}

	if (viewForPlayer.GameName() != gameName) ||
		(viewForPlayer.RulesetDescription() != testRuleset.FrontendDescription()) ||
		(viewForPlayer.Turn() != testTurn) ||
		(viewForPlayer.NumberOfReadyHints() != testReadyHints) ||
		(viewForPlayer.MaximumNumberOfHints() != testMaximumHints) ||
		(viewForPlayer.NumberOfMistakesMade() != testMistakesMade) ||
		(viewForPlayer.NumberOfMistakesIndicatingGameOver() != testMistakesForGameOver) ||
		(viewForPlayer.DeckSize() != testDeckSize) {
		unitTest.Fatalf(
			"player view %+v not as expected"+
				" (name %v,"+
				" ruleset description %v,"+
				" turn %v,"+
				" ready hints %v,"+
				" maximum hints %v,"+
				" mistakes made %v,"+
				" mistakes for game over %v,"+
				" deck size %v)",
			viewForPlayer,
			gameName,
			testRuleset.FrontendDescription(),
			testTurn,
			testReadyHints,
			testMaximumHints,
			testMistakesMade,
			testMistakesForGameOver,
			testDeckSize)
	}

	expectedChatLogLength := len(testChatLog)
	actualChatLog := viewForPlayer.ChatLog()
	if len(actualChatLog) != expectedChatLogLength {
		unitTest.Fatalf(
			"player view %+v did not have expected chat log %+v",
			viewForPlayer,
			testChatLog)
	}

	for messageIndex := 0; messageIndex < expectedChatLogLength; messageIndex++ {
		if actualChatLog[messageIndex] != testChatLog[messageIndex] {
			unitTest.Fatalf(
				"player view %+v did not have expected chat log %+v",
				viewForPlayer,
				testChatLog)
		}
	}

	expectedActionLogLength := len(testActionLog)
	actualActionLog := viewForPlayer.ActionLog()
	if len(actualActionLog) != expectedActionLogLength {
		unitTest.Fatalf(
			"player view %+v did not have expected action log %+v",
			viewForPlayer,
			testActionLog)
	}

	for messageIndex := 0; messageIndex < expectedActionLogLength; messageIndex++ {
		if actualActionLog[messageIndex] != testActionLog[messageIndex] {
			unitTest.Fatalf(
				"player view %+v did not have expected action log %+v",
				viewForPlayer,
				testActionLog)
		}
	}

	expectedColorsForHint := testRuleset.ColorsAvailableAsHint()
	expectedNumberOfColorsForHint := len(expectedColorsForHint)
	actualColorsForHint := viewForPlayer.ColorsAvailableAsHint()
	if len(actualColorsForHint) != expectedNumberOfColorsForHint {
		unitTest.Fatalf(
			"player view %+v did not have expected colors for hint %+v",
			viewForPlayer,
			actualColorsForHint)
	}

	for colorIndex := 0; colorIndex < expectedNumberOfColorsForHint; colorIndex++ {
		if actualColorsForHint[colorIndex] != expectedColorsForHint[colorIndex] {
			unitTest.Fatalf(
				"player view %+v did not have expected colors for hint %+v",
				viewForPlayer,
				actualColorsForHint)
		}
	}

	expectedIndicesForHint := testRuleset.IndicesAvailableAsHint()
	expectedNumberOfIndicesForHint := len(expectedIndicesForHint)
	actualIndicesForHint := viewForPlayer.IndicesAvailableAsHint()
	if len(actualIndicesForHint) != expectedNumberOfIndicesForHint {
		unitTest.Fatalf(
			"player view %+v did not have expected indices for hint %+v",
			viewForPlayer,
			actualIndicesForHint)
	}

	for indexIndex := 0; indexIndex < expectedNumberOfIndicesForHint; indexIndex++ {
		if actualIndicesForHint[indexIndex] != expectedIndicesForHint[indexIndex] {
			unitTest.Fatalf(
				"player view %+v did not have expected indices for hint %+v",
				viewForPlayer,
				actualIndicesForHint)
		}
	}

	expectedNumberOfPlayers := len(testPlayersInCurrentTurnOrder)
	actualPlayersInCurrentTurnOrder, actualCurrentPlayerIndex :=
		viewForPlayer.CurrentTurnOrder()
	if (len(actualPlayersInCurrentTurnOrder) != expectedNumberOfPlayers) ||
		(actualCurrentPlayerIndex != expectedTurnIndex) {
		unitTest.Fatalf(
			"player view current turn order %v, index %v did not have expected order %v, index %v",
			actualPlayersInCurrentTurnOrder,
			actualCurrentPlayerIndex,
			testPlayersInCurrentTurnOrder,
			expectedTurnIndex)
	}

	for turnIndex := 0; turnIndex < expectedNumberOfPlayers; turnIndex++ {
		actualPlayer := actualPlayersInCurrentTurnOrder[turnIndex]
		expectedPlayer := testPlayersInCurrentTurnOrder[turnIndex]
		if actualPlayer != expectedPlayer {
			unitTest.Fatalf(
				"player view current turn order %v, index %v did not have expected order %v, index %v",
				actualPlayersInCurrentTurnOrder,
				actualCurrentPlayerIndex,
				testPlayersInCurrentTurnOrder,
				expectedTurnIndex)
		}
	}
}

func TestGameIsFinishedWhenEnoughMistakes(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[0]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset

	// We also mock that some cards were played to test that the score from
	// played cards is ignored if the game ends because of mistakes.
	expectedPlayedCards := make(map[string][]card.Readonly, 0)
	playedColor := testRuleset.ColorSuits()[0]
	possibleIndices := testRuleset.DistinctPossibleIndices()
	expectedPlayedCards[playedColor] =
		[]card.Readonly{
			card.NewReadonly(playedColor, possibleIndices[0]),
			card.NewReadonly(playedColor, possibleIndices[1]),
		}

	mockReadAndWriteState.ReturnForPlayedForColor = expectedPlayedCards

	mockReadAndWriteState.ReturnForNumberOfMistakesMade =
		testRuleset.NumberOfMistakesIndicatingGameOver()

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	viewForPlayer, errorFromViewState :=
		gameCollection.ViewState(
			gameName,
			playerName)

	if errorFromViewState != nil {
		unitTest.Fatalf(
			"ViewState(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromViewState)
	}

	actualGameIsFinished := viewForPlayer.GameIsFinished()
	if !actualGameIsFinished {
		unitTest.Fatalf(
			"GameIsFinished() produced %v when the number of mistakes was too high",
			actualGameIsFinished)
	}

	actualScore := viewForPlayer.Score()
	expectedScore := 0
	if actualScore != expectedScore {
		unitTest.Fatalf(
			"player view %+v returned %v for Score() rather than expected %v because game ended due to mistakes",
			viewForPlayer,
			actualScore,
			expectedScore)
	}
}

func TestGameIsNotFinishedWhenDeckNotEmpty(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[0]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset

	mockReadAndWriteState.ReturnForDeckSize = 1

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	viewForPlayer, errorFromViewState :=
		gameCollection.ViewState(
			gameName,
			playerName)

	if errorFromViewState != nil {
		unitTest.Fatalf(
			"ViewState(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromViewState)
	}

	actualGameIsFinished := viewForPlayer.GameIsFinished()
	if actualGameIsFinished {
		unitTest.Fatalf(
			"GameIsFinished() produced %v when the deck was not empty",
			actualGameIsFinished)
	}
}

func TestGameIsFinishedWhenTurnsWithEmptyDeckTooLarge(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[0]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurnsTakenWithEmptyDeck =
		len(testPlayersInOriginalOrder)

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	viewForPlayer, errorFromViewState :=
		gameCollection.ViewState(
			gameName,
			playerName)

	if errorFromViewState != nil {
		unitTest.Fatalf(
			"ViewState(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromViewState)
	}

	actualGameIsFinished := viewForPlayer.GameIsFinished()
	if !actualGameIsFinished {
		unitTest.Fatalf(
			"GameIsFinished() produced %v when the number of turns with an"+
				" empty deck is equal to the number of participants",
			actualGameIsFinished)
	}
}

func TestPlayedSequencesWhenSomeAreEmpty(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[0]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset

	expectedPlayedCards := make(map[string][]card.Readonly, 0)
	colorSuits := testRuleset.ColorSuits()
	numberOfSuits := len(colorSuits)
	if numberOfSuits < 3 {
		unitTest.Fatalf(
			"testRuleset.ColorSuits() %v has not enough colors (test needs at least 3)",
			testRuleset.ColorSuits())
	}

	sequenceIndices := testRuleset.DistinctPossibleIndices()
	if len(sequenceIndices) < 4 {
		unitTest.Fatalf(
			"testRuleset.DistinctPossibleIndices() %v has not enough indices (test needs at least 4)",
			testRuleset.DistinctPossibleIndices())
	}

	colorWithSeveralCards := colorSuits[0]
	severalCardsForColor :=
		[]card.Readonly{
			card.NewReadonly(colorWithSeveralCards, sequenceIndices[0]),
			card.NewReadonly(colorWithSeveralCards, sequenceIndices[1]),
			card.NewReadonly(colorWithSeveralCards, sequenceIndices[3]),
		}

	numberWhichIsSeveral := len(severalCardsForColor)
	expectedPlayedCards[colorWithSeveralCards] = severalCardsForColor

	colorWithSingleCard := colorSuits[2]
	singleCardForColor := card.NewReadonly(colorWithSingleCard, sequenceIndices[0])
	expectedPlayedCards[colorWithSingleCard] =
		[]card.Readonly{
			singleCardForColor,
		}

	// The score should be equal to the number of cards played.
	expectedScore := numberWhichIsSeveral + 1

	mockReadAndWriteState.ReturnForPlayedForColor = expectedPlayedCards

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	viewForPlayer, errorFromViewState :=
		gameCollection.ViewState(
			gameName,
			playerName)

	if errorFromViewState != nil {
		unitTest.Fatalf(
			"ViewState(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromViewState)
	}

	actualScore := viewForPlayer.Score()
	if actualScore != expectedScore {
		unitTest.Fatalf(
			"player view %+v returned %v for Score() rather than expected %v",
			viewForPlayer,
			actualScore,
			expectedScore)
	}

	actualPlayedCards := viewForPlayer.PlayedCards()

	if len(actualPlayedCards) != numberOfSuits {
		unitTest.Fatalf(
			"player view %+v did not have expected %v sequences of played cards",
			viewForPlayer,
			numberOfSuits)
	}

	foundColorWithSeveralCards := false
	foundColorWithSingleCard := false

	for suitIndex := 0; suitIndex < numberOfSuits; suitIndex++ {
		actualPile := actualPlayedCards[suitIndex]

		if len(actualPile) > 0 {
			pileColor := actualPile[0].ColorSuit()

			if pileColor == colorWithSeveralCards {
				foundColorWithSeveralCards = true

				if len(actualPile) != numberWhichIsSeveral {
					unitTest.Fatalf(
						"player view %+v did not have expected sequence %+v for color %v",
						viewForPlayer,
						severalCardsForColor,
						colorWithSeveralCards)
				}

				for indexInPile := 0; indexInPile < numberWhichIsSeveral; indexInPile++ {
					if actualPile[indexInPile] != severalCardsForColor[indexInPile] {
						unitTest.Fatalf(
							"player view %+v did not have expected sequence %+v for color %v",
							viewForPlayer,
							severalCardsForColor,
							colorWithSeveralCards)
					}
				}
			} else if pileColor == colorWithSingleCard {
				foundColorWithSingleCard = true

				if len(actualPile) != 1 {
					unitTest.Fatalf(
						"player view %+v did not have expected sequence %+v for color %v",
						viewForPlayer,
						singleCardForColor,
						colorWithSingleCard)
				}

				if actualPile[0] != singleCardForColor {
					unitTest.Fatalf(
						"player view %+v did not have expected sequence %+v for color %v",
						viewForPlayer,
						singleCardForColor,
						colorWithSingleCard)
				}
			}
		}
	}

	if !foundColorWithSeveralCards {
		unitTest.Fatalf(
			"player view %+v did not have expected sequence %+v for color %v",
			viewForPlayer,
			severalCardsForColor,
			colorWithSeveralCards)
	}

	if !foundColorWithSingleCard {
		unitTest.Fatalf(
			"player view %+v did not have expected sequence %+v for color %v",
			viewForPlayer,
			singleCardForColor,
			colorWithSingleCard)
	}
}

func TestPlayedSequencesWhenAllAreNonempty(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[0]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset

	expectedPlayedCards := make(map[string][]card.Readonly, 0)
	colorSuits := testRuleset.ColorSuits()
	numberOfSuits := len(colorSuits)
	if numberOfSuits < 2 {
		unitTest.Fatalf(
			"testRuleset.ColorSuits() %v has not enough colors (test needs at least 2)",
			testRuleset.ColorSuits())
	}

	sequenceIndices := testRuleset.DistinctPossibleIndices()
	if len(sequenceIndices) < numberOfSuits {
		unitTest.Fatalf(
			"testRuleset.DistinctPossibleIndices() %v has not enough indices (test needs at least %v)",
			testRuleset.DistinctPossibleIndices(),
			numberOfSuits)
	}

	// The score should be equal to the number of cards played.
	expectedScore := 0
	for colorCount := 0; colorCount < numberOfSuits; colorCount++ {
		colorSuit := colorSuits[colorCount]
		sequenceForColor := make([]card.Readonly, colorCount+1)
		for cardCount := 0; cardCount <= colorCount; cardCount++ {
			sequenceForColor[cardCount] = card.NewReadonly(colorSuit, sequenceIndices[cardCount])
			expectedScore++
		}

		expectedPlayedCards[colorSuit] = sequenceForColor
	}

	mockReadAndWriteState.ReturnForPlayedForColor = expectedPlayedCards

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	viewForPlayer, errorFromViewState :=
		gameCollection.ViewState(
			gameName,
			playerName)

	if errorFromViewState != nil {
		unitTest.Fatalf(
			"ViewState(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromViewState)
	}

	actualScore := viewForPlayer.Score()
	if actualScore != expectedScore {
		unitTest.Fatalf(
			"player view %+v returned %v for Score() rather than expected %v",
			viewForPlayer,
			actualScore,
			expectedScore)
	}

	actualPlayedCards := viewForPlayer.PlayedCards()

	if len(actualPlayedCards) != numberOfSuits {
		unitTest.Fatalf(
			"player view %+v did not have expected %v sequences of played cards",
			viewForPlayer,
			numberOfSuits)
	}

	for suitIndex := 0; suitIndex < numberOfSuits; suitIndex++ {
		actualPile := actualPlayedCards[suitIndex]

		if len(actualPile) <= 0 {
			unitTest.Fatalf(
				"player view %+v did not have expected sequences %+v (at least one empty pile)",
				viewForPlayer,
				expectedPlayedCards)
		}

		pileColor := actualPile[0].ColorSuit()

		expectedPile := expectedPlayedCards[pileColor]
		expectedPileSize := len(expectedPile)
		if len(actualPile) != expectedPileSize {
			unitTest.Fatalf(
				"player view %+v did not have expected sequence %+v for color %v",
				viewForPlayer,
				expectedPile,
				pileColor)
		}

		for indexInPile := 0; indexInPile < expectedPileSize; indexInPile++ {
			if actualPile[indexInPile] != expectedPile[indexInPile] {
				unitTest.Fatalf(
					"player view %+v did not have expected sequence %+v for color %v",
					viewForPlayer,
					expectedPile,
					pileColor)
			}
		}
	}
}

func TestDiscardedCards(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[0]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset

	expectedDiscardedCards := make(map[card.Readonly]int, 0)
	colorSuits := testRuleset.ColorSuits()
	numberOfSuits := len(colorSuits)
	if numberOfSuits < 3 {
		unitTest.Fatalf(
			"testRuleset.ColorSuits() %v has not enough colors (test needs at least 3)",
			testRuleset.ColorSuits())
	}
	sequenceIndices := testRuleset.DistinctPossibleIndices()
	if len(sequenceIndices) < 4 {
		unitTest.Fatalf(
			"testRuleset.DistinctPossibleIndices() %v has not enough indices (test needs at least 4)",
			testRuleset.DistinctPossibleIndices())
	}

	// We set up with several copies each of cards with the same color and different indices.
	expectedDiscardedCards[card.NewReadonly(colorSuits[0], sequenceIndices[0])] = 3
	expectedDiscardedCards[card.NewReadonly(colorSuits[0], sequenceIndices[1])] = 1
	expectedDiscardedCards[card.NewReadonly(colorSuits[0], sequenceIndices[3])] = 1

	// We also add several copies of a single sequence index of a different color.
	expectedDiscardedCards[card.NewReadonly(colorSuits[2], sequenceIndices[1])] = 2

	mockReadAndWriteState.ReturnForNumberOfDiscardedCards = expectedDiscardedCards

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	viewForPlayer, errorFromViewState :=
		gameCollection.ViewState(
			gameName,
			playerName)

	if errorFromViewState != nil {
		unitTest.Fatalf(
			"ViewState(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromViewState)
	}

	actualDiscardedCards := viewForPlayer.DiscardedCards()

	for _, discardedCard := range actualDiscardedCards {
		expectedDiscardedCards[discardedCard] -= 1
	}

	for _, remainingCount := range expectedDiscardedCards {
		if remainingCount != 0 {
			unitTest.Fatalf(
				"player view %+v after removing actual discards %v from expected had %v remaining expected",
				viewForPlayer,
				actualDiscardedCards,
				expectedDiscardedCards)
		}
	}
}

func TestPlayerIsForbiddenFromSeeingOwnHand(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[0]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	viewForPlayer, errorFromViewState :=
		gameCollection.ViewState(
			gameName,
			playerName)

	if errorFromViewState != nil {
		unitTest.Fatalf(
			"ViewState(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromViewState)
	}

	visibleHand, playerChatColor, errorFromVisibleHand := viewForPlayer.VisibleHand(playerName)

	if errorFromVisibleHand == nil {
		unitTest.Fatalf(
			"player view %+v produced nil error when trying to view own hand, saw %v with color %v",
			viewForPlayer,
			visibleHand,
			playerChatColor)
	}
}

func TestErrorFromPlayerProviderAffectsVisibleHandCorrectly(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
		}
	viewingPlayer := testPlayersInOriginalOrder[1]
	playerWithVisibleHand := testPlayersInOriginalOrder[0]
	gameCollection, mockPersister, mockPlayerProvider :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockPlayerProvider.MockPlayers = make(map[string]*mockPlayerState, 0)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	viewForPlayer, errorFromViewState :=
		gameCollection.ViewState(
			gameName,
			viewingPlayer)

	if errorFromViewState != nil {
		unitTest.Fatalf(
			"ViewState(%v, %v) produced error %v",
			gameName,
			viewingPlayer,
			errorFromViewState)
	}

	actualVisibleHand, actualPlayerColor, errorFromVisibleHand :=
		viewForPlayer.VisibleHand(playerWithVisibleHand)

	if errorFromVisibleHand == nil {
		unitTest.Fatalf(
			"VisibleHand(%v) from player view %+v did not produced expected error, instead produced %+v with color %v",
			playerWithVisibleHand,
			viewForPlayer,
			actualVisibleHand,
			actualPlayerColor)
	}
}

func TestPlayerSeesOtherHandCorrectly(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
		}
	viewingPlayer := testPlayersInOriginalOrder[1]
	playerWithVisibleHand := testPlayersInOriginalOrder[0]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset

	colorSuits := testRuleset.ColorSuits()
	numberOfSuits := len(colorSuits)
	if numberOfSuits < 2 {
		unitTest.Fatalf(
			"testRuleset.ColorSuits() %v has not enough colors (test needs at least 2)",
			testRuleset.ColorSuits())
	}
	sequenceIndices := testRuleset.DistinctPossibleIndices()
	if len(sequenceIndices) < 2 {
		unitTest.Fatalf(
			"testRuleset.DistinctPossibleIndices() %v has not enough indices (test needs at least 2)",
			testRuleset.DistinctPossibleIndices())
	}

	firstPlayerHand :=
		[]card.Readonly{
			card.NewReadonly(colorSuits[0], sequenceIndices[0]),
			card.NewReadonly(colorSuits[0], sequenceIndices[0]),
			card.NewReadonly(colorSuits[1], sequenceIndices[0]),
			card.NewReadonly(colorSuits[0], sequenceIndices[1]),
			card.NewReadonly(colorSuits[1], sequenceIndices[1]),
		}

	lastPlayerHand :=
		[]card.Readonly{
			card.NewReadonly(colorSuits[1], sequenceIndices[0]),
			card.NewReadonly(colorSuits[1], sequenceIndices[0]),
			card.NewReadonly(colorSuits[0], sequenceIndices[0]),
			card.NewReadonly(colorSuits[0], sequenceIndices[1]),
			card.NewReadonly(colorSuits[1], sequenceIndices[1]),
		}

	expectedVisibleHands := make(map[string][]card.Readonly, 0)
	expectedVisibleHands[playerWithVisibleHand] = firstPlayerHand
	expectedVisibleHands[playerNamesAvailableInTest[2]] = lastPlayerHand

	mockReadAndWriteState.ReturnForVisibleHand = expectedVisibleHands

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	viewForPlayer, errorFromViewState :=
		gameCollection.ViewState(
			gameName,
			viewingPlayer)

	if errorFromViewState != nil {
		unitTest.Fatalf(
			"ViewState(%v, %v) produced error %v",
			gameName,
			viewingPlayer,
			errorFromViewState)
	}

	actualVisibleHand, _, errorFromVisibleHand :=
		viewForPlayer.VisibleHand(playerWithVisibleHand)

	if errorFromVisibleHand != nil {
		unitTest.Fatalf(
			"VisibleHand(%v) from player view %+v produced error %v",
			playerWithVisibleHand,
			viewForPlayer,
			errorFromVisibleHand)
	}

	assertReadonlyCardSlicesMatch(
		"view visible hand",
		unitTest,
		actualVisibleHand,
		firstPlayerHand)
}

func TestPlayerSeesOwnInferredHandCorrectly(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
		}
	viewingPlayer := testPlayersInOriginalOrder[0]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset

	colorSuits := testRuleset.ColorSuits()
	numberOfSuits := len(colorSuits)
	if numberOfSuits < 2 {
		unitTest.Fatalf(
			"testRuleset.ColorSuits() %v has not enough colors (test needs at least 2)",
			testRuleset.ColorSuits())
	}
	sequenceIndices := testRuleset.DistinctPossibleIndices()
	if len(sequenceIndices) < 2 {
		unitTest.Fatalf(
			"testRuleset.DistinctPossibleIndices() %v has not enough indices (test needs at least 2)",
			testRuleset.DistinctPossibleIndices())
	}

	viewingPlayerHand :=
		[]card.Inferred{
			card.NewInferred(colorSuits, sequenceIndices),
			card.NewInferred([]string{colorSuits[0]}, sequenceIndices),
			card.NewInferred(colorSuits, []int{sequenceIndices[0]}),
			card.NewInferred([]string{colorSuits[0]}, []int{sequenceIndices[0]}),
			card.NewInferred(
				[]string{colorSuits[0], colorSuits[1]},
				[]int{sequenceIndices[0], sequenceIndices[1]}),
		}

	otherPlayerHand := []card.Inferred{}

	expectedInferredHands := make(map[string][]card.Inferred, 0)
	expectedInferredHands[viewingPlayer] = viewingPlayerHand
	expectedInferredHands[playerNamesAvailableInTest[1]] = otherPlayerHand

	mockReadAndWriteState.ReturnForInferredHand = expectedInferredHands

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	viewForPlayer, errorFromViewState :=
		gameCollection.ViewState(
			gameName,
			viewingPlayer)

	if errorFromViewState != nil {
		unitTest.Fatalf(
			"ViewState(%v, %v) produced error %v",
			gameName,
			viewingPlayer,
			errorFromViewState)
	}

	actualInferredHand, errorFromInferredHand :=
		viewForPlayer.KnowledgeOfOwnHand(viewingPlayer)

	if errorFromInferredHand != nil {
		unitTest.Fatalf(
			"KnowledgeOfOwnHand(%v) from player view %+v produced error %v",
			viewingPlayer,
			viewForPlayer,
			errorFromInferredHand)
	}

	expectedHandSize := len(viewingPlayerHand)

	if len(actualInferredHand) != expectedHandSize {
		unitTest.Fatalf(
			"inferred hand %+v did not match expected %+v in length",
			actualInferredHand,
			errorFromViewState)
	}

	for indexInHand := 0; indexInHand < expectedHandSize; indexInHand++ {
		assertInferredCardPossibilitiesCorrect(
			fmt.Sprintf("own inferred hand at index %v", indexInHand),
			unitTest,
			actualInferredHand[indexInHand],
			viewingPlayerHand[indexInHand].PossibleColors(),
			viewingPlayerHand[indexInHand].PossibleIndices())
	}
}
