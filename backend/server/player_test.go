package server_test

import (
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

var testSegmentTranslator server.EndpointSegmentTranslator = &server.Base32Translator{}

type mockPlayerState struct {
	name  string
	color string
}

// Name returns the private name field.
func (playerState *mockPlayerState) Name() string {
	return playerState.name
}

// Color returns the private color field.
func (playerState *mockPlayerState) Color() string {
	return playerState.color
}

var testPlayerStates []player.ReadonlyState = []player.ReadonlyState{
	&mockPlayerState{
		name:  "Player One",
		color: colorsAvailableInTest[0],
	},
	// Player Two has the same color as Player One
	&mockPlayerState{
		name:  "Player Two",
		color: colorsAvailableInTest[0],
	},
	&mockPlayerState{
		name:  "Player Three",
		color: colorsAvailableInTest[1],
	},
}

var testPlayerList endpoint.PlayerList = endpoint.PlayerList{
	Players: []endpoint.PlayerState{
		endpoint.PlayerState{
			Identifier: testSegmentTranslator.ToSegment(testPlayerStates[0].Name()),
			Name:       testPlayerStates[0].Name(),
			Color:      testPlayerStates[0].Color(),
		},
		endpoint.PlayerState{
			Identifier: testSegmentTranslator.ToSegment(testPlayerStates[1].Name()),
			Name:       testPlayerStates[1].Name(),
			Color:      testPlayerStates[1].Color(),
		},
		endpoint.PlayerState{
			Identifier: testSegmentTranslator.ToSegment(testPlayerStates[2].Name()),
			Name:       testPlayerStates[2].Name(),
			Color:      testPlayerStates[2].Color(),
		},
	},
}

type functionNameAndArgument struct {
	FunctionName     string
	FunctionArgument interface{}
}

type mockPlayerCollection struct {
	FunctionsAndArgumentsReceived []functionNameAndArgument
	ErrorToReturn                 error
	ReturnForAll                  []player.ReadonlyState
	ReturnForGet                  player.ReadonlyState
	ReturnForAvailableChatColors  []string
}

func (mockCollection *mockPlayerCollection) recordFunctionAndArgument(
	functionName string, functionArgument interface{}) {
	mockCollection.FunctionsAndArgumentsReceived =
		append(
			mockCollection.FunctionsAndArgumentsReceived,
			functionNameAndArgument{
				FunctionName:     functionName,
				FunctionArgument: functionArgument,
			})
}

func (mockCollection *mockPlayerCollection) getFirstAndEnsureOnly(
	unitTest *testing.T,
	testIdentifier string) functionNameAndArgument {
	if len(mockCollection.FunctionsAndArgumentsReceived) != 1 {
		unitTest.Fatalf(
			testIdentifier+
				"/mock player collection recorded %v function calls, expected 1.",
			mockCollection.FunctionsAndArgumentsReceived)
	}

	return mockCollection.FunctionsAndArgumentsReceived[0]
}

// Add gets mocked.
func (mockCollection *mockPlayerCollection) Add(
	playerName string,
	chatColor string) error {
	mockCollection.recordFunctionAndArgument(
		"Add",
		[]string{playerName, chatColor})
	return mockCollection.ErrorToReturn
}

// UpdateColor gets mocked.
func (mockCollection *mockPlayerCollection) UpdateColor(
	playerName string,
	chatColor string) error {
	mockCollection.recordFunctionAndArgument(
		"UpdateColor",
		[]string{playerName, chatColor})
	return mockCollection.ErrorToReturn
}

// All gets mocked.
func (mockCollection *mockPlayerCollection) All() []player.ReadonlyState {
	mockCollection.recordFunctionAndArgument(
		"All",
		nil)
	return mockCollection.ReturnForAll
}

// Get gets mocked.
func (mockCollection *mockPlayerCollection) Get(playerIdentifier string) (player.ReadonlyState, error) {
	mockCollection.recordFunctionAndArgument(
		"playerIdentifier",
		playerIdentifier)
	return mockCollection.ReturnForGet, mockCollection.ErrorToReturn
}

// Reset gets mocked.
func (mockCollection *mockPlayerCollection) Reset() {
	mockCollection.recordFunctionAndArgument(
		"Reset",
		nil)
}

// AvailableChatColors gets mocked.
func (mockCollection *mockPlayerCollection) AvailableChatColors() []string {
	mockCollection.recordFunctionAndArgument(
		"AvailableChatColors",
		nil)
	return mockCollection.ReturnForAvailableChatColors
}

// newServer prepares a mock player collection and uses it to prepare a
// server.State with the default endpoint segment translator for the tests,
// in a consistent way for the tests of the player endpoints, returning the
// mock collection and the server state.
func newServer() (*mockPlayerCollection, *server.State) {
	return newServerForTranslator(testSegmentTranslator)
}

// newServerForTranslator prepares a mock player collection and uses it to
// prepare a server.State with the given endpoint segment translator in a
// consistent way for the tests of the player endpoints, returning the
// mock collection and the server state.
func newServerForTranslator(
	segmentTranslator server.EndpointSegmentTranslator) (*mockPlayerCollection, *server.State) {
	mockCollection := &mockPlayerCollection{}

	serverState :=
		server.New(
			"test",
			segmentTranslator,
			mockCollection,
			nil)

	return mockCollection, serverState
}

func TestGetNoSegmentBadRequest(unitTest *testing.T) {
	mockCollection, testServer := newServer()

	getResponse := server.MockGet(testServer, "/backend/player")

	testIdentifier := "GET with empty list of relevant segments"

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		getResponse,
		nil,
		http.StatusBadRequest)

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection,
		testIdentifier)
}

