package persister_test

import (
	"context"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game"
)

// The last element is here to ensure that the game name used to test
// interactions in a single game does not contaminate these tests, if
// the other tests happen first but do not successfully clean up.
// Conversely, these tests do not have an effect on the other tests
// even if they fail to clean up.
var creationRetrievalDeletionTestGameNames = []string{
	testGameNamePrefix + "A valid game",
	testGameNamePrefix + "Name with SQL injection'-- including this",
	testGameNamePrefix + "Another valid game",
	singleInteractionTestGameName,
}

func TestRandomSeedCausesNoPanic(unitTest *testing.T) {
	statePersisters := preparePersisters(unitTest, creationRetrievalDeletionTestGameNames)

	for _, statePersister := range statePersisters {
		testIdentifier := "Positive seed/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			// This is a really trivial test, but it is just nice to have 100% test coverage.
			statePersister.GamePersister.RandomSeed()
		})
	}
}

func TestReturnErrorWhenGameDoesNotExist(unitTest *testing.T) {
	statePersisters := preparePersisters(unitTest, creationRetrievalDeletionTestGameNames)

	for _, statePersister := range statePersisters {
		testIdentifier :=
			"ReadAndWriteGame(unknown game)/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			invalidName := testGameNamePrefix + "Not a valid game"
			gameState, errorFromGet :=
				statePersister.GamePersister.ReadAndWriteGame(
					context.Background(),
					invalidName)

			if errorFromGet == nil {
				unitTest.Fatalf(
					"ReadAndWriteGame(unknown game name %v) did not return an error, did return game state %v",
					invalidName,
					gameState)
			}
		})
	}
}

func TestReturnErrorWhenLeavingNonexistentGame(unitTest *testing.T) {
	statePersisters := preparePersisters(unitTest, creationRetrievalDeletionTestGameNames)
	playerName := "Player One"
	nonexistentGame := testGameNamePrefix + "Non-existent game"

	for _, statePersister := range statePersisters {
		testIdentifier :=
			"ReadAndWriteGame(unknown game)/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			errorFromtLeavingNonexistentGame :=
				statePersister.GamePersister.RemoveGameFromListForPlayer(
					context.Background(),
					nonexistentGame,
					playerName)

			if errorFromtLeavingNonexistentGame == nil {
				unitTest.Fatalf(
					"RemoveGameFromListForPlayer(%v, %v) produced nil error",
					nonexistentGame,
					playerName)
			}
		})
	}
}

func TestReturnEmptyListWhenPlayerHasNoGames(unitTest *testing.T) {
	statePersisters := preparePersisters(unitTest, creationRetrievalDeletionTestGameNames)

	for _, statePersister := range statePersisters {
		testIdentifier :=
			"ReadAllWithPlayer(unknown player)/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			invalidName := "Not A. Participant"
			gameStates, errorFromReadAll :=
				statePersister.GamePersister.ReadAllWithPlayer(
					context.Background(),
					invalidName)

			if errorFromReadAll != nil {
				unitTest.Fatalf(
					"ReadAllWithPlayer(unknown player name %v) produced error %v",
					invalidName,
					errorFromReadAll)
			}

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

func TestRejectAddGameWithNoName(unitTest *testing.T) {
	statePersisters := preparePersisters(unitTest, creationRetrievalDeletionTestGameNames)

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

	for _, statePersister := range statePersisters {
		testIdentifier :=
			"Reject Add(game with no name)/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			errorFromInvalidAdd :=
				statePersister.GamePersister.AddGame(
					context.Background(),
					"",
					logLengthForTest,
					nil,
					defaultTestRuleset,
					threePlayersWithNilHands,
					nil)

			// If there was no error, then something went wrong.
			if errorFromInvalidAdd == nil {
				unitTest.Fatalf(
					"AddGame([empty game name], %v, nil, %v, %v, nil) did not produce an error",
					logLengthForTest,
					defaultTestRuleset,
					threePlayersWithNilHands)
			}
		})
	}
}

