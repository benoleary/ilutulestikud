package game_test

import (
	"fmt"
	"testing"
)

func TestWrapperFunctions(unitTest *testing.T) {
	gameName := "Test game"
	playerName := playerNamesAvailableInTest[0]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, playerNamesAvailableInTest)

	mockReadAndWriteState :=
		NewMockGameState(unitTest, fmt.Errorf("No write function should be called"))
	mockReadAndWriteState.ReturnForName = gameName
	mockReadAndWriteState.ReturnForRuleset = testRuleset

	testTurn := 3
	mockReadAndWriteState.ReturnForTurn = testTurn
	mockReadAndWriteState.ReturnForPlayerNames = playerNamesAvailableInTest

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
		(viewForPlayer.Score() != 0) ||
		(viewForPlayer.NumberOfReadyHints() != 0) ||
		(viewForPlayer.NumberOfSpentHints() != 0) ||
		(viewForPlayer.NumberOfMistakesStillAllowed() != 0) ||
		(viewForPlayer.NumberOfMistakesMade() != 0) ||
		(viewForPlayer.DeckSize() != 0) {
		unitTest.Fatalf("test not yet ready")
	}
}
