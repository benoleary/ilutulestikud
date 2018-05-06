package player_test

import (
	"fmt"
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

type mockPlayerState struct {
	mockName  string
	mockColor string
}

func (mockState *mockPlayerState) Name() string {
	return mockState.mockName
}

func (mockState *mockPlayerState) Color() string {
	return mockState.mockColor
}

type mockPersister struct {
	TestReference           *testing.T
	ReturnForAll            []player.ReadonlyState
	ReturnForGet            player.ReadonlyState
	ReturnForNontestError   error
	TestErrorForAll         error
	TestErrorForGet         error
	TestErrorForAdd         error
	TestErrorForUpdateColor error
	TestErrorForReset       error
	ArgumentsForAdd         map[string][]string
}

func NewMockFullTestError(testReference *testing.T, testError error) *mockPersister {
	return &mockPersister{
		TestReference:           testReference,
		ReturnForAll:            nil,
		ReturnForGet:            nil,
		ReturnForNontestError:   nil,
		TestErrorForAll:         testError,
		TestErrorForGet:         testError,
		TestErrorForAdd:         testError,
		TestErrorForUpdateColor: testError,
		TestErrorForReset:       testError,
		ArgumentsForAdd:         make(map[string][]string, 0),
	}
}

func (mockImplementation *mockPersister) All() []player.ReadonlyState {
	if mockImplementation.TestErrorForAll != nil {
		mockImplementation.TestReference.Errorf("%v", mockImplementation.TestErrorForAll)
	}

	return mockImplementation.ReturnForAll
}

func (mockImplementation *mockPersister) Get(playerName string) (player.ReadonlyState, error) {
	if mockImplementation.TestErrorForGet != nil {
		mockImplementation.TestReference.Errorf("%v", mockImplementation.TestErrorForGet)
	}

	return mockImplementation.ReturnForGet, mockImplementation.ReturnForNontestError
}

func (mockImplementation *mockPersister) Add(playerName string, chatColor string) error {
	if mockImplementation.TestErrorForAdd != nil {
		mockImplementation.TestReference.Errorf("%v", mockImplementation.TestErrorForAdd)
	}

	existingColors := mockImplementation.ArgumentsForAdd[playerName]
	mockImplementation.ArgumentsForAdd[playerName] = append(existingColors, chatColor)

	return mockImplementation.ReturnForNontestError
}

func (mockImplementation *mockPersister) UpdateColor(playerName string, chatColor string) error {
	if mockImplementation.TestErrorForUpdateColor != nil {
		mockImplementation.TestReference.Errorf("%v", mockImplementation.TestErrorForUpdateColor)
	}

	return mockImplementation.ReturnForNontestError
}

func (mockImplementation *mockPersister) Reset() {
	if mockImplementation.TestErrorForReset != nil {
		mockImplementation.TestReference.Errorf("%v", mockImplementation.TestErrorForReset)
	}
}

func prepareCollection(
	unitTest *testing.T,
	initialPlayerNames []string,
	availableColors []string,
	mockImplementation *mockPersister) (*player.StateCollection, map[string]bool) {
	stateCollection, errorFromCreation :=
		player.NewCollection(
			mockImplementation,
			defaultTestPlayerNames,
			colorsAvailableInTest)

	if errorFromCreation != nil {
		unitTest.Fatalf(
			"Error when preparing collection: %v",
			errorFromCreation)
	}

	colorSet := make(map[string]bool, 0)
	for _, availableColor := range availableColors {
		colorSet[availableColor] = true
	}

	return stateCollection, colorSet
}

func TestConstructorAddsCorrectly(unitTest *testing.T) {
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
		mockImplementation := &mockPersister{}

		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			stateCollection, validColors :=
				prepareCollection(
					unitTest,
					testCase.initialPlayerNames,
					colorsAvailableInTest,
					mockImplementation)

			numberOfAddedPlayers := len(mockImplementation.ArgumentsForAdd)

			if numberOfAddedPlayers != len(testCase.initialPlayerNames) {
				unitTest.Errorf(
					"Number of initial players (expected %v) did not match number of players added (added %v)",
					testCase.initialPlayerNames,
					mockImplementation.ArgumentsForAdd)
			}

			for _, initialPlayerName := range testCase.initialPlayerNames {
				addArguments, hasAddedArguments :=
					mockImplementation.ArgumentsForAdd[initialPlayerName]

				if !hasAddedArguments {
					unitTest.Errorf(
						"No Add arguments for player name %v",
						initialPlayerName)
				}

				if len(addArguments) != 1 {
					unitTest.Errorf(
						"Wrong number of Add arguments for player name %v - expected 1, arguments slice %v",
						initialPlayerName,
						addArguments)
				}

				colorOfAdd := addArguments[0]
				if !validColors[colorOfAdd] {
					unitTest.Errorf(
						"Add for player %v had invalid color %v (valid colors are %v)",
						initialPlayerName,
						colorOfAdd,
						validColors)
				}
			}
		})
	}
}

