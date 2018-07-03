package game_test

import (
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game/card"
)

func TestRecordChatMessageReturnsErrorIfStateProducesError(unitTest *testing.T) {
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
	mockReadAndWriteState.ReturnForNontestError = fmt.Errorf("expected error")
	mockReadAndWriteState.TestErrorForRecordChatMessage = nil

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	chatMessage := "Some irrelevant text"
	errorFromRecordChatMessage := executorForPlayer.RecordChatMessage(chatMessage)

	if errorFromRecordChatMessage == nil {
		unitTest.Fatalf(
			"RecordChatMessage(%v) produced nil error",
			chatMessage)
	}
}

func TestRecordChatMessageReturnsNoErrorIfStateProducesNoError(unitTest *testing.T) {
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
	mockReadAndWriteState.ReturnForNontestError = nil
	mockReadAndWriteState.TestErrorForRecordChatMessage = nil

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	chatMessage := "Some irrelevant text"
	errorFromRecordChatMessage :=
		executorForPlayer.RecordChatMessage(chatMessage)

	if errorFromRecordChatMessage != nil {
		unitTest.Fatalf(
			"RecordChatMessage(%v) produced error %v",
			chatMessage,
			errorFromRecordChatMessage)
	}
}

func TestRejectTakeTurnByDiscardingIfTooManyMistakesMade(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForNumberOfMistakesMade =
		testRuleset.NumberOfMistakesIndicatingGameOver()

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	indexInHandToDiscard := 1
	errorFromTakeTurnByDiscarding :=
		executorForPlayer.TakeTurnByDiscarding(indexInHandToDiscard)

	if errorFromTakeTurnByDiscarding == nil {
		unitTest.Fatalf(
			"TakeTurnByDiscarding(%v) produced nil error when not player's turn",
			indexInHandToDiscard)
	}
}

func TestRejectTakeTurnByDiscardingIfNotPlayerTurn(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 1

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	indexInHandToDiscard := 1
	errorFromTakeTurnByDiscarding :=
		executorForPlayer.TakeTurnByDiscarding(indexInHandToDiscard)

	if errorFromTakeTurnByDiscarding == nil {
		unitTest.Fatalf(
			"TakeTurnByDiscarding(%v) produced nil error when not player's turn",
			indexInHandToDiscard)
	}
}

func TestRejectTakeTurnByDiscardingIfErrorGettingHand(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 3
	mockReadAndWriteState.ReturnForNontestError = fmt.Errorf("expected error")

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	indexInHandToDiscard := 1
	errorFromTakeTurnByDiscarding :=
		executorForPlayer.TakeTurnByDiscarding(indexInHandToDiscard)

	if errorFromTakeTurnByDiscarding == nil {
		unitTest.Fatalf(
			"TakeTurnByDiscarding(%v) produced nil error instead of error around error from getting hand",
			indexInHandToDiscard)
	}
}

func TestRejectTakeTurnByDiscardingIfPlayerHandAlreadyTooSmall(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 3
	reducedSizeHands := make(map[string][]card.Readonly, 1)
	reducedSizeHands[playerName] = []card.Readonly{}
	mockReadAndWriteState.ReturnForVisibleHand = reducedSizeHands

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	indexInHandToDiscard := 1
	errorFromTakeTurnByDiscarding :=
		executorForPlayer.TakeTurnByDiscarding(indexInHandToDiscard)

	if errorFromTakeTurnByDiscarding == nil {
		unitTest.Fatalf(
			"TakeTurnByDiscarding(%v) produced nil error when player hand is already too small",
			indexInHandToDiscard)
	}
}

func TestRejectTakeTurnByDiscardingIfIndexNegative(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 3

	// This test does not need the cards to be initialized correctly,
	// just that the hand slice is the correct length.
	correctHandSize :=
		testRuleset.NumberOfCardsInPlayerHand(len(testPlayersInOriginalOrder))
	correctSizeHands := make(map[string][]card.Readonly, 1)
	correctSizeHands[playerName] = make([]card.Readonly, correctHandSize)
	mockReadAndWriteState.ReturnForVisibleHand = correctSizeHands

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	indexInHandToDiscard := -1
	errorFromTakeTurnByDiscarding :=
		executorForPlayer.TakeTurnByDiscarding(indexInHandToDiscard)

	if errorFromTakeTurnByDiscarding == nil {
		unitTest.Fatalf(
			"TakeTurnByDiscarding(%v) produced nil error when index is negative",
			indexInHandToDiscard)
	}
}

