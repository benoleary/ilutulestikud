package persister_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game"
)

func TestRandomSeedCausesNoPanic(unitTest *testing.T) {
	statePersisters := preparePersisters()

	for _, statePersister := range statePersisters {
		testIdentifier := "Positive seed/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			// This is a really trivial test, but it is just nice to have 100% test coverage.
			statePersister.GamePersister.RandomSeed()
		})
	}
}

func TestReturnErrorWhenGameDoesNotExist(unitTest *testing.T) {
	statePersisters := preparePersisters()

	for _, statePersister := range statePersisters {
		testIdentifier :=
			"ReadAndWriteGame(unknown game)/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			invalidName := "Not a valid game"
			gameState, errorFromGet :=
				statePersister.GamePersister.ReadAndWriteGame(invalidName)

			if errorFromGet == nil {
				unitTest.Fatalf(
					"ReadAndWriteGame(unknown game name %v) did not return an error, did return game state %v",
					invalidName,
					gameState)
			}
		})
	}
}

func TestReturnEmptyListWhenPlayerHasNoGames(unitTest *testing.T) {
	statePersisters := preparePersisters()

	for _, statePersister := range statePersisters {
		testIdentifier :=
			"ReadAllWithPlayer(unknown player)/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			invalidName := "Not A. Participant"
			gameStates :=
				statePersister.GamePersister.ReadAllWithPlayer(invalidName)

			if gameStates == nil {
				unitTest.Fatalf(
					"ReadAllWithPlayer(unknown player name %v) returned nil list",
					invalidName)
			}

			if len(gameStates) != 0 {
				unitTest.Fatalf(
					"ReadAllWithPlayer(unknown player name %v) returned non-empty list %v",
					invalidName,
					gameStates)
			}
		})
	}
}

func TestRejectAddGameWithExistingName(unitTest *testing.T) {
	statePersisters := preparePersisters()

	twoPlayersWithNilHands :=
		[]game.PlayerNameWithHand{
			game.PlayerNameWithHand{
				PlayerName:  defaultTestPlayers[0],
				InitialHand: nil,
			},
			game.PlayerNameWithHand{
				PlayerName:  defaultTestPlayers[2],
				InitialHand: nil,
			},
		}

	threePlayersWithNilHands :=
		[]game.PlayerNameWithHand{
			game.PlayerNameWithHand{
				PlayerName:  defaultTestPlayers[0],
				InitialHand: nil,
			},
			game.PlayerNameWithHand{
				PlayerName:  defaultTestPlayers[1],
				InitialHand: nil,
			},
			game.PlayerNameWithHand{
				PlayerName:  defaultTestPlayers[2],
				InitialHand: nil,
			},
		}

	expectedGamesMappedToPlayers := make(map[string]map[string]bool, 0)

	for _, statePersister := range statePersisters {
		for _, gameName := range []string{"A valid game", "Another valid game"} {
			testIdentifier :=
				"Reject Add(game with existing name)/" + statePersister.PersisterDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				expectedGamesMappedToPlayers[gameName] =
					playerNameSet(twoPlayersWithNilHands)

				errorFromInitialAdd :=
					statePersister.GamePersister.AddGame(
						gameName,
						defaultTestRuleset,
						twoPlayersWithNilHands,
						nil)

				if errorFromInitialAdd != nil {
					unitTest.Fatalf(
						"AddGame(%v, %v, %v, nil) produced an error: %v",
						gameName,
						defaultTestRuleset,
						twoPlayersWithNilHands,
						errorFromInitialAdd)
				}

				// We check that the persister still produces valid states.
				assertReturnedGamesAreConsistent(
					testIdentifier,
					unitTest,
					expectedGamesMappedToPlayers,
					statePersister.GamePersister)

				initialState :=
					getStateAndAssertNoError(
						testIdentifier+"/ReadAndWriteGame(initial game)",
						unitTest,
						gameName,
						statePersister.GamePersister)

				errorFromSecondAdd :=
					statePersister.GamePersister.AddGame(
						gameName,
						defaultTestRuleset,
						threePlayersWithNilHands,
						nil)

				assertGameNameAndParticipantsAreCorrect(
					testIdentifier+"/ReadAndWriteGame(initial game)",
					unitTest,
					gameName,
					expectedGamesMappedToPlayers[gameName],
					initialState)

				// We check that the persister still produces valid states.
				assertReturnedGamesAreConsistent(
					testIdentifier,
					unitTest,
					expectedGamesMappedToPlayers,
					statePersister.GamePersister)

				// If there was no error, then something went wrong.
				if errorFromSecondAdd == nil {
					unitTest.Fatalf(
						"AddGame(%v, %v, %v, nil) did not produce an error",
						gameName,
						defaultTestRuleset,
						defaultTestPlayers)
				}

				// We check that the player is unchanged.
				existingStateAfterAddWithNewColor :=
					getStateAndAssertNoError(
						testIdentifier+"/ReadAndWriteGame(initial game)",
						unitTest,
						gameName,
						statePersister.GamePersister)

				assertGameNameAndParticipantsAreCorrect(
					testIdentifier+"/ReadAndWriteGame(initial game)",
					unitTest,
					gameName,
					expectedGamesMappedToPlayers[gameName],
					existingStateAfterAddWithNewColor)
			})
		}
	}
}