func TestRejectAddGameWithExistingName(unitTest *testing.T) {
	statePersisters := preparePersisters(unitTest, creationRetrievalDeletionTestGameNames)

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

	for _, statePersister := range statePersisters {
		expectedGamesMappedToPlayers := make(map[string]map[string]bool, 0)

		for _, gameName := range creationRetrievalDeletionTestGameNames[:2] {
			testIdentifier :=
				"Reject Add(game with existing name)/" + statePersister.PersisterDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				expectedGamesMappedToPlayers[gameName] =
					playerNameSet(twoPlayersWithNilHands)

				errorFromInitialAdd :=
					statePersister.GamePersister.AddGame(
						context.Background(),
						gameName,
						logLengthForTest,
						nil,
						defaultTestRuleset,
						twoPlayersWithNilHands,
						nil)

				if errorFromInitialAdd != nil {
					unitTest.Fatalf(
						"AddGame(%v, %v, nil, %v, %v, nil) produced an error: %v",
						gameName,
						logLengthForTest,
						defaultTestRuleset,
						twoPlayersWithNilHands,
						errorFromInitialAdd)
				}

				// We check that the persister still produces valid states.
				assertReturnedGamesAreConsistent(
					testIdentifier+"/Consistency check after initial add",
					unitTest,
					expectedGamesMappedToPlayers,
					statePersister.GamePersister)

				initialState :=
					getStateAndAssertNoError(
						testIdentifier+"/ReadAndWriteGame(initial game) after initial add",
						unitTest,
						gameName,
						statePersister.GamePersister)

				errorFromSecondAdd :=
					statePersister.GamePersister.AddGame(
						context.Background(),
						gameName,
						logLengthForTest,
						nil,
						defaultTestRuleset,
						threePlayersWithNilHands,
						nil)

				assertGameNameAndParticipantsAreCorrect(
					testIdentifier+"/Name and participants check of initial state after second add",
					unitTest,
					gameName,
					expectedGamesMappedToPlayers[gameName],
					initialState)

				// We check that the persister still produces valid states.
				assertReturnedGamesAreConsistent(
					testIdentifier+"/Consistency check after second add",
					unitTest,
					expectedGamesMappedToPlayers,
					statePersister.GamePersister)

				// If there was no error, then something went wrong.
				if errorFromSecondAdd == nil {
					unitTest.Fatalf(
						"AddGame(%v, %v, %v, %v, nil) did not produce an error",
						gameName,
						logLengthForTest,
						defaultTestRuleset,
						defaultTestPlayers)
				}

				// We check that the player is unchanged.
				existingStateAfterAddWithNewColor :=
					getStateAndAssertNoError(
						testIdentifier+"/ReadAndWriteGame(retrieved game) after second add",
						unitTest,
						gameName,
						statePersister.GamePersister)

				assertGameNameAndParticipantsAreCorrect(
					testIdentifier+"/Name and participants check of retrieved state after second add",
					unitTest,
					gameName,
					expectedGamesMappedToPlayers[gameName],
					existingStateAfterAddWithNewColor)
			})
		}
	}
}