func TestRejectTakeTurnByDiscardingIfIndexTooLarge(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 3

	// This test does not need the cards to be initialized correctly,
	// just that the hand slice is the correct length.
	correctHandSize :=
		testRuleset.NumberOfCardsInPlayerHand(len(testPlayersInOriginalOrder))
	correctSizeHands := make(map[string][]card.Readonly, 1)
	correctSizeHands[playerName] = make([]card.Readonly, correctHandSize)
	mockReadAndWriteState.ReturnForVisibleHand = correctSizeHands

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	indexInHandToDiscard := correctHandSize
	errorFromTakeTurnByDiscarding :=
		executorForPlayer.TakeTurnByDiscarding(indexInHandToDiscard)

	if errorFromTakeTurnByDiscarding == nil {
		unitTest.Fatalf(
			"TakeTurnByDiscarding(%v) produced nil error when index is negative",
			indexInHandToDiscard)
	}
}

func TestTakeTurnByDiscardingWhenAlreadyMaximumHints(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, playerProvider :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 3
	mockReadAndWriteState.ReturnForNumberOfReadyHints =
		testRuleset.MaximumNumberOfHints()

	indexInHandToDiscard := 2

	// This test does not need the cards to be initialized correctly,
	// just that the hand slice is the correct length.
	correctHandSize :=
		testRuleset.NumberOfCardsInPlayerHand(len(testPlayersInOriginalOrder))
	correctSizeHands := make(map[string][]card.Readonly, 1)
	correctSizeHands[playerName] = make([]card.Readonly, correctHandSize)
	expectedDiscardedCard := card.NewReadonly("some_color", 123)
	correctSizeHands[playerName][indexInHandToDiscard] = expectedDiscardedCard
	mockReadAndWriteState.ReturnForVisibleHand = correctSizeHands

	mockReadAndWriteState.TestErrorForEnactTurnByDiscardingAndReplacing = nil

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	errorFromTakeTurnByDiscarding :=
		executorForPlayer.TakeTurnByDiscarding(indexInHandToDiscard)

	if errorFromTakeTurnByDiscarding != nil {
		unitTest.Fatalf(
			"TakeTurnByDiscarding(%v) produced unexpected error %v",
			indexInHandToDiscard,
			errorFromTakeTurnByDiscarding)
	}

	actualArguments :=
		mockReadAndWriteState.ArgumentsFromEnactTurnByDiscardingAndReplacing

	if len(actualArguments) != 1 {
		unitTest.Fatalf(
			"list of argument sets %v did not have exactly 1 element",
			actualArguments)
	}

	expectedPlayerState, errorFromGetPlayer := playerProvider.Get(playerName)
	if errorFromGetPlayer != nil {
		unitTest.Fatalf(
			"mock player provider's Get(%v) produced unexpected error %v",
			playerName,
			errorFromGetPlayer)
	}

	actualArgument := actualArguments[0]
	expectedActionMessage :=
		fmt.Sprintf(
			"discards card %v %v",
			expectedDiscardedCard.ColorSuit(),
			expectedDiscardedCard.SequenceIndex())
	if (actualArgument.MessageString != expectedActionMessage) ||
		(actualArgument.PlayerState != expectedPlayerState) ||
		(actualArgument.IndexInt != indexInHandToDiscard) ||
		(actualArgument.HintsInt != 0) ||
		(actualArgument.MistakesInt != 0) {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) resulted in wrong call to EnactTurnByDiscardingAndReplacing(...): %v",
			gameName,
			playerName,
			actualArguments[0])
	}

	assertInferredCardPossibilitiesCorrect(
		"knowledge of drawn card when discarding at maximum hints",
		unitTest,
		actualArgument.DrawnInferred,
		testRuleset.ColorSuits(),
		testRuleset.DistinctPossibleIndices())
}

