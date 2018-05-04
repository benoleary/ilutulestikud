package persister_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/player"
	"github.com/benoleary/ilutulestikud/backend/player/persister"
)

var colorsAvailableInTest []string = defaults.AvailableColors()
var defaultTestPlayerNames []string = []string{"Player One", "Player Two", "Player Three"}

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

func preparePersisters() []persisterAndDescription {
	return []persisterAndDescription{
		persisterAndDescription{
			PlayerPersister:      persister.NewInMemoryPersister(),
			PersisterDescription: "in-memory persister",
		},
	}
}

func TestReturnErrorWhenPlayerNotFoundInternally(unitTest *testing.T) {
	statePersisters := preparePersisters()

	for _, statePersister := range statePersisters {
		testIdentifier := "Get(unknown player)/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			invalidName := "Not A. Participant"
			playerState, errorFromGet := statePersister.PlayerPersister.Get(invalidName)

			if errorFromGet == nil {
				unitTest.Fatalf(
					"Get(unknown player name %v) did not return an error, did return player state %v",
					invalidName,
					playerState)
			}
		})
	}
}

func TestRejectNewPlayerWithNoName(unitTest *testing.T) {
	statePersisters := preparePersisters()
	playerName := ""
	chatColor := colorsAvailableInTest[0]

	for _, statePersister := range statePersisters {
		testIdentifier :=
			"Reject Add(player with no name)/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			errorFromAdd := statePersister.PlayerPersister.Add(playerName, chatColor)

			// We check that the persister still produces valid states.
			assertPlayerNamesAreCorrectAndGetIsConsistentWithAll(
				testIdentifier,
				unitTest,
				defaultTestPlayerNames,
				statePersister.PlayerPersister)

			// If there was no error, then something went wrong.
			if errorFromAdd == nil {
				unitTest.Fatalf(
					"Add(%v, %v) did not produce an error",
					playerName,
					chatColor)
			}

			// We check that the player was not added.
			playerState, errorFromGet := statePersister.PlayerPersister.Get(playerName)

			// If there was no error, then something went wrong.
			if errorFromGet == nil {
				unitTest.Fatalf(
					"Get(%v) did not produce an error, instead retrieved %v",
					playerName,
					playerState)
			}
		})
	}
}

func TestRejectAddPlayerWithExistingName(unitTest *testing.T) {
	statePersisters := preparePersisters()

	for _, statePersister := range statePersisters {
		for _, playerName := range defaultTestPlayerNames {
			testIdentifier :=
				"Reject Add(player with existing name)/" + statePersister.PersisterDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				errorFromInitialAdd :=
					statePersister.PlayerPersister.Add(playerName, colorsAvailableInTest[0])

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
					statePersister.PlayerPersister.Add(playerName, colorsAvailableInTest[1])

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
						"Add(existing player %v, new color %v) changed the player state from %v to %v",
						playerName,
						colorsAvailableInTest[1],
						initialState,
						existingStateAfterAddWithNewColor)
				}
			})
		}
	}
}

func TestAddPlayerWithValidColorAndTestGet(unitTest *testing.T) {
	statePersisters := preparePersisters()

	chatColor := colorsAvailableInTest[1]

	testCases := []struct {
		testName   string
		playerName string
	}{
		{
			testName:   "Simple ASCII",
			playerName: "New Player",
		},
		{
			testName:   "Non-ASCII and punctuation",
			playerName: "?ß@äô#\"'\"\\\\\\",
		},
		{
			testName:   "Slashes",
			playerName: "/Slashes/are/reserved/for/parsing/URI/segments/",
		},
	}

	for _, statePersister := range statePersisters {
		for _, testCase := range testCases {
			testIdentifier :=
				statePersister.PersisterDescription +
					"/Add(" + testCase.playerName + ", with valid color) and Get(same player)"

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				errorFromAdd := statePersister.PlayerPersister.Add(testCase.playerName, chatColor)

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
						"Add(%v, %v) then Get(%v) produced a state %v which does not have the correct color",
						testCase.playerName,
						chatColor,
						testCase.playerName,
						newState)
				}
			})
		}
	}
}

func TestRejectUpdateInvalidPlayer(unitTest *testing.T) {
	statePersisters := preparePersisters()

	playerName := "Not A. Participant"
	chatColor := colorsAvailableInTest[0]

	for _, statePersister := range statePersisters {
		testIdentifier :=
			"UpdateColor(valid player, invalid color)/" + statePersister.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			errorFromUpdate :=
				statePersister.PlayerPersister.UpdateColor(playerName, chatColor)

			if errorFromUpdate == nil {
				unitTest.Fatalf(
					"UpdateColor(%v, %v) did not produce an error",
					playerName,
					chatColor)
			}

			// We check that the persister still produces valid states.
			assertPlayerNamesAreCorrectAndGetIsConsistentWithAll(
				testIdentifier,
				unitTest,
				defaultTestPlayerNames,
				statePersister.PlayerPersister)

			// We check that the player was not added.
			playerState, errorFromGet := statePersister.PlayerPersister.Get(playerName)

			// If there was no error, then something went wrong.
			if errorFromGet == nil {
				unitTest.Fatalf(
					"Get(%v) did not produce an error, instead retrieved %v",
					playerName,
					playerState)
			}
		})
	}
}