func TestReturnFromAllIsCorrect(unitTest *testing.T) {
	testCases := []struct {
		testName              string
		expectedReturnFromAll []player.ReadonlyState
	}{
		{
			testName:              "Nil player list",
			expectedReturnFromAll: nil,
		},
		{
			testName:              "Empty list",
			expectedReturnFromAll: []player.ReadonlyState{},
		},
		{
			testName: "Three players",
			expectedReturnFromAll: []player.ReadonlyState{
				&mockPlayerState{
					mockName:  "Mock Player One",
					mockColor: colorsAvailableInTest[0],
				},
				&mockPlayerState{
					mockName:  "Mock Player Two",
					mockColor: colorsAvailableInTest[1],
				},
				&mockPlayerState{
					mockName:  "Mock Player Three",
					mockColor: colorsAvailableInTest[0], // Same as Mock Player One
				},
			},
		},
	}

	for _, testCase := range testCases {
		mockImplementation :=
			NewMockFullTestError(unitTest, fmt.Errorf("Only All should be called"))
		mockImplementation.TestErrorForAll = nil
		mockImplementation.ReturnForAll = testCase.expectedReturnFromAll

		expectedNumberOfPlayers := len(testCase.expectedReturnFromAll)

		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			stateCollection, validColors :=
				prepareCollection(
					unitTest,
					nil,
					colorsAvailableInTest,
					mockImplementation)

			actualReturnFromAll := stateCollection.All()

			if len(actualReturnFromAll) != expectedNumberOfPlayers {
				unitTest.Errorf(
					"Number of players from All unexpected: expected %v; actual %v",
					testCase.expectedReturnFromAll,
					actualReturnFromAll)
			}

			for playerIndex := 0; playerIndex < expectedNumberOfPlayers; playerIndex++ {
				expectedPlayer := testCase.expectedReturnFromAll[playerIndex]
				actualPlayer := actualReturnFromAll[playerIndex]

				// We did not set up any expected nil.
				if (actualPlayer == nil) ||
					(actualPlayer.Name() != expectedPlayer.Name()) ||
					(actualPlayer.Color() != expectedPlayer.Color()) {
					unitTest.Errorf(
						"Actual return from All did not match expected in index %v: expected %v; actual %v",
						playerIndex,
						testCase.expectedReturnFromAll,
						actualReturnFromAll)
				}
			}
		})
	}
}

func TestReturnFromGetIsCorrect(unitTest *testing.T) {
	testCases := []struct {
		testName              string
		expectedReturnFromGet player.ReadonlyState
		expectedErrorFromGet  error
	}{
		{
			testName:              "Nil player, string error",
			expectedReturnFromGet: nil,
			expectedErrorFromGet:  nil,
		},
		{
			testName: "Valid player, nil error",
			expectedReturnFromGet: &mockPlayerState{
				mockName:  "Mock Player",
				mockColor: colorsAvailableInTest[0],
			},
			expectedErrorFromGet: nil,
		},
	}

	for _, testCase := range testCases {
		mockImplementation :=
			NewMockFullTestError(unitTest, fmt.Errorf("Only Get should be called"))
		mockImplementation.TestErrorForGet = testCase.expectedErrorFromGet
		mockImplementation.ReturnForGet = testCase.expectedReturnFromGet

		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			stateCollection, validColors :=
				prepareCollection(
					unitTest,
					nil,
					colorsAvailableInTest,
					mockImplementation)

			actualReturnFromGet, actualErrorFromGet :=
				stateCollection.Get("Does not matter for the mock")

			if actualErrorFromGet != testCase.expectedErrorFromGet {
				unitTest.Errorf(
					"Unexpected error from Get: expected %v; actual %v",
					testCase.expectedErrorFromGet,
					actualErrorFromGet)
			}

			if testCase.expectedReturnFromGet == nil {
				if actualReturnFromGet != nil {
					unitTest.Errorf(
						"Unexpected player.State from Get: expected nil; actual %v",
						actualReturnFromGet)
				}
			} else {
				if actualReturnFromGet == nil {
					unitTest.Errorf(
						"Unexpected player.State from Get: expected %v; actual nil",
						testCase.expectedReturnFromGet)
				} else if (actualReturnFromGet.Name() != testCase.expectedReturnFromGet.Name()) ||
					(actualReturnFromGet.Color() != testCase.expectedReturnFromGet.Color()) {
					unitTest.Errorf(
						"Unexpected player.State from Get: expected %v; actual %v",
						testCase.expectedReturnFromGet,
						actualReturnFromGet)
				}
			}
		})
	}
}