func TestTakeTurnByDiscardingWhenLessThanMaximumHints(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, playerProvider :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 3
	mockReadAndWriteState.ReturnForNumberOfReadyHints =
		testRuleset.MaximumNumberOfHints() - 2

	indexInHandToDiscard := 2

	// This test does not need the cards to be initialized correctly,
	// just that the hand slice is the correct length.
	correctHandSize :=
		testRuleset.NumberOfCardsInPlayerHand(len(testPlayersInOriginalOrder))
	correctSizeHands := make(map[string][]card.Readonly, 1)
	correctSizeHands[playerName] = make([]card.Readonly, correctHandSize)
	expectedDiscardedCard := card.NewReadonly("some_color", 123)
	correctSizeHands[playerName][indexInHandToDiscard] = expectedDiscardedCard
	mockReadAndWriteState.ReturnForVisibleHand = correctSizeHands

	mockReadAndWriteState.TestErrorForEnactTurnByDiscardingAndReplacing = nil

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	errorFromTakeTurnByDiscarding :=
		executorForPlayer.TakeTurnByDiscarding(indexInHandToDiscard)

	if errorFromTakeTurnByDiscarding != nil {
		unitTest.Fatalf(
			"TakeTurnByDiscarding(%v) produced unexpected error %v",
			indexInHandToDiscard,
			errorFromTakeTurnByDiscarding)
	}

	actualArguments :=
		mockReadAndWriteState.ArgumentsFromEnactTurnByDiscardingAndReplacing

	if len(actualArguments) != 1 {
		unitTest.Fatalf(
			"list of argument sets %v did not have exactly 1 element",
			actualArguments)
	}

	expectedPlayerState, errorFromGetPlayer := playerProvider.Get(playerName)
	if errorFromGetPlayer != nil {
		unitTest.Fatalf(
			"mock player provider's Get(%v) produced unexpected error %v",
			playerName,
			errorFromGetPlayer)
	}

	actualArgument := actualArguments[0]
	expectedActionMessage :=
		fmt.Sprintf(
			"discards card %v %v",
			expectedDiscardedCard.ColorSuit(),
			expectedDiscardedCard.SequenceIndex())
	if (actualArgument.MessageString != expectedActionMessage) ||
		(actualArgument.PlayerState != expectedPlayerState) ||
		(actualArgument.IndexInt != indexInHandToDiscard) ||
		(actualArgument.HintsInt != 1) ||
		(actualArgument.MistakesInt != 0) {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) resulted in wrong call to EnactTurnByDiscardingAndReplacing(...): %v",
			gameName,
			playerName,
			actualArguments[0])
	}

	assertInferredCardPossibilitiesCorrect(
		"knowledge of drawn card when discarding with less than maximum hints",
		unitTest,
		actualArgument.DrawnInferred,
		testRuleset.ColorSuits(),
		testRuleset.DistinctPossibleIndices())
}

func TestRejectTakeTurnByPlayingIfTooManyMistakesMade(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForNumberOfMistakesMade =
		testRuleset.NumberOfMistakesIndicatingGameOver()

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	indexInHandToPlay := 1
	errorFromTakeTurnByPlaying :=
		executorForPlayer.TakeTurnByPlaying(indexInHandToPlay)

	if errorFromTakeTurnByPlaying == nil {
		unitTest.Fatalf(
			"TakeTurnByPlaying(%v) produced nil error when not player's turn",
			indexInHandToPlay)
	}
}

func TestRejectTakeTurnByPlayingIfNotPlayerTurn(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 1

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	indexInHandToPlay := 1
	errorFromTakeTurnByPlaying :=
		executorForPlayer.TakeTurnByPlaying(indexInHandToPlay)

	if errorFromTakeTurnByPlaying == nil {
		unitTest.Fatalf(
			"TakeTurnByPlaying(%v) produced nil error when not player's turn",
			indexInHandToPlay)
	}
}

func TestRejectTakeTurnByPlayingIfErrorGettingHand(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 3
	mockReadAndWriteState.ReturnForNontestError = fmt.Errorf("expected error")

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	indexInHandToPlay := 1
	errorFromTakeTurnByPlaying :=
		executorForPlayer.TakeTurnByPlaying(indexInHandToPlay)

	if errorFromTakeTurnByPlaying == nil {
		unitTest.Fatalf(
			"TakeTurnByPlaying(%v) produced nil error instead of error around error from getting hand",
			indexInHandToPlay)
	}
}