func TestGetInvalidSegmentNotFound(unitTest *testing.T) {
	mockCollection, testServer := newServer()

	getResponse :=
		server.MockGet(testServer, "/backend/player/invalid-segment")

	testIdentifier := "GET invalid-segment"

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		getResponse,
		nil,
		http.StatusNotFound)

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection,
		testIdentifier)
}

func TestPostNoSegmentBadRequest(unitTest *testing.T) {
	mockCollection, testServer := newServer()

	bodyObject := endpoint.PlayerState{
		Name:  "Player Name",
		Color: "Chat color",
	}

	postResponse, encodingError :=
		server.MockPost(testServer, "/backend/player", bodyObject)

	testIdentifier := "POST with empty list of relevant segments"

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		postResponse,
		encodingError,
		http.StatusBadRequest)

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection,
		testIdentifier)
}

func TestPostInvalidSegmentNotFound(unitTest *testing.T) {
	mockCollection, testServer := newServer()

	bodyObject := endpoint.PlayerState{
		Name:  "Player Name",
		Color: "Chat color",
	}

	postResponse, encodingError :=
		server.MockPost(testServer, "/backend/player/invalid-segment", bodyObject)

	testIdentifier := "POST invalid-segment"

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		postResponse,
		encodingError,
		http.StatusNotFound)

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection,
		testIdentifier)
}

func TestPlayerListDelivered(unitTest *testing.T) {
	mockCollection, testServer := newServer()

	mockCollection.ReturnForAll = testPlayerStates

	getResponse :=
		server.MockGet(testServer, "/backend/player/registered-players")

	testIdentifier := "GET registered-players"

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		getResponse,
		nil,
		http.StatusOK)

	functionRecord :=
		mockCollection.getFirstAndEnsureOnly(
			unitTest,
			testIdentifier)

	assertFunctionRecordIsCorrect(
		unitTest,
		functionRecord,
		functionNameAndArgument{
			FunctionName:     "All",
			FunctionArgument: nil,
		},
		testIdentifier)

	bodyDecoder := json.NewDecoder(getResponse.Body)

	var responsePlayerList endpoint.PlayerList
	parsingError := bodyDecoder.Decode(&responsePlayerList)
	if parsingError != nil {
		unitTest.Fatalf(
			testIdentifier+"/error parsing JSON from HTTP response body: %v",
			parsingError)
	}

	if responsePlayerList.Players == nil {
		unitTest.Fatalf(
			testIdentifier+"/returned %v which has a nil list of players.",
			responsePlayerList)
	}

	expectedPlayers := testPlayerList.Players

	if len(responsePlayerList.Players) != len(expectedPlayers) {
		unitTest.Fatalf(
			testIdentifier+
				"/returned %v which does not match the expected list of players %v.",
			responsePlayerList,
			expectedPlayers)
	}

	// The list of expected players contains no duplicates, so it suffices to compare lengths
	// and that every expected players is found.
	for _, expectedPlayer := range expectedPlayers {
		foundPlayer := false
		for _, actualPlayer := range responsePlayerList.Players {
			if (actualPlayer.Identifier == expectedPlayer.Identifier) &&
				(actualPlayer.Name == expectedPlayer.Name) &&
				(actualPlayer.Color == expectedPlayer.Color) {
				foundPlayer = true
			}
		}

		if !foundPlayer {
			unitTest.Fatalf(
				testIdentifier+
					"/returned %v which does not match the expected list of players %v"+
					" (did not find %v).",
				responsePlayerList,
				expectedPlayers,
				expectedPlayer)
		}
	}
}

