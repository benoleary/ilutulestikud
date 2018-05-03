package player_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/player"
	"github.com/benoleary/ilutulestikud/backend/server"
	"github.com/benoleary/ilutulestikud/backend/server/endpoint/parsing"
)

var testPlayerList endpoint.PlayerList = endpoint.PlayerList{
	Players: []endpoint.PlayerState{
		endpoint.PlayerState{
			Identifier: segmentTranslatorForTest().ToSegment(testPlayerStates[0].Name()),
			Name:       testPlayerStates[0].Name(),
			Color:      testPlayerStates[0].Color(),
		},
		endpoint.PlayerState{
			Identifier: segmentTranslatorForTest().ToSegment(testPlayerStates[1].Name()),
			Name:       testPlayerStates[1].Name(),
			Color:      testPlayerStates[1].Color(),
		},
		endpoint.PlayerState{
			Identifier: segmentTranslatorForTest().ToSegment(testPlayerStates[2].Name()),
			Name:       testPlayerStates[2].Name(),
			Color:      testPlayerStates[2].Color(),
		},
	},
}

type mockPlayerCollection struct {
	FunctionsAndArgumentsReceived []functionNameAndArgument
	ErrorToReturn                 error
	ReturnForAll                  []player.ReadonlyState
	ReturnForAvailableChatColors  []string
}

func (mockCollection *mockPlayerCollection) recordFunctionAndArgument(
	functionName string,
	functionArgument interface{}) {
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
		stringPair{first: playerName, second: chatColor})
	return mockCollection.ErrorToReturn
}

// UpdateColor gets mocked.
func (mockCollection *mockPlayerCollection) UpdateColor(
	playerName string,
	chatColor string) error {
	mockCollection.recordFunctionAndArgument(
		"UpdateColor",
		stringPair{first: playerName, second: chatColor})
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
	return nil, mockCollection.ErrorToReturn
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

// newPlayerCollectionAndServer prepares a mock player collection and uses it to prepare a
// server.State with the default endpoint segment translator for the tests,
// in a consistent way for the tests of the player endpoints, returning the
// mock collection and the server state.
func newPlayerCollectionAndServer() (*mockPlayerCollection, *server.State) {
	return newPlayerCollectionAndServerForTranslator(segmentTranslatorForTest())
}

// newPlayerCollectionAndServerForTranslator prepares a mock player collection and uses it to
// prepare a server.State with the given endpoint segment translator in a
// consistent way for the tests of the player endpoints, returning the
// mock collection and the server state.
func newPlayerCollectionAndServerForTranslator(
	segmentTranslator parsing.SegmentTranslator) (
	*mockPlayerCollection, *server.State) {
	mockCollection := &mockPlayerCollection{}

	serverState :=
		server.New(
			"test",
			segmentTranslator,
			mockCollection,
			nil)

	return mockCollection, serverState
}

func TestGetPlayerNoFurtherSegmentBadRequest(unitTest *testing.T) {
	testIdentifier := "GET with no segments after player"
	mockCollection, testServer := newPlayerCollectionAndServer()

	getResponse := mockGet(testServer, "/backend/player")

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		getResponse,
		nil,
		http.StatusBadRequest)

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestGetPlayerInvalidSegmentNotFound(unitTest *testing.T) {
	testIdentifier := "GET player/invalid-segment"
	mockCollection, testServer := newPlayerCollectionAndServer()

	getResponse :=
		mockGet(testServer, "/backend/player/invalid-segment")

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		getResponse,
		nil,
		http.StatusNotFound)

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestPostPlayerNoFurtherSegmentBadRequest(unitTest *testing.T) {
	testIdentifier := "POST with no segments after player"
	mockCollection, testServer := newPlayerCollectionAndServer()

	bodyObject := endpoint.PlayerState{
		Name:  "Player Name",
		Color: "Chat color",
	}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/player", bodyObject)

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		postResponse,
		encodingError,
		http.StatusBadRequest)

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestPostPlayerInvalidSegmentNotFound(unitTest *testing.T) {
	testIdentifier := "POST player/invalid-segment"
	mockCollection, testServer := newPlayerCollectionAndServer()

	bodyObject := endpoint.PlayerState{
		Name:  "Player Name",
		Color: "Chat color",
	}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/player/invalid-segment", bodyObject)

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		postResponse,
		encodingError,
		http.StatusNotFound)

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestPlayerListDelivered(unitTest *testing.T) {
	testIdentifier := "GET registered-players"
	mockCollection, testServer := newPlayerCollectionAndServer()

	mockCollection.ReturnForAll = testPlayerStates

	getResponse :=
		mockGet(testServer, "/backend/player/registered-players")

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

func TestAvailableColorsCorrectlyDelivered(unitTest *testing.T) {
	testIdentifier := "GET available-colors"
	mockCollection, testServer := newPlayerCollectionAndServer()

	expectedColors :=
		[]string{
			"red",
			"green",
			"blue",
		}

	mockCollection.ReturnForAvailableChatColors = expectedColors

	getResponse :=
		mockGet(testServer, "/backend/player/available-colors")

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

func TestRejectInvalidNewPlayerWithMalformedRequest(unitTest *testing.T) {
	testIdentifier := "Reject invalid POST new-player with malformed JSON body"

	// There is no point testing with valid JSON objects which do not correspond
	// to the expected JSON object, as the JSON will just be parsed with empty
	// strings for the missing attributes and extra attributes will just be
	// ignored. The tests of the player state collection can cover the cases of
	// empty player names and colors.
	bodyString := "{\"Identifier\" :\"Something\", \"Name\":}"

	mockCollection, testServer := newPlayerCollectionAndServer()

	mockCollection.ErrorToReturn = errors.New("error")

	postResponse :=
		mockPostWithDirectBody(
			testServer,
			"/backend/player/new-player",
			bodyString)

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		postResponse,
		nil,
		http.StatusBadRequest)

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestRejectNewPlayerIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier := "Reject POST new-player if collection rejects it"
	mockCollection, testServer := newPlayerCollectionAndServer()

	mockCollection.ErrorToReturn = errors.New("error")
	mockCollection.ReturnForAll = testPlayerStates

	bodyObject := endpoint.PlayerState{
		Name:  "A. Player Name",
		Color: "The color",
	}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/player/new-player", bodyObject)

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
			FunctionArgument: stringPair{first: bodyObject.Name, second: bodyObject.Color},
		},
		testIdentifier)
}