func TestAddGamesThenLeaveGamesThenDeleteGames(unitTest *testing.T) {
	statePersisters := preparePersisters(unitTest, creationRetrievalDeletionTestGameNames)
	leavingPlayer := defaultTestPlayers[0]
	stayingPlayer := defaultTestPlayers[2]
	nonparticipatingPlayer := "Not A. Participant"

	twoPlayersWithNilHands :=
		[]game.PlayerNameWithHand{
			game.PlayerNameWithHand{
				PlayerName:  leavingPlayer,
				InitialHand: nil,
			},
			game.PlayerNameWithHand{
				PlayerName:  stayingPlayer,
				InitialHand: nil,
			},
		}

	threePlayersWithNilHands :=
		[]game.PlayerNameWithHand{
			game.PlayerNameWithHand{
				PlayerName:  stayingPlayer,
				InitialHand: nil,
			},
			game.PlayerNameWithHand{
				PlayerName:  defaultTestPlayers[1],
				InitialHand: nil,
			},
			game.PlayerNameWithHand{
				PlayerName:  leavingPlayer,
				InitialHand: nil,
			},
		}

	for _, statePersister := range statePersisters {
		testIdentifier :=
			"Add games then leave games/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			firstGameName := creationRetrievalDeletionTestGameNames[0]
			errorFromFirstAdd :=
				statePersister.GamePersister.AddGame(
					context.Background(),
					firstGameName,
					logLengthForTest,
					nil,
					defaultTestRuleset,
					twoPlayersWithNilHands,
					nil)

			if errorFromFirstAdd != nil {
				unitTest.Fatalf(
					"AddGame(%v, %v, nil, %v, %v, nil) produced an error: %v",
					firstGameName,
					logLengthForTest,
					defaultTestRuleset,
					twoPlayersWithNilHands,
					errorFromFirstAdd)
			}

			// We just check that no error happens when fetching the game.
			getStateAndAssertNoError(
				testIdentifier+"/ReadAndWriteGame(first game before leaving)",
				unitTest,
				firstGameName,
				statePersister.GamePersister)

			noGameNames := make(map[string]bool, 0)
			justFirstGameName := make(map[string]bool, 1)
			justFirstGameName[firstGameName] = true
			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for leaver after adding first game",
				unitTest,
				leavingPlayer,
				justFirstGameName,
				statePersister.GamePersister)
			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for stayer after adding first game",
				unitTest,
				stayingPlayer,
				justFirstGameName,
				statePersister.GamePersister)
			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for non-participant after adding first game",
				unitTest,
				nonparticipatingPlayer,
				noGameNames,
				statePersister.GamePersister)

			errorFromNonparticipantLeavingFirstGame :=
				statePersister.GamePersister.RemoveGameFromListForPlayer(
					context.Background(),
					firstGameName,
					nonparticipatingPlayer)

			if errorFromNonparticipantLeavingFirstGame == nil {
				unitTest.Fatalf(
					"RemoveGameFromListForPlayer(%v, %v) produced nil error",
					firstGameName,
					nonparticipatingPlayer)
			}

			secondGameName := creationRetrievalDeletionTestGameNames[1]
			errorFromSecondAdd :=
				statePersister.GamePersister.AddGame(
					context.Background(),
					secondGameName,
					logLengthForTest,
					nil,
					defaultTestRuleset,
					threePlayersWithNilHands,
					nil)

			if errorFromSecondAdd != nil {
				unitTest.Fatalf(
					"AddGame(%v, %v, nil, %v, %v, nil) produced an error: %v",
					secondGameName,
					logLengthForTest,
					defaultTestRuleset,
					threePlayersWithNilHands,
					errorFromSecondAdd)
			}

			// We just check that no error happens when fetching the game.
			getStateAndAssertNoError(
				testIdentifier+"/ReadAndWriteGame(second game before leaving)",
				unitTest,
				secondGameName,
				statePersister.GamePersister)

			firstTwoGameNames := make(map[string]bool, 2)
			firstTwoGameNames[firstGameName] = true
			firstTwoGameNames[secondGameName] = true
			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for leaver after adding second game",
				unitTest,
				leavingPlayer,
				firstTwoGameNames,
				statePersister.GamePersister)
			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for stayer after adding second game",
				unitTest,
				stayingPlayer,
				firstTwoGameNames,
				statePersister.GamePersister)

			thirdGameName := creationRetrievalDeletionTestGameNames[2]
			errorFromThirdAdd :=
				statePersister.GamePersister.AddGame(
					context.Background(),
					thirdGameName,
					logLengthForTest,
					nil,
					defaultTestRuleset,
					threePlayersWithNilHands,
					nil)

			if errorFromThirdAdd != nil {
				unitTest.Fatalf(
					"AddGame(%v, %v, nil, %v, %v, nil) produced an error: %v",
					thirdGameName,
					logLengthForTest,
					defaultTestRuleset,
					threePlayersWithNilHands,
					errorFromThirdAdd)
			}

			// We just check that no error happens when fetching the game.
			getStateAndAssertNoError(
				testIdentifier+"/ReadAndWriteGame(third game before leaving)",
				unitTest,
				thirdGameName,
				statePersister.GamePersister)

			allThreeGameNames := make(map[string]bool, 3)
			allThreeGameNames[firstGameName] = true
			allThreeGameNames[secondGameName] = true
			allThreeGameNames[thirdGameName] = true
			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for leaver after adding third game",
				unitTest,
				leavingPlayer,
				allThreeGameNames,
				statePersister.GamePersister)
			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for stayer after adding third game",
				unitTest,
				stayingPlayer,
				allThreeGameNames,
				statePersister.GamePersister)

			errorFromLeavingFirstGame :=
				statePersister.GamePersister.RemoveGameFromListForPlayer(
					context.Background(),
					firstGameName,
					leavingPlayer)

			if errorFromLeavingFirstGame != nil {
				unitTest.Fatalf(
					"RemoveGameFromListForPlayer(%v, %v) produced an error: %v",
					firstGameName,
					leavingPlayer,
					errorFromLeavingFirstGame)
			}

			// We check that the game state had no change in its name or participant list.
			firstStateAfterLeaving :=
				getStateAndAssertNoError(
					testIdentifier+"/ReadAndWriteGame(first game after leaving)",
					unitTest,
					firstGameName,
					statePersister.GamePersister)

			assertGameNameAndParticipantsAreCorrect(
				testIdentifier+"/ReadAndWriteGame(first game after leaving)",
				unitTest,
				firstGameName,
				playerNameSet(twoPlayersWithNilHands),
				firstStateAfterLeaving)

			lastTwoGameNames := make(map[string]bool, 2)
			lastTwoGameNames[secondGameName] = true
			lastTwoGameNames[thirdGameName] = true
			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for leaver after leaving first game",
				unitTest,
				leavingPlayer,
				lastTwoGameNames,
				statePersister.GamePersister)
			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for stayer after leaver leaving first game",
				unitTest,
				stayingPlayer,
				allThreeGameNames,
				statePersister.GamePersister)

			errorFromLeavingThirdGame :=
				statePersister.GamePersister.RemoveGameFromListForPlayer(
					context.Background(),
					thirdGameName,
					leavingPlayer)

			if errorFromLeavingThirdGame != nil {
				unitTest.Fatalf(
					"RemoveGameFromListForPlayer(%v, %v) produced an error: %v",
					thirdGameName,
					leavingPlayer,
					errorFromLeavingThirdGame)
			}

			// We check that the game state had no change in its name or participant list.
			thirdStateAfterLeaving :=
				getStateAndAssertNoError(
					testIdentifier+"/ReadAndWriteGame(third game after leaving)",
					unitTest,
					thirdGameName,
					statePersister.GamePersister)

			assertGameNameAndParticipantsAreCorrect(
				testIdentifier+"/ReadAndWriteGame(third game after leaving)",
				unitTest,
				thirdGameName,
				playerNameSet(threePlayersWithNilHands),
				thirdStateAfterLeaving)

			justSecondGameName := make(map[string]bool, 1)
			justSecondGameName[secondGameName] = true
			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for leaver after leaving third game",
				unitTest,
				leavingPlayer,
				justSecondGameName,
				statePersister.GamePersister)
			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for stayer after leaver leaving third game",
				unitTest,
				stayingPlayer,
				allThreeGameNames,
				statePersister.GamePersister)

			errorFromLeavingSecondGame :=
				statePersister.GamePersister.RemoveGameFromListForPlayer(
					context.Background(),
					secondGameName,
					leavingPlayer)

			if errorFromLeavingSecondGame != nil {
				unitTest.Fatalf(
					"RemoveGameFromListForPlayer(%v, %v) produced an error: %v",
					secondGameName,
					leavingPlayer,
					errorFromLeavingSecondGame)
			}

			// We check that the game state had no change in its name or participant list.
			secondStateAfterLeaving :=
				getStateAndAssertNoError(
					testIdentifier+"/ReadAndWriteGame(second game after leaving)",
					unitTest,
					secondGameName,
					statePersister.GamePersister)

			assertGameNameAndParticipantsAreCorrect(
				testIdentifier+"/ReadAndWriteGame(second game after leaving)",
				unitTest,
				secondGameName,
				playerNameSet(threePlayersWithNilHands),
				secondStateAfterLeaving)

			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for leaver after leaving second game",
				unitTest,
				leavingPlayer,
				noGameNames,
				statePersister.GamePersister)
			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for stayer after leaver leaving second game",
				unitTest,
				stayingPlayer,
				allThreeGameNames,
				statePersister.GamePersister)

			errorFromLeavingSecondGameAgain :=
				statePersister.GamePersister.RemoveGameFromListForPlayer(
					context.Background(),
					secondGameName,
					leavingPlayer)

			if errorFromLeavingSecondGameAgain == nil {
				unitTest.Fatalf(
					"RemoveGameFromListForPlayer(%v, %v) a second time produced nil error",
					secondGameName,
					leavingPlayer)
			}

			errorFromDeletingSecondGame :=
				statePersister.GamePersister.Delete(
					context.Background(),
					secondGameName)

			if errorFromDeletingSecondGame != nil {
				unitTest.Fatalf(
					"Delete(%v) produced error %v",
					secondGameName,
					errorFromDeletingSecondGame)
			}

			firstAndThirdGameNames := make(map[string]bool, 2)
			firstAndThirdGameNames[firstGameName] = true
			firstAndThirdGameNames[thirdGameName] = true

			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for stayer after deleting second game",
				unitTest,
				stayingPlayer,
				firstAndThirdGameNames,
				statePersister.GamePersister)

			secondGame, errorFromGetDeletedSecond :=
				statePersister.GamePersister.ReadAndWriteGame(
					context.Background(),
					secondGameName)
			if errorFromGetDeletedSecond == nil {
				unitTest.Fatalf(
					"ReadAndWriteGame(deleted game %v) produced state %v and nil error",
					secondGameName,
					secondGame)
			}

			errorFromDeletingThirdGame :=
				statePersister.GamePersister.Delete(
					context.Background(),
					thirdGameName)

			if errorFromDeletingThirdGame != nil {
				unitTest.Fatalf(
					"Delete(%v) produced error %v",
					thirdGameName,
					errorFromDeletingThirdGame)
			}

			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for stayer after deleting second and third games",
				unitTest,
				stayingPlayer,
				justFirstGameName,
				statePersister.GamePersister)

			thirdGame, errorFromGetDeletedThird :=
				statePersister.GamePersister.ReadAndWriteGame(
					context.Background(),
					thirdGameName)
			if errorFromGetDeletedThird == nil {
				unitTest.Fatalf(
					"ReadAndWriteGame(deleted game %v) produced state %v and nil error",
					thirdGameName,
					thirdGame)
			}

			errorFromDeletingThirdGameAgain :=
				statePersister.GamePersister.Delete(context.Background(), thirdGameName)

			if errorFromDeletingThirdGameAgain != nil {
				unitTest.Fatalf(
					"Delete(%v) a second time produced error %v",
					thirdGameName,
					errorFromDeletingThirdGameAgain)
			}

			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for stayer after deleting second game once and third game twice",
				unitTest,
				stayingPlayer,
				justFirstGameName,
				statePersister.GamePersister)

			errorFromDeletingFirstGame :=
				statePersister.GamePersister.Delete(
					context.Background(),
					firstGameName)

			if errorFromDeletingFirstGame != nil {
				unitTest.Fatalf(
					"Delete(%v) produced error %v",
					firstGameName,
					errorFromDeletingFirstGame)
			}

			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for stayer after deleting all games",
				unitTest,
				stayingPlayer,
				noGameNames,
				statePersister.GamePersister)

			firstGame, errorFromGetDeletedFirst :=
				statePersister.GamePersister.ReadAndWriteGame(
					context.Background(),
					firstGameName)
			if errorFromGetDeletedFirst == nil {
				unitTest.Fatalf(
					"ReadAndWriteGame(deleted game %v) produced state %v and nil error",
					firstGameName,
					firstGame)
			}

			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for stayer after deleting all games",
				unitTest,
				stayingPlayer,
				noGameNames,
				statePersister.GamePersister)

			errorFromFirstAddAgain :=
				statePersister.GamePersister.AddGame(
					context.Background(),
					firstGameName,
					logLengthForTest,
					nil,
					defaultTestRuleset,
					twoPlayersWithNilHands,
					nil)

			if errorFromFirstAddAgain != nil {
				unitTest.Fatalf(
					"AddGame(%v, %v, nil, %v, %v, nil) produced an error: %v",
					firstGameName,
					logLengthForTest,
					defaultTestRuleset,
					twoPlayersWithNilHands,
					errorFromFirstAddAgain)
			}

			// We just check that no error happens when fetching the game.
			getStateAndAssertNoError(
				testIdentifier+"/ReadAndWriteGame(first game again after deleting all games)",
				unitTest,
				firstGameName,
				statePersister.GamePersister)

			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for leaver after adding first game again after deleting all games",
				unitTest,
				leavingPlayer,
				justFirstGameName,
				statePersister.GamePersister)
			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for stayer after adding first game again after deleting all games",
				unitTest,
				stayingPlayer,
				justFirstGameName,
				statePersister.GamePersister)

			errorFromDeletingFirstGameAgain :=
				statePersister.GamePersister.Delete(context.Background(), firstGameName)

			// This time we expect no error.
			if errorFromDeletingFirstGameAgain != nil {
				unitTest.Fatalf(
					"Delete(%v) produced unexpected error %v instead of expected nil",
					firstGameName,
					errorFromDeletingFirstGameAgain)
			}

			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for stayer after deleting all games again",
				unitTest,
				stayingPlayer,
				noGameNames,
				statePersister.GamePersister)

			firstGameAgain, errorFromGetDeletedFirstAgain :=
				statePersister.GamePersister.ReadAndWriteGame(
					context.Background(),
					firstGameName)
			if errorFromGetDeletedFirstAgain == nil {
				unitTest.Fatalf(
					"ReadAndWriteGame(twice-deleted game %v) produced state %v and nil error",
					firstGameName,
					firstGameAgain)
			}

			assertReadAllWithPlayerGameNamesCorrect(
				testIdentifier+"/all for stayer after deleting all games again",
				unitTest,
				stayingPlayer,
				noGameNames,
				statePersister.GamePersister)
		})
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

		actualGame, errorFromGet :=
			gamePersister.ReadAndWriteGame(context.Background(), expectedGame)

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
		assertReadAllWithPlayerGameNamesCorrect(
			testIdentifier,
			unitTest,
			expectedPlayerName,
			expectedGameNames,
			gamePersister)
	}
}

