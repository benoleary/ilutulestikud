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

	// This test does not need the cards to be initialized correctly,
	// just that the hand slice is the correct length.
	correctHandSize :=
		testRuleset.NumberOfCardsInPlayerHand(len(testPlayersInOriginalOrder))
	correctSizeHands := make(map[string][]card.Readonly, 1)
	correctSizeHands[playerName] = make([]card.Readonly, correctHandSize)
	expectedDiscardedCard := card.NewReadonly("some_color", 123)
	correctSizeHands[playerName][0] = expectedDiscardedCard
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

	indexInHandToDiscard := 0
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
		(actualArgument.IndexInt != 0) ||
		(actualArgument.HintsInt != 0) ||
		(actualArgument.MistakesInt != 0) {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) resulted in wrong call to EnactTurnByDiscardingAndReplacing(...): %v",
			gameName,
			playerName,
			actualArguments[0])
	}

	unitTest.Fatalf("test does not yet check that the replacement inferred card is correct.")
}
