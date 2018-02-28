package player_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/benoleary/ilutulestikud/backendjson"

	"github.com/benoleary/ilutulestikud/player"
)

// This just tests that the factory method does not cause any panics, and returns a non-nil pointer.
func TestNewHandler(unitTest *testing.T) {
	actualState := player.NewHandler()
	if actualState == nil {
		unitTest.Fatalf("New handler was nil.")
	}
}

func TestGetNoSegmentBadRequest(unitTest *testing.T) {
	playerHandler := player.NewHandler()
	_, actualCode := playerHandler.HandleGet(make([]string, 0))

	if actualCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"GET with empty list of relevant segments did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			actualCode)
	}
}

func TestGetInvalidSegmentNotFound(unitTest *testing.T) {
	playerHandler := player.NewHandler()
	_, actualCode := playerHandler.HandleGet([]string{"invalid-segment"})

	if actualCode != http.StatusNotFound {
		unitTest.Fatalf(
			"GET invalid-segment did not return expected HTTP code %v, instead was %v.",
			http.StatusNotFound,
			actualCode)
	}
}

func TestPostNoSegmentBadRequest(unitTest *testing.T) {
	bytesBuffer := new(bytes.Buffer)
	json.NewEncoder(bytesBuffer).Encode(backendjson.PlayerState{
		Name:  "Player Name",
		Color: "Chat color",
	})
	playerHandler := player.NewHandler()

	_, actualCode := playerHandler.HandlePost(json.NewDecoder(bytesBuffer), make([]string, 0))

	if actualCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"POST with empty list of relevant segments did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			actualCode)
	}
}

func TestPostInvalidSegmentNotFound(unitTest *testing.T) {
	bytesBuffer := new(bytes.Buffer)
	json.NewEncoder(bytesBuffer).Encode(backendjson.PlayerState{
		Name:  "Player Name",
		Color: "Chat color",
	})
	playerHandler := player.NewHandler()

	_, actualCode := playerHandler.HandlePost(json.NewDecoder(bytesBuffer), []string{"invalid-segment"})

	if actualCode != http.StatusNotFound {
		unitTest.Fatalf(
			"POST invalid-segment did not return expected HTTP code %v, instead was %v.",
			http.StatusNotFound,
			actualCode)
	}
}

