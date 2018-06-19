package game_test

import (
	"fmt"
	"testing"
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
	errorFromRecordChatMessage := executorForPlayer.RecordChatMessage(chatMessage)

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
	errorFromTakeTurnByDiscarding := executorForPlayer.TakeTurnByDiscarding(indexInHandToDiscard)

	if errorFromTakeTurnByDiscarding == nil {
		unitTest.Fatalf(
			"TakeTurnByDiscarding(%v) produced nil error when not player's turn",
			indexInHandToDiscard)
	}
}