func TestRejectNewPlayerIfIdentifierHasSegmentDelimiter(unitTest *testing.T) {
	testIdentifier := "Reject POST new-player if identifier has segment delimiter"

	// We could use a no-operation translator so that it is clear that the test case has
	// a '/', but this way ensures that a translation does happen.
	mockCollection, testServer :=
		newPlayerCollectionAndServerForTranslator(&parsing.Base64Translator{})

	mockCollection.ErrorToReturn = nil
	mockCollection.ReturnForAll = testPlayerStates

	// breaksBase64 is a string which encodes in base 64 to a string which contains
	// a '/' character, which should in turn break the system which expects to be able
	// to parse identifiers from URI segments delimited by the '/' character.
	// It should unescape to \/\\\? as a literal.
	breaksBase64 := "\\/\\\\\\?"

	bodyObject := endpoint.PlayerState{
		Name:  breaksBase64,
		Color: "The color",
	}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/player/new-player", bodyObject)

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
			FunctionArgument: stringPair{first: bodyObject.Name, second: bodyObject.Color},
		},
		testIdentifier)
}

func TestAcceptValidNewPlayer(unitTest *testing.T) {
	testIdentifier := "POST new-player"
	mockCollection, testServer := newPlayerCollectionAndServer()

	mockCollection.ErrorToReturn = nil
	mockCollection.ReturnForAll = testPlayerStates

	bodyObject := endpoint.PlayerState{
		Name:  "A. Player Name",
		Color: "The color",
	}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/player/new-player", bodyObject)

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
			FunctionArgument: stringPair{first: bodyObject.Name, second: bodyObject.Color},
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

func TestRejectInvalidUpdatePlayerWithMalformedRequest(unitTest *testing.T) {
	testIdentifier := "Reject invalid POST update-player with malformed JSON body"

	// There is no point testing with valid JSON objects which do not correspond
	// to the expected JSON object, as the JSON will just be parsed with empty
	// strings for the missing attributes and extra attributes will just be
	// ignored. The tests of the player state collection can cover the cases of
	// empty player names and colors.
	bodyString := "{\"Identifier\" :\"Something\", \"Name\":}"

	mockCollection, testServer := newPlayerCollectionAndServer()

	mockCollection.ErrorToReturn = errors.New("error")

	postResponse :=
		mockPostWithDirectBody(
			testServer,
			"/backend/player/update-player",
			bodyString)

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		postResponse,
		nil,
		http.StatusBadRequest)

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestRejectUpdatePlayerIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier := "Reject POST update-player if collection rejects it"
	mockCollection, testServer := newPlayerCollectionAndServer()

	mockCollection.ErrorToReturn = errors.New("error")
	mockCollection.ReturnForAll = testPlayerStates

	bodyObject := endpoint.PlayerState{
		Name:  "A. Player Name",
		Color: "The color",
	}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/player/update-player", bodyObject)

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
			FunctionArgument: stringPair{first: bodyObject.Name, second: bodyObject.Color},
		},
		testIdentifier)
}

func TestAcceptValidUpdatePlayer(unitTest *testing.T) {
	testIdentifier := "POST update-player"
	mockCollection, testServer := newPlayerCollectionAndServer()

	mockCollection.ErrorToReturn = nil
	mockCollection.ReturnForAll = testPlayerStates

	bodyObject := endpoint.PlayerState{
		Name:  "A. Player Name",
		Color: "The color",
	}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/player/update-player", bodyObject)

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
			FunctionArgument: stringPair{first: bodyObject.Name, second: bodyObject.Color},
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

func TestResetPlayers(unitTest *testing.T) {
	testIdentifier := "POST reset-players"
	mockCollection, testServer := newPlayerCollectionAndServer()

	postResponse, encodingError :=
		mockPost(testServer, "/backend/player/reset-players", nil)

	if encodingError != nil {
		unitTest.Fatalf(
			testIdentifier+"/encoding nil produced error",
			encodingError)
	}

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		postResponse,
		encodingError,
		http.StatusOK)

	expectedRecords := []functionNameAndArgument{
		functionNameAndArgument{
			FunctionName:     "Reset",
			FunctionArgument: nil,
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