func TestAvailableColorListNotEmpty(unitTest *testing.T) {
	mockCollection, testServer := newServer()

	expectedColors :=
		[]string{
			"red",
			"green",
			"blue",
		}

	mockCollection.ReturnForAvailableChatColors = expectedColors

	getResponse :=
		server.MockGet(testServer, "/backend/player/available-colors")

	testIdentifier := "GET available-colors"

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		getResponse,
		nil,
		http.StatusOK)

	functionRecord :=
		mockCollection.getFirstAndEnsureOnly(
			unitTest,
			testIdentifier)

	assertFunctionRecordIsCorrect(
		unitTest,
		functionRecord,
		functionNameAndArgument{
			FunctionName:     "AvailableChatColors",
			FunctionArgument: nil,
		},
		testIdentifier)

	bodyDecoder := json.NewDecoder(getResponse.Body)

	var responseColorList endpoint.ChatColorList
	parsingError := bodyDecoder.Decode(&responseColorList)
	if parsingError != nil {
		unitTest.Fatalf(
			testIdentifier+"/error parsing JSON from HTTP response body: %v",
			parsingError)
	}

	if responseColorList.Colors == nil {
		unitTest.Fatalf(
			testIdentifier+"/returned %v which has a nil list of colors.",
			responseColorList)
	}

	if len(responseColorList.Colors) != len(expectedColors) {
		unitTest.Fatalf(
			testIdentifier+
				"/returned %v which does not match the expected list of colors %v.",
			responseColorList,
			expectedColors)
	}

	// The list of expected colors contains no duplicates, so it suffices to compare lengths
	// and that every expected color is found.
	for _, expectedColor := range expectedColors {
		foundColor := false
		for _, actualColor := range responseColorList.Colors {
			if actualColor == expectedColor {
				foundColor = true
			}
		}

		if !foundColor {
			unitTest.Fatalf(
				testIdentifier+
					"/returned %v which does not match the expected list of colors %v"+
					" (did not find %v).",
				responseColorList,
				expectedColors,
				expectedColor)
		}
	}
}

func TestRejectInvalidNewPlayerObject(unitTest *testing.T) {
	baseIdentifier := "Reject invalid POST new-player with body = "

	testCases := []struct {
		testIdentifier string
		bodyObject     interface{}
		postFromString bool
	}{
		{
			testIdentifier: baseIdentifier + "nil",
			bodyObject:     nil,
			postFromString: false,
		},
		{
			testIdentifier: baseIdentifier + "malformed JSON",
			bodyObject:     "{\"Identifier\" :\"Something\", \"Name\":}",
			postFromString: true,
		},
		{
			testIdentifier: baseIdentifier + "endpoint.ChatColorList",
			bodyObject: &endpoint.ChatColorList{
				Colors: []string{
					"Player 1",
					"Player 2",
				},
			},
			postFromString: false,
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.testIdentifier, func(unitTest *testing.T) {
			mockCollection, testServer := newServer()

			mockCollection.ErrorToReturn = errors.New("error")

			var postResponse *httptest.ResponseRecorder
			var encodingError error
			if testCase.postFromString {
				postResponse =
					server.MockPostWithDirectBody(
						testServer,
						"/backend/player/new-player",
						testCase.bodyObject.(string))
				encodingError = nil
			} else {
				postResponse, encodingError =
					server.MockPost(testServer, "/backend/player/new-player", testCase.bodyObject)
			}

			unitTest.Logf(
				testCase.testIdentifier+"/object %v generated encoding error %v.",
				testCase.bodyObject,
				encodingError)

			assertResponseIsCorrect(
				unitTest,
				testCase.testIdentifier,
				postResponse,
				encodingError,
				http.StatusBadRequest)

			assertNoFunctionWasCalled(
				unitTest,
				mockCollection,
				testCase.testIdentifier)
		})
	}

}

