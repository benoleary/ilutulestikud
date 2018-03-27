package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/player"
	"github.com/benoleary/ilutulestikud/backend/server"
)

const breaksBase64 = "\\/\\\\\\?" // This should unescape to \/\\\? in the tests.

var colorsAvailableInTest []string = defaults.AvailableColors()

// newServerForIdentifier prepares a server.State in a consistent way for the
// tests of the player endpoints.
func newServerForIdentifier(
	nameToIdentifier endpoint.NameToIdentifier) *server.State {
	playerPersister := player.NewInMemoryPersister(nameToIdentifier)
	playerCollection :=
		player.NewCollection(
			playerPersister,
			defaults.InitialPlayerNames(),
			colorsAvailableInTest)

	return server.New("test",
		playerCollection,
		nil)
}

// newServer prepares a server.State in a consistent way for the tests of the
// player endpoints.
func newServer() *server.State {
	return newServerForIdentifier(&endpoint.Base32NameEncoder{})
}

func TestGetNoSegmentBadRequest(unitTest *testing.T) {
	testServer := newServer()

	getResponse := server.MockGet(testServer, "/backend/player")

	if getResponse.Code != http.StatusBadRequest {
		unitTest.Fatalf(
			"GET with empty list of relevant segments did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			getResponse.Code)
	}
}

func TestGetInvalidSegmentNotFound(unitTest *testing.T) {
	testServer := newServer()

	getResponse := server.MockGet(testServer, "/backend/player/invalid-segment")

	if getResponse.Code != http.StatusNotFound {
		unitTest.Fatalf(
			"GET invalid-segment did not return expected HTTP code %v, instead was %v.",
			http.StatusNotFound,
			getResponse.Code)
	}
}

func TestPostNoSegmentBadRequest(unitTest *testing.T) {
	testServer := newServer()

	bodyObject := endpoint.PlayerState{
		Name:  "Player Name",
		Color: "Chat color",
	}

	postResponse, encodingError :=
		server.MockPost(testServer, "/backend/player", bodyObject)

	if encodingError != nil {
		unitTest.Fatalf(
			"Encoding error for POST: %v",
			encodingError)
	}

	if postResponse.Code != http.StatusBadRequest {
		unitTest.Fatalf(
			"POST with empty list of relevant segments did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			postResponse.Code)
	}
}

func TestPostInvalidSegmentNotFound(unitTest *testing.T) {
	_, playerHandler := newCollectionAndHandler()
	bytesBuffer := new(bytes.Buffer)
	json.NewEncoder(bytesBuffer).Encode(endpoint.PlayerState{
		Name:  "Player Name",
		Color: "Chat color",
	})

	_, actualCode := playerHandler.HandlePost(json.NewDecoder(bytesBuffer), []string{"invalid-segment"})

	if actualCode != http.StatusNotFound {
		unitTest.Fatalf(
			"POST invalid-segment did not return expected HTTP code %v, instead was %v.",
			http.StatusNotFound,
			actualCode)
	}
}

func TestDefaultPlayerListNotEmpty(unitTest *testing.T) {
	_, playerHandler := newCollectionAndHandler()
	actualInterface, actualCode := playerHandler.HandleGet([]string{"registered-players"})
	assertAtLeastOnePlayerReturnedInList(
		unitTest,
		actualCode,
		actualInterface,
		"GET registered-players")
}

func TestAvailableColorListNotEmpty(unitTest *testing.T) {
	_, playerHandler := newCollectionAndHandler()
	actualInterface, actualCode := playerHandler.HandleGet([]string{"available-colors"})

	if actualCode != http.StatusOK {
		unitTest.Fatalf(
			"GET available-colors did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			actualCode)
	}

	actualAvailableColorList, isTypeCorrect := actualInterface.(endpoint.ChatColorList)

	if !isTypeCorrect {
		unitTest.Fatalf(
			"GET available-colors did not return expected endpoint.ChatColorList, instead was %v.",
			actualInterface)
	}

	if actualAvailableColorList.Colors == nil {
		unitTest.Fatalf(
			"GET available-colors returned %v which has a nil list of colors.",
			actualInterface)
	}

	if len(actualAvailableColorList.Colors) <= 0 {
		unitTest.Fatalf(
			"GET available-colors returned %v which has an empty list of colors.",
			actualAvailableColorList)
	}
}

func TestReturnErrorWhenPlayerNotFoundInternally(unitTest *testing.T) {
	playerCollection, _ := newCollectionAndHandler()

	invalidIdentifier := "not a valid identifier"
	internalPlayer, internalIdentificationError :=
		playerCollection.Get(invalidIdentifier)

	if internalIdentificationError == nil {
		unitTest.Fatalf(
			"Internal get of invalid player identifier %v did not return an error, did return player state %v",
			invalidIdentifier,
			internalPlayer)
	}
}

func TestRejectInvalidNewPlayer(unitTest *testing.T) {
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
			bytesBuffer := new(bytes.Buffer)
			if testCase.arguments.bodyObject != nil {
				json.NewEncoder(bytesBuffer).Encode(testCase.arguments.bodyObject)
			}

			_, playerHandler := newCollectionAndHandler()
			_, postCode :=
				playerHandler.HandlePost(json.NewDecoder(bytesBuffer), []string{"new-player"})

			if postCode != http.StatusBadRequest {
				unitTest.Fatalf(
					"POST new-player with invalid JSON %v did not return expected HTTP code %v, instead was %v.",
					testCase.arguments.bodyObject,
					http.StatusBadRequest,
					postCode)
			}
		})
	}
}