func TestRejectTakeTurnByPlayingIfPlayerHandAlreadyTooSmall(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 3
	reducedSizeHands := make(map[string][]card.Readonly, 1)
	reducedSizeHands[playerName] = []card.Readonly{}
	mockReadAndWriteState.ReturnForVisibleHand = reducedSizeHands

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	indexInHandToPlay := 1
	errorFromTakeTurnByPlaying :=
		executorForPlayer.TakeTurnByPlaying(indexInHandToPlay)

	if errorFromTakeTurnByPlaying == nil {
		unitTest.Fatalf(
			"TakeTurnByPlaying(%v) produced nil error when player hand is already too small",
			indexInHandToPlay)
	}
}

func TestRejectTakeTurnByPlayingIfIndexNegative(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 3

	// This test does not need the cards to be initialized correctly,
	// just that the hand slice is the correct length.
	correctHandSize :=
		testRuleset.NumberOfCardsInPlayerHand(len(testPlayersInOriginalOrder))
	correctSizeHands := make(map[string][]card.Readonly, 1)
	correctSizeHands[playerName] = make([]card.Readonly, correctHandSize)
	mockReadAndWriteState.ReturnForVisibleHand = correctSizeHands

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	indexInHandToPlay := -1
	errorFromTakeTurnByPlaying :=
		executorForPlayer.TakeTurnByPlaying(indexInHandToPlay)

	if errorFromTakeTurnByPlaying == nil {
		unitTest.Fatalf(
			"TakeTurnByPlaying(%v) produced nil error when index is negative",
			indexInHandToPlay)
	}
}

func TestRejectTakeTurnByPlayingIfIndexTooLarge(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 3

	// This test does not need the cards to be initialized correctly,
	// just that the hand slice is the correct length.
	correctHandSize :=
		testRuleset.NumberOfCardsInPlayerHand(len(testPlayersInOriginalOrder))
	correctSizeHands := make(map[string][]card.Readonly, 1)
	correctSizeHands[playerName] = make([]card.Readonly, correctHandSize)
	mockReadAndWriteState.ReturnForVisibleHand = correctSizeHands

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	indexInHandToPlay := correctHandSize
	errorFromTakeTurnByPlaying :=
		executorForPlayer.TakeTurnByPlaying(indexInHandToPlay)

	if errorFromTakeTurnByPlaying == nil {
		unitTest.Fatalf(
			"TakeTurnByPlaying(%v) produced nil error when index is negative",
			indexInHandToPlay)
	}
}

