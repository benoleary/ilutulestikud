package game_test

import (
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game/card"
)

func TestViewErrorWhenPersisterGivesError(unitTest *testing.T) {
	gameName := "Test game"
	playerName := playerNamesAvailableInTest[0]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, playerNamesAvailableInTest)

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForNontestError = fmt.Errorf("Expected error for test")

	viewForPlayer, errorFromViewState :=
		gameCollection.ViewState(
			gameName,
			playerName)

	if errorFromViewState == nil {
		unitTest.Fatalf(
			"ViewState(%v, %v) did not produce expected error, instead produced %v",
			gameName,
			playerName,
			viewForPlayer)
	}
}

func TestViewErrorWhenPlayerNotParticipant(unitTest *testing.T) {
	gameName := "Test game"
	playerName := "Test Player"
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, playerNamesAvailableInTest)

	mockPersister.TestErrorForReadAndWriteGame = nil

	mockReadAndWriteState :=
		NewMockGameState(unitTest, fmt.Errorf("No function should be called"))
	mockReadAndWriteState.TestErrorForName = nil
	mockReadAndWriteState.MockName = "test game"
	mockReadAndWriteState.TestErrorForPlayerNames = nil
	mockReadAndWriteState.ReturnForPlayerNames = []string{
		"A. Different Player",
		"A. Nother Different Player",
	}

	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	viewForPlayer, errorFromViewState :=
		gameCollection.ViewState(
			gameName,
			playerName)

	if errorFromViewState == nil {
		unitTest.Fatalf(
			"ViewState(%v, %v) did not produce expected error, instead produced %v",
			gameName,
			playerName,
			viewForPlayer)
	}
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

func TestRejectAddNewWhenInvalid(unitTest *testing.T) {
	validGameName := "Test game"
	errorWhenPlayerProviderShouldNotAllowGet :=
		fmt.Errorf("mock player provider should not allow Get(...)")
	fullDeck := testRuleset.CopyOfFullCardset()

	validPlayerNameList :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
		}

	testCases := []struct {
		testName                   string
		gameName                   string
		playerNames                []string
		initialDeck                []card.Readonly
		errorFromPlayerProviderGet error
	}{
		{
			testName:                   "Empty game name",
			gameName:                   "",
			playerNames:                validPlayerNameList,
			initialDeck:                fullDeck,
			errorFromPlayerProviderGet: errorWhenPlayerProviderShouldNotAllowGet,
		},
		{
			testName:                   "Nil players",
			gameName:                   validGameName,
			playerNames:                nil,
			initialDeck:                fullDeck,
			errorFromPlayerProviderGet: errorWhenPlayerProviderShouldNotAllowGet,
		},
		{
			testName:                   "No players",
			gameName:                   validGameName,
			playerNames:                []string{},
			initialDeck:                fullDeck,
			errorFromPlayerProviderGet: errorWhenPlayerProviderShouldNotAllowGet,
		},
		{
			testName: "Too few players",
			gameName: validGameName,
			playerNames: []string{
				playerNamesAvailableInTest[0],
			},
			initialDeck:                fullDeck,
			errorFromPlayerProviderGet: errorWhenPlayerProviderShouldNotAllowGet,
		},
		{
			testName:                   "Too many players",
			gameName:                   validGameName,
			playerNames:                playerNamesAvailableInTest,
			initialDeck:                fullDeck,
			errorFromPlayerProviderGet: errorWhenPlayerProviderShouldNotAllowGet,
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
			initialDeck:                fullDeck,
			errorFromPlayerProviderGet: nil,
		},
		{
			testName: "Unknown player",
			gameName: validGameName,
			playerNames: []string{
				playerNamesAvailableInTest[2],
				playerNamesAvailableInTest[1],
				"Unknown Player",
				playerNamesAvailableInTest[3],
			},
			initialDeck:                fullDeck,
			errorFromPlayerProviderGet: errorWhenPlayerProviderShouldNotAllowGet,
		},
		{
			testName: "Too few cards",
			gameName: validGameName,
			playerNames: []string{
				playerNamesAvailableInTest[2],
				playerNamesAvailableInTest[1],
				playerNamesAvailableInTest[3],
			},
			initialDeck: []card.Readonly{
				fullDeck[0],
				fullDeck[1],
			},
			errorFromPlayerProviderGet: nil,
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			gameCollection, _, _ :=
				prepareCollection(unitTest, playerNamesAvailableInTest)

			errorFromAddNew :=
				gameCollection.AddNewWithGivenDeck(
					testCase.gameName,
					testRuleset,
					testCase.playerNames,
					testCase.initialDeck)

			if errorFromAddNew == nil {
				unitTest.Fatalf(
					"AddNewWithGivenDeck(%v, %v, %v, %v) did not produce expected error",
					testCase.gameName,
					testRuleset,
					testCase.playerNames,
					testCase.initialDeck)
			}
		})
	}
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
