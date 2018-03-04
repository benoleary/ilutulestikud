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

// newHandler prepares a GetAndPostHandler for the tests.
func newHandler() *player.GetAndPostHandler {
	playerFactory :=
		player.NewInMemoryCollection(
			defaults.InitialPlayerNames(),
			colorsAvailableInTest)
	return player.NewGetAndPostHandler(playerFactory)
}

func TestGetNoSegmentBadRequest(unitTest *testing.T) {
	playerHandler := newHandler()
	_, actualCode := playerHandler.HandleGet(make([]string, 0))

	if actualCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"GET with empty list of relevant segments did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			actualCode)
	}
}

func TestGetInvalidSegmentNotFound(unitTest *testing.T) {
	playerHandler := newHandler()
	_, actualCode := playerHandler.HandleGet([]string{"invalid-segment"})

	if actualCode != http.StatusNotFound {
		unitTest.Fatalf(
			"GET invalid-segment did not return expected HTTP code %v, instead was %v.",
			http.StatusNotFound,
			actualCode)
	}
}

func TestPostNoSegmentBadRequest(unitTest *testing.T) {
	playerHandler := newHandler()
	bytesBuffer := new(bytes.Buffer)
	json.NewEncoder(bytesBuffer).Encode(endpoint.PlayerState{
		Name:  "Player Name",
		Color: "Chat color",
	})

	_, actualCode := playerHandler.HandlePost(json.NewDecoder(bytesBuffer), make([]string, 0))

	if actualCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"POST with empty list of relevant segments did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			actualCode)
	}
}

func TestPostInvalidSegmentNotFound(unitTest *testing.T) {
	playerHandler := newHandler()
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
	playerHandler := newHandler()
	actualInterface, actualCode := playerHandler.HandleGet([]string{"registered-players"})
	assertAtLeastOnePlayerReturnedInList(
		unitTest,
		actualCode,
		actualInterface,
		"GET registered-players")
}

