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

type collectionAndDescription struct {
	PlayerCollection      *player.StateCollection
	CollectionDescription string
}

func prepareCollections(
	unitTest *testing.T,
	initialPlayerNames []string,
	availableColors []string) []collectionAndDescription {
	stateCollections, errorFromCreation :=
		prepareCollectionsOrReturnError(
			initialPlayerNames,
			availableColors)

	if errorFromCreation != nil {
		unitTest.Fatalf(
			"Error when preparing collections: %v",
			errorFromCreation)
	}

	return stateCollections
}

func prepareCollectionsOrReturnError(
	initialPlayerNames []string,
	availableColors []string) ([]collectionAndDescription, error) {
	statePersisters := []persisterAndDescription{
		persisterAndDescription{
			PlayerPersister:      persister.NewInMemoryPersister(),
			PersisterDescription: "in-memory persister",
		},
	}

	numberOfPersisters := len(statePersisters)

	stateCollections := make([]collectionAndDescription, numberOfPersisters)

	for persisterIndex := 0; persisterIndex < numberOfPersisters; persisterIndex++ {
		statePersister := statePersisters[persisterIndex]
		stateCollection, creationError :=
			player.NewCollection(
				statePersister.PlayerPersister,
				initialPlayerNames,
				availableColors)

		if creationError != nil {
			return nil, creationError
		}

		stateCollections[persisterIndex] = collectionAndDescription{
			PlayerCollection:      stateCollection,
			CollectionDescription: "collection around " + statePersister.PersisterDescription,
		}
	}

	return stateCollections, nil
}

func TestAllCorrectlyReturnsInitialPlayers(unitTest *testing.T) {
	testCases := []struct {
		testName           string
		initialPlayerNames []string
	}{
		{
			testName:           "Nil initial player list",
			initialPlayerNames: nil,
		},
		{
			testName:           "Empty initial player list",
			initialPlayerNames: []string{},
		},
		{
			testName:           "Default initial player list",
			initialPlayerNames: defaultTestPlayerNames,
		},
	}

	for _, testCase := range testCases {
		collectionTypes :=
			prepareCollections(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

		for _, collectionType := range collectionTypes {
			testIdentifier := testCase.testName + "/" + collectionType.CollectionDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
					testIdentifier,
					unitTest,
					testCase.initialPlayerNames,
					colorsAvailableInTest,
					collectionType.PlayerCollection)
			})
		}
	}
}

func TestReturnErrorWhenPlayerNotFoundInternally(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

	for _, collectionType := range collectionTypes {
		testIdentifier := "Get(unknown player)/" + collectionType.CollectionDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			invalidName := "Not A. Participant"
			playerState, errorFromGet :=
				collectionType.PlayerCollection.Get(invalidName)

			if errorFromGet == nil {
				unitTest.Fatalf(
					"Get(unknown player name %v) did not return an error, did return player state %v",
					invalidName,
					playerState)
			}
		})
	}
}

func TestErrorIfEmptyAvailableColors(unitTest *testing.T) {
	stateCollections, errorFromCreation :=
		prepareCollectionsOrReturnError(
			defaultTestPlayerNames,
			[]string{})

	if errorFromCreation == nil {
		unitTest.Fatalf(
			"No error when preparing collections with empty list of colors, returned %v",
			stateCollections)
	}
}

func TestNonemptyAvailableColors(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

	for _, collectionType := range collectionTypes {
		testIdentifier := "Non-empty available colors/" + collectionType.CollectionDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			availableColors := collectionType.PlayerCollection.AvailableChatColors()

			numberOfExpectedColors := len(colorsAvailableInTest)

			if len(availableColors) != numberOfExpectedColors {
				unitTest.Fatalf(
					"AvailableChatColors() set up with %v returned list %v which has wrong size",
					colorsAvailableInTest,
					availableColors)
			}

			expectedColorMap := mapStringsToTrue(colorsAvailableInTest)

			for colorIndex := 0; colorIndex < numberOfExpectedColors; colorIndex++ {
				if !expectedColorMap[availableColors[colorIndex]] {
					unitTest.Fatalf(
						"AvailableChatColors() set up with %v returned list %v which had unexpected color %v",
						colorsAvailableInTest,
						availableColors,
						availableColors[colorIndex])
				}
			}
		})
	}
}