func TestRejectNewPlayerIfCollectionRejectsIt(unitTest *testing.T) {
	bodyObject := endpoint.PlayerState{
		Name:  "A. Player Name",
		Color: "The color",
	}

	mockCollection, testServer := newServer()

	mockCollection.ErrorToReturn = errors.New("error")
	mockCollection.ReturnForAll = testPlayerStates

	postResponse, encodingError :=
		server.MockPost(testServer, "/backend/player/new-player", bodyObject)

	testIdentifier := "Reject POST new-player if collection rejects it"

	unitTest.Logf(
		testIdentifier+"/object %v generated encoding error %v.",
		bodyObject,
		encodingError)

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		postResponse,
		encodingError,
		http.StatusBadRequest)

	functionRecord :=
		mockCollection.getFirstAndEnsureOnly(
			unitTest,
			testIdentifier)

	assertFunctionRecordIsCorrect(
		unitTest,
		functionRecord,
		functionNameAndArgument{
			FunctionName:     "Add",
			FunctionArgument: bodyObject,
		},
		testIdentifier)
}

func TestRejectNewPlayerIfIdentifierHasSegmentDelimiter(unitTest *testing.T) {
	bodyObject := endpoint.PlayerState{
		Name:  "A. Player Name",
		Color: "The color",
	}

	mockCollection, testServer := newServer()

	mockCollection.ErrorToReturn = nil
	mockCollection.ReturnForAll = testPlayerStates

	postResponse, encodingError :=
		server.MockPost(testServer, "/backend/player/new-player", bodyObject)

	testIdentifier := "Reject POST new-player if identifier has segment delimiter"

	unitTest.Logf(
		testIdentifier+"/object %v generated encoding error %v.",
		bodyObject,
		encodingError)

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		postResponse,
		encodingError,
		http.StatusBadRequest)

	functionRecord :=
		mockCollection.getFirstAndEnsureOnly(
			unitTest,
			testIdentifier)

	assertFunctionRecordIsCorrect(
		unitTest,
		functionRecord,
		functionNameAndArgument{
			FunctionName:     "Add",
			FunctionArgument: bodyObject,
		},
		testIdentifier)
}

func TestAcceptValidNewPlayer(unitTest *testing.T) {
	bodyObject := endpoint.PlayerState{
		Name:  "A. Player Name",
		Color: "The color",
	}

	mockCollection, testServer := newServer()

	mockCollection.ErrorToReturn = nil
	mockCollection.ReturnForAll = testPlayerStates

	postResponse, encodingError :=
		server.MockPost(testServer, "/backend/player/new-player", bodyObject)

	testIdentifier := "POST new-player"

	unitTest.Logf(
		testIdentifier+"/object %v generated encoding error %v.",
		bodyObject,
		encodingError)

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		postResponse,
		encodingError,
		http.StatusOK)

	expectedRecords := []functionNameAndArgument{
		functionNameAndArgument{
			FunctionName:     "Add",
			FunctionArgument: []string{bodyObject.Name, bodyObject.Color},
		},
		functionNameAndArgument{
			FunctionName:     "All",
			FunctionArgument: testPlayerList,
		},
	}

	assertFunctionRecordsAreCorrect(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		expectedRecords,
		testIdentifier)
}

func TestRejectInvalidUpdatePlayerObject(unitTest *testing.T) {
	baseIdentifier := "Reject invalid POST update-player with body = "

	testCases := []struct {
		testIdentifier string
		bodyObject     interface{}
		postFromString bool
	}{
		{
			testIdentifier: baseIdentifier + "nil",
			bodyObject:     nil,
			postFromString: false,
		},
		{
			testIdentifier: baseIdentifier + "malformed JSON",
			bodyObject:     "{\"Identifier\" :\"Something\", \"Name\":}",
			postFromString: true,
		},
		{
			testIdentifier: baseIdentifier + "endpoint.ChatColorList",
			bodyObject: &endpoint.ChatColorList{
				Colors: []string{
					"Player 1",
					"Player 2",
				},
			},
			postFromString: false,
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.testIdentifier, func(unitTest *testing.T) {
			mockCollection, testServer := newServer()

			mockCollection.ErrorToReturn = errors.New("error")

			var postResponse *httptest.ResponseRecorder
			var encodingError error
			if testCase.postFromString {
				postResponse =
					server.MockPostWithDirectBody(
						testServer,
						"/backend/player/update-player",
						testCase.bodyObject.(string))
				encodingError = nil
			} else {
				postResponse, encodingError =
					server.MockPost(testServer, "/backend/player/new-player", testCase.bodyObject)
			}

			unitTest.Logf(
				testCase.testIdentifier+"/object %v generated encoding error %v.",
				testCase.bodyObject,
				encodingError)

			assertResponseIsCorrect(
				unitTest,
				testCase.testIdentifier,
				postResponse,
				encodingError,
				http.StatusBadRequest)

			assertNoFunctionWasCalled(
				unitTest,
				mockCollection,
				testCase.testIdentifier)
		})
	}

}