func TestErrorIfEmptyAvailableColors(unitTest *testing.T) {
	stateCollection, errorFromCreation :=
		player.NewCollection(
			persister.NewInMemoryPersister(),
			defaultTestPlayerNames,
			[]string{})

	if errorFromCreation == nil {
		unitTest.Fatalf(
			"No error when preparing collection with empty list of colors, returned %v",
			stateCollection)
	}
}

func TestNonemptyAvailableColors(unitTest *testing.T) {
	stateCollection := prepareCollection(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

	availableColors := stateCollection.AvailableChatColors()

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
}

func TestRejectNewPlayerWithNoName(unitTest *testing.T) {
	stateCollection := prepareCollection(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

	playerName := ""
	chatColor := colorsAvailableInTest[0]

	testIdentifier := "Reject Add(player with no name)"

	errorFromAdd := stateCollection.Add(playerName, chatColor)

	// We check that the collection still produces valid states.
	assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
		testIdentifier,
		unitTest,
		defaultTestPlayerNames,
		colorsAvailableInTest,
		stateCollection)

	// If there was no error, then something went wrong.
	if errorFromAdd == nil {
		unitTest.Fatalf(
			"Add(%v, %v) did not produce an error",
			playerName,
			chatColor)
	}

	// We check that the player was not added.
	playerState, errorFromGet :=
		stateCollection.Get(playerName)

	// If there was no error, then something went wrong.
	if errorFromGet == nil {
		unitTest.Fatalf(
			"Get(%v) did not produce an error, instead retrieved %v",
			playerName,
			playerState)
	}
}

func TestAddNewPlayerWithInvalidColor(unitTest *testing.T) {
	stateCollection := prepareCollection(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

	playerName := "A. New Player"
	invalidColor := "Not a valid color"

	testIdentifier := "Reject Add(player with invalid color)"

	errorFromAdd := stateCollection.Add(playerName, invalidColor)

	// We check that the collection still produces valid states.
	assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
		testIdentifier,
		unitTest,
		defaultTestPlayerNames,
		colorsAvailableInTest,
		stateCollection)

	// If there was no error, then something went wrong.
	if errorFromAdd == nil {
		unitTest.Fatalf(
			"Add(%v, %v) did not produce an error",
			playerName,
			invalidColor)
	}

	// We check that the player was not added.
	playerState, errorFromGet := stateCollection.Get(playerName)

	// If there was no error, then something went wrong.
	if errorFromGet == nil {
		unitTest.Fatalf(
			"Get(%v) did not produce an error, instead retrieved %v",
			playerName,
			playerState)
	}
}

func TestRejectAddPlayerWithExistingName(unitTest *testing.T) {
	stateCollection := prepareCollection(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

	for _, playerName := range defaultTestPlayerNames {
		testIdentifier := "Reject Add(player with existing name)"

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			initialState := getStateAndAssertNoError(
				testIdentifier+"/Get(initial player)",
				unitTest,
				playerName,
				stateCollection)

			errorFromAddWithNoColor := stateCollection.Add(playerName, "")

			// We check that the collection still produces valid states.
			assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
				testIdentifier,
				unitTest,
				defaultTestPlayerNames,
				colorsAvailableInTest,
				stateCollection)

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
				stateCollection)

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

			errorFromAddWithNewColor := stateCollection.Add(playerName, newColor)

			// We check that the collection still produces valid states.
			assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
				testIdentifier,
				unitTest,
				defaultTestPlayerNames,
				colorsAvailableInTest,
				stateCollection)

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
				stateCollection)

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