func assertReturnedGamesAreConsistent(
	testIdentifier string,
	unitTest *testing.T,
	expectedGamesMappedToPlayers map[string]map[string]bool,
	gamePersister game.StatePersister) {
	gamesForPlayer := make(map[string]map[string]bool, 0)

	for expectedGame, expectedParticipants := range expectedGamesMappedToPlayers {
		// We update the games expected to be found for each player.
		for expectedPlayerName, isParticipant := range expectedParticipants {
			gameNameSet, _ := gamesForPlayer[expectedPlayerName]

			if gameNameSet == nil {
				gameNameSet = make(map[string]bool, 0)
				gamesForPlayer[expectedPlayerName] = gameNameSet
			}

			gameNameSet[expectedGame] = isParticipant
		}

		actualGame, errorFromGet := gamePersister.ReadAndWriteGame(expectedGame)

		if errorFromGet != nil {
			unitTest.Fatalf(
				testIdentifier+"/ReadAndWriteGame(%v) produced an error %v",
				expectedGame,
				errorFromGet)
		}

		if actualGame == nil {
			unitTest.Fatalf(
				testIdentifier + "/ReadAndWriteGame(%v) produced a nil game")
		}

		actualReadonlyState := actualGame.Read()

		if actualReadonlyState == nil {
			unitTest.Fatalf(
				testIdentifier+"/ReadAndWriteGame(%v) produced a game %v with a nil read-only state",
				actualGame)
		}

		assertGameNameAndParticipantsAreCorrect(
			testIdentifier+"/ReadAndWriteGame("+expectedGame+")",
			unitTest,
			expectedGame,
			expectedParticipants,
			actualReadonlyState)
	}

	for expectedPlayerName, expectedGameNames := range gamesForPlayer {
		statesFromAllWithPlayer := gamePersister.ReadAllWithPlayer(expectedPlayerName)

		if len(statesFromAllWithPlayer) != len(expectedGameNames) {
			unitTest.Fatalf(
				testIdentifier+
					"/ReadAllWithPlayer(%v) produced %v which did not have the expected game names %v",
				statesFromAllWithPlayer,
				expectedGameNames)
		}

		for _, gameState := range statesFromAllWithPlayer {
			if !expectedGameNames[gameState.Name()] {
				unitTest.Fatalf(
					testIdentifier+
						"/ReadAllWithPlayer(%v) produced %v which did not have the expected game names %v",
					statesFromAllWithPlayer,
					expectedGameNames)
			}
		}
	}
}

func assertGameNameAndParticipantsAreCorrect(
	testIdentifier string,
	unitTest *testing.T,
	expectedName string,
	expectedPlayerNames map[string]bool,
	readonlyGame game.ReadonlyState) {
	if readonlyGame.Name() != expectedName {
		unitTest.Fatalf(
			testIdentifier+"/expected name %v, actual state %v",
			expectedName,
			readonlyGame)
	}

	assertPlayersMatchNames(
		testIdentifier,
		unitTest,
		expectedPlayerNames,
		readonlyGame.PlayerNames())
}

func getStateAndAssertNoError(
	testIdentifier string,
	unitTest *testing.T,
	gameName string,
	gamePersister game.StatePersister) game.ReadonlyState {
	actualGame, errorFromGet := gamePersister.ReadAndWriteGame(gameName)

	if errorFromGet != nil {
		unitTest.Fatalf(
			testIdentifier+
				"/ReadAndWriteGame(%v) produced an error %v",
			gameName,
			errorFromGet)
	}

	if actualGame == nil {
		unitTest.Fatalf(
			testIdentifier+"/nil state from ReadAndWriteGame(%v)",
			gameName)
	}

	readonlyGame := actualGame.Read()

	if readonlyGame == nil {
		unitTest.Fatalf(
			testIdentifier+"/nil read-only state from read-and-write state from ReadAndWriteGame(%v)",
			gameName)
	}

	if readonlyGame.Name() != gameName {
		unitTest.Fatalf(
			testIdentifier+"/ReadAndWriteGame(%v) produced game with different name %v",
			gameName,
			readonlyGame)
	}

	return readonlyGame
}