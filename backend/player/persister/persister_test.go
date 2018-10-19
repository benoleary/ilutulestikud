package persister_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/cloud"
	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/player"
	"github.com/benoleary/ilutulestikud/backend/player/persister"
)

const testPrefix = "Tough_no_player_name_can_start_like_this_"
const invalidName = testPrefix + "Not A. Participant"

var colorsAvailableInTest []string = defaults.AvailableColors()
var defaultTestPlayerNames []string = []string{
	testPrefix + "Player One",
	testPrefix + "Player Two",
	testPrefix + "Player Three",
}

func mapStringsToTrue(stringsToMap []string) map[string]bool {
	stringMap := make(map[string]bool, 0)
	for _, stringToMap := range stringsToMap {
		stringMap[stringToMap] = true
	}

	return stringMap
}

type persisterAndDescription struct {
	PlayerPersister      player.StatePersister
	PersisterDescription string
}

func preparePersisters(unitTest *testing.T) []persisterAndDescription {
	postgresqlUsername := os.Getenv("POSTGRESQL_USERNAME")
	postgresqlPassword := os.Getenv("POSTGRESQL_PASSWORD")
	postgresqlPlayerdb := os.Getenv("POSTGRESQL_PLAYERDB")
	postgresqlLocation := os.Getenv("POSTGRESQL_LOCATION")
	connectionString :=
		fmt.Sprintf(
			"user=%v password=%v dbname=%v %v",
			postgresqlUsername,
			postgresqlPassword,
			postgresqlPlayerdb,
			postgresqlLocation)

	persistersAndDescriptions := []persisterAndDescription{
		persisterAndDescription{
			PlayerPersister:      persister.NewInMemory(),
			PersisterDescription: "in-memory persister",
		},
		persisterAndDescription{
			PlayerPersister:      persister.NewInPostgresql(connectionString),
			PersisterDescription: "in-PostgreSQL persister",
		},
		persisterAndDescription{
			PlayerPersister:      persister.NewInCloudDatastore(cloud.IlutulestikudIdentifier),
			PersisterDescription: "in-Cloud-Datastore persister",
		},
	}

	for _, playerPersister := range persistersAndDescriptions {
		errorFromDeletionOfExisting :=
			playerPersister.PlayerPersister.Delete(context.Background(), invalidName)
		unitTest.Logf(
			"Error from persister %v deleting %v when setting up"+
				" (to ensure that it does not exist before the test) was %v",
			playerPersister.PersisterDescription,
			invalidName,
			errorFromDeletionOfExisting)
	}

	return persistersAndDescriptions
}

func TestReturnErrorWhenPlayerNotFoundWithGet(unitTest *testing.T) {
	statePersisters := preparePersisters(unitTest)

	for _, statePersister := range statePersisters {
		testIdentifier := "Get(unknown player)/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			playerState, errorFromGet :=
				statePersister.PlayerPersister.Get(
					context.Background(),
					invalidName)

			if errorFromGet == nil {
				unitTest.Fatalf(
					"Get(unknown player name %v) did not return an error, did return player state %v",
					invalidName,
					playerState)
			}
		})
	}
}

