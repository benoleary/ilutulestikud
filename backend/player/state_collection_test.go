package player_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/player"
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
	initialPlayerNames []string,
	availableColors []string) []collectionAndDescription {
	statePersisters := []persisterAndDescription{
		persisterAndDescription{
			PlayerPersister:      player.NewInMemoryPersister(),
			PersisterDescription: "in-memory persister",
		},
	}

	numberOfPersisters := len(statePersisters)

	stateCollections := make([]collectionAndDescription, numberOfPersisters)

	for persisterIndex := 0; persisterIndex < numberOfPersisters; persisterIndex++ {
		statePersister := statePersisters[persisterIndex]
		stateCollection :=
			player.NewCollection(
				statePersister.PlayerPersister,
				initialPlayerNames,
				availableColors)
		stateCollections[persisterIndex] = collectionAndDescription{
			PlayerCollection:      stateCollection,
			CollectionDescription: "collection around " + statePersister.PersisterDescription,
		}
	}

	return stateCollections
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
			prepareCollections(defaultTestPlayerNames, colorsAvailableInTest)

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
		prepareCollections(defaultTestPlayerNames, colorsAvailableInTest)

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

func TestEmptyAvailableColors(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(defaultTestPlayerNames, []string{})

	for _, collectionType := range collectionTypes {
		testIdentifier := "Empty available colors/" + collectionType.CollectionDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			availableColors := collectionType.PlayerCollection.AvailableChatColors()

			if len(availableColors) != 0 {
				unitTest.Fatalf(
					"AvailableChatColors() when set up with empty list returned non-empty list %v",
					availableColors)
			}
		})
	}
}

func TestNonemptyAvailableColors(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(defaultTestPlayerNames, colorsAvailableInTest)

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

			for colorIndex := 0; colorIndex < numberOfExpectedColors; colorIndex++ {
				if availableColors[colorIndex] != colorsAvailableInTest[colorIndex] {
					unitTest.Fatalf(
						"AvailableChatColors() set up with %v returned list %v which did not match in element %v",
						colorsAvailableInTest,
						availableColors,
						colorIndex)
				}
			}
		})
	}
}

func TestRejectNewPlayerWithNoName(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(defaultTestPlayerNames, colorsAvailableInTest)
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
					"Get(%v) did not produce an error",
					playerName)
			}
		})
	}
}

func TestAddNewPlayerWithInvalidColor(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(defaultTestPlayerNames, colorsAvailableInTest)
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
					"Get(%v) did not produce an error",
					playerName)
			}
		})
	}
}

func TestRejectAddPlayerWithExistingName(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(defaultTestPlayerNames, colorsAvailableInTest)

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
				if errorFromAddWithNewColor != nil {
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
		prepareCollections(defaultTestPlayerNames, colorsAvailableInTest)

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

				newStateHasValidColor := mapStringsToTrue(colorsAvailableInTest)[newState.Color()]
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
		prepareCollections(defaultTestPlayerNames, colorsAvailableInTest)

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
		prepareCollections(defaultTestPlayerNames, colorsAvailableInTest)

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
					"Get(%v) did not produce an error",
					playerName)
			}
		})
	}
}

