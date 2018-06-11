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

	unitTest.Fatalf("test not yet ready")
}