func TestRejectAddPlayerWithExistingName(unitTest *testing.T) {
	statePersisters := preparePersisters(unitTest)

	for _, statePersister := range statePersisters {
		for _, playerName := range defaultTestPlayerNames {
			testIdentifier :=
				"Reject Add(player with existing name)/" + statePersister.PersisterDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				errorFromInitialAdd :=
					statePersister.PlayerPersister.Add(
						context.Background(),
						playerName,
						colorsAvailableInTest[0])

				if errorFromInitialAdd != nil {
					unitTest.Fatalf(
						"Add(%v, %v) produced an error: %v",
						playerName,
						colorsAvailableInTest[0],
						errorFromInitialAdd)
				}

				// We check that the persister still produces valid states.
				assertPlayerNamesAreCorrectAndGetIsConsistentWithAll(
					testIdentifier,
					unitTest,
					defaultTestPlayerNames,
					statePersister.PlayerPersister)

				initialState :=
					getStateAndAssertNoError(
						testIdentifier+"/Get(initial player)",
						unitTest,
						playerName,
						statePersister.PlayerPersister)

				errorFromSecondAdd :=
					statePersister.PlayerPersister.Add(
						context.Background(),
						playerName,
						colorsAvailableInTest[1])

				// We check that the persister still produces valid states.
				assertPlayerNamesAreCorrectAndGetIsConsistentWithAll(
					testIdentifier,
					unitTest,
					defaultTestPlayerNames,
					statePersister.PlayerPersister)

				// If there was no error, then something went wrong.
				if errorFromSecondAdd == nil {
					unitTest.Fatalf(
						"Add(%v, %v) did not produce an error",
						playerName,
						colorsAvailableInTest[1])
				}

				// We check that the player is unchanged.
				existingStateAfterAddWithNewColor :=
					getStateAndAssertNoError(
						testIdentifier+"/Get(initial player)",
						unitTest,
						playerName,
						statePersister.PlayerPersister)

				if (existingStateAfterAddWithNewColor.Name() != initialState.Name()) ||
					(existingStateAfterAddWithNewColor.Color() != initialState.Color()) {
					unitTest.Fatalf(
						"Add(existing player %v, new color %v) changed the player state from %+v to %+v",
						playerName,
						colorsAvailableInTest[1],
						initialState,
						existingStateAfterAddWithNewColor)
				}

				// We also clean up afterwards, but report any error as a test failure.
				errorFromDelete :=
					statePersister.PlayerPersister.Delete(
						context.Background(),
						playerName)
				if errorFromDelete != nil {
					unitTest.Fatalf(
						"Delete(%v) produced error %v",
						playerName,
						errorFromDelete)
				}
			})
		}
	}
}

func TestAddPlayerWithValidColorAndTestGet(unitTest *testing.T) {
	statePersisters := preparePersisters(unitTest)

	chatColor := colorsAvailableInTest[1]

	testCases := []struct {
		testName   string
		playerName string
	}{
		{
			testName:   "Simple ASCII",
			playerName: testPrefix + "New Player",
		},
		{
			testName:   "Non-ASCII and punctuation",
			playerName: testPrefix + "?ß@äô#\"'\"\\\\\\",
		},
		{
			testName:   "Slashes",
			playerName: testPrefix + "/Slashes/are/reserved/for/parsing/URI/segments/",
		},
		{
			testName:   "SQL injection",
			playerName: testPrefix + "'; --",
		},
	}

	for _, statePersister := range statePersisters {
		for _, testCase := range testCases {
			testIdentifier :=
				statePersister.PersisterDescription +
					"/Add(" + testCase.playerName + ", with valid color) and Get(same player)"

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				errorFromAdd :=
					statePersister.PlayerPersister.Add(
						context.Background(),
						testCase.playerName,
						chatColor)

				if errorFromAdd != nil {
					unitTest.Fatalf(
						"Add(%v, %v) produced an error %v",
						testCase.playerName,
						chatColor,
						errorFromAdd)
				}

				// We check that the persister still produces valid states.
				assertPlayerNamesAreCorrectAndGetIsConsistentWithAll(
					testIdentifier,
					unitTest,
					defaultTestPlayerNames,
					statePersister.PlayerPersister)

				// We check that the player can be retrieved.
				newState :=
					getStateAndAssertNoError(
						testIdentifier+"/Retrieve with Get(...)",
						unitTest,
						testCase.playerName,
						statePersister.PlayerPersister)

				if newState.Color() != chatColor {
					unitTest.Fatalf(
						"Add(%v, %v) then Get(%v) produced a state %v which does not have"+
							" the correct color",
						testCase.playerName,
						chatColor,
						testCase.playerName,
						newState)
				}

				// We also clean up afterwards, but report any error as a test failure.
				errorFromDelete :=
					statePersister.PlayerPersister.Delete(
						context.Background(),
						testCase.playerName)
				if errorFromDelete != nil {
					unitTest.Fatalf(
						"Delete(%v) produced error %v",
						testCase.playerName,
						errorFromDelete)
				}
			})
		}
	}
}