func TestDefaultPlayerListNotEmpty(unitTest *testing.T) {
	playerHandler := player.NewHandler()
	actualInterface, actualCode := playerHandler.HandleGet([]string{"registered-players"})

	if actualCode != http.StatusOK {
		unitTest.Fatalf(
			"GET registered-players did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			actualCode)
	}

	actualPlayerStateList, isTypeCorrect := actualInterface.(backendjson.PlayerStateList)

	if !isTypeCorrect {
		unitTest.Fatalf(
			"GET registered-players did not return expected backendjson.PlayerStateList, instead was %v.",
			actualInterface)
	}

	if actualPlayerStateList.Players == nil {
		unitTest.Fatalf(
			"GET registered-players returned %v which has a nil list of players.",
			actualInterface)
	}

	if len(actualPlayerStateList.Players) <= 0 {
		unitTest.Fatalf(
			"GET registered-players returned %v which has an empty list of players.",
			actualPlayerStateList)
	}
}

func TestAvailableColorListNotEmpty(unitTest *testing.T) {
	playerHandler := player.NewHandler()
	actualInterface, actualCode := playerHandler.HandleGet([]string{"available-colors"})

	if actualCode != http.StatusOK {
		unitTest.Fatalf(
			"GET available-colors did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			actualCode)
	}

	actualAvailableColorList, isTypeCorrect := actualInterface.(backendjson.ChatColorList)

	if !isTypeCorrect {
		unitTest.Fatalf(
			"GET available-colors did not return expected backendjson.ChatColorList, instead was %v.",
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
		handler   *player.Handler
		arguments testArguments
		expected  expectedReturns
	}{
		{
			name:    "Nil object",
			handler: player.NewHandler(),
			arguments: testArguments{
				bodyObject: nil,
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
		{
			name:    "Wrong object",
			handler: player.NewHandler(),
			arguments: testArguments{
				bodyObject: &backendjson.ChatColorList{
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
			json.NewEncoder(bytesBuffer).Encode(testCase.arguments.bodyObject)

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

func TestRegisterAndRetrieveNewPlayer(unitTest *testing.T) {
	type testArguments struct {
		playerName string
		chatColor  string
	}

	type expectedReturns struct {
		codeFromPost int
		codeFromGet  int
	}

	testCases := []struct {
		name      string
		handler   *player.Handler
		arguments testArguments
		expected  expectedReturns
	}{
		{
			name:    "Ascii only",
			handler: player.NewHandler(),
			arguments: testArguments{
				playerName: "Easy Test Name",
				chatColor:  "Plain color",
			},
			expected: expectedReturns{
				codeFromPost: http.StatusOK,
				codeFromGet:  http.StatusOK,
			},
		},
		{
			name:    "Punctuation and non-standard characters",
			handler: player.NewHandler(),
			arguments: testArguments{
				playerName: "?ß@äô#\"'\"",
				chatColor:  "\\\\\\",
			},
			expected: expectedReturns{
				codeFromPost: http.StatusOK,
				codeFromGet:  http.StatusOK,
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			bytesBuffer := new(bytes.Buffer)
			json.NewEncoder(bytesBuffer).Encode(backendjson.PlayerState{
				Name:  testCase.arguments.playerName,
				Color: testCase.arguments.chatColor,
			})

			// First we add the new player.
			postInterface, postCode :=
				testCase.handler.HandlePost(json.NewDecoder(bytesBuffer), []string{"new-player"})

			if postCode != http.StatusOK {
				unitTest.Fatalf("POST new-player did not return expected HTTP code %v, instead was %v.", http.StatusOK, postCode)
			}

			postPlayerStateList, isPostTypeCorrect := postInterface.(backendjson.PlayerStateList)

			if !isPostTypeCorrect {
				unitTest.Fatalf("POST new-player did not return expected backendjson.PlayerStateList, instead was %v.", postInterface)
			}

			if postPlayerStateList.Players == nil {
				unitTest.Fatalf("POST new-player returned %v which has a nil list of players.", postInterface)
			}

			if len(postPlayerStateList.Players) <= 0 {
				unitTest.Fatalf("POST new-player returned %v which has an empty list of players.", postPlayerStateList)
			}

			// First we check that we can retrieve the new player within the program.
			internalPlayer, isFoundInternally := testCase.handler.GetPlayerByName(testCase.arguments.playerName)

			if !isFoundInternally {
				unitTest.Fatalf("Did not find player %v.", testCase.arguments.playerName)
			}

			if testCase.arguments.playerName != internalPlayer.Name {
				unitTest.Fatalf("Player %v was found but had name %v.", testCase.arguments.playerName, internalPlayer.Name)
			}

			// Finally we check that the new player exists in the list of registered players given out by the endpoint.
			getInterface, getCode :=
				testCase.handler.HandleGet([]string{"registered-players"})

			if getCode != http.StatusOK {
				unitTest.Fatalf("GET registered-players did not return expected HTTP code %v, instead was %v.", http.StatusOK, getCode)
			}

			getPlayerStateList, isGetTypeCorrect := getInterface.(backendjson.PlayerStateList)

			if !isGetTypeCorrect {
				unitTest.Fatalf("GET registered-players did not return expected backendjson.PlayerStateList, instead was %v.", postInterface)
			}

			if getPlayerStateList.Players == nil {
				unitTest.Fatalf("GET registered-players returned %v which has a nil list of players.", postInterface)
			}

			if len(getPlayerStateList.Players) <= 0 {
				unitTest.Fatalf("GET registered-players returned %v which has an empty list of players.", postPlayerStateList)
			}

			hasNewPlayer := false
			for _, registerPlayer := range getPlayerStateList.Players {
				if testCase.arguments.playerName == registerPlayer.Name {
					hasNewPlayer = true
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
