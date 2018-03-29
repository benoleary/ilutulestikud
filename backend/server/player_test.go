package server_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/player"
	"github.com/benoleary/ilutulestikud/backend/server"
)

var colorsAvailableInTest []string = defaults.AvailableColors()

type functionNameAndArgument struct {
	FunctionName     string
	FunctionArgument interface{}
}

type mockPlayerCollection struct {
	FunctionsAndArgumentsReceived           []functionNameAndArgument
	ErrorToReturn                           error
	ReturnForAdd                            string
	ReturnForGet                            player.ReadonlyState
	ReturnForRegisteredPlayersForEndpoint   endpoint.PlayerList
	ReturnForAvailableChatColorsForEndpoint endpoint.ChatColorList
}

func (mockCollection *mockPlayerCollection) recordFunctionAndArgument(
	functionName string, functionArgument interface{}) {
	mockCollection.FunctionsAndArgumentsReceived =
		append(mockCollection.FunctionsAndArgumentsReceived, functionNameAndArgument{
			FunctionName:     functionName,
			FunctionArgument: functionArgument,
		})
}

func (mockCollection *mockPlayerCollection) clearFunctionsAndArguments() {
	mockCollection.FunctionsAndArgumentsReceived = make([]functionNameAndArgument, 0)
}

func (mockCollection *mockPlayerCollection) popSingleAndEnsureClear(
	unitTest *testing.T,
	testIdentifier string) functionNameAndArgument {
	if len(mockCollection.FunctionsAndArgumentsReceived) != 1 {
		unitTest.Fatalf(
			testIdentifier+"/mock player collection recorded %v function calls, expected 1.",
			mockCollection.FunctionsAndArgumentsReceived)
	}

	nameAndArgument := mockCollection.FunctionsAndArgumentsReceived[0]
	mockCollection.clearFunctionsAndArguments()

	return nameAndArgument
}

// Add gets mocked.
func (mockCollection *mockPlayerCollection) Add(
	playerInformation endpoint.PlayerState) (string, error) {
	mockCollection.recordFunctionAndArgument("Add", playerInformation)
	return mockCollection.ReturnForAdd, mockCollection.ErrorToReturn
}

// UpdateFromPresentAttributes gets mocked.
func (mockCollection *mockPlayerCollection) UpdateFromPresentAttributes(
	updaterReference endpoint.PlayerState) error {
	mockCollection.recordFunctionAndArgument("UpdateFromPresentAttributes", updaterReference)
	return mockCollection.ErrorToReturn
}

// Get gets mocked.
func (mockCollection *mockPlayerCollection) Get(playerIdentifier string) (player.ReadonlyState, error) {
	mockCollection.recordFunctionAndArgument("playerIdentifier", playerIdentifier)
	return mockCollection.ReturnForGet, mockCollection.ErrorToReturn
}

// Reset gets mocked.
func (mockCollection *mockPlayerCollection) Reset() {
	mockCollection.recordFunctionAndArgument("Reset", nil)
}

// RegisteredPlayersForEndpoint gets mocked.
func (mockCollection *mockPlayerCollection) RegisteredPlayersForEndpoint() endpoint.PlayerList {
	mockCollection.recordFunctionAndArgument("RegisteredPlayersForEndpoint", nil)
	return mockCollection.ReturnForRegisteredPlayersForEndpoint
}

// AvailableChatColorsForEndpoint gets mocked.
func (mockCollection *mockPlayerCollection) AvailableChatColorsForEndpoint() endpoint.ChatColorList {
	mockCollection.recordFunctionAndArgument("AvailableChatColorsForEndpoint", nil)
	return mockCollection.ReturnForAvailableChatColorsForEndpoint
}

// newServerForIdentifier prepares a server.State in a consistent way for the
// tests of the player endpoints.
func newServerForIdentifier(
	nameToIdentifier endpoint.NameToIdentifier) (*mockPlayerCollection, *server.State) {
	playerPersister := player.NewInMemoryPersister(nameToIdentifier)
	playerCollection := &mockPlayerCollection{}

	serverState :=
		server.New("test",
			playerCollection,
			nil)

	return playerCollection, serverState
}