func TestAddPlayerWithValidColorAndTestGet(unitTest *testing.T) {
	stateCollection := prepareCollection(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

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

	for _, testCase := range testCases {
		testIdentifier := "Add(" + testCase.playerName + ", with valid color) and Get(same player)"

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			errorFromAdd :=
				stateCollection.Add(testCase.playerName, chatColor)

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
				stateCollection)

			// We check that the player can be retrieved.
			newState := getStateAndAssertNoError(
				testIdentifier+"/Retrieve with Get(...)",
				unitTest,
				testCase.playerName,
				stateCollection)

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

func TestAddPlayerWithNoColorAndTestGetHasValidColor(unitTest *testing.T) {
	stateCollection := prepareCollection(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

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

	for _, testCase := range testCases {
		testIdentifier := "Add(" + testCase.playerName + ", with no color) and Get(same player)"

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			errorFromAdd := stateCollection.Add(testCase.playerName, "")

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
				stateCollection)

			// We check that the player can be retrieved.
			newState := getStateAndAssertNoError(
				testIdentifier+"/Get(newly-added player)",
				unitTest,
				testCase.playerName,
				stateCollection)

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

func TestRejectUpdateInvalidPlayer(unitTest *testing.T) {
	stateCollection := prepareCollection(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

	playerName := "Not A. Participant"
	chatColor := colorsAvailableInTest[0]

	testIdentifier := "UpdateColor(valid player, invalid color)"

	errorFromUpdate := stateCollection.UpdateColor(playerName, chatColor)

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
		stateCollection)

	// We check that the player was not added.
	playerState, errorFromGet :=
		stateCollection.Get(playerName)

	// If there was no error, then something went wrong.
	if errorFromGet == nil {
		unitTest.Fatalf(
			"Get(%v) did not produce an error, instead retrieved %v",
			playerName,
			playerState)
	}
}

func TestRejectUpdatePlayerWithInvalidColor(unitTest *testing.T) {
	stateCollection := prepareCollection(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

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

	for _, testCase := range testCases {
		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			// We save the state to later check that the color was not changed.
			initialState := getStateAndAssertNoError(
				testCase.testName+"/Get(valid player) before update",
				unitTest,
				testCase.playerName,
				stateCollection)

			errorFromUpdate :=
				stateCollection.UpdateColor(testCase.playerName, testCase.chatColor)

			if errorFromUpdate == nil {
				unitTest.Fatalf(
					"UpdateColor(%v, %v) did not produce an error",
					testCase.playerName,
					testCase.chatColor)
			}

			// We check that the collection still produces valid states.
			assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
				testCase.testName,
				unitTest,
				defaultTestPlayerNames,
				colorsAvailableInTest,
				stateCollection)

			// We check that the player can be retrieved.
			stateAfterUpdate := getStateAndAssertNoError(
				testCase.testName+"/Get(same player) after update",
				unitTest,
				testCase.playerName,
				stateCollection)

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

func TestUpdateAllPlayersToFirstColor(unitTest *testing.T) {
	stateCollection := prepareCollection(unitTest, defaultTestPlayerNames, colorsAvailableInTest)

	firstColor := colorsAvailableInTest[0]

	for _, playerName := range defaultTestPlayerNames {
		testIdentifier := "Update player to first color"

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			errorFromAddWithNoColor := stateCollection.UpdateColor(playerName, firstColor)

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
				stateCollection)

			// We check that the player has the correct color.
			updatedState := getStateAndAssertNoError(
				testIdentifier+"/Get(updated player)",
				unitTest,
				playerName,
				stateCollection)

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
		stateCollection := prepareCollection(unitTest, defaultTestPlayerNames, colorsAvailableInTest)
		testIdentifier :=
			testCase.testName

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			if testCase.shouldAddBeforeReset {
				errorFromAdd := stateCollection.Add(playerNameToAdd, chatColorForAdd)

				if errorFromAdd != nil {
					unitTest.Fatalf(
						"Add(%v, %v) produced an error: %v",
						playerNameToAdd,
						chatColorForAdd,
						errorFromAdd)
				}
			}

			if testCase.shouldUpdateBeforeReset {
				errorFromUpdate := stateCollection.UpdateColor(playerNameToUpdate, chatColorForUpdate)
				if errorFromUpdate != nil {
					unitTest.Fatalf(
						"UpdateColor(%v, %v) produced an error: %v",
						playerNameToUpdate,
						chatColorForUpdate,
						errorFromUpdate)
				}
			}

			// Now we can reset.
			stateCollection.Reset()

			// We check that the collection still produces valid states.
			assertPlayerNamesAreCorrectAndColorsAreValidAndGetIsConsistentWithAll(
				testIdentifier,
				unitTest,
				defaultTestPlayerNames,
				colorsAvailableInTest,
				stateCollection)

			// We check that if a player had been added, it is no longer retrievable.
			addedState, errorFromGet :=
				stateCollection.Get(playerNameToAdd)

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