func assertReadAllWithPlayerGameNamesCorrect(
	testIdentifier string,
	unitTest *testing.T,
	playerName string,
	expectedGameNames map[string]bool,
	gamePersister game.StatePersister) {
	statesFromAllWithPlayer, errorFromReadAll :=
		gamePersister.ReadAllWithPlayer(context.Background(), playerName)

	if errorFromReadAll != nil {
		unitTest.Fatalf(
			"ReadAllWithPlayer(player name %v) produced error %v",
			playerName,
			errorFromReadAll)
	}

	if len(statesFromAllWithPlayer) != len(expectedGameNames) {
		for _, stateFromAllWithPlayer := range statesFromAllWithPlayer {
			unitTest.Logf(
				testIdentifier+
					"/ReadAllWithPlayer(%v) listing states:\n%v\n",
				playerName,
				stateFromAllWithPlayer)
		}

		unitTest.Fatalf(
			testIdentifier+
				"/ReadAllWithPlayer(%v) produced %v which did not have the expected game names %v",
			playerName,
			statesFromAllWithPlayer,
			expectedGameNames)
	}

	for _, gameState := range statesFromAllWithPlayer {
		if !expectedGameNames[gameState.Name()] {
			unitTest.Fatalf(
				testIdentifier+
					"/ReadAllWithPlayer(%v) produced %v which did not have the expected game names %v",
				playerName,
				statesFromAllWithPlayer,
				expectedGameNames)
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
	actualGame, errorFromGet :=
		gamePersister.ReadAndWriteGame(context.Background(), gameName)

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