// newServer prepares a server.State in a consistent way for the tests of the
// player endpoints.
func newServer() (*mockPlayerCollection, *server.State) {
	return newServerForIdentifier(&endpoint.Base32NameEncoder{})
}

func TestGetNoSegmentBadRequest(unitTest *testing.T) {
	_, testServer := newServer()

	getResponse := server.MockGet(testServer, "/backend/player")

	if getResponse.Code != http.StatusBadRequest {
		unitTest.Fatalf(
			"GET with empty list of relevant segments did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			getResponse.Code)
	}
}

func TestGetInvalidSegmentNotFound(unitTest *testing.T) {
	_, testServer := newServer()

	getResponse := server.MockGet(testServer, "/backend/player/invalid-segment")

	if getResponse.Code != http.StatusNotFound {
		unitTest.Fatalf(
			"GET invalid-segment did not return expected HTTP code %v, instead was %v.",
			http.StatusNotFound,
			getResponse.Code)
	}
}

func TestPostNoSegmentBadRequest(unitTest *testing.T) {
	_, testServer := newServer()

	bodyObject := endpoint.PlayerState{
		Name:  "Player Name",
		Color: "Chat color",
	}

	postResponse, encodingError :=
		server.MockPost(testServer, "/backend/player", bodyObject)
	assertPostResponseCorrect(
		unitTest,
		"POST with empty list of relevant segments",
		postResponse,
		encodingError,
		http.StatusBadRequest)
}

func TestPostInvalidSegmentNotFound(unitTest *testing.T) {
	_, testServer := newServer()

	bodyObject := endpoint.PlayerState{
		Name:  "Player Name",
		Color: "Chat color",
	}

	postResponse, encodingError :=
		server.MockPost(testServer, "/backend/player/invalid-segment", bodyObject)
	assertPostResponseCorrect(
		unitTest,
		"POST invalid-segment",
		postResponse,
		encodingError,
		http.StatusNotFound)
}

func TestDefaultPlayerListNotEmpty(unitTest *testing.T) {
	playerCollection, testServer := newServer()

	playerCollection.ReturnForRegisteredPlayersForEndpoint = endpoint.PlayerList{
		Players: []endpoint.PlayerState{
			endpoint.PlayerState{
				Name: "Player One",
			},
			endpoint.PlayerState{
				Name: "Player Two",
			},
			endpoint.PlayerState{
				Name: "Player Three",
			},
		},
	}

	getResponse := server.MockGet(testServer, "/backend/player/registered-players")

	assertAtLeastOnePlayerReturnedInList(
		unitTest,
		getResponse,
		"GET registered-players")
}

func TestAvailableColorListNotEmpty(unitTest *testing.T) {
	playerCollection, testServer := newServer()

	playerCollection.ReturnForAvailableChatColorsForEndpoint = endpoint.ChatColorList{
		Colors: []string{
			"red",
			"green",
			"blue",
		},
	}

	getResponse := server.MockGet(testServer, "/backend/player/available-colors")

	if getResponse.Code != http.StatusOK {
		unitTest.Fatalf(
			"GET available-colors did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			getResponse.Code)
	}

	bodyDecoder := json.NewDecoder(getResponse.Body)

	var responseColorList endpoint.ChatColorList
	parsingError := bodyDecoder.Decode(&responseColorList)
	if parsingError != nil {
		unitTest.Fatalf(
			"Error parsing JSON from HTTP response body: %v",
			parsingError)
	}

	if responseColorList.Colors == nil {
		unitTest.Fatalf(
			"GET available-colors returned %v which has a nil list of colors.",
			responseColorList)
	}

	if len(responseColorList.Colors) <= 0 {
		unitTest.Fatalf(
			"GET available-colors returned %v which has an empty list of colors.",
			responseColorList)
	}
}

func TestRejectInvalidNewPlayer(unitTest *testing.T) {
	playerCollection, testServer := newServer()

	bodyObject :=
		endpoint.ChatColorList{
			Colors: []string{
				"Player 1",
				"Player 2"},
		}

	playerCollection.ErrorToReturn = errors.New("error")

	postResponse, encodingError :=
		server.MockPost(testServer, "/backend/player/new-player", bodyObject)

	unitTest.Logf(
		"POST new-player with invalid JSON %v generated encoding error %v.",
		bodyObject,
		encodingError)

	if postResponse.Code != http.StatusBadRequest {
		unitTest.Fatalf(
			"POST new-player with invalid JSON %v did not return expected HTTP code %v, instead was %v.",
			bodyObject,
			http.StatusBadRequest,
			postResponse.Code)
	}
}

func TestRejectNewPlayerWithNameWhichBreaksEncoding(unitTest *testing.T) {
	// We need a special kind of name and encoding.
	playerName := server.BreaksBase64
	playerCollection, testServer := newServerForIdentifier(&endpoint.Base64NameEncoder{})

	bodyObject := endpoint.PlayerState{
		Name:  playerName,
		Color: "First color",
	}

	postResponse, encodingError :=
		server.MockPost(testServer, "/backend/player/new-player", bodyObject)
	assertPostResponseCorrect(
		unitTest,
		"POST new-player with encoding-breaking player name",
		postResponse,
		encodingError,
		http.StatusBadRequest)
}

func TestRejectNewPlayerWithExistingName(unitTest *testing.T) {
	playerName := "A. Player Name"
	firstBodyObject := endpoint.PlayerState{
		Name:  playerName,
		Color: "First color",
	}

	playerCollection, testServer := newServer()

	firstPostResponse, firstEncodingError :=
		server.MockPost(testServer, "/backend/player/new-player", firstBodyObject)
	assertPostResponseCorrect(
		unitTest,
		"POST new-player first time for name "+playerName,
		firstPostResponse,
		firstEncodingError,
		http.StatusOK)

	functionRecord :=
		playerCollection.popSingleAndEnsureClear(
			unitTest,
			"TestRejectNewPlayerWithExistingName/initial registration")

	expectedFunctionRecord := functionNameAndArgument{
		FunctionName:     "Add",
		FunctionArgument: firstBodyObject,
	}

	if functionRecord != expectedFunctionRecord {
		unitTest.Fatalf(
			"POST new-player with valid JSON %v did not trigger expected function %v, instead recorded %v",
			firstBodyObject,
			expectedFunctionRecord,
			functionRecord)
	}

	secondBodyObject := endpoint.PlayerState{
		Name:  playerName,
		Color: "Second color",
	}

	secondPostResponse, secondEncodingError :=
		server.MockPost(testServer, "/backend/player/new-player", secondBodyObject)
	assertPostResponseCorrect(
		unitTest,
		"POST new-player second time for name "+playerName,
		secondPostResponse,
		secondEncodingError,
		http.StatusBadRequest)
}

// Remove, just test each endpoint called mock functions correctly.
func TestRegisterAndRetrieveNewPlayer(unitTest *testing.T) {
	type testArguments struct {
		playerName string
		chatColor  string
	}

	testCases := []struct {
		name      string
		arguments testArguments
	}{
		{
			name: "Ascii only, with color",
			arguments: testArguments{
				playerName: "Easy Test Name",
				chatColor:  "Plain color",
			},
		},
		{
			name: "Ascii only, no color",
			arguments: testArguments{
				playerName: "Easy Test Name",
			},
		},
		{
			name: "Punctuation and non-standard characters",
			arguments: testArguments{
				playerName: "?ß@äô#\"'\"",
				chatColor:  "\\\\\\",
			},
		},
		{
			name: "URI segment delimiter",
			arguments: testArguments{
				playerName: "/Slashes/are/reserved/for/parsing/URI/segments/",
				chatColor:  "irrelevant",
			},
		},
		{
			name: "Produces identifier with '/' in base64",
			arguments: testArguments{
				playerName: server.BreaksBase64,
				chatColor:  server.BreaksBase64,
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			playerCollection, testServer := newServer()

			bodyObject := endpoint.PlayerState{
				Name:  testCase.arguments.playerName,
				Color: testCase.arguments.chatColor,
			}

			postResponse, encodingError :=
				server.MockPost(testServer, "/backend/player/new-player", bodyObject)
			assertPostResponseCorrect(
				unitTest,
				"POST new-player",
				postResponse,
				encodingError,
				http.StatusOK)

			// Then we check that the POST returned a valid response.
			assertAtLeastOnePlayerReturnedInList(
				unitTest,
				postResponse,
				"POST new-player")

			// Finally we check that the player was registered properly.
			assertPlayerIsCorrectExternallyAndInternally(
				unitTest,
				playerCollection,
				playerHandler,
				testCase.arguments.playerName,
				testCase.arguments.chatColor,
				"Register new player")
		})
	}
}

func TestRejectInvalidUpdatePlayer(unitTest *testing.T) {
	endpointPlayer := endpoint.PlayerState{
		Name:  "Test Player",
		Color: "Test color",
	}

	type testArguments struct {
		bodyObject interface{}
	}

	type expectedReturns struct {
		codeFromPost int
	}

	testCases := []struct {
		name      string
		arguments testArguments
		expected  expectedReturns
	}{
		{
			name: "Nil object",
			arguments: testArguments{
				bodyObject: nil,
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
		{
			name: "Wrong object",
			arguments: testArguments{
				bodyObject: &endpoint.ChatColorList{
					Colors: []string{"Player 1", "Player 2"},
				},
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			_, playerHandler := newCollectionAndHandler()

			registrationBytesBuffer := new(bytes.Buffer)
			json.NewEncoder(registrationBytesBuffer).Encode(endpointPlayer)

			// First we add the player.
			playerHandler.HandlePost(json.NewDecoder(registrationBytesBuffer), []string{"new-player"})

			// We do not check that the POST succeeded, nor that return list is correct, nor do we check
			// that the player was correctly register: these are all covered by another test.

			// Now we try to update the player.
			updateBytesBuffer := new(bytes.Buffer)
			if testCase.arguments.bodyObject != nil {
				json.NewEncoder(updateBytesBuffer).Encode(testCase.arguments.bodyObject)
			}

			_, postCode :=
				playerHandler.HandlePost(json.NewDecoder(updateBytesBuffer), []string{"update-player"})

			if postCode != http.StatusBadRequest {
				unitTest.Fatalf(
					"POST update-player with invalid JSON %v did not return expected HTTP code %v, instead was %v.",
					testCase.arguments.bodyObject,
					http.StatusBadRequest,
					postCode)
			}
		})
	}
}

func TestUpdatePlayer(unitTest *testing.T) {
	playerName := "Test Player"
	originalColor := "white"
	newColor := "grey"

	type testArguments struct {
		playerName string
		chatColor  string
	}

	type expectedReturns struct {
		codeFromPost      int
		codeFromGet       int
		playerAfterUpdate *endpoint.PlayerState
	}

	testCases := []struct {
		name      string
		arguments testArguments
		expected  expectedReturns
	}{
		{
			name: "Non-existent player",
			arguments: testArguments{
				playerName: "Non-existent player",
				chatColor:  newColor,
			},
			expected: expectedReturns{
				codeFromPost:      http.StatusBadRequest,
				codeFromGet:       http.StatusNotFound,
				playerAfterUpdate: nil,
			},
		},
		{
			name: "No-op with empty color",
			arguments: testArguments{
				playerName: playerName,
				chatColor:  "",
			},
			expected: expectedReturns{
				codeFromPost: http.StatusOK,
				codeFromGet:  http.StatusOK,
				playerAfterUpdate: &endpoint.PlayerState{
					Name:  playerName,
					Color: originalColor,
				},
			},
		},
		{
			name: "Simple color change",
			arguments: testArguments{
				playerName: playerName,
				chatColor:  newColor,
			},
			expected: expectedReturns{
				codeFromPost: http.StatusOK,
				codeFromGet:  http.StatusOK,
				playerAfterUpdate: &endpoint.PlayerState{
					Name:  playerName,
					Color: newColor,
				},
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			playerCollection, playerHandler := newCollectionAndHandler()

			registrationBytesBuffer := new(bytes.Buffer)
			json.NewEncoder(registrationBytesBuffer).Encode(endpoint.PlayerState{
				Name:  playerName,
				Color: originalColor,
			})

			// First we add the player.
			registrationInterface, registratonCode :=
				playerHandler.HandlePost(json.NewDecoder(registrationBytesBuffer), []string{"new-player"})

			// Checks that the POST succeeded and that the return list is correct are covered by other
			// tests, but we need to parse the response to find the identifier generated for the new
			// player.
			playerList := assertAtLeastOnePlayerReturnedInList(
				unitTest,
				registratonCode,
				registrationInterface,
				"POST to create new player before updating")

			playerIdentifier := testCase.arguments.playerName
			for _, playerState := range playerList.Players {
				if playerState.Name == testCase.arguments.playerName {
					playerIdentifier = playerState.Identifier
				}
			}

			// Now we update the player.
			updateBytesBuffer := new(bytes.Buffer)
			json.NewEncoder(updateBytesBuffer).Encode(endpoint.PlayerState{
				Identifier: playerIdentifier,
				Name:       testCase.arguments.playerName,
				Color:      testCase.arguments.chatColor,
			})

			updateInterface, updateCode :=
				playerHandler.HandlePost(json.NewDecoder(updateBytesBuffer), []string{"update-player"})

			if updateCode != testCase.expected.codeFromPost {
				unitTest.Fatalf(
					"POST update-player did not return expected HTTP code %v, instead was %v.",
					testCase.expected.codeFromPost,
					updateCode)
			}

			// We check that we get a valid response body only when we expect a valid response code.
			if testCase.expected.codeFromPost == http.StatusOK {
				assertAtLeastOnePlayerReturnedInList(
					unitTest,
					updateCode,
					updateInterface,
					"POST update-player")
			}

			// If the test expects a valid player to have been updated, we check that it really is
			// there and is as expected.
			if testCase.expected.playerAfterUpdate != nil {
				assertPlayerIsCorrectExternallyAndInternally(
					unitTest,
					playerCollection,
					playerHandler,
					testCase.expected.playerAfterUpdate.Name,
					testCase.expected.playerAfterUpdate.Color,
					"Update valid player")
			}
		})
	}
}

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

func assertPostResponseCorrect(
	unitTest *testing.T,
	testIdentifier string,
	responseRecorder *httptest.ResponseRecorder,
	encodingError error,
	expectedCode int) {
	if encodingError != nil {
		unitTest.Fatalf(
			testIdentifier+"/encoding error: %v",
			encodingError)
	}

	if postResponse.Code != expectedCode {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			expectedCode,
			postResponse.Code)
	}
}

func assertAtLeastOnePlayerReturnedInList(
	unitTest *testing.T,
	responseRecorder *httptest.ResponseRecorder,
	endpointIdentifier string) endpoint.PlayerList {
	if responseRecorder.Code != http.StatusOK {
		unitTest.Fatalf(
			"GET registered-players did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			responseCode)
	}

	bodyDecoder := json.NewDecoder(responseRecorder.Body)

	var responsePlayerList endpoint.PlayerList
	parsingError := httpBodyDecoder.Decode(&responsePlayerList)
	if parsingError != nil {
		unitTest.Fatalf(
			"Error parsing JSON from HTTP response body: %v",
			parsingError)
	}

	if responsePlayerList.Players == nil {
		unitTest.Fatalf(
			endpointIdentifier+" returned %v which has a nil list of players.",
			responseInterface)
	}

	if len(responsePlayerList.Players) <= 0 {
		unitTest.Fatalf(
			endpointIdentifier+" returned %v which has an empty list of players.",
			responsePlayerList)
	}

	return responsePlayerList
}

func assertPlayerIsCorrect(
	unitTest *testing.T,
	expectedPlayerIdentifier string,
	actualPlayerIdentifier string,
	expectedPlayerName string,
	actualPlayerName string,
	expectedChatColor string,
	actualChatColor string,
	testIdentifier string) {
	if actualPlayerIdentifier != expectedPlayerIdentifier {
		unitTest.Fatalf(
			testIdentifier+": player with identifier %v was found but had identifier %v.",
			expectedPlayerIdentifier,
			actualPlayerIdentifier)
	}

	if actualPlayerName != expectedPlayerName {
		unitTest.Fatalf(
			testIdentifier+": player %v was found but had name %v.",
			expectedPlayerName,
			actualPlayerName)
	}

	if actualChatColor != expectedChatColor {
		if expectedChatColor != "" {
			unitTest.Fatalf(
				testIdentifier+": player %v was found but had color %v instead of expected %v.",
				expectedPlayerName,
				actualChatColor,
				expectedChatColor)
		}

		// Otherwise we check that the player was assigned a valid color.
		isValidColor := false
		availableColors := colorsAvailableInTest
		for _, availableColor := range availableColors {
			if actualChatColor == availableColor {
				isValidColor = true
				break
			}
		}

		if !isValidColor {
			unitTest.Fatalf(
				testIdentifier+": player %v was found but had color %v which is not in list of allowed colors %v.",
				expectedPlayerName,
				actualChatColor,
				availableColors)
		}
	}
}

func assertPlayerIsCorrectExternallyAndInternally(
	unitTest *testing.T,
	playerCollection *player.StateCollection,
	testHandler *player.GetAndPostHandler,
	expectedPlayerName string,
	expectedChatColor string,
	testIdentifier string) {
	// We have to look externally first so that we can find the identifier matching the name.
	getInterface, getCode :=
		testHandler.HandleGet([]string{"registered-players"})

	getPlayerList := assertAtLeastOnePlayerReturnedInList(
		unitTest,
		getCode,
		getInterface,
		testIdentifier+"/GET registered-players")

	// This should be wrong, but it will be over-written if the player exists anyway.
	expectedIdentifier := expectedPlayerName

	hasNewPlayer := false
	for _, registeredPlayer := range getPlayerList.Players {
		if expectedPlayerName == registeredPlayer.Name {
			hasNewPlayer = true
			expectedIdentifier = registeredPlayer.Identifier

			assertPlayerIsCorrect(
				unitTest,
				expectedIdentifier,
				registeredPlayer.Identifier,
				expectedPlayerName,
				registeredPlayer.Name,
				expectedChatColor,
				registeredPlayer.Color,
				testIdentifier+"/GET registered-players")
		}
	}

	if !hasNewPlayer {
		unitTest.Fatalf(
			testIdentifier+"/GET registered-players did not have %v in its list of players %v.",
			expectedPlayerName,
			getPlayerList.Players)
	}

	// We can check the internal function now that we have the identifier.
	internalPlayer, internalIdentificationError :=
		playerCollection.Get(expectedIdentifier)

	if internalIdentificationError != nil {
		unitTest.Fatalf(
			testIdentifier+"/internal: did not find player %v (error = %v).",
			expectedPlayerName,
			internalIdentificationError)
	}

	if internalPlayer == nil {
		unitTest.Fatalf(
			testIdentifier+"/internal: found nil for player %v.",
			expectedPlayerName)
	}

	assertPlayerIsCorrect(
		unitTest,
		expectedIdentifier,
		internalPlayer.Identifier(),
		expectedPlayerName,
		internalPlayer.Name(),
		expectedChatColor,
		internalPlayer.Color(),
		testIdentifier+"/internal player.GetAndPostHandler.GetPlayerByName")
}