func TestAddNewPlayersThenDeleteThem(unitTest *testing.T) {
	statePersisters := preparePersisters(unitTest)
	firstPlayer := testPrefix + "First Player"
	firstColor := colorsAvailableInTest[0]
	secondPlayer := testPrefix + "Second Player"
	secondColor := colorsAvailableInTest[1]

	for _, statePersister := range statePersisters {
		testIdentifier :=
			statePersister.PersisterDescription +
				"/Add(" + firstPlayer + ") then Add(" + secondPlayer +
				") then Delete(" + firstPlayer + ") then Delete(" + secondPlayer + ")"

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			preexistingPlayers, errorFromPreexistingAll :=
				statePersister.PlayerPersister.All(context.Background())
			if errorFromPreexistingAll != nil {
				unitTest.Fatalf(
					"All() produced an error %v",
					errorFromPreexistingAll)
			}

			numberOfPreexistingPlayers := len(preexistingPlayers)

			errorFromFirstAdd :=
				statePersister.PlayerPersister.Add(
					context.Background(),
					firstPlayer,
					firstColor)
			if errorFromFirstAdd != nil {
				unitTest.Fatalf(
					"Add(%v, %v) produced an error %v",
					firstPlayer,
					colorsAvailableInTest[0],
					errorFromFirstAdd)
			}

			statesFromAllAfterFirstAdd, errorFromAllAfterFirstAdd :=
				statePersister.PlayerPersister.All(context.Background())
			if errorFromAllAfterFirstAdd != nil {
				unitTest.Fatalf(
					"All() after first Add(%v, %v) produced an error %v",
					firstPlayer,
					colorsAvailableInTest[0],
					errorFromAllAfterFirstAdd)
			}

			if len(statesFromAllAfterFirstAdd) != (numberOfPreexistingPlayers + 1) {
				unitTest.Fatalf(
					"Adding first player to a new persister resulted in states %+v",
					statesFromAllAfterFirstAdd)
			}

			// We check that the persister still produces valid states.
			assertPlayerNamesAreCorrectAndGetIsConsistentWithAll(
				testIdentifier,
				unitTest,
				defaultTestPlayerNames,
				statePersister.PlayerPersister)

			// We check that the first player can be retrieved.
			firstStateBeforeSecondAdd :=
				getStateAndAssertNoError(
					testIdentifier+"/Retrieve with Get(...)",
					unitTest,
					firstPlayer,
					statePersister.PlayerPersister)

			if firstStateBeforeSecondAdd.Color() != firstColor {
				unitTest.Fatalf(
					"Add(%v, %v) then Get(%v) produced a state %+v which does not have"+
						" the correct color",
					firstPlayer,
					firstColor,
					firstPlayer,
					firstStateBeforeSecondAdd)
			}

			errorFromSecondAdd :=
				statePersister.PlayerPersister.Add(
					context.Background(),
					secondPlayer,
					secondColor)

			if errorFromSecondAdd != nil {
				unitTest.Fatalf(
					"Add(%v, %v) produced an error %v",
					secondPlayer,
					secondColor,
					errorFromSecondAdd)
			}

			statesFromAllAfterSecondAdd, errorFromAllAfterSecondAdd :=
				statePersister.PlayerPersister.All(context.Background())
			if errorFromAllAfterSecondAdd != nil {
				unitTest.Fatalf(
					"All() after second Add(%v, %v) produced an error %v",
					firstPlayer,
					colorsAvailableInTest[0],
					errorFromAllAfterSecondAdd)
			}

			if len(statesFromAllAfterSecondAdd) != (numberOfPreexistingPlayers + 2) {
				unitTest.Fatalf(
					"Adding second player after adding first player to a new persister"+
						" resulted in states %+v",
					statesFromAllAfterSecondAdd)
			}

			// We check that the persister still produces valid states.
			assertPlayerNamesAreCorrectAndGetIsConsistentWithAll(
				testIdentifier,
				unitTest,
				defaultTestPlayerNames,
				statePersister.PlayerPersister)

			// We check that the first player can still be retrieved.
			firstStateAfterSecondAdd :=
				getStateAndAssertNoError(
					testIdentifier+"/Retrieve with Get(...)",
					unitTest,
					firstPlayer,
					statePersister.PlayerPersister)

			if firstStateAfterSecondAdd.Color() != firstColor {
				unitTest.Fatalf(
					"First state %+v changed color from expected %v",
					firstStateAfterSecondAdd,
					firstColor)
			}

			// We check that the second player can be retrieved.
			secondStateBeforeFirstDelete :=
				getStateAndAssertNoError(
					testIdentifier+"/Retrieve with Get(...)",
					unitTest,
					secondPlayer,
					statePersister.PlayerPersister)

			if secondStateBeforeFirstDelete.Color() != secondColor {
				unitTest.Fatalf(
					"Add(%v, %v) then Get(%v) produced a state %v which does not have"+
						" the correct color",
					secondPlayer,
					secondColor,
					secondPlayer,
					secondStateBeforeFirstDelete)
			}

			errorFromFirstDelete :=
				statePersister.PlayerPersister.Delete(
					context.Background(),
					firstPlayer)

			if errorFromFirstDelete != nil {
				unitTest.Fatalf(
					"Delete(%v) produced an error %v",
					firstPlayer,
					errorFromFirstDelete)
			}

			statesFromAllAfterFirstDelete, errorFromAllAfterFirstDelete :=
				statePersister.PlayerPersister.All(context.Background())
			if errorFromAllAfterFirstDelete != nil {
				unitTest.Fatalf(
					"All() after first Delete(%v) produced an error %v",
					firstPlayer,
					errorFromAllAfterFirstDelete)
			}

			if len(statesFromAllAfterFirstDelete) != (numberOfPreexistingPlayers + 1) {
				unitTest.Fatalf(
					"Deleting the first player after adding two players to a new persister"+
						" resulted in states %+v",
					statesFromAllAfterFirstDelete)
			}

			// We check that the second player can still be retrieved.
			secondStateAfterFirstDelete :=
				getStateAndAssertNoError(
					testIdentifier+"/Retrieve with Get(...)",
					unitTest,
					secondPlayer,
					statePersister.PlayerPersister)

			if secondStateAfterFirstDelete.Color() != secondColor {
				unitTest.Fatalf(
					"Second state %+v changed color from expected %v",
					secondStateAfterFirstDelete,
					secondColor)
			}

			errorFromSecondDelete :=
				statePersister.PlayerPersister.Delete(
					context.Background(),
					secondPlayer)

			if errorFromSecondDelete != nil {
				unitTest.Fatalf(
					"Delete(%v) produced an error %v",
					secondPlayer,
					errorFromSecondDelete)
			}

			statesFromAllAfterSecondDelete, errorFromAllAfterSecondDelete :=
				statePersister.PlayerPersister.All(context.Background())
			if errorFromAllAfterSecondDelete != nil {
				unitTest.Fatalf(
					"All() after second Delete(%v) produced an error %v",
					secondPlayer,
					errorFromAllAfterSecondDelete)
			}

			if len(statesFromAllAfterSecondDelete) != numberOfPreexistingPlayers {
				unitTest.Fatalf(
					"Deleting the second player after deleting first player after"+
						" adding two players to a new persister resulted in states %+v",
					statesFromAllAfterSecondDelete)
			}
		})
	}
}