func TestMistakeWhenTakingTurnByPlaying(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, playerProvider :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 3
	mockReadAndWriteState.ReturnForNumberOfReadyHints =
		testRuleset.MaximumNumberOfHints()

	indexInHandToAttemptToPlay := 2

	// This test does not need all the cards to be initialized correctly,
	// just that the hand slice is the correct length and that the played
	// card is valid.
	testSuit := testRuleset.ColorSuits()[1]
	correctHandSize :=
		testRuleset.NumberOfCardsInPlayerHand(len(testPlayersInOriginalOrder))
	correctSizeHands := make(map[string][]card.Readonly, 1)
	correctSizeHands[playerName] = make([]card.Readonly, correctHandSize)
	expectedDiscardedCard := card.NewReadonly(testSuit, 3)
	correctSizeHands[playerName][indexInHandToAttemptToPlay] = expectedDiscardedCard
	mockReadAndWriteState.ReturnForVisibleHand = correctSizeHands

	alreadyPlayed := make(map[string][]card.Readonly, 0)
	alreadyPlayed[testSuit] = []card.Readonly{card.NewReadonly(testSuit, 1)}
	mockReadAndWriteState.ReturnForPlayedForColor = alreadyPlayed

	mockReadAndWriteState.TestErrorForEnactTurnByDiscardingAndReplacing = nil

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	errorFromTakeTurnByPlaying :=
		executorForPlayer.TakeTurnByPlaying(indexInHandToAttemptToPlay)

	if errorFromTakeTurnByPlaying != nil {
		unitTest.Fatalf(
			"TakeTurnByPlaying(%v) produced unexpected error %v",
			indexInHandToAttemptToPlay,
			errorFromTakeTurnByPlaying)
	}

	actualArguments :=
		mockReadAndWriteState.ArgumentsFromEnactTurnByDiscardingAndReplacing

	if len(actualArguments) != 1 {
		unitTest.Fatalf(
			"list of argument sets %v did not have exactly 1 element",
			actualArguments)
	}

	expectedPlayerState, errorFromGetPlayer := playerProvider.Get(playerName)
	if errorFromGetPlayer != nil {
		unitTest.Fatalf(
			"mock player provider's Get(%v) produced unexpected error %v",
			playerName,
			errorFromGetPlayer)
	}

	actualArgument := actualArguments[0]
	expectedActionMessage :=
		fmt.Sprintf(
			"mistakenly tries to play card %v %v",
			expectedDiscardedCard.ColorSuit(),
			expectedDiscardedCard.SequenceIndex())
	if (actualArgument.MessageString != expectedActionMessage) ||
		(actualArgument.PlayerState != expectedPlayerState) ||
		(actualArgument.IndexInt != indexInHandToAttemptToPlay) ||
		(actualArgument.HintsInt != 0) ||
		(actualArgument.MistakesInt != 1) {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) resulted in wrong call to EnactTurnByDiscardingAndReplacing(...): %v",
			gameName,
			playerName,
			actualArguments[0])
	}

	assertInferredCardPossibilitiesCorrect(
		"knowledge of drawn card when discarding at maximum hints",
		unitTest,
		actualArgument.DrawnInferred,
		testRuleset.ColorSuits(),
		testRuleset.DistinctPossibleIndices())
}

func TestTakeTurnByPlayingWithNoBonusHintWhenLessThanMaximumHints(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, playerProvider :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 3
	mockReadAndWriteState.ReturnForNumberOfReadyHints =
		testRuleset.MaximumNumberOfHints() - 2

	indexInHandToPlay := 0

	// This test does not need all the cards to be initialized correctly,
	// just that the hand slice is the correct length and that the played
	// card is valid. With the standard ruleset, there is a bonus hint if
	// the sequence index is five or greater.
	testSuit := testRuleset.ColorSuits()[1]
	sequenceIndexToPlay := 4
	correctHandSize :=
		testRuleset.NumberOfCardsInPlayerHand(len(testPlayersInOriginalOrder))
	correctSizeHands := make(map[string][]card.Readonly, 1)
	correctSizeHands[playerName] = make([]card.Readonly, correctHandSize)
	expectedPlayedCard := card.NewReadonly(testSuit, sequenceIndexToPlay)
	correctSizeHands[playerName][indexInHandToPlay] = expectedPlayedCard
	mockReadAndWriteState.ReturnForVisibleHand = correctSizeHands

	alreadyPlayed := make(map[string][]card.Readonly, 0)
	alreadyPlayed[testSuit] =
		[]card.Readonly{card.NewReadonly(testSuit, sequenceIndexToPlay-1)}
	mockReadAndWriteState.ReturnForPlayedForColor = alreadyPlayed

	mockReadAndWriteState.TestErrorForEnactTurnByPlayingAndReplacing = nil

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	errorFromTakeTurnByPlaying :=
		executorForPlayer.TakeTurnByPlaying(indexInHandToPlay)

	if errorFromTakeTurnByPlaying != nil {
		unitTest.Fatalf(
			"TakeTurnByPlaying(%v) produced unexpected error %v",
			indexInHandToPlay,
			errorFromTakeTurnByPlaying)
	}

	actualArguments :=
		mockReadAndWriteState.ArgumentsFromEnactTurnByPlayingAndReplacing

	if len(actualArguments) != 1 {
		unitTest.Fatalf(
			"list of argument sets %v did not have exactly 1 element",
			actualArguments)
	}

	expectedPlayerState, errorFromGetPlayer := playerProvider.Get(playerName)
	if errorFromGetPlayer != nil {
		unitTest.Fatalf(
			"mock player provider's Get(%v) produced unexpected error %v",
			playerName,
			errorFromGetPlayer)
	}

	actualArgument := actualArguments[0]
	expectedActionMessage :=
		fmt.Sprintf(
			"successfully plays card %v %v",
			expectedPlayedCard.ColorSuit(),
			expectedPlayedCard.SequenceIndex())
	if (actualArgument.MessageString != expectedActionMessage) ||
		(actualArgument.PlayerState != expectedPlayerState) ||
		(actualArgument.IndexInt != indexInHandToPlay) ||
		(actualArgument.HintsInt != 0) ||
		(actualArgument.MistakesInt != 0) {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) resulted in wrong call to EnactTurnByPlayingAndReplacing(...): %v",
			gameName,
			playerName,
			actualArguments[0])
	}

	assertInferredCardPossibilitiesCorrect(
		"knowledge of drawn card when playing with no bonus hint and less than maximum hints",
		unitTest,
		actualArgument.DrawnInferred,
		testRuleset.ColorSuits(),
		testRuleset.DistinctPossibleIndices())
}