func TestRejectNewPlayerWithNoName(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(unitTest, defaultTestPlayerNames, colorsAvailableInTest)
	playerName := ""
	chatColor := colorsAvailableInTest[0]

	for _, collectionType := range collectionTypes {
		testIdentifier :=
			"Reject Add(player with no name)/" + collectionType.CollectionDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			errorFromAdd :=
				collectionType.PlayerCollection.Add(playerName, chatColor)

			// We check that the collection still produces valid states.
			assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
				testIdentifier,
				unitTest,
				defaultTestPlayerNames,
				colorsAvailableInTest,
				collectionType.PlayerCollection)

			// If there was no error, then something went wrong.
			if errorFromAdd == nil {
				unitTest.Fatalf(
					"Add(%v, %v) did not produce an error",
					playerName,
					chatColor)
			}

			// We check that the player was not added.
			playerState, errorFromGet :=
				collectionType.PlayerCollection.Get(playerName)

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

func TestAddNewPlayerWithInvalidColor(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(unitTest, defaultTestPlayerNames, colorsAvailableInTest)
	playerName := "A. New Player"
	invalidColor := "Not a valid color"

	for _, collectionType := range collectionTypes {
		testIdentifier :=
			"Reject Add(player with invalid color)/" + collectionType.CollectionDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			errorFromAdd :=
				collectionType.PlayerCollection.Add(playerName, invalidColor)

			// We check that the collection still produces valid states.
			assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
				testIdentifier,
				unitTest,
				defaultTestPlayerNames,
				colorsAvailableInTest,
				collectionType.PlayerCollection)

			// If there was no error, then something went wrong.
			if errorFromAdd == nil {
				unitTest.Fatalf(
					"Add(%v, %v) did not produce an error",
					playerName,
					invalidColor)
			}

			// We check that the player was not added.
			playerState, errorFromGet :=
				collectionType.PlayerCollection.Get(playerName)

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
	collectionTypes :=
		prepareCollections(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

	for _, collectionType := range collectionTypes {
		for _, playerName := range defaultTestPlayerNames {
			testIdentifier :=
				"Reject Add(player with existing name)/" + collectionType.CollectionDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				initialState := getStateAndAssertNoError(
					testIdentifier+"/Get(initial player)",
					unitTest,
					playerName,
					collectionType.PlayerCollection)

				errorFromAddWithNoColor := collectionType.PlayerCollection.Add(playerName, "")

				// We check that the collection still produces valid states.
				assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
					testIdentifier,
					unitTest,
					defaultTestPlayerNames,
					colorsAvailableInTest,
					collectionType.PlayerCollection)

				// If there was no error, then something went wrong.
				if errorFromAddWithNoColor == nil {
					unitTest.Fatalf(
						"Add(%v, [empty string for color]) did not produce an error",
						playerName)
				}

				// We check that the player is unchanged.
				existingStateAfterAddWithNoColor := getStateAndAssertNoError(
					testIdentifier+"/Get(initial player)",
					unitTest,
					playerName,
					collectionType.PlayerCollection)

				if (existingStateAfterAddWithNoColor.Name() != existingStateAfterAddWithNoColor.Name()) ||
					(existingStateAfterAddWithNoColor.Color() != existingStateAfterAddWithNoColor.Color()) {
					unitTest.Fatalf(
						"Add(existing player %v, empty color string) changed the player state from %v to %v",
						playerName,
						initialState,
						existingStateAfterAddWithNoColor)
				}

				newColor := colorsAvailableInTest[0]
				if newColor == initialState.Color() {
					newColor = colorsAvailableInTest[1]
				}

				errorFromAddWithNewColor := collectionType.PlayerCollection.Add(playerName, newColor)

				// We check that the collection still produces valid states.
				assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
					testIdentifier,
					unitTest,
					defaultTestPlayerNames,
					colorsAvailableInTest,
					collectionType.PlayerCollection)

				// If there was no error, then something went wrong.
				if errorFromAddWithNewColor == nil {
					unitTest.Fatalf(
						"Add(%v, %v) did not produce an error",
						playerName,
						newColor)
				}

				// We check that the player is unchanged.
				existingStateAfterAddWithNewColor := getStateAndAssertNoError(
					testIdentifier+"/Get(initial player)",
					unitTest,
					playerName,
					collectionType.PlayerCollection)

				if (existingStateAfterAddWithNewColor.Name() != initialState.Name()) ||
					(existingStateAfterAddWithNewColor.Color() != initialState.Color()) {
					unitTest.Fatalf(
						"Add(existing player %v, new color %v) changed the player state from %v to %v",
						playerName,
						newColor,
						initialState,
						existingStateAfterAddWithNewColor)
				}
			})
		}
	}
}

