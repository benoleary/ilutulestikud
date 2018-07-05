package player_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	player_state "github.com/benoleary/ilutulestikud/backend/player"
	"github.com/benoleary/ilutulestikud/backend/server/endpoint/parsing"
	player_endpoint "github.com/benoleary/ilutulestikud/backend/server/endpoint/player"
)

// segmentTranslatorForTest returns the standard base-32 translator.
func segmentTranslatorForTest() parsing.SegmentTranslator {
	return &parsing.Base32Translator{}
}

func decoderAroundDefaultPlayer(
	unitTest *testing.T,
	testIdentifier string) *json.Decoder {
	defaultBodyObject := parsing.PlayerState{
		Name:  "Player Name",
		Color: "Chat color",
	}

	return DecoderAroundInterface(unitTest, testIdentifier, defaultBodyObject)
}

var testPlayerList parsing.PlayerList = parsing.PlayerList{
	Players: []parsing.PlayerState{
		parsing.PlayerState{
			Identifier: segmentTranslatorForTest().ToSegment(testPlayerStates[0].Name()),
			Name:       testPlayerStates[0].Name(),
			Color:      testPlayerStates[0].Color(),
		},
		parsing.PlayerState{
			Identifier: segmentTranslatorForTest().ToSegment(testPlayerStates[1].Name()),
			Name:       testPlayerStates[1].Name(),
			Color:      testPlayerStates[1].Color(),
		},
		parsing.PlayerState{
			Identifier: segmentTranslatorForTest().ToSegment(testPlayerStates[2].Name()),
			Name:       testPlayerStates[2].Name(),
			Color:      testPlayerStates[2].Color(),
		},
	},
}