func TestTakeTurnByPlayingWithBonusHintWhenAlreadyAtMaximumHints(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, playerProvider :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 3
	mockReadAndWriteState.ReturnForNumberOfReadyHints =
		testRuleset.MaximumNumberOfHints()

	indexInHandToPlay := 3

	// This test does not need all the cards to be initialized correctly,
	// just that the hand slice is the correct length and that the played
	// card is valid. With the standard ruleset, there is a bonus hint if
	// the sequence index is five or greater.
	testSuit := testRuleset.ColorSuits()[1]
	sequenceIndexToPlay := 5
	correctHandSize :=
		testRuleset.NumberOfCardsInPlayerHand(len(testPlayersInOriginalOrder))
	correctSizeHands := make(map[string][]card.Readonly, 1)
	correctSizeHands[playerName] = make([]card.Readonly, correctHandSize)
	expectedPlayedCard := card.NewReadonly(testSuit, sequenceIndexToPlay)
	correctSizeHands[playerName][indexInHandToPlay] = expectedPlayedCard
	mockReadAndWriteState.ReturnForVisibleHand = correctSizeHands

	alreadyPlayed := make(map[string][]card.Readonly, 0)
	alreadyPlayed[testSuit] =
		[]card.Readonly{card.NewReadonly(testSuit, sequenceIndexToPlay-1)}
	mockReadAndWriteState.ReturnForPlayedForColor = alreadyPlayed

	mockReadAndWriteState.TestErrorForEnactTurnByPlayingAndReplacing = nil

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	errorFromTakeTurnByPlaying :=
		executorForPlayer.TakeTurnByPlaying(indexInHandToPlay)

	if errorFromTakeTurnByPlaying != nil {
		unitTest.Fatalf(
			"TakeTurnByPlaying(%v) produced unexpected error %v",
			indexInHandToPlay,
			errorFromTakeTurnByPlaying)
	}

	actualArguments :=
		mockReadAndWriteState.ArgumentsFromEnactTurnByPlayingAndReplacing

	if len(actualArguments) != 1 {
		unitTest.Fatalf(
			"list of argument sets %v did not have exactly 1 element",
			actualArguments)
	}

	expectedPlayerState, errorFromGetPlayer := playerProvider.Get(playerName)
	if errorFromGetPlayer != nil {
		unitTest.Fatalf(
			"mock player provider's Get(%v) produced unexpected error %v",
			playerName,
			errorFromGetPlayer)
	}

	actualArgument := actualArguments[0]
	expectedActionMessage :=
		fmt.Sprintf(
			"successfully plays card %v %v",
			expectedPlayedCard.ColorSuit(),
			expectedPlayedCard.SequenceIndex())
	if (actualArgument.MessageString != expectedActionMessage) ||
		(actualArgument.PlayerState != expectedPlayerState) ||
		(actualArgument.IndexInt != indexInHandToPlay) ||
		(actualArgument.HintsInt != 0) ||
		(actualArgument.MistakesInt != 0) {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) resulted in wrong call to EnactTurnByPlayingAndReplacing(...): %v",
			gameName,
			playerName,
			actualArguments[0])
	}

	assertInferredCardPossibilitiesCorrect(
		"knowledge of drawn card when playing with bonus but already at maximum hints",
		unitTest,
		actualArgument.DrawnInferred,
		testRuleset.ColorSuits(),
		testRuleset.DistinctPossibleIndices())
}