func TestAddPlayerWithValidColorAndTestGet(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

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

	for _, collectionType := range collectionTypes {
		for _, testCase := range testCases {
			testIdentifier :=
				collectionType.CollectionDescription +
					"/Add(" + testCase.playerName + ", with valid color) and Get(same player)"

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				errorFromAdd :=
					collectionType.PlayerCollection.Add(testCase.playerName, chatColor)

				if errorFromAdd != nil {
					unitTest.Fatalf(
						"Add(%v, %v) produced an error %v",
						testCase.playerName,
						chatColor,
						errorFromAdd)
				}

				// We check that the collection still produces valid states.
				assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
					testIdentifier,
					unitTest,
					defaultTestPlayerNames,
					colorsAvailableInTest,
					collectionType.PlayerCollection)

				// We check that the player can be retrieved.
				newState := getStateAndAssertNoError(
					testIdentifier+"/Retrieve with Get(...)",
					unitTest,
					testCase.playerName,
					collectionType.PlayerCollection)

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

func TestAddPlayerWithNoColorAndTestGetHasValidColor(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

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

	for _, collectionType := range collectionTypes {
		for _, testCase := range testCases {
			testIdentifier :=
				collectionType.CollectionDescription +
					"/Add(" + testCase.playerName + ", with no color) and Get(same player)"

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				errorFromAdd := collectionType.PlayerCollection.Add(testCase.playerName, "")

				if errorFromAdd != nil {
					unitTest.Fatalf(
						"Add(%v, [empty string for color]) produced an error %v",
						testCase.playerName,
						errorFromAdd)
				}

				// We check that the collection still produces valid states.
				assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
					testIdentifier,
					unitTest,
					defaultTestPlayerNames,
					colorsAvailableInTest,
					collectionType.PlayerCollection)

				// We check that the player can be retrieved.
				newState := getStateAndAssertNoError(
					testIdentifier+"/Get(newly-added player)",
					unitTest,
					testCase.playerName,
					collectionType.PlayerCollection)

				newStateHasValidColor := mapStringsToTrue(colorsAvailableInTest)[newState.Color()]
				if !newStateHasValidColor {
					unitTest.Fatalf(
						"Add(%v, [empty string for color]) then Get(%v) produced a state %v which does not have a color in the list %v",
						testCase.playerName,
						newState,
						testCase.playerName,
						colorsAvailableInTest)
				}
			})
		}
	}
}

func TestRejectUpdateInvalidPlayer(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

	playerName := "Not A. Participant"
	chatColor := colorsAvailableInTest[0]

	for _, collectionType := range collectionTypes {
		testIdentifier :=
			"UpdateColor(valid player, invalid color)/" + collectionType.CollectionDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			errorFromUpdate :=
				collectionType.PlayerCollection.UpdateColor(playerName, chatColor)

			if errorFromUpdate == nil {
				unitTest.Fatalf(
					"UpdateColor(%v, %v) did not produce an error",
					playerName,
					chatColor)
			}

			// We check that the collection still produces valid states.
			assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
				testIdentifier,
				unitTest,
				defaultTestPlayerNames,
				colorsAvailableInTest,
				collectionType.PlayerCollection)

			// We check that the player was not added.
			playerState, errorFromGet :=
				collectionType.PlayerCollection.Get(playerName)

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

func TestRejectUpdatePlayerWithInvalidColor(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

	testCases := []struct {
		testName   string
		playerName string
		chatColor  string
	}{
		{
			testName:   "Valid player but missing color",
			playerName: defaultTestPlayerNames[0],
			chatColor:  "",
		},
		{
			testName:   "Valid player but invalid color",
			playerName: defaultTestPlayerNames[1],
			chatColor:  "not a color",
		},
	}

	for _, collectionType := range collectionTypes {
		for _, testCase := range testCases {
			testIdentifier :=
				testCase.testName + "/" + collectionType.CollectionDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				// We save the state to later check that the color was not changed.
				initialState := getStateAndAssertNoError(
					testIdentifier+"/Get(valid player) before update",
					unitTest,
					testCase.playerName,
					collectionType.PlayerCollection)

				errorFromUpdate :=
					collectionType.PlayerCollection.UpdateColor(testCase.playerName, testCase.chatColor)

				if errorFromUpdate == nil {
					unitTest.Fatalf(
						"UpdateColor(%v, %v) did not produce an error",
						testCase.playerName,
						testCase.chatColor)
				}

				// We check that the collection still produces valid states.
				assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
					testIdentifier,
					unitTest,
					defaultTestPlayerNames,
					colorsAvailableInTest,
					collectionType.PlayerCollection)

				// We check that the player can be retrieved.
				stateAfterUpdate := getStateAndAssertNoError(
					testIdentifier+"/Get(same player) after update",
					unitTest,
					testCase.playerName,
					collectionType.PlayerCollection)

				if stateAfterUpdate.Color() != initialState.Color() {
					unitTest.Fatalf(
						"UpdateColor(%v, %v) changed the state from %v to %v",
						testCase.playerName,
						testCase.chatColor,
						initialState,
						stateAfterUpdate)
				}
			})
		}
	}
}