func TestRejectNewPlayerWithNameWhichBreaksEncoding(unitTest *testing.T) {
	// We need a special kind of name and encoding.
	playerName := breaksBase64
	_, playerHandler := newCollectionAndHandlerForIdentifier(&endpoint.Base64NameEncoder{})

	bodyObject := endpoint.PlayerState{
		Name:  playerName,
		Color: "First color",
	}

	bytesBuffer := new(bytes.Buffer)
	json.NewEncoder(bytesBuffer).Encode(bodyObject)

	_, invalidRegistrationCode :=
		playerHandler.HandlePost(json.NewDecoder(bytesBuffer), []string{"new-player"})

	if invalidRegistrationCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"POST new-player with valid JSON %v but encoding-breaking player name %v"+
				" did not return expected HTTP code %v, instead was %v.",
			bodyObject,
			playerName,
			http.StatusBadRequest,
			invalidRegistrationCode)
	}
}

func TestRejectNewPlayerWithExistingName(unitTest *testing.T) {
	playerName := "A. Player Name"
	firstBodyObject := endpoint.PlayerState{
		Name:  playerName,
		Color: "First color",
	}

	firstBytesBuffer := new(bytes.Buffer)
	json.NewEncoder(firstBytesBuffer).Encode(firstBodyObject)

	_, playerHandler := newCollectionAndHandler()
	_, validRegistrationCode :=
		playerHandler.HandlePost(json.NewDecoder(firstBytesBuffer), []string{"new-player"})

	if validRegistrationCode != http.StatusOK {
		unitTest.Fatalf(
			"POST new-player with valid JSON %v did not return expected HTTP code %v, instead was %v.",
			firstBodyObject,
			http.StatusOK,
			validRegistrationCode)
	}

	secondBodyObject := endpoint.PlayerState{
		Name:  playerName,
		Color: "Second color",
	}

	secondBytesBuffer := new(bytes.Buffer)
	json.NewEncoder(secondBytesBuffer).Encode(secondBodyObject)

	_, invalidRegistrationCode :=
		playerHandler.HandlePost(json.NewDecoder(secondBytesBuffer), []string{"new-player"})

	if invalidRegistrationCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"POST new-player with valid JSON %v but second request for same player name %v"+
				" did not return expected HTTP code %v, instead was %v.",
			secondBodyObject,
			playerName,
			http.StatusBadRequest,
			invalidRegistrationCode)
	}
}

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
				playerName: breaksBase64,
				chatColor:  breaksBase64,
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			playerCollection, playerHandler := newCollectionAndHandler()

			bytesBuffer := new(bytes.Buffer)
			json.NewEncoder(bytesBuffer).Encode(endpoint.PlayerState{
				Name:  testCase.arguments.playerName,
				Color: testCase.arguments.chatColor,
			})

			// First we add the new player.
			postInterface, postCode :=
				playerHandler.HandlePost(json.NewDecoder(bytesBuffer), []string{"new-player"})

			// Then we check that the POST returned a valid response.
			assertAtLeastOnePlayerReturnedInList(
				unitTest,
				postCode,
				postInterface,
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

func assertAtLeastOnePlayerReturnedInList(
	unitTest *testing.T,
	responseCode int,
	responseInterface interface{},
	endpointIdentifier string) endpoint.PlayerList {
	if responseCode != http.StatusOK {
		unitTest.Fatalf(
			"GET registered-players did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			responseCode)
	}

	responsePlayerList, isTypeCorrect := responseInterface.(endpoint.PlayerList)

	if !isTypeCorrect {
		unitTest.Fatalf(
			endpointIdentifier+" did not return expected endpoint.PlayerList, instead was %v.",
			responseInterface)
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