func TestRejectUpdateInvalidPlayer(unitTest *testing.T) {
	statePersisters := preparePersisters(unitTest)

	chatColor := colorsAvailableInTest[0]

	for _, statePersister := range statePersisters {
		testIdentifier :=
			"UpdateColor(valid player, invalid color)/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			errorFromUpdate :=
				statePersister.PlayerPersister.UpdateColor(
					context.Background(),
					invalidName,
					chatColor)

			if errorFromUpdate == nil {
				unitTest.Fatalf(
					"UpdateColor(%v, %v) did not produce an error",
					invalidName,
					chatColor)
			}

			// We check that the persister still produces valid states.
			assertPlayerNamesAreCorrectAndGetIsConsistentWithAll(
				testIdentifier,
				unitTest,
				defaultTestPlayerNames,
				statePersister.PlayerPersister)

			// We check that the player was not added.
			playerState, errorFromGet :=
				statePersister.PlayerPersister.Get(
					context.Background(),
					invalidName)

			// If there was no error, then something went wrong.
			if errorFromGet == nil {
				unitTest.Fatalf(
					"Get(%v) did not produce an error, instead retrieved %v",
					invalidName,
					playerState)
			}
		})
	}
}

func TestUpdateAllPlayersToNewColor(unitTest *testing.T) {
	statePersisters := preparePersisters(unitTest)

	initialColor := colorsAvailableInTest[0]
	newColor := colorsAvailableInTest[1]

	for _, statePersister := range statePersisters {
		testIdentifier :=
			"Update player to new color/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			// First we have to add all the required players.
			for _, playerName := range defaultTestPlayerNames {
				errorFromAdd :=
					statePersister.PlayerPersister.Add(
						context.Background(),
						playerName,
						initialColor)

				if errorFromAdd != nil {
					unitTest.Fatalf(
						"Add(%v, %v) produced an error %v",
						playerName,
						initialColor,
						errorFromAdd)
				}
			}

			for _, playerName := range defaultTestPlayerNames {
				errorFromUpdateColor :=
					statePersister.PlayerPersister.UpdateColor(
						context.Background(),
						playerName,
						newColor)

				if errorFromUpdateColor != nil {
					unitTest.Fatalf(
						"UpdateColor(%v, %v) produced an error: %v",
						playerName,
						newColor,
						errorFromUpdateColor)
				}

				// We check that the persister still produces valid states.
				assertPlayerNamesAreCorrectAndGetIsConsistentWithAll(
					testIdentifier,
					unitTest,
					defaultTestPlayerNames,
					statePersister.PlayerPersister)

				// We check that the player has the correct color.
				updatedState :=
					getStateAndAssertNoError(
						testIdentifier+"/Get(updated player)",
						unitTest,
						playerName,
						statePersister.PlayerPersister)

				if (updatedState.Name() != playerName) ||
					(updatedState.Color() != newColor) {
					unitTest.Fatalf(
						"UpdateColor(%v, %v) then Get(%v) produced state %v",
						playerName,
						newColor,
						playerName,
						updatedState)
				}
			}

			// Finally, we have to clean up the added test players.
			for _, playerName := range defaultTestPlayerNames {
				errorFromDelete :=
					statePersister.PlayerPersister.Delete(
						context.Background(),
						playerName)

				if errorFromDelete != nil {
					unitTest.Fatalf(
						"Delete(%v) produced error %v",
						playerName,
						errorFromDelete)
				}
			}
		})
	}
}