func TestUpdateAllPlayersToFirstColor(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

	firstColor := colorsAvailableInTest[0]

	for _, collectionType := range collectionTypes {
		for _, playerName := range defaultTestPlayerNames {
			testIdentifier :=
				"Update player to first color/" + collectionType.CollectionDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				errorFromAddWithNoColor := collectionType.PlayerCollection.UpdateColor(playerName, firstColor)

				if errorFromAddWithNoColor != nil {
					unitTest.Fatalf(
						"UpdateColor(%v, %v) produced an error: %v",
						playerName,
						firstColor,
						errorFromAddWithNoColor)
				}

				// We check that the collection still produces valid states.
				assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
					testIdentifier,
					unitTest,
					defaultTestPlayerNames,
					colorsAvailableInTest,
					collectionType.PlayerCollection)

				// We check that the player has the correct color.
				updatedState := getStateAndAssertNoError(
					testIdentifier+"/Get(updated player)",
					unitTest,
					playerName,
					collectionType.PlayerCollection)

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
		collectionTypes :=
			prepareCollections(unitTest, defaultTestPlayerNames, colorsAvailableInTest)
		for _, collectionType := range collectionTypes {
			testIdentifier :=
				testCase.testName + "/" + collectionType.CollectionDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				if testCase.shouldAddBeforeReset {
					errorFromAdd :=
						collectionType.PlayerCollection.Add(playerNameToAdd, chatColorForAdd)

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
						collectionType.PlayerCollection.UpdateColor(playerNameToUpdate, chatColorForUpdate)
					if errorFromUpdate != nil {
						unitTest.Fatalf(
							"UpdateColor(%v, %v) produced an error: %v",
							playerNameToUpdate,
							chatColorForUpdate,
							errorFromUpdate)
					}
				}

				// Now we can reset.
				collectionType.PlayerCollection.Reset()

				// We check that the collection still produces valid states.
				assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
					testIdentifier,
					unitTest,
					defaultTestPlayerNames,
					colorsAvailableInTest,
					collectionType.PlayerCollection)

				// We check that if a player had been added, it is no longer retrievable.
				addedState, errorFromGet :=
					collectionType.PlayerCollection.Get(playerNameToAdd)

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

func assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
	testIdentifier string,
	unitTest *testing.T,
	playerNames []string,
	validColors []string,
	playerCollection *player.StateCollection) {
	// First we set up a map of valid colors, ignoring possible duplication.
	if len(validColors) <= 0 {
		unitTest.Fatalf(testIdentifier + "/no valid colors provided to check against player states")
	}

	validColorMap := mapStringsToTrue(validColors)

	numberOfPlayerNames := len(playerNames)

	statesFromAll := playerCollection.All()

	if len(playerNames) != numberOfPlayerNames {
		unitTest.Fatalf(
			testIdentifier+
				"/All() returned %v which has the wrong number of players to match the given names %v",
			statesFromAll,
			playerNames)
	}

	setOfNamesFromAll := make(map[string]bool, 0)
	for _, stateFromAll := range statesFromAll {
		stateColor := stateFromAll.Color()
		if !validColorMap[stateColor] {
			unitTest.Fatalf(
				testIdentifier+
					"/player %v has color not contained in list of valid colors %v",
				stateFromAll,
				validColors)
		}

		stateName := stateFromAll.Name()
		if setOfNamesFromAll[stateName] {
			unitTest.Fatalf(
				testIdentifier+
					"/player name %v duplicated in return from All()",
				stateName)
		}

		setOfNamesFromAll[stateName] = true
	}

	// Now we check that Get(...) is consistent with each player from All().
	for _, stateFromAll := range statesFromAll {
		nameFromAll := stateFromAll.Name()
		stateFromGet, errorFromGet := playerCollection.Get(nameFromAll)
		if errorFromGet != nil {
			unitTest.Fatalf(
				testIdentifier+
					"/Get(%v) produced error %v",
				nameFromAll,
				errorFromGet)
		}

		if (stateFromGet.Name() != nameFromAll) ||
			(stateFromGet.Color() != stateFromAll.Color()) {
			unitTest.Fatalf(
				testIdentifier+
					"/State from Get(...) %v did not match state from All() %v",
				stateFromAll,
				stateFromGet)
		}
	}
}

func getStateAndAssertNoError(
	testIdentifier string,
	unitTest *testing.T,
	playerName string,
	playerCollection *player.StateCollection) player.ReadonlyState {
	playerState, errorGettingState :=
		playerCollection.Get(playerName)
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