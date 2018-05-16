package player_test

import (
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/player"
)

var colorsAvailableInTest []string = defaults.AvailableColors
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
	testReference           *testing.T
	ReturnForAll            []player.ReadonlyState
	ReturnForGet            player.ReadonlyState
	ReturnForNontestError   error
	TestErrorForAll         error
	TestErrorForGet         error
	TestErrorForAdd         error
	TestErrorForUpdateColor error
	TestErrorForReset       error
	ArgumentsForAdd         []mockPlayerState
}

func NewMockPersister(testReference *testing.T, testError error) *mockPersister {
	return &mockPersister{
		testReference:           testReference,
		ReturnForAll:            nil,
		ReturnForGet:            nil,
		ReturnForNontestError:   nil,
		TestErrorForAll:         testError,
		TestErrorForGet:         testError,
		TestErrorForAdd:         testError,
		TestErrorForUpdateColor: testError,
		TestErrorForReset:       testError,
		ArgumentsForAdd:         make([]mockPlayerState, 0),
	}
}

func (mockImplementation *mockPersister) All() []player.ReadonlyState {
	if mockImplementation.TestErrorForAll != nil {
		mockImplementation.testReference.Errorf(
			"All(): %v",
			mockImplementation.TestErrorForAll)
	}

	return mockImplementation.ReturnForAll
}

func (mockImplementation *mockPersister) Get(playerName string) (player.ReadonlyState, error) {
	if mockImplementation.TestErrorForGet != nil {
		mockImplementation.testReference.Errorf(
			"Get(%v): %v",
			playerName,
			mockImplementation.TestErrorForGet)
	}

	return mockImplementation.ReturnForGet, mockImplementation.ReturnForNontestError
}

func (mockImplementation *mockPersister) Add(playerName string, chatColor string) error {
	if mockImplementation.TestErrorForAdd != nil {
		mockImplementation.testReference.Errorf(
			"Add(%v, %v): %v",
			playerName,
			chatColor,
			mockImplementation.TestErrorForAdd)
	}

	mockImplementation.ArgumentsForAdd =
		append(mockImplementation.ArgumentsForAdd, mockPlayerState{
			mockName:  playerName,
			mockColor: chatColor,
		})

	return mockImplementation.ReturnForNontestError
}

func (mockImplementation *mockPersister) UpdateColor(playerName string, chatColor string) error {
	if mockImplementation.TestErrorForUpdateColor != nil {
		mockImplementation.testReference.Errorf(
			"UpdateColor(%v, %v): %v",
			playerName,
			chatColor,
			mockImplementation.TestErrorForUpdateColor)
	}

	return mockImplementation.ReturnForNontestError
}

func (mockImplementation *mockPersister) Reset() {
	if mockImplementation.TestErrorForReset != nil {
		mockImplementation.testReference.Errorf(
			"Reset(): %v",
			mockImplementation.TestErrorForReset)
	}
}

func prepareCollection(
	unitTest *testing.T,
	initialPlayerNames []string,
	availableColors []string,
	mockImplementation *mockPersister) (*player.StateCollection, map[string]bool) {
	// We allow the set-up to call Add(...) and All(), and then restore the settings afterwards.
	originalTestErrorForAdd := mockImplementation.TestErrorForAdd
	originalTestErrorForAll := mockImplementation.TestErrorForAll
	mockImplementation.TestErrorForAdd = nil
	mockImplementation.TestErrorForAll = nil

	stateCollection, errorFromCreation :=
		player.NewCollection(
			mockImplementation,
			initialPlayerNames,
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

	mockImplementation.TestErrorForAdd = originalTestErrorForAdd
	mockImplementation.TestErrorForAll = originalTestErrorForAll

	return stateCollection, colorSet
}

func TestConstructorAndResetBothAddCorrectly(unitTest *testing.T) {
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
		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			// Allowing Add(...) and All() is taken care of in prepareCollection(...).
			mockImplementation :=
				NewMockPersister(unitTest, fmt.Errorf("Only Add(...) and All() should be called"))
			stateCollection, validColors :=
				prepareCollection(
					unitTest,
					testCase.initialPlayerNames,
					colorsAvailableInTest,
					mockImplementation)

			assertPersisterAddCalledCorrectly(
				testCase.testName,
				unitTest,
				testCase.initialPlayerNames,
				validColors,
				mockImplementation.ArgumentsForAdd)

			// We clear the record of calls to the persister's function, and allow Reset()
			// to be called, along with Add(...) and All(), so that the initial players
			// can be restored in the persister.
			mockImplementation.ArgumentsForAdd = make([]mockPlayerState, 0)
			mockImplementation.TestErrorForReset = nil
			mockImplementation.TestErrorForAdd = nil
			mockImplementation.TestErrorForAll = nil

			stateCollection.Reset()

			assertPersisterAddCalledCorrectly(
				testCase.testName,
				unitTest,
				testCase.initialPlayerNames,
				validColors,
				mockImplementation.ArgumentsForAdd)
		})
	}
}

func TestReturnFromAllIsCorrect(unitTest *testing.T) {
	testCases := []struct {
		testName       string
		expectedReturn []player.ReadonlyState
	}{
		{
			testName:       "Nil player list",
			expectedReturn: nil,
		},
		{
			testName:       "Empty list",
			expectedReturn: []player.ReadonlyState{},
		},
		{
			testName: "Three players",
			expectedReturn: []player.ReadonlyState{
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
			NewMockPersister(unitTest, fmt.Errorf("Only All() should be called"))
		mockImplementation.TestErrorForAll = nil
		mockImplementation.ReturnForAll = testCase.expectedReturn

		expectedNumberOfPlayers := len(testCase.expectedReturn)

		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			stateCollection, _ :=
				prepareCollection(
					unitTest,
					nil,
					colorsAvailableInTest,
					mockImplementation)

			actualReturnFromAll := stateCollection.All()

			if len(actualReturnFromAll) != expectedNumberOfPlayers {
				unitTest.Errorf(
					"Number of players from All() unexpected: expected %v; actual %v",
					testCase.expectedReturn,
					actualReturnFromAll)
			}

			for playerIndex := 0; playerIndex < expectedNumberOfPlayers; playerIndex++ {
				expectedPlayer := testCase.expectedReturn[playerIndex]
				actualPlayer := actualReturnFromAll[playerIndex]

				// We did not set up any expected nil.
				if (actualPlayer == nil) ||
					(actualPlayer.Name() != expectedPlayer.Name()) ||
					(actualPlayer.Color() != expectedPlayer.Color()) {
					unitTest.Errorf(
						"Actual return from All() did not match expected in index %v: expected %v; actual %v",
						playerIndex,
						testCase.expectedReturn,
						actualReturnFromAll)
				}
			}
		})
	}
}

func TestReturnFromGetIsCorrect(unitTest *testing.T) {
	testCases := []struct {
		testName       string
		expectedReturn player.ReadonlyState
		expectedError  error
	}{
		{
			testName:       "Nil player, string error",
			expectedReturn: nil,
			expectedError:  fmt.Errorf("Expected error from Get(...)"),
		},
		{
			testName: "Valid player, nil error",
			expectedReturn: &mockPlayerState{
				mockName:  "Mock Player",
				mockColor: colorsAvailableInTest[0],
			},
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		mockImplementation :=
			NewMockPersister(unitTest, fmt.Errorf("Only Get(...) should be called"))
		mockImplementation.TestErrorForGet = nil
		mockImplementation.ReturnForGet = testCase.expectedReturn
		mockImplementation.ReturnForNontestError = testCase.expectedError

		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			stateCollection, _ :=
				prepareCollection(
					unitTest,
					nil,
					colorsAvailableInTest,
					mockImplementation)

			irrelevantPlayerName := "Does not matter for the mock"

			actualReturn, actualError :=
				stateCollection.Get(irrelevantPlayerName)

			if actualError != testCase.expectedError {
				unitTest.Errorf(
					"Unexpected error from Get(%v): expected %v; actual %v",
					irrelevantPlayerName,
					testCase.expectedError,
					actualError)
			}

			if testCase.expectedReturn == nil {
				if actualReturn != nil {
					unitTest.Errorf(
						"Unexpected player.State from Get(%v): expected nil; actual %v",
						irrelevantPlayerName,
						actualReturn)
				}
			} else {
				if actualReturn == nil {
					unitTest.Errorf(
						"Unexpected player.State from Get(%v): expected %v; actual nil",
						irrelevantPlayerName,
						testCase.expectedReturn)
				} else if (actualReturn.Name() != testCase.expectedReturn.Name()) ||
					(actualReturn.Color() != testCase.expectedReturn.Color()) {
					unitTest.Errorf(
						"Unexpected player.State from Get(%v): expected %v; actual %v",
						irrelevantPlayerName,
						testCase.expectedReturn,
						actualReturn)
				}
			}
		})
	}
}