func TestTakeTurnByPlayingWithBonusHintWhenLessThanMaximumHints(unitTest *testing.T) {
	gameName := "Test game"
	testPlayersInOriginalOrder :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[3],
		}
	playerName := testPlayersInOriginalOrder[2]
	gameCollection, mockPersister, playerProvider :=
		prepareCollection(unitTest, testPlayersInOriginalOrder)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForPlayerNames = testPlayersInOriginalOrder
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForTurn = 3
	mockReadAndWriteState.ReturnForNumberOfReadyHints =
		testRuleset.MaximumNumberOfHints() - 2

	indexInHandToPlay := 0

	// This test does not need all the cards to be initialized correctly,
	// just that the hand slice is the correct length and that the played
	// card is valid. With the standard ruleset, there is a bonus hint if
	// the sequence index is five or greater.
	testSuit := testRuleset.ColorSuits()[1]
	sequenceIndexToPlay := 5
	correctHandSize :=
		testRuleset.NumberOfCardsInPlayerHand(len(testPlayersInOriginalOrder))
	correctSizeHands := make(map[string][]card.Readonly, 1)
	correctSizeHands[playerName] = make([]card.Readonly, correctHandSize)
	expectedPlayedCard := card.NewReadonly(testSuit, sequenceIndexToPlay)
	correctSizeHands[playerName][indexInHandToPlay] = expectedPlayedCard
	mockReadAndWriteState.ReturnForVisibleHand = correctSizeHands

	alreadyPlayed := make(map[string][]card.Readonly, 0)
	alreadyPlayed[testSuit] =
		[]card.Readonly{card.NewReadonly(testSuit, sequenceIndexToPlay-1)}
	mockReadAndWriteState.ReturnForPlayedForColor = alreadyPlayed

	mockReadAndWriteState.TestErrorForEnactTurnByPlayingAndReplacing = nil

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction != nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v",
			gameName,
			playerName,
			errorFromExecuteAction)
	}

	errorFromTakeTurnByPlaying :=
		executorForPlayer.TakeTurnByPlaying(indexInHandToPlay)

	if errorFromTakeTurnByPlaying != nil {
		unitTest.Fatalf(
			"TakeTurnByPlaying(%v) produced unexpected error %v",
			indexInHandToPlay,
			errorFromTakeTurnByPlaying)
	}

	actualArguments :=
		mockReadAndWriteState.ArgumentsFromEnactTurnByPlayingAndReplacing

	if len(actualArguments) != 1 {
		unitTest.Fatalf(
			"list of argument sets %v did not have exactly 1 element",
			actualArguments)
	}

	expectedPlayerState, errorFromGetPlayer := playerProvider.Get(playerName)
	if errorFromGetPlayer != nil {
		unitTest.Fatalf(
			"mock player provider's Get(%v) produced unexpected error %v",
			playerName,
			errorFromGetPlayer)
	}

	actualArgument := actualArguments[0]
	expectedActionMessage :=
		fmt.Sprintf(
			"successfully plays card %v %v",
			expectedPlayedCard.ColorSuit(),
			expectedPlayedCard.SequenceIndex())
	if (actualArgument.MessageString != expectedActionMessage) ||
		(actualArgument.PlayerState != expectedPlayerState) ||
		(actualArgument.IndexInt != indexInHandToPlay) ||
		(actualArgument.HintsInt != 1) ||
		(actualArgument.MistakesInt != 0) {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) resulted in wrong call to EnactTurnByPlayingAndReplacing(...): %v",
			gameName,
			playerName,
			actualArguments[0])
	}

	assertInferredCardPossibilitiesCorrect(
		"knowledge of drawn card when playing with bonus and less than maximum hints",
		unitTest,
		actualArgument.DrawnInferred,
		testRuleset.ColorSuits(),
		testRuleset.DistinctPossibleIndices())
}
