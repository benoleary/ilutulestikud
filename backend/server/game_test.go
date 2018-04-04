package server_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/server"
)

type mockGameCollection struct {
	FunctionsAndArgumentsReceived []functionNameAndArgument
	ErrorToReturn                 error
}

func (mockCollection *mockGameCollection) recordFunctionAndArgument(
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

func (mockCollection *mockGameCollection) getFirstAndEnsureOnly(
	unitTest *testing.T,
	testIdentifier string) functionNameAndArgument {
	if len(mockCollection.FunctionsAndArgumentsReceived) != 1 {
		unitTest.Fatalf(
			testIdentifier+
				"/mock game collection recorded %v function calls, expected 1.",
			mockCollection.FunctionsAndArgumentsReceived)
	}

	return mockCollection.FunctionsAndArgumentsReceived[0]
}

// ViewState gets mocked.
func (mockCollection *mockGameCollection) ViewState(
	gameName string,
	playerName string) (*game.PlayerView, error) {
	mockCollection.recordFunctionAndArgument(
		"ViewState",
		stringPair{first: gameName, second: playerName})
	return nil, mockCollection.ErrorToReturn
}

// ViewAllWithPlayer gets mocked.
func (mockCollection *mockGameCollection) ViewAllWithPlayer(
	playerName string) ([]*game.PlayerView, error) {
	mockCollection.recordFunctionAndArgument(
		"ViewAllWithPlayer",
		playerName)
	return nil, mockCollection.ErrorToReturn
}

// PerformAction gets mocked.
func (mockCollection *mockGameCollection) PerformAction(
	playerAction endpoint.PlayerAction) error {
	mockCollection.recordFunctionAndArgument(
		"PerformAction",
		playerAction)
	return mockCollection.ErrorToReturn
}

// AddNew gets mocked.
func (mockCollection *mockGameCollection) AddNew(
	gameDefinition endpoint.GameDefinition) error {
	mockCollection.recordFunctionAndArgument(
		"AddNew",
		gameDefinition)
	return mockCollection.ErrorToReturn
}

// newGameCollectionAndServer prepares a mock game collection and uses it to prepare a
// server.State with the default endpoint segment translator for the tests,
// in a consistent way for the tests of the player endpoints, returning the
// mock collection and the server state.
func newGameCollectionAndServer() (*mockGameCollection, *server.State) {
	return newGameCollectionAndServerForTranslator(segmentTranslatorForTest())
}

// newGameCollectionAndServerForTranslator prepares a mock game collection and uses it to
// prepare a server.State with the given endpoint segment translator in a
// consistent way for the tests of the game endpoints, returning the
// mock collection and the server state.
func newGameCollectionAndServerForTranslator(
	segmentTranslator server.EndpointSegmentTranslator) (*mockGameCollection, *server.State) {
	mockCollection := &mockGameCollection{}

	serverState :=
		server.New(
			"test",
			segmentTranslator,
			nil,
			mockCollection)

	return mockCollection, serverState
}

func TestGetGameNoFurtherSegmentBadRequest(unitTest *testing.T) {
	testIdentifier := "GET with no segments after game"
	mockCollection, testServer := newGameCollectionAndServer()

	getResponse := mockGet(testServer, "/backend/game")

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

func TestGetGameInvalidSegmentNotFound(unitTest *testing.T) {
	testIdentifier := "GET game/invalid-segment"
	mockCollection, testServer := newGameCollectionAndServer()

	getResponse :=
		mockGet(testServer, "/backend/game/invalid-segment")

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

func TestPostGameNoFurtherSegmentBadRequest(unitTest *testing.T) {
	testIdentifier := "POST with no segments after game"
	mockCollection, testServer := newGameCollectionAndServer()

	bodyObject := endpoint.PlayerState{
		Name:  "Player Name",
		Color: "Chat color",
	}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/game", bodyObject)

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

func TestPostGameInvalidSegmentNotFound(unitTest *testing.T) {
	testIdentifier := "POST game/invalid-segment"
	mockCollection, testServer := newGameCollectionAndServer()

	bodyObject := endpoint.PlayerState{
		Name:  "Player Name",
		Color: "Chat color",
	}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/game/invalid-segment", bodyObject)

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

func TestAvailableRulesetsCorrectlyDelivered(unitTest *testing.T) {
	testIdentifier := "GET available-rulesets"
	mockCollection, testServer := newPlayerCollectionAndServer()

	expectedRulesetIdentifiers := game.ValidRulesetIdentifiers()

	getResponse :=
		mockGet(testServer, "/backend/game/available-rulesets")

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		getResponse,
		nil,
		http.StatusOK)

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)

	bodyDecoder := json.NewDecoder(getResponse.Body)

	var responseRulesetList endpoint.RulesetList
	parsingError := bodyDecoder.Decode(&responseRulesetList)
	if parsingError != nil {
		unitTest.Fatalf(
			testIdentifier+"/error parsing JSON from HTTP response body: %v",
			parsingError)
	}

	if responseRulesetList.Rulesets == nil {
		unitTest.Fatalf(
			testIdentifier+"/returned %v which has a nil list of rulesets.",
			responseRulesetList)
	}

	if len(responseRulesetList.Rulesets) != len(expectedRulesetIdentifiers) {
		unitTest.Fatalf(
			testIdentifier+
				"/returned %v which does not match the expected list of ruleset identifiers %v.",
			responseRulesetList,
			expectedRulesetIdentifiers)
	}

	// The list of expected rulesets contains no duplicates, so it suffices to compare lengths
	// and that every expected ruleset is found.
	for _, expectedRulesetIdentifier := range expectedRulesetIdentifiers {
		expectedRuleset, identificationError := game.RulesetFromIdentifier(expectedRulesetIdentifier)
		if identificationError != nil {
			unitTest.Fatalf(
				testIdentifier+
					"/valid ruleset identifier %v produced an error when fetching ruleset: %v",
				expectedRulesetIdentifier,
				identificationError)
		}

		foundRuleset := false
		for _, actualRuleset := range responseRulesetList.Rulesets {
			if (actualRuleset.Identifier == expectedRulesetIdentifier) &&
				(actualRuleset.Description == expectedRuleset.FrontendDescription()) &&
				(actualRuleset.MinimumNumberOfPlayers == expectedRuleset.MinimumNumberOfPlayers()) &&
				(actualRuleset.MaximumNumberOfPlayers == expectedRuleset.MaximumNumberOfPlayers()) {
				foundRuleset = true
			}
		}

		if !foundRuleset {
			unitTest.Fatalf(
				testIdentifier+
					"/returned %v which does not match the expected list of ruleset identifiers %v"+
					" (did not find %v).",
				responseRulesetList,
				expectedRulesetIdentifiers,
				expectedRuleset)
		}
	}
}

func TestGetAllGamesWithPlayerNoFurtherSegmentBadRequest(unitTest *testing.T) {
	testIdentifier := "GET with no segments after all-games-with-player"
	mockCollection, testServer := newGameCollectionAndServer()

	getResponse := mockGet(testServer, "/backend/game/all-games-with-player")

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

func TestGetAllGamesWithPlayerInvalidPlayerIdentifierBadRequest(unitTest *testing.T) {
	testIdentifier := "GET all-games-with-player with invalid identifier"
	mockCollection, testServer := newGameCollectionAndServer()

	// The character '+' is not a valid base-32 character.
	getResponse := mockGet(testServer, "/backend/game/all-games-with-player/++++")

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