func TestRejectUpdatePlayerWithInvalidColor(unitTest *testing.T) {
	collectionTypes :=
		prepareCollections(defaultTestPlayerNames, colorsAvailableInTest)

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
		prepareCollections(defaultTestPlayerNames, colorsAvailableInTest)

	firstColor := colorsAvailableInTest[0]

	for _, collectionType := range collectionTypes {
		for _, playerName := range defaultTestPlayerNames {
			testIdentifier :=
				"Update player to first color/" + collectionType.CollectionDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				errorFromAddWithNoColor := collectionType.PlayerCollection.UpdateColor(playerName, firstColor)

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

func assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
	testIdentifier string,
	unitTest *testing.T,
	playerNames []string,
	validColors []string,
	playerCollection *player.StateCollection) {
	// First we set up a map of valid colors, ignoring possible duplication.
	validColorMap := mapStringsToTrue(validColors)

	unitTest.Fatalf(testIdentifier + "/no valid colors provided to check against player states")

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

// just copy-paste dumps of old player.getAndPostHandler tests, need to be adapted.

func TestResetPlayers(unitTest *testing.T) {
	initialPlayers := []string{"Initial One", "Initial Two"}
	newPlayer := "New Player"

	type testArguments struct {
		shouldUpdate   bool
		shouldRegister bool
	}

	testCases := []struct {
		name      string
		arguments testArguments
	}{
		{
			name: "Reset on initial",
			arguments: testArguments{
				shouldUpdate:   false,
				shouldRegister: false,
			},
		},
		{
			name: "Reset after update of initial player",
			arguments: testArguments{
				shouldUpdate:   true,
				shouldRegister: false,
			},
		},
		{
			name: "Reset after registration of new player",
			arguments: testArguments{
				shouldUpdate:   false,
				shouldRegister: true,
			},
		},
		{
			name: "Reset after update of initial player and registration of new player",
			arguments: testArguments{
				shouldUpdate:   true,
				shouldRegister: true,
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			playerPersister := player.NewInMemoryPersister(&endpoint.Base64NameEncoder{})
			playerCollection :=
				player.NewCollection(
					playerPersister,
					initialPlayers,
					colorsAvailableInTest)
			playerGetAndPostHandler := player.NewGetAndPostHandler(playerCollection)

			initialPlayerStates := playerCollection.All()
			expectedPlayerNames := make(map[string]bool, 2)
			foundOne := false
			foundTwo := false
			identifierOne := ""
			colorOne := ""
			for _, initialPlayerState := range initialPlayerStates {
				expectedPlayerNames[initialPlayerState.Name()] = true

				if initialPlayerState.Name() == initialPlayers[0] {
					foundOne = true
					identifierOne = initialPlayerState.Identifier()
					colorOne = initialPlayerState.Color()
				} else if initialPlayerState.Name() == initialPlayers[1] {
					foundTwo = true
				}
			}

			if !foundOne {
				unitTest.Fatalf(
					"Initial player %v could not be found internally",
					initialPlayers[0])
			}

			if !foundTwo {
				unitTest.Fatalf(
					"Initial player %v could not be found internally",
					initialPlayers[1])
			}

			if testCase.arguments.shouldUpdate {
				// We update the first player to have a different color from the list.
				if colorOne == colorsAvailableInTest[0] {
					colorOne = colorsAvailableInTest[1]
				} else {
					colorOne = colorsAvailableInTest[0]
				}

				updateBytesBuffer := new(bytes.Buffer)
				json.NewEncoder(updateBytesBuffer).Encode(endpoint.PlayerState{
					Identifier: identifierOne,
					Color:      colorOne,
				})

				// Now we update the player.
				_, postCode :=
					playerGetAndPostHandler.HandlePost(
						json.NewDecoder(updateBytesBuffer),
						[]string{"update-player"})

				if postCode != http.StatusOK {
					unitTest.Fatalf(
						"POST update-player did not return expected HTTP code %v, instead was %v.",
						http.StatusOK,
						postCode)
				}
			}

			if testCase.arguments.shouldRegister {
				registrationBytesBuffer := new(bytes.Buffer)
				json.NewEncoder(registrationBytesBuffer).Encode(endpoint.PlayerState{
					Name:  newPlayer,
					Color: colorsAvailableInTest[0],
				})

				// Now we add the player.
				_, postCode :=
					playerGetAndPostHandler.HandlePost(
						json.NewDecoder(registrationBytesBuffer),
						[]string{"new-player"})

				if postCode != http.StatusOK {
					unitTest.Fatalf(
						"POST new-player did not return expected HTTP code %v, instead was %v.",
						http.StatusOK,
						postCode)
				}
			}

			// Now that the system has been set up, we reset it.
			resetInterface, resetCode :=
				playerGetAndPostHandler.HandlePost(nil, []string{"reset-players"})

			// Then we check that the POST returned a valid response.
			resetResponseList := assertAtLeastOnePlayerReturnedInList(
				unitTest,
				resetCode,
				resetInterface,
				"POST reset-players")

			// Before we check that only initial players are returned, we check that each
			// initial player is present and as expected.
			for _, expectedPlayerName := range initialPlayers {
				assertPlayerIsCorrectExternallyAndInternally(
					unitTest,
					playerCollection,
					playerGetAndPostHandler,
					expectedPlayerName,
					"",
					"Reset player "+expectedPlayerName)
			}

			getInterface, getCode :=
				playerGetAndPostHandler.HandleGet([]string{"registered-players"})

			getListAfterReset := assertAtLeastOnePlayerReturnedInList(
				unitTest,
				getCode,
				getInterface,
				"GET registered-players after reset")

			// We check that the response to the reset POST and the response to the GET
			// afterwards contain exclusively the initial players.
			for _, playerList := range []endpoint.PlayerList{resetResponseList, getListAfterReset} {
				for _, playerState := range playerList.Players {
					if !expectedPlayerNames[playerState.Name] {
						unitTest.Fatalf(
							"Found player %v after reset, when initial players are %v.",
							playerState.Name,
							expectedPlayerNames)
					}
				}
			}
		})
	}
}
