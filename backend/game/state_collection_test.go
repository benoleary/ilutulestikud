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

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForName = "test game"
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

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForName = gameName
	mockReadAndWriteState.ReturnForRuleset = testRuleset
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
	gameName := "Mock with specified player"
	playerName := "Test Player"
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, playerNamesAvailableInTest)

	mockGameWithPlayer := NewMockGameState(unitTest)
	mockGameWithPlayer.ReturnForName = gameName
	mockGameWithPlayer.ReturnForCreationTime = time.Now().Add(-2 * time.Second)
	mockGameWithPlayer.ReturnForRuleset = testRuleset
	mockGameWithPlayer.ReturnForPlayerNames = playerNamesAvailableInTest

	mockGameWithoutPlayer := NewMockGameState(unitTest)
	mockGameWithoutPlayer.ReturnForName = "Mock without specified player"
	mockGameWithoutPlayer.ReturnForCreationTime = time.Now().Add(-1 * time.Second)
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

	mockFirstGame := NewMockGameState(unitTest)
	mockFirstGame.ReturnForName = "first test game"
	mockFirstGame.ReturnForCreationTime = testStartTime.Add(-2 * time.Second)
	mockFirstGame.ReturnForRuleset = testRuleset
	mockFirstGame.ReturnForPlayerNames = playerNamesAvailableInTest

	mockSecondGame := NewMockGameState(unitTest)
	mockSecondGame.ReturnForName = "second test game"
	mockSecondGame.ReturnForCreationTime = testStartTime.Add(-1 * time.Second)
	mockSecondGame.ReturnForRuleset = testRuleset
	mockSecondGame.ReturnForPlayerNames = []string{
		playerNamesAvailableInTest[1],
		playerName,
	}

	mockThirdGame := NewMockGameState(unitTest)
	mockThirdGame.ReturnForName = "third test game"
	mockThirdGame.ReturnForCreationTime = testStartTime
	mockThirdGame.ReturnForRuleset = testRuleset
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
			initialDeck:                testRuleset.CopyOfFullCardset(),
			errorFromPlayerProviderGet: errorWhenPlayerProviderShouldNotAllowGet,
		},
		{
			testName:                   "Nil players",
			gameName:                   validGameName,
			playerNames:                nil,
			initialDeck:                testRuleset.CopyOfFullCardset(),
			errorFromPlayerProviderGet: errorWhenPlayerProviderShouldNotAllowGet,
		},
		{
			testName:                   "No players",
			gameName:                   validGameName,
			playerNames:                []string{},
			initialDeck:                testRuleset.CopyOfFullCardset(),
			errorFromPlayerProviderGet: errorWhenPlayerProviderShouldNotAllowGet,
		},
		{
			testName: "Too few players",
			gameName: validGameName,
			playerNames: []string{
				playerNamesAvailableInTest[0],
			},
			initialDeck:                testRuleset.CopyOfFullCardset(),
			errorFromPlayerProviderGet: errorWhenPlayerProviderShouldNotAllowGet,
		},
		{
			testName:                   "Too many players",
			gameName:                   validGameName,
			playerNames:                playerNamesAvailableInTest,
			initialDeck:                testRuleset.CopyOfFullCardset(),
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
			initialDeck:                testRuleset.CopyOfFullCardset(),
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
			initialDeck:                testRuleset.CopyOfFullCardset(),
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
			initialDeck:                testRuleset.CopyOfFullCardset()[0:2],
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

func TestAddNewWithGivenShuffle(unitTest *testing.T) {
	gameName := "Test game"

	// We need at least 16 cards for the test
	// ((3 players * 5 cards per hand) + 1 for remaining deck).
	// We choose a sequence that should prove that the cards
	// have been propagated correctly.
	testDeck := []card.Readonly{
		card.NewReadonly("some color for player 1", 2),
		card.NewReadonly("another color for player 1", 3),
		card.NewReadonly("player 1 color", 1),
		card.NewReadonly("some color for player 1", 1),
		card.NewReadonly("player 1 color", 1),
		card.NewReadonly("a color for player 2", 2),
		card.NewReadonly("another color for player 2", 2),
		card.NewReadonly("player 2 color", 2),
		card.NewReadonly("some color for player 2", 2),
		card.NewReadonly("player 2 color", 2),
		card.NewReadonly("player 3 hand color", 1),
		card.NewReadonly("player 3 hand color", 2),
		card.NewReadonly("player 3 hand color", 3),
		card.NewReadonly("player 3 hand color", 2),
		card.NewReadonly("player 3 hand color", 1),
		card.NewReadonly("color which should end up in remaining deck", 1),
		card.NewReadonly("color which should end up in remaining deck", 2),
		card.NewReadonly("color which should end up in remaining deck", 1),
		card.NewReadonly("another color which should end up in remaining deck", 1),
		card.NewReadonly("color which should end up in remaining deck", 1),
	}

	// We need a deep copy of the deck as adding the game
	// modifies the deck given to it.
	copyOfInputDeck := make([]card.Readonly, len(testDeck))
	copy(copyOfInputDeck, testDeck)

	gameParticipants :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
		}

	gameCollection, mockGamePersister, _ :=
		prepareCollection(unitTest, playerNamesAvailableInTest)

	mockGamePersister.TestErrorForAddGame = nil

	errorFromAddNew :=
		gameCollection.AddNewWithGivenDeck(
			gameName,
			testRuleset,
			gameParticipants,
			testDeck)

	baseIdentifier :=
		fmt.Sprintf(
			"AddNewWithGivenDeck(%v, [%v], %v, %v)",
			gameName,
			testRuleset,
			gameParticipants,
			copyOfInputDeck)

	if errorFromAddNew != nil {
		unitTest.Fatalf(
			baseIdentifier+" produced error %v",
			errorFromAddNew)
	}

	actualPersistanceCalls := mockGamePersister.ArgumentsForAddGame

	if len(actualPersistanceCalls) != 1 {
		unitTest.Fatalf(
			baseIdentifier+" resulted in wrong number of calls to persister: %v",
			actualPersistanceCalls)
	}

	actualPersistanceCall := actualPersistanceCalls[0]

	if (actualPersistanceCall.gameName != gameName) ||
		(actualPersistanceCall.gameRuleset != testRuleset) {
		unitTest.Fatalf(
			baseIdentifier+" resulted in wrong call to persister: %v",
			actualPersistanceCall)
	}

	numberOfExpectedPlayers := len(gameParticipants)
	lengthOfExpectedHand :=
		testRuleset.NumberOfCardsInPlayerHand(numberOfExpectedPlayers)
	expectedColors := testRuleset.ColorSuits()
	expectedIndices := testRuleset.DistinctPossibleIndices()
	playersAndHandsFromCall := actualPersistanceCall.playersInTurnOrderWithInitialHands

	if len(playersAndHandsFromCall) != numberOfExpectedPlayers {
		unitTest.Fatalf(
			baseIdentifier+" resulted in wrong call to persister: %v",
			actualPersistanceCall)
	}

	for playerIndex := 0; playerIndex < numberOfExpectedPlayers; playerIndex++ {
		if playersAndHandsFromCall[playerIndex].PlayerName != gameParticipants[playerIndex] {
			unitTest.Fatalf(
				baseIdentifier+" resulted in wrong call to persister: %v",
				actualPersistanceCall)
		}

		// We take blocks of cards from the start of the deck for each player, as
		// that is how the collection should deal them to the players: each player
		// gets all the cards for their hand before the next player gets any.
		indexOfFirstCardInHand := playerIndex * lengthOfExpectedHand
		indexOfCardAfterLastCardInHand := indexOfFirstCardInHand + lengthOfExpectedHand
		expectedVisibleHand :=
			copyOfInputDeck[indexOfFirstCardInHand:indexOfCardAfterLastCardInHand]

		// We make fresh inferred cards around the visible cards, as we expect the
		// collection to have done.
		expectedPlayerHand := make([]card.InHand, lengthOfExpectedHand)
		for indexInHand := 0; indexInHand < lengthOfExpectedHand; indexInHand++ {
			expectedPlayerHand[indexInHand] =
				card.InHand{
					Readonly: expectedVisibleHand[indexInHand],
					Inferred: card.NewInferred(
						expectedColors,
						expectedIndices),
				}
		}

		testIdentifier :=
			fmt.Sprintf(
				baseIdentifier+" call to persister/checking hand of player %v",
				playerIndex)
		assertInHandCardSlicesMatch(
			testIdentifier,
			unitTest,
			playersAndHandsFromCall[playerIndex].InitialHand,
			expectedPlayerHand)
	}

	assertReadonlyCardSlicesMatch(
		baseIdentifier+" call to persister/checking initial deck",
		unitTest,
		copyOfInputDeck[numberOfExpectedPlayers*lengthOfExpectedHand:],
		actualPersistanceCall.initialDeck)
}

func TestAddNewWithDefaultShuffle(unitTest *testing.T) {
	gameName := "Test game"

	expectedFullCardset := testRuleset.CopyOfFullCardset()

	gameParticipants :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
		}

	gameCollection, mockGamePersister, _ :=
		prepareCollection(unitTest, playerNamesAvailableInTest)

	mockGamePersister.TestErrorForAddGame = nil
	mockGamePersister.TestErrorForRandomSeed = nil
	mockGamePersister.ReturnForRandomSeed = 1

	errorFromAddNew :=
		gameCollection.AddNew(
			gameName,
			testRuleset,
			gameParticipants)

	baseIdentifier :=
		fmt.Sprintf(
			"AddNew(%v, [%v], %v)",
			gameName,
			testRuleset,
			gameParticipants)

	if errorFromAddNew != nil {
		unitTest.Fatalf(
			baseIdentifier+" produced error %v",
			errorFromAddNew)
	}

	actualPersistanceCalls := mockGamePersister.ArgumentsForAddGame

	if len(actualPersistanceCalls) != 1 {
		unitTest.Fatalf(
			baseIdentifier+" resulted in wrong number of calls to persister: %v",
			actualPersistanceCalls)
	}

	actualPersistanceCall := actualPersistanceCalls[0]

	if (actualPersistanceCall.gameName != gameName) ||
		(actualPersistanceCall.gameRuleset != testRuleset) {
		unitTest.Fatalf(
			baseIdentifier+" resulted in wrong call to persister: %v",
			actualPersistanceCall)
	}

	numberOfExpectedPlayers := len(gameParticipants)
	lengthOfExpectedHand :=
		testRuleset.NumberOfCardsInPlayerHand(numberOfExpectedPlayers)
	expectedColors := testRuleset.ColorSuits()
	expectedIndices := testRuleset.DistinctPossibleIndices()
	playersAndHandsFromCall := actualPersistanceCall.playersInTurnOrderWithInitialHands

	if len(playersAndHandsFromCall) != numberOfExpectedPlayers {
		unitTest.Fatalf(
			baseIdentifier+" resulted in wrong call to persister: %v",
			actualPersistanceCall)
	}

	// In this case, we cannot be sure of which cards should appear where.
	// We check that the inferred information is correct (conveniently independent
	// of what the cards are at the moment when the initial deal has just been made)
	// and collect all the cards together to check against the initial deck.
	shuffledCards := make(map[card.Readonly]int, 0)
	numberOfCardsInTotal := 0
	for playerIndex := 0; playerIndex < numberOfExpectedPlayers; playerIndex++ {
		nameAndHand := playersAndHandsFromCall[playerIndex]
		if (nameAndHand.PlayerName != gameParticipants[playerIndex]) ||
			(len(nameAndHand.InitialHand) != lengthOfExpectedHand) {
			unitTest.Fatalf(
				baseIdentifier+" resulted in wrong call to persister: %v",
				actualPersistanceCall)
		}

		for _, inHandCard := range playersAndHandsFromCall[playerIndex].InitialHand {
			assertInferredCardPossibilitiesCorrect(
				baseIdentifier+"/inferred card in hand",
				unitTest,
				inHandCard.Inferred,
				expectedColors,
				expectedIndices)

			numberOfCopiesBeforeThis := shuffledCards[inHandCard.Readonly]
			shuffledCards[inHandCard.Readonly] = numberOfCopiesBeforeThis + 1
			numberOfCardsInTotal++
		}
	}

	for _, remainingCard := range actualPersistanceCall.initialDeck {
		numberOfCopiesBeforeThis := shuffledCards[remainingCard]
		shuffledCards[remainingCard] = numberOfCopiesBeforeThis + 1
		numberOfCardsInTotal++
	}

	if numberOfCardsInTotal != len(expectedFullCardset) {
		unitTest.Fatalf(
			baseIdentifier+" had wrong number of cards in total %v, expected %v",
			numberOfCardsInTotal,
			len(expectedFullCardset))
	}

	for _, expectedCard := range expectedFullCardset {
		numberOfCopiesBeforeThis := shuffledCards[expectedCard]
		shuffledCards[expectedCard] = numberOfCopiesBeforeThis - 1
	}

	for cardKey, numberOfCopies := range shuffledCards {
		if numberOfCopies != 0 {
			unitTest.Fatalf(
				baseIdentifier+" had %v copies of %v after crossing off all expected.",
				numberOfCopies,
				cardKey)
		}
	}
}

func TestExecutorErrorWhenPersisterGivesError(unitTest *testing.T) {
	gameName := "Test game"
	playerName := playerNamesAvailableInTest[0]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, playerNamesAvailableInTest)

	mockPersister.TestErrorForReadAndWriteGame = nil
	mockPersister.ReturnForNontestError = fmt.Errorf("Expected error for test")

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction == nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) did not produce expected error, instead produced %v",
			gameName,
			playerName,
			executorForPlayer)
	}
}

func TestExecutorErrorWhenPlayerNotRegistered(unitTest *testing.T) {
	gameName := "Test game"
	playerName := "Not Registered"
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, playerNamesAvailableInTest)

	mockPersister.TestErrorForReadAndWriteGame = nil

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForName = gameName
	mockReadAndWriteState.ReturnForPlayerNames = []string{
		playerNamesAvailableInTest[0],
		playerNamesAvailableInTest[1],
		playerNamesAvailableInTest[2],
	}

	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction == nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) did not produce expected error, instead produced %v",
			gameName,
			playerName,
			executorForPlayer)
	}
}

func TestExecutorErrorWhenPlayerNotParticipant(unitTest *testing.T) {
	gameName := "Test game"
	playerName := playerNamesAvailableInTest[0]
	gameCollection, mockPersister, _ :=
		prepareCollection(unitTest, playerNamesAvailableInTest)

	mockPersister.TestErrorForReadAndWriteGame = nil

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForName = gameName
	mockReadAndWriteState.ReturnForPlayerNames = []string{
		playerNamesAvailableInTest[1],
		playerNamesAvailableInTest[2],
	}

	mockPersister.ReturnForReadAndWriteGame = mockReadAndWriteState

	executorForPlayer, errorFromExecuteAction :=
		gameCollection.ExecuteAction(
			gameName,
			playerName)

	if errorFromExecuteAction == nil {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) did not produce expected error, instead produced %v",
			gameName,
			playerName,
			executorForPlayer)
	}
}

func TestExecutorCorrectWhenPersisterGivesValidGame(unitTest *testing.T) {
	gameName := "Test game"
	playerName := playerNamesAvailableInTest[0]
	gameCollection, gamePersister, _ :=
		prepareCollection(unitTest, playerNamesAvailableInTest)

	mockReadAndWriteState := NewMockGameState(unitTest)
	mockReadAndWriteState.ReturnForName = gameName
	mockReadAndWriteState.ReturnForRuleset = testRuleset
	mockReadAndWriteState.ReturnForPlayerNames = playerNamesAvailableInTest

	gamePersister.TestErrorForReadAndWriteGame = nil
	gamePersister.ReturnForReadAndWriteGame = mockReadAndWriteState

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

	// We do not fully test the executor as that is done in another test file.
	// We just test recording a chat message.
	testMessage := "Test message!"
	expectedError := fmt.Errorf("expected error")
	mockReadAndWriteState.ReturnForNontestError = expectedError
	mockReadAndWriteState.TestErrorForRecordChatMessage = nil

	errorFromRecordChatMessage :=
		executorForPlayer.RecordChatMessage(testMessage)

	if errorFromRecordChatMessage != expectedError {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) produced error %v which was not expected error %v",
			gameName,
			playerName,
			errorFromRecordChatMessage,
			expectedError)
	}

	actualArguments := mockReadAndWriteState.ArgumentsFromRecordChatMessage
	if len(actualArguments) != 1 {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) resulted in wrong number of calls to RecordChatMessage(...): %v",
			gameName,
			playerName,
			actualArguments)
	}

	expectedArguments :=
		argumentsForRecordChatMessage{
			NameString:    playerName,
			ColorString:   mockChatColor,
			MessageString: testMessage,
		}

	if actualArguments[0] != expectedArguments {
		unitTest.Fatalf(
			"ExecuteAction(%v, %v) resulted in wrong call to RecordChatMessage(...): %v",
			gameName,
			playerName,
			actualArguments[0])
	}
}

func TestReturnErrorFromPersisterDelete(unitTest *testing.T) {
	gameCollection, gamePersister, _ :=
		prepareCollection(unitTest, playerNamesAvailableInTest)
	gamePersister.TestErrorForDelete = nil

	testCases := []struct {
		testName      string
		expectedError error
	}{
		{
			testName:      "String error",
			expectedError: fmt.Errorf("Expected error from Delete(...)"),
		},
		{
			testName:      "Nil error",
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		gamePersister.ReturnForNontestError = testCase.expectedError

		unitTest.Run(testCase.testName, func(unitTest *testing.T) {

			gameName := "Mock Player"

			actualError := gameCollection.Delete(gameName)

			if actualError != testCase.expectedError {
				unitTest.Errorf(
					"Delete(game name %v) returned error %v - expected %v",
					gameName,
					actualError,
					testCase.expectedError)
			}
		})
	}
}