func assertPlayerNamesAreCorrectAndGetIsConsistentWithAll(
	testIdentifier string,
	unitTest *testing.T,
	playerNames []string,
	playerPersister player.StatePersister) {
	numberOfPlayerNames := len(playerNames)

	statesFromAll, errorFromAll :=
		playerPersister.All(context.Background())
	if errorFromAll != nil {
		unitTest.Fatalf(
			testIdentifier+
				"/All() returned %+v and error %v",
			statesFromAll,
			errorFromAll)
	}

	if len(playerNames) != numberOfPlayerNames {
		unitTest.Fatalf(
			testIdentifier+
				"/All() returned %v which has the wrong number of players to match the given names %v",
			statesFromAll,
			playerNames)
	}

	setOfNamesFromAll := make(map[string]bool, 0)
	for _, stateFromAll := range statesFromAll {
		if stateFromAll == nil {
			unitTest.Fatalf(
				testIdentifier+"/nil state in return from All(): %v",
				statesFromAll)
		}

		stateName := stateFromAll.Name()
		if setOfNamesFromAll[stateName] {
			unitTest.Fatalf(
				testIdentifier+"/player name %v duplicated in return from All() %v",
				stateName,
				statesFromAll)
		}

		setOfNamesFromAll[stateName] = true
	}

	// Now we check that Get(...) is consistent with each player from All().
	for _, stateFromAll := range statesFromAll {
		// At this point we can be sure that there are no nils in statesFromAll.
		nameFromAll := stateFromAll.Name()
		stateFromGet, errorFromGet :=
			playerPersister.Get(
				context.Background(),
				nameFromAll)
		if errorFromGet != nil {
			unitTest.Fatalf(
				testIdentifier+"/Get(%v) produced error %v",
				nameFromAll,
				errorFromGet)
		}

		if stateFromGet == nil {
			unitTest.Fatalf(
				testIdentifier+"/nil state from Get(%v)",
				nameFromAll)
		}

		if (stateFromGet.Name() != nameFromAll) ||
			(stateFromGet.Color() != stateFromAll.Color()) {
			unitTest.Fatalf(
				testIdentifier+"/State from Get(...) %v did not match state from All() %v",
				stateFromAll,
				stateFromGet)
		}
	}
}

func getStateAndAssertNoError(
	testIdentifier string,
	unitTest *testing.T,
	playerName string,
	playerPersister player.StatePersister) player.ReadonlyState {
	playerState, errorGettingState :=
		playerPersister.Get(
			context.Background(),
			playerName)
	if errorGettingState != nil {
		unitTest.Fatalf(
			testIdentifier+"/Get(%v) produced an error %v",
			playerName,
			errorGettingState)
	}

	if playerState == nil {
		unitTest.Fatalf(
			testIdentifier+"/nil state from Get(%v)",
			playerName)
	}

	if playerState.Name() != playerName {
		unitTest.Fatalf(
			testIdentifier+"/Get(%v) produced player with different name %v",
			playerName,
			playerState)
	}

	return playerState
}