type mockPlayerCollection struct {
	FunctionsAndArgumentsReceived []functionNameAndArgument
	ErrorToReturn                 error
	ReturnForAll                  []player_state.ReadonlyState
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
func (mockCollection *mockPlayerCollection) All() []player_state.ReadonlyState {
	mockCollection.recordFunctionAndArgument(
		"All",
		nil)
	return mockCollection.ReturnForAll
}

// Get gets mocked.
func (mockCollection *mockPlayerCollection) Get(playerIdentifier string) (player_state.ReadonlyState, error) {
	mockCollection.recordFunctionAndArgument(
		"playerIdentifier",
		playerIdentifier)
	return nil, mockCollection.ErrorToReturn
}

// Delete gets mocked.
func (mockCollection *mockPlayerCollection) Delete(playerName string) error {
	mockCollection.recordFunctionAndArgument(
		"Delete",
		playerName)
	return mockCollection.ErrorToReturn
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

// newPlayerCollectionAndHandler prepares a mock player collection and uses it to
// prepare a player_endpoint.Handler with the default endpoint segment translator
// for the tests, in a consistent way for the tests of the player endpoints,
// returning the mock collection and the endpoint handler.
func newPlayerCollectionAndHandler() (*mockPlayerCollection, *player_endpoint.Handler) {
	return newPlayerCollectionAndHandlerForTranslator(segmentTranslatorForTest())
}

// newPlayerCollectionAndHandlerForTranslator prepares a mock player collection and
// uses it to prepare a player_endpoint.Handler with the given endpoint segment
// translator in a consistent way for the tests of the player endpoints, returning
// the mock collection and the endpoint handler.
func newPlayerCollectionAndHandlerForTranslator(
	segmentTranslator parsing.SegmentTranslator) (
	*mockPlayerCollection, *player_endpoint.Handler) {
	mockCollection := &mockPlayerCollection{}

	handlerForPlayer := player_endpoint.New(mockCollection, segmentTranslator)

	return mockCollection, handlerForPlayer
}

func TestGetPlayerNilFutherSegmentSliceBadRequest(unitTest *testing.T) {
	testIdentifier := "GET with nil segment slice after player"
	mockCollection, testHandler := newPlayerCollectionAndHandler()
	_, responseCode := testHandler.HandleGet(nil)

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			responseCode)
	}

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestGetPlayerEmptyFutherSegmentSliceBadRequest(unitTest *testing.T) {
	testIdentifier := "GET with empty segment slice after player"
	mockCollection, testHandler := newPlayerCollectionAndHandler()
	_, responseCode := testHandler.HandleGet([]string{})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			responseCode)
	}

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestGetPlayerInvalidSegmentNotFound(unitTest *testing.T) {
	testIdentifier := "GET player/invalid-segment"
	mockCollection, testHandler := newPlayerCollectionAndHandler()

	_, responseCode :=
		testHandler.HandleGet([]string{"invalid-segment"})

	if responseCode != http.StatusNotFound {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusNotFound,
			responseCode)
	}

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestPostPlayerNilFutherSegmentSliceBadRequest(unitTest *testing.T) {
	testIdentifier := "POST with nil segment slice after player"
	mockCollection, testHandler := newPlayerCollectionAndHandler()

	_, responseCode :=
		testHandler.HandlePost(decoderAroundDefaultPlayer(unitTest, testIdentifier), nil)

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			responseCode)
	}

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestPostPlayerEmptyFutherSegmentSliceBadRequest(unitTest *testing.T) {
	testIdentifier := "POST with empty segment slice after player"
	mockCollection, testHandler := newPlayerCollectionAndHandler()

	_, responseCode :=
		testHandler.HandlePost(
			decoderAroundDefaultPlayer(unitTest, testIdentifier),
			[]string{})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			responseCode)
	}

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestPostPlayerInvalidSegmentNotFound(unitTest *testing.T) {
	testIdentifier := "POST player/invalid-segment"
	mockCollection, testHandler := newPlayerCollectionAndHandler()

	_, responseCode :=
		testHandler.HandlePost(
			decoderAroundDefaultPlayer(unitTest, testIdentifier),
			[]string{"invalid-segment"})

	if responseCode != http.StatusNotFound {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusNotFound,
			responseCode)
	}

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestPlayerListDelivered(unitTest *testing.T) {
	testIdentifier := "GET registered-players"
	mockCollection, testHandler := newPlayerCollectionAndHandler()
	mockCollection.ReturnForAll = testPlayerStates

	returnedInterface, responseCode :=
		testHandler.HandleGet([]string{"registered-players"})

	if returnedInterface == nil {
		unitTest.Fatalf(
			testIdentifier+"/returned nil along with code %v",
			responseCode)
	}

	if responseCode != http.StatusOK {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			responseCode)
	}

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

	responsePlayerList, isInterfaceCorrect := returnedInterface.(parsing.PlayerList)

	if !isInterfaceCorrect {
		unitTest.Fatalf(
			testIdentifier+"/received %v instead of expected parsing.PlayerList",
			returnedInterface)
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

	expectedColors :=
		[]string{
			"red",
			"green",
			"blue",
		}

	mockCollection, testHandler := newPlayerCollectionAndHandler()
	mockCollection.ReturnForAvailableChatColors = expectedColors
	mockCollection.ReturnForAll = testPlayerStates

	returnedInterface, responseCode :=
		testHandler.HandleGet([]string{"available-colors"})

	if returnedInterface == nil {
		unitTest.Fatalf(
			testIdentifier+"/returned nil along with code %v",
			responseCode)
	}

	if responseCode != http.StatusOK {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			responseCode)
	}

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

	responseColorList, isInterfaceCorrect := returnedInterface.(parsing.ChatColorList)

	if !isInterfaceCorrect {
		unitTest.Fatalf(
			testIdentifier+"/received %v instead of expected parsing.ChatColorList",
			returnedInterface)
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
	mockCollection, testHandler := newPlayerCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("error")

	// There is no point testing with valid JSON objects which do not correspond
	// to the expected JSON object, as the JSON will just be parsed with empty
	// strings for the missing attributes and extra attributes will just be
	// ignored. The tests of the player state collection can cover the cases of
	// empty player names and colors.
	bodyString := "{\"Identifier\" :\"Something\", \"Name\":}"

	bodyDecoder := json.NewDecoder(bytes.NewReader(bytes.NewBufferString(bodyString).Bytes()))

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"new-player"})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			responseCode)
	}

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestRejectNewPlayerIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier := "Reject POST new-player if collection rejects it"
	mockCollection, testHandler := newPlayerCollectionAndHandler()

	mockCollection.ErrorToReturn = errors.New("error")
	mockCollection.ReturnForAll = testPlayerStates

	bodyObject := parsing.PlayerState{
		Name:  "A. Player Name",
		Color: "The color",
	}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"new-player"})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			responseCode)
	}

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
	mockCollection, testHandler :=
		newPlayerCollectionAndHandlerForTranslator(&parsing.Base64Translator{})
	mockCollection.ErrorToReturn = nil
	mockCollection.ReturnForAll = testPlayerStates

	// breaksBase64 is a string which encodes in base 64 to a string which contains
	// a '/' character, which should in turn break the system which expects to be able
	// to parse identifiers from URI segments delimited by the '/' character.
	// It should unescape to \/\\\? as a literal.
	breaksBase64 := "\\/\\\\\\?"

	bodyObject := parsing.PlayerState{
		Name:  breaksBase64,
		Color: "The color",
	}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"new-player"})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			responseCode)
	}

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
	mockCollection, testHandler := newPlayerCollectionAndHandler()
	mockCollection.ErrorToReturn = nil
	mockCollection.ReturnForAll = testPlayerStates

	bodyObject := parsing.PlayerState{
		Name:  "A. Player Name",
		Color: "The color",
	}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"new-player"})

	if responseCode != http.StatusOK {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			responseCode)
	}

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
	mockCollection, testHandler := newPlayerCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("error")

	// There is no point testing with valid JSON objects which do not correspond
	// to the expected JSON object, as the JSON will just be parsed with empty
	// strings for the missing attributes and extra attributes will just be
	// ignored. The tests of the player state collection can cover the cases of
	// empty player names and colors.
	bodyString := "{\"Identifier\" :\"Something\", \"Name\":}"

	bodyDecoder := json.NewDecoder(bytes.NewReader(bytes.NewBufferString(bodyString).Bytes()))

	_, responseCode := testHandler.HandlePost(bodyDecoder, []string{"update-player"})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			responseCode)
	}

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestRejectUpdatePlayerIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier := "Reject POST update-player if collection rejects it"
	mockCollection, testHandler := newPlayerCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("error")
	mockCollection.ReturnForAll = testPlayerStates

	bodyObject := parsing.PlayerState{
		Name:  "A. Player Name",
		Color: "The color",
	}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"update-player"})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			responseCode)
	}

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
	mockCollection, testHandler := newPlayerCollectionAndHandler()
	mockCollection.ErrorToReturn = nil
	mockCollection.ReturnForAll = testPlayerStates

	bodyObject := parsing.PlayerState{
		Name:  "A. Player Name",
		Color: "The color",
	}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"update-player"})

	if responseCode != http.StatusOK {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			responseCode)
	}

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
	mockCollection, testHandler := newPlayerCollectionAndHandler()

	_, responseCode :=
		testHandler.HandlePost(nil, []string{"reset-players"})

	if responseCode != http.StatusOK {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			responseCode)
	}

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
