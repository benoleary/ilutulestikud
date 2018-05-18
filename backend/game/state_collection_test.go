package game_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game"
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

	// We do not fully test the view as that is done in another test file.
	if viewForPlayer.GameName() != gameName {
		unitTest.Fatalf(
			"ViewState(%v, %v) %v did not have expected game name %v",
			gameName,
			playerName,
			viewForPlayer,
			gameName)
	}
}

func TestErrorWhenViewErrorOnStateFromAll(unitTest *testing.T) {
	gameName := "Test game"
	playerName := "Test Player"
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, playerNamesAvailableInTest)

	mockGameWithPlayer :=
		NewMockGameState(unitTest, fmt.Errorf("mock game with player"))
	mockGameWithPlayer.TestErrorForName = nil
	mockGameWithPlayer.MockName = gameName
	mockGameWithPlayer.TestErrorForCreationTime = nil
	mockGameWithPlayer.ReturnForCreationTime = time.Now().Add(-2 * time.Second)
	mockGameWithPlayer.TestErrorForRuleset = nil
	mockGameWithPlayer.ReturnForRuleset = testRuleset
	mockGameWithPlayer.TestErrorForPlayerNames = nil
	mockGameWithPlayer.ReturnForPlayerNames = playerNamesAvailableInTest

	mockGameWithoutPlayer :=
		NewMockGameState(unitTest, fmt.Errorf("mock game without player"))
	mockGameWithoutPlayer.TestErrorForName = nil
	mockGameWithoutPlayer.MockName = "test game"
	mockGameWithoutPlayer.TestErrorForCreationTime = nil
	mockGameWithoutPlayer.ReturnForCreationTime = time.Now().Add(-1 * time.Second)
	mockGameWithoutPlayer.TestErrorForPlayerNames = nil
	mockGameWithoutPlayer.ReturnForPlayerNames = []string{
		playerNamesAvailableInTest[1],
		playerNamesAvailableInTest[2],
	}

	mockPersister.TestErrorForReadAllWithPlayer = nil
	mockPersister.ReturnForReadAllWithPlayer = []game.ReadonlyState{
		mockGameWithPlayer,
		mockGameWithoutPlayer,
	}

	viewsForPlayer, errorFromViewAll :=
		gameCollection.ViewAllWithPlayer(playerName)

	if errorFromViewAll == nil {
		unitTest.Fatalf(
			"ViewAllWithPlayer(%v) did not produce expected error, instead produced %v",
			playerName,
			viewsForPlayer)
	}
}

func TestViewsCorrectFromAllForPlayer(unitTest *testing.T) {
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, playerNamesAvailableInTest)
	playerName := playerNamesAvailableInTest[0]

	testStartTime := time.Now()

	mockFirstGame :=
		NewMockGameState(unitTest, fmt.Errorf("first mock game"))
	mockFirstGame.TestErrorForName = nil
	mockFirstGame.MockName = "first test game"
	mockFirstGame.TestErrorForCreationTime = nil
	mockFirstGame.ReturnForCreationTime = testStartTime.Add(-2 * time.Second)
	mockFirstGame.TestErrorForRuleset = nil
	mockFirstGame.ReturnForRuleset = testRuleset
	mockFirstGame.TestErrorForPlayerNames = nil
	mockFirstGame.ReturnForPlayerNames = playerNamesAvailableInTest

	mockSecondGame :=
		NewMockGameState(unitTest, fmt.Errorf("second mock game"))
	mockSecondGame.TestErrorForName = nil
	mockSecondGame.MockName = "second test game"
	mockSecondGame.TestErrorForCreationTime = nil
	mockSecondGame.ReturnForCreationTime = testStartTime.Add(-1 * time.Second)
	mockSecondGame.TestErrorForRuleset = nil
	mockSecondGame.ReturnForRuleset = testRuleset
	mockSecondGame.TestErrorForPlayerNames = nil
	mockSecondGame.ReturnForPlayerNames = []string{
		playerNamesAvailableInTest[1],
		playerName,
	}

	mockThirdGame :=
		NewMockGameState(unitTest, fmt.Errorf("third mock game"))
	mockThirdGame.TestErrorForName = nil
	mockThirdGame.MockName = "third test game"
	mockThirdGame.TestErrorForCreationTime = nil
	mockThirdGame.ReturnForCreationTime = testStartTime
	mockThirdGame.TestErrorForRuleset = nil
	mockThirdGame.ReturnForRuleset = testRuleset
	mockThirdGame.TestErrorForPlayerNames = nil
	mockThirdGame.ReturnForPlayerNames = []string{
		playerName,
		playerNamesAvailableInTest[2],
	}

	mockPersister.TestErrorForReadAllWithPlayer = nil

	// We return the games out of order to double-check that sorting works.
	mockPersister.ReturnForReadAllWithPlayer = []game.ReadonlyState{
		mockThirdGame,
		mockFirstGame,
		mockSecondGame,
	}

	expectedGames := []game.ReadonlyState{
		mockFirstGame,
		mockSecondGame,
		mockThirdGame,
	}

	viewsForPlayer, errorFromViewAll :=
		gameCollection.ViewAllWithPlayer(playerName)

	if errorFromViewAll != nil {
		unitTest.Fatalf(
			"ViewAllWithPlayer(%v) produced error %v",
			playerName,
			errorFromViewAll)
	}

	numberOfExpectedGames := len(expectedGames)

	if len(viewsForPlayer) != numberOfExpectedGames {
		unitTest.Fatalf(
			"ViewAllWithPlayer(%v) %v had wrong number of games: expected %v",
			playerName,
			viewsForPlayer,
			expectedGames)
	}

	for gameIndex := 0; gameIndex < numberOfExpectedGames; gameIndex++ {
		viewForPlayer := viewsForPlayer[gameIndex]
		expectedGame := expectedGames[gameIndex]
		// We do not fully test the view as that is done in another test file.
		if viewForPlayer.GameName() != expectedGame.Name() {
			unitTest.Fatalf(
				"ViewAllWithPlayer(%v) %v had wrong name in position %v: actual %v, expected %v",
				playerName,
				viewsForPlayer,
				gameIndex,
				viewForPlayer.GameName(),
				expectedGame.Name())
		}
	}
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