func TestAvailableColorListNotEmpty(unitTest *testing.T) {
	playerHandler := newHandler()
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

func TestRejectInvalidNewPlayer(unitTest *testing.T) {
	type testArguments struct {
		bodyObject interface{}
	}

	type expectedReturns struct {
		codeFromPost int
	}

	testCases := []struct {
		name      string
		handler   *player.GetAndPostHandler
		arguments testArguments
		expected  expectedReturns
	}{
		{
			name:    "Nil object",
			handler: newHandler(),
			arguments: testArguments{
				bodyObject: nil,
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
		{
			name:    "Wrong object",
			handler: newHandler(),
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

			_, postCode :=
				testCase.handler.HandlePost(json.NewDecoder(bytesBuffer), []string{"new-player"})

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

func TestRejectNewPlayerWithExistingName(unitTest *testing.T) {
	playerName := "A. Player Name"
	firstBodyObject := endpoint.PlayerState{
		Name:  playerName,
		Color: "First color",
	}

	firstBytesBuffer := new(bytes.Buffer)
	json.NewEncoder(firstBytesBuffer).Encode(firstBodyObject)

	playerHandler := newHandler()
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
			playerName,
			secondBodyObject,
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
		handler   *player.GetAndPostHandler
		arguments testArguments
	}{
		{
			name:    "Ascii only, with color",
			handler: newHandler(),
			arguments: testArguments{
				playerName: "Easy Test Name",
				chatColor:  "Plain color",
			},
		},
		{
			name:    "Ascii only, no color",
			handler: newHandler(),
			arguments: testArguments{
				playerName: "Easy Test Name",
			},
		},
		{
			name:    "Punctuation and non-standard characters",
			handler: newHandler(),
			arguments: testArguments{
				playerName: "?ß@äô#\"'\"",
				chatColor:  "\\\\\\",
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			bytesBuffer := new(bytes.Buffer)
			json.NewEncoder(bytesBuffer).Encode(endpoint.PlayerState{
				Name:  testCase.arguments.playerName,
				Color: testCase.arguments.chatColor,
			})

			// First we add the new player.
			postInterface, postCode :=
				testCase.handler.HandlePost(json.NewDecoder(bytesBuffer), []string{"new-player"})

			assertAtLeastOnePlayerReturnedInList(
				unitTest,
				postCode,
				postInterface,
				"POST new-player")

			// First we check that we can retrieve the player within the program.
			internalPlayer, isFoundInternally :=
				testCase.handler.GetPlayerByName(testCase.arguments.playerName)

			if !isFoundInternally {
				unitTest.Fatalf("Did not find player %v.", testCase.arguments.playerName)
			}

			if internalPlayer == nil {
				unitTest.Fatalf("Found nil for player %v.", testCase.arguments.playerName)
			}

			assertPlayerIsCorrect(
				unitTest,
				testCase.arguments.playerName,
				internalPlayer.Name(),
				testCase.arguments.chatColor,
				internalPlayer.Color(),
				"Internal player.GetAndPostHandler.GetPlayerByName")

			// Finally we check that the player exists in the list of registered players given out by the endpoint.
			getInterface, getCode :=
				testCase.handler.HandleGet([]string{"registered-players"})

			getPlayerStateList := assertAtLeastOnePlayerReturnedInList(
				unitTest,
				getCode,
				getInterface,
				"GET registered-players")

			hasNewPlayer := false
			for _, registeredPlayer := range getPlayerStateList.Players {
				if testCase.arguments.playerName == registeredPlayer.Name {
					hasNewPlayer = true

					assertPlayerIsCorrect(
						unitTest,
						testCase.arguments.playerName,
						registeredPlayer.Name,
						testCase.arguments.chatColor,
						registeredPlayer.Color,
						"GET registered-players")
				}
			}

			if !hasNewPlayer {
				unitTest.Fatalf(
					"GET registered-players did not have %v in its list of players %v.",
					testCase.arguments.playerName,
					getPlayerStateList.Players)
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
		handler   *player.GetAndPostHandler
		arguments testArguments
		expected  expectedReturns
	}{
		{
			name:    "Non-existent player",
			handler: newHandler(),
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
			name:    "No-op with empty color",
			handler: newHandler(),
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
			name:    "Simple color change",
			handler: newHandler(),
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
			registrationBytesBuffer := new(bytes.Buffer)
			json.NewEncoder(registrationBytesBuffer).Encode(endpoint.PlayerState{
				Name:  playerName,
				Color: originalColor,
			})

			// First we add the player.
			testCase.handler.HandlePost(json.NewDecoder(registrationBytesBuffer), []string{"new-player"})

			// We do not check that the POST succeeded, nor that return list is correct, nor do we check
			// that the player was correctly register: these are all covered by another test.

			// Now we update the player.
			updateBytesBuffer := new(bytes.Buffer)
			json.NewEncoder(updateBytesBuffer).Encode(endpoint.PlayerState{
				Name:  testCase.arguments.playerName,
				Color: testCase.arguments.chatColor,
			})

			postInterface, postCode :=
				testCase.handler.HandlePost(json.NewDecoder(updateBytesBuffer), []string{"update-player"})

			if postCode != testCase.expected.codeFromPost {
				unitTest.Fatalf(
					"POST update-player did not return expected HTTP code %v, instead was %v.",
					http.StatusOK,
					postCode)
			}

			// We check that we get a valid response body only when we expect a valid response code.
			if testCase.expected.codeFromPost == http.StatusOK {
				assertAtLeastOnePlayerReturnedInList(
					unitTest,
					postCode,
					postInterface,
					"POST update-player")
			}

			// If the test expects a valid player to have been updated, we check that it reallyis
			// there and is as expected.
			if testCase.expected.playerAfterUpdate != nil {
				// First we check that we can retrieve the player within the program.
				internalPlayer, isFoundInternally :=
					testCase.handler.GetPlayerByName(testCase.arguments.playerName)

				if !isFoundInternally {
					unitTest.Fatalf("Did not find player %v.", testCase.arguments.playerName)
				}

				if internalPlayer == nil {
					unitTest.Fatalf("Found nil for player %v.", testCase.arguments.playerName)
				}

				assertPlayerIsCorrect(
					unitTest,
					testCase.expected.playerAfterUpdate.Name,
					internalPlayer.Name(),
					testCase.expected.playerAfterUpdate.Color,
					internalPlayer.Color(),
					"Internal player.GetAndPostHandler.GetPlayerByName")

				// Finally we check that the player exists in the list of registered players given out by the endpoint.
				getInterface, getCode :=
					testCase.handler.HandleGet([]string{"registered-players"})

				getPlayerStateList := assertAtLeastOnePlayerReturnedInList(
					unitTest,
					getCode,
					getInterface,
					"GET registered-players")

				hasNewPlayer := false
				for _, registeredPlayer := range getPlayerStateList.Players {
					if testCase.arguments.playerName == registeredPlayer.Name {
						hasNewPlayer = true

						assertPlayerIsCorrect(
							unitTest,
							testCase.arguments.playerName,
							registeredPlayer.Name,
							testCase.arguments.chatColor,
							registeredPlayer.Color,
							"GET registered-players")
					}
				}

				if !hasNewPlayer {
					unitTest.Fatalf(
						"GET registered-players did not have %v in its list of players %v.",
						testCase.arguments.playerName,
						getPlayerStateList.Players)
				}
			}
		})
	}
}

func assertAtLeastOnePlayerReturnedInList(
	unitTest *testing.T,
	responseCode int,
	responseInterface interface{},
	endpointIdentifier string) endpoint.PlayerStateList {
	if responseCode != http.StatusOK {
		unitTest.Fatalf(
			"GET registered-players did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			responseCode)
	}

	responsePlayerStateList, isTypeCorrect := responseInterface.(endpoint.PlayerStateList)

	if !isTypeCorrect {
		unitTest.Fatalf(
			endpointIdentifier+" did not return expected endpoint.PlayerStateList, instead was %v.",
			responseInterface)
	}

	if responsePlayerStateList.Players == nil {
		unitTest.Fatalf(
			endpointIdentifier+" returned %v which has a nil list of players.",
			responseInterface)
	}

	if len(responsePlayerStateList.Players) <= 0 {
		unitTest.Fatalf(
			endpointIdentifier+" returned %v which has an empty list of players.",
			responsePlayerStateList)
	}

	return responsePlayerStateList
}

func assertPlayerIsCorrect(
	unitTest *testing.T,
	expectedPlayerName string,
	actualPlayerName string,
	expectedChatColor string,
	actualChatColor string,
	testIdentifier string) {
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
			if availableColor == actualChatColor {
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
