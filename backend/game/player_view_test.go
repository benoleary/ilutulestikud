package game_test

import (
	"fmt"
	"testing"

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

	mockReadAndWriteState :=
		NewMockGameState(unitTest, fmt.Errorf("No write function should be called"))

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

	testScore := 7
	mockReadAndWriteState.ReturnForScore = testScore

	testReadyHints := 5
	testSpentHints := (testRuleset.MaximumNumberOfHints() - testReadyHints)
	mockReadAndWriteState.ReturnForNumberOfReadyHints = testReadyHints

	testMistakesMade := 5
	testMistakesAllowed := (testRuleset.MaximumNumberOfMistakesAllowed() - testMistakesMade)
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
		(viewForPlayer.Score() != testScore) ||
		(viewForPlayer.NumberOfReadyHints() != testReadyHints) ||
		(viewForPlayer.NumberOfSpentHints() != testSpentHints) ||
		(viewForPlayer.NumberOfMistakesStillAllowed() != testMistakesAllowed) ||
		(viewForPlayer.NumberOfMistakesMade() != testMistakesMade) ||
		(viewForPlayer.DeckSize() != testDeckSize) {
		unitTest.Fatalf(
			"player view %+v not as expected"+
				" (name %v,"+
				" ruleset description %v,"+
				" turn %v,"+
				" score %v,"+
				" ready hints %v,"+
				" spent hints %v,"+
				" mistakes allowed %v,"+
				" mistakes made %v,"+
				" deck size %v)",
			viewForPlayer,
			gameName,
			testRuleset.FrontendDescription(),
			testTurn,
			testScore,
			testReadyHints,
			testSpentHints,
			testMistakesAllowed,
			testMistakesMade,
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

	unitTest.Fatalf("test not yet ready")
}