func TestAvailableColorsIsCorrectAndFreshCopy(unitTest *testing.T) {
	mockImplementation :=
		NewMockPersister(unitTest, fmt.Errorf("No functions should be called"))

	stateCollection, validColors :=
		prepareCollection(
			unitTest,
			nil,
			colorsAvailableInTest,
			mockImplementation)

	firstColors := stateCollection.AvailableChatColors()

	assertColorsAreCorrect(
		"First slice from AvailableChatColors()",
		unitTest,
		firstColors,
		validColors)

	firstColors[0] = "not even a valid color"
	if validColors[firstColors[0]] {
		unitTest.Fatalf(
			"Somehow %v is in the valid color map %v",
			firstColors[0],
			validColors)
	}

	secondColors := stateCollection.AvailableChatColors()

	assertColorsAreCorrect(
		"Second slice from AvailableChatColors()",
		unitTest,
		secondColors,
		validColors)
}

func TestRejectAddWithEmptyPlayerName(unitTest *testing.T) {
	mockImplementation :=
		NewMockPersister(unitTest, fmt.Errorf("No functions should be called"))

	stateCollection, _ :=
		prepareCollection(
			unitTest,
			nil,
			colorsAvailableInTest,
			mockImplementation)

	actualError := stateCollection.Add("", colorsAvailableInTest[0])

	if actualError == nil {
		unitTest.Fatalf(
			"No error from Add(empty player name, chat color %v)",
			colorsAvailableInTest[0])
	}
}

func TestRejectAddWithInvalidColor(unitTest *testing.T) {
	mockImplementation :=
		NewMockPersister(unitTest, fmt.Errorf("No functions should be called"))

	stateCollection, validColors :=
		prepareCollection(
			unitTest,
			nil,
			colorsAvailableInTest,
			mockImplementation)

	playerName := "Mock Player"

	invalidColor := "not a valid color"
	if validColors[invalidColor] {
		unitTest.Fatalf(
			"Somehow %v is in the valid color map %v",
			invalidColor,
			validColors)
	}

	actualError := stateCollection.Add(playerName, invalidColor)

	if actualError == nil {
		unitTest.Fatalf(
			"No error from Add(player name %v, chat color %v)",
			playerName,
			invalidColor)
	}
}

func TestReturnErrorFromPersisterAdd(unitTest *testing.T) {
	testCases := []struct {
		testName      string
		expectedError error
	}{
		{
			testName:      "String error",
			expectedError: fmt.Errorf("Expected error from Add(...)"),
		},
		{
			testName:      "Nil error",
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		mockImplementation :=
			NewMockPersister(unitTest, fmt.Errorf("Only Add(...) and All() should be called"))
		mockImplementation.TestErrorForAdd = nil
		mockImplementation.TestErrorForAll = nil
		mockImplementation.ReturnForNontestError = testCase.expectedError

		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			stateCollection, _ :=
				prepareCollection(
					unitTest,
					nil,
					colorsAvailableInTest,
					mockImplementation)

			playerName := "Mock Player"
			chatColor := colorsAvailableInTest[0]

			actualError := stateCollection.Add(playerName, chatColor)

			if actualError != testCase.expectedError {
				unitTest.Errorf(
					"Add(player name %v, chat color %v) returned error %v - expected %v",
					playerName,
					chatColor,
					actualError,
					testCase.expectedError)
			}
		})
	}
}

func TestAddPlayerWithNoColorGetsValidColor(unitTest *testing.T) {
	mockImplementation :=
		NewMockPersister(unitTest, fmt.Errorf("Only Add(...) and All() should be called"))
	mockImplementation.TestErrorForAdd = nil
	mockImplementation.TestErrorForAll = nil

	stateCollection, validColors :=
		prepareCollection(
			unitTest,
			nil,
			colorsAvailableInTest,
			mockImplementation)

	playerName := "Mock Player"

	errorFromAdd := stateCollection.Add(playerName, "")

	if errorFromAdd != nil {
		unitTest.Fatalf(
			"Add(player name %v, empty chat color) produced unexpected error %v",
			playerName,
			errorFromAdd)
	}

	if len(mockImplementation.ArgumentsForAdd) != 1 {
		unitTest.Fatalf(
			"Add(player name %v, empty chat color) did not call the persister's add once, but with %v",
			playerName,
			mockImplementation.ArgumentsForAdd)
	}

	assignedColor := mockImplementation.ArgumentsForAdd[0].mockColor
	if !validColors[assignedColor] {
		unitTest.Fatalf(
			"Assigned color %v is not in the valid color map %v",
			assignedColor,
			validColors)
	}
}

