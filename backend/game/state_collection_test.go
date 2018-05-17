package game_test

import (
	"testing"
)

func TestViewErrorWhenPersisterGivesError(unitTest *testing.T) {
	unitTest.Fatalf("Not implemented yet")
}

func TestViewErrorWhenPlayerNotParticipant(unitTest *testing.T) {
	unitTest.Fatalf("Not implemented yet")
}

func TestViewCorrectWhenPersisterGivesValidGame(unitTest *testing.T) {
	unitTest.Fatalf("Not implemented yet")
}

func TestErrorWhenViewErrorOnStateFromAll(unitTest *testing.T) {
	unitTest.Fatalf("Not implemented yet")
}

func TestViewsCorrectFromAllForPlayer(unitTest *testing.T) {
	unitTest.Fatalf("Not implemented yet")
}

func TestRejectAddNewForEmptyGameName(unitTest *testing.T) {
	validGameName := "Test game"

	validPlayerNameList :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
		}

	testCases := []struct {
		testName    string
		gameName    string
		playerNames []string
	}{
		{
			testName:    "Empty game name",
			gameName:    "",
			playerNames: validPlayerNameList,
		},
		{
			testName:    "Nil players",
			gameName:    validGameName,
			playerNames: nil,
		},
		{
			testName:    "No players",
			gameName:    validGameName,
			playerNames: []string{},
		},
		{
			testName: "Too few players",
			gameName: validGameName,
			playerNames: []string{
				playerNamesAvailableInTest[0],
			},
		},
		{
			testName:    "Too many players",
			gameName:    validGameName,
			playerNames: playerNamesAvailableInTest,
		},
		{
			testName: "Repeated player",
			gameName: validGameName,
			playerNames: []string{
				playerNamesAvailableInTest[2],
				playerNamesAvailableInTest[1],
				playerNamesAvailableInTest[1],
				playerNamesAvailableInTest[3],
			},
		},
		{
			testName: "Unregistered player",
			gameName: validGameName,
			playerNames: []string{
				playerNamesAvailableInTest[2],
				playerNamesAvailableInTest[1],
				"Not A. Registered Player",
				playerNamesAvailableInTest[3],
			},
		},
	}

	unitTest.Fatalf("Not implemented yet")
}

func TestRejectAddNewWhenErrorCreatingHands(unitTest *testing.T) {
	unitTest.Fatalf("Not implemented yet")
}

func TestAddNewWithDefaultShuffle(unitTest *testing.T) {
	unitTest.Fatalf("Not implemented yet")
}

func TestAddNewWithGivenShuffle(unitTest *testing.T) {
	unitTest.Fatalf("Not implemented yet")
}

func TestExecutorErrorWhenPersisterGivesError(unitTest *testing.T) {
	unitTest.Fatalf("Not implemented yet")
}

func TestExecutorErrorWhenPlayerNotParticipant(unitTest *testing.T) {
	unitTest.Fatalf("Not implemented yet")
}

func TestExecutorCorrectWhenPersisterGivesValidGame(unitTest *testing.T) {
	unitTest.Fatalf("Not implemented yet")
}