func TestRejectUpdatePlayerIfCollectionRejectsIt(unitTest *testing.T) {
	bodyObject := endpoint.PlayerState{
		Name:  "A. Player Name",
		Color: "The color",
	}

	mockCollection, testServer := newServer()

	mockCollection.ErrorToReturn = errors.New("error")
	mockCollection.ReturnForAll = testPlayerStates

	postResponse, encodingError :=
		server.MockPost(testServer, "/backend/player/update-player", bodyObject)

	testIdentifier := "Reject POST update-player if collection rejects it"

	unitTest.Logf(
		testIdentifier+"/object %v generated encoding error %v.",
		bodyObject,
		encodingError)

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		postResponse,
		encodingError,
		http.StatusBadRequest)

	functionRecord :=
		mockCollection.getFirstAndEnsureOnly(
			unitTest,
			testIdentifier)

	assertFunctionRecordIsCorrect(
		unitTest,
		functionRecord,
		functionNameAndArgument{
			FunctionName:     "UpdateColor",
			FunctionArgument: []string{bodyObject.Name, bodyObject.Color},
		},
		testIdentifier)
}

func TestAcceptValidUpdatePlayer(unitTest *testing.T) {
	bodyObject := endpoint.PlayerState{
		Name:  "A. Player Name",
		Color: "The color",
	}

	mockCollection, testServer := newServer()

	mockCollection.ErrorToReturn = nil
	mockCollection.ReturnForAll = testPlayerStates

	postResponse, encodingError :=
		server.MockPost(testServer, "/backend/player/update-player", bodyObject)

	testIdentifier := "POST update-player"

	unitTest.Logf(
		testIdentifier+"/object %v generated encoding error %v.",
		bodyObject,
		encodingError)

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		postResponse,
		encodingError,
		http.StatusOK)

	expectedRecords := []functionNameAndArgument{
		functionNameAndArgument{
			FunctionName:     "UpdateColor",
			FunctionArgument: []string{bodyObject.Name, bodyObject.Color},
		},
		functionNameAndArgument{
			FunctionName:     "All",
			FunctionArgument: nil,
		},
	}

	assertFunctionRecordsAreCorrect(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		expectedRecords,
		testIdentifier)
}

// Still need to test reset.

func assertNoFunctionWasCalled(
	unitTest *testing.T,
	mockCollection *mockPlayerCollection,
	testIdentifier string) {
	if len(mockCollection.FunctionsAndArgumentsReceived) != 0 {
		unitTest.Fatalf(
			testIdentifier+": unexpectedly called player collection methods %v",
			mockCollection.FunctionsAndArgumentsReceived)
	}
}

func assertFunctionRecordIsCorrect(
	unitTest *testing.T,
	actualRecord functionNameAndArgument,
	expectedRecord functionNameAndArgument,
	testIdentifier string) {
	if actualRecord != expectedRecord {
		unitTest.Fatalf(
			testIdentifier+"/function record mismatch: actual = %v, expected = %v",
			actualRecord,
			expectedRecord)
	}
}

func assertFunctionRecordsAreCorrect(
	unitTest *testing.T,
	actualRecords []functionNameAndArgument,
	expectedRecords []functionNameAndArgument,
	testIdentifier string) {
	expectedNumberOfRecords := len(expectedRecords)

	if len(actualRecords) != expectedNumberOfRecords {
		unitTest.Fatalf(
			testIdentifier+"/function record list length mismatch: actual = %v, expected = %v",
			actualRecords,
			expectedRecords)
	}

	for recordIndex := 0; recordIndex < expectedNumberOfRecords; recordIndex++ {
		actualRecord := actualRecords[recordIndex]
		expectedRecord := expectedRecords[recordIndex]
		if actualRecord != expectedRecord {
			unitTest.Fatalf(
				testIdentifier+
					"/function record[%v] mismatch: actual = %v, expected = %v (list: actual = %v, expected = %v)",
				recordIndex,
				actualRecord,
				expectedRecord,
				actualRecords,
				expectedRecords)
		}
	}
}

func assertResponseIsCorrect(
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

	if responseRecorder == nil {
		unitTest.Fatalf(testIdentifier + "/endpoint returned nil response.")
	}

	if responseRecorder.Code != expectedCode {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			expectedCode,
			responseRecorder.Code)
	}
}
