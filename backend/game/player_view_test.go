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
		NewMockGameState(unitTest, fmt.Errorf("No function should be called"))
	mockReadAndWriteState.TestErrorForName = nil
	mockReadAndWriteState.MockName = gameName
	mockReadAndWriteState.TestErrorForRuleset = nil
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.TestErrorForPlayerNames = nil
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
		(viewForPlayer.RulesetDescription() != gameName) ||
		(viewForPlayer.Turn() != 0) ||
		(viewForPlayer.Score() != 0) ||
		(viewForPlayer.NumberOfReadyHints() != 0) ||
		(viewForPlayer.NumberOfSpentHints() != 0) ||
		(viewForPlayer.NumberOfMistakesStillAllowed() != 0) ||
		(viewForPlayer.NumberOfMistakesMade() != 0) ||
		(viewForPlayer.DeckSize() != 0) {
		unitTest.Fatalf("test not yet ready")
	}
}