func TestRejectUpdateWithInvalidColor(unitTest *testing.T) {
	mockImplementation :=
		NewMockPersister(unitTest, fmt.Errorf("No functions should be called"))

	stateCollection, validColors :=
		prepareCollection(
			unitTest,
			nil,
			colorsAvailableInTest,
			mockImplementation)

	playerName := "Mock Player"

	invalidColor := "not a valid color"
	if validColors[invalidColor] {
		unitTest.Fatalf(
			"Somehow %v is in the valid color map %v",
			invalidColor,
			validColors)
	}

	errorFromUpdateColor := stateCollection.UpdateColor(playerName, invalidColor)

	if errorFromUpdateColor == nil {
		unitTest.Fatalf(
			"No error from UpdateColor(player name %v, chat color %v)",
			playerName,
			invalidColor)
	}
}

func TestReturnErrorFromPersisterUpdateColor(unitTest *testing.T) {
	testCases := []struct {
		testName      string
		expectedError error
	}{
		{
			testName:      "String error",
			expectedError: fmt.Errorf("Expected error from UpdateColor(...)"),
		},
		{
			testName:      "Nil error",
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		mockImplementation :=
			NewMockPersister(unitTest, fmt.Errorf("Only UpdateColor(...) should be called"))
		mockImplementation.TestErrorForUpdateColor = nil
		mockImplementation.ReturnForNontestError = testCase.expectedError

		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			stateCollection, _ :=
				prepareCollection(
					unitTest,
					nil,
					colorsAvailableInTest,
					mockImplementation)

			playerName := "Mock Player"
			chatColor := colorsAvailableInTest[0]

			actualError := stateCollection.UpdateColor(playerName, chatColor)

			if actualError != testCase.expectedError {
				unitTest.Errorf(
					"UpdateColor(player name %v, chat color %v) returned error %v - expected %v",
					playerName,
					chatColor,
					actualError,
					testCase.expectedError)
			}
		})
	}
}

func assertPersisterAddCalledCorrectly(
	testIdentifier string,
	unitTest *testing.T,
	initialPlayerNames []string,
	validColors map[string]bool,
	argumentsForPersisterAdd []mockPlayerState) {
	numberOfAddedPlayers := len(argumentsForPersisterAdd)

	if numberOfAddedPlayers != len(initialPlayerNames) {
		unitTest.Errorf(
			"Number of initial players (expected %v) did not match number of players added (added %v)",
			initialPlayerNames,
			argumentsForPersisterAdd)
	}

	for _, initialPlayerName := range initialPlayerNames {
		numberOfAdds := 0

		for _, argumentsFromSingleAdd := range argumentsForPersisterAdd {
			if argumentsFromSingleAdd.mockName == initialPlayerName {
				numberOfAdds++

				if !validColors[argumentsFromSingleAdd.mockColor] {
					unitTest.Errorf(
						"Add(...) for player %v had invalid color %v (valid colors are %v)",
						initialPlayerName,
						argumentsFromSingleAdd.mockColor,
						validColors)
				}
			}
		}

		if numberOfAdds != 1 {
			unitTest.Errorf(
				"Wrong number of Add(...) arguments for player name %v - expected 1, arguments slice %v",
				initialPlayerName,
				argumentsForPersisterAdd)
		}
	}
}

func assertColorsAreCorrect(
	testIdentifier string,
	unitTest *testing.T,
	actualColors []string,
	validColors map[string]bool) {
	if len(actualColors) != len(validColors) {
		unitTest.Fatalf(
			testIdentifier+"/actual colors %v had wrong length, expected %v",
			actualColors,
			colorsAvailableInTest)
	}

	for _, actualColor := range actualColors {
		if !validColors[actualColor] {
			unitTest.Fatalf(
				testIdentifier+"/actual colors %v had unexpected color %v, expected %v",
				actualColors,
				actualColor,
				colorsAvailableInTest)
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