func TestUpdateAllPlayersToFirstColor(unitTest *testing.T) {
	statePersisters := preparePersisters()

	firstColor := colorsAvailableInTest[0]

	for _, statePersister := range statePersisters {
		for _, playerName := range defaultTestPlayerNames {
			testIdentifier :=
				"Update player to first color/" + statePersister.PersisterDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				errorFromAddWithNoColor :=
					statePersister.PlayerPersister.UpdateColor(playerName, firstColor)

				if errorFromAddWithNoColor != nil {
					unitTest.Fatalf(
						"UpdateColor(%v, %v) produced an error: %v",
						playerName,
						firstColor,
						errorFromAddWithNoColor)
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
					(updatedState.Color() != firstColor) {
					unitTest.Fatalf(
						"UpdateColor(%v, %v) then Get(%v) produced state %v",
						playerName,
						firstColor,
						playerName,
						updatedState)
				}
			})
		}
	}
}

func TestReset(unitTest *testing.T) {
	playerNameToAdd := "Added player"
	chatColorForAdd := colorsAvailableInTest[0]
	playerNameToUpdate := defaultTestPlayerNames[0]
	chatColorForUpdate := colorsAvailableInTest[1]

	testCases := []struct {
		testName                string
		shouldAddBeforeReset    bool
		shouldUpdateBeforeReset bool
	}{
		{
			testName:                "No add, no update",
			shouldAddBeforeReset:    false,
			shouldUpdateBeforeReset: false,
		},
		{
			testName:                "Just add, no update",
			shouldAddBeforeReset:    true,
			shouldUpdateBeforeReset: false,
		},
		{
			testName:                "No add, just update",
			shouldAddBeforeReset:    false,
			shouldUpdateBeforeReset: true,
		},
		{
			testName:                "Both add and update",
			shouldAddBeforeReset:    true,
			shouldUpdateBeforeReset: true,
		},
	}

	for _, testCase := range testCases {
		statePersisters := preparePersisters()

		for _, statePersister := range statePersisters {
			testIdentifier :=
				testCase.testName + "/" + statePersister.PersisterDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				if testCase.shouldAddBeforeReset {
					errorFromAdd :=
						statePersister.PlayerPersister.Add(playerNameToAdd, chatColorForAdd)

					if errorFromAdd != nil {
						unitTest.Fatalf(
							"Add(%v, %v) produced an error: %v",
							playerNameToAdd,
							chatColorForAdd,
							errorFromAdd)
					}
				}

				if testCase.shouldUpdateBeforeReset {
					errorFromUpdate :=
						statePersister.PlayerPersister.UpdateColor(playerNameToUpdate, chatColorForUpdate)
					if errorFromUpdate != nil {
						unitTest.Fatalf(
							"UpdateColor(%v, %v) produced an error: %v",
							playerNameToUpdate,
							chatColorForUpdate,
							errorFromUpdate)
					}
				}

				// Now we can reset.
				statePersister.PlayerPersister.Reset()

				// We check that the persister still produces valid states.
				assertPlayerNamesAreCorrectAndGetIsConsistentWithAll(
					testIdentifier,
					unitTest,
					defaultTestPlayerNames,
					statePersister.PlayerPersister)

				// We check that if a player had been added, it is no longer retrievable.
				addedState, errorFromGet :=
					statePersister.PlayerPersister.Get(playerNameToAdd)

				// If there was no error, then something went wrong.
				if errorFromGet == nil {
					unitTest.Fatalf(
						"Get(%v) did not produce an error, instead retrieved %v",
						playerNameToAdd,
						addedState)
				}
			})
		}
	}
}

func assertPlayerNamesAreCorrectAndGetIsConsistentWithAll(
	testIdentifier string,
	unitTest *testing.T,
	playerNames []string,
	playerPersister player.StatePersister) {
	numberOfPlayerNames := len(playerNames)

	statesFromAll := playerPersister.All()

	if len(playerNames) != numberOfPlayerNames {
		unitTest.Fatalf(
			testIdentifier+
				"/All() returned %v which has the wrong number of players to match the given names %v",
			statesFromAll,
			playerNames)
	}

	setOfNamesFromAll := make(map[string]bool, 0)
	for _, stateFromAll := range statesFromAll {
		stateName := stateFromAll.Name()
		if setOfNamesFromAll[stateName] {
			unitTest.Fatalf(
				testIdentifier+"/player name %v duplicated in return from All()",
				stateName)
		}

		setOfNamesFromAll[stateName] = true
	}

	// Now we check that Get(...) is consistent with each player from All().
	for _, stateFromAll := range statesFromAll {
		nameFromAll := stateFromAll.Name()
		stateFromGet, errorFromGet := playerPersister.Get(nameFromAll)
		if errorFromGet != nil {
			unitTest.Fatalf(
				testIdentifier+"/Get(%v) produced error %v",
				nameFromAll,
				errorFromGet)
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
	playerState, errorGettingState := playerPersister.Get(playerName)
	if errorGettingState != nil {
		unitTest.Fatalf(
			testIdentifier+"/Get(%v) produced an error %v",
			playerName,
			errorGettingState)
	}

	if playerState.Name() != playerName {
		unitTest.Fatalf(
			testIdentifier+"/Get(%v) produced player with different name %v",
			playerName,
			playerState)
	}

	return playerState
}
