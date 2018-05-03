package game_test

// This file tests the github.com/benoleary/ilutulestikud/backend/server/endpoint/game package,
// but does not import it directly, as it is most convenient to use a server.State which contains
// a struct which came from that package. On the other hand, the
// github.com/benoleary/ilutulestikud/backend/game package must be imported for the purposes of
// comparisons within the tests.
import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/chat"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/player"
	"github.com/benoleary/ilutulestikud/backend/server"
	"github.com/benoleary/ilutulestikud/backend/server/endpoint/parsing"
)

// mockGameDefinition takes up to five players, not as an array so that
// the default comparison works.
type mockGameDefinition struct {
	GameName           string
	RulesetDescription string
	FirstPlayerName    string
	SecondPlayerName   string
	ThirdPlayerName    string
	FourthPlayerName   string
	FifthPlayerName    string
}

type mockGameState struct {
	mockName    string
	mockPlayers []player.ReadonlyState
	mockTurn    int
	mockChatLog *chat.Log
}

// Name gets mocked.
func (mockGame mockGameState) Name() string {
	return mockGame.mockName
}

// Ruleset gets mocked.
func (mockGame mockGameState) Ruleset() game.Ruleset {
	return &game.StandardWithoutRainbowRuleset{}
}

// Players gets mocked.
func (mockGame mockGameState) Players() []player.ReadonlyState {
	return mockGame.mockPlayers
}

// Turn gets mocked.
func (mockGame mockGameState) Turn() int {
	return mockGame.mockTurn
}

// CreationTime gets mocked.
func (mockGame mockGameState) CreationTime() time.Time {
	return time.Now()
}

// ChatLog gets mocked.
func (mockGame mockGameState) ChatLog() *chat.Log {
	return mockGame.mockChatLog
}

// Score gets mocked.
func (mockGame mockGameState) Score() int {
	return 1
}

// NumberOfReadyHints gets mocked.
func (mockGame mockGameState) NumberOfReadyHints() int {
	return 2
}

// NumberOfMistakesMade gets mocked.
func (mockGame mockGameState) NumberOfMistakesMade() int {
	return 3
}

// DeckSize gets mocked.
func (mockGame mockGameState) DeckSize() int {
	return 4
}

// LastPlayedForColor gets mocked.
func (mockGame mockGameState) LastPlayedForColor(colorSuit string) game.ReadonlyCard {
	return nil
}

// NumberOfDiscardedCards gets mocked.
func (mockGame mockGameState) NumberOfDiscardedCards(colorSuit string, sequenceIndex int) int {
	return 5
}

// VisibleCardInHand gets mocked.
func (mockGame mockGameState) VisibleCardInHand(
	holdingPlayerName string,
	indexInHand int) (game.ReadonlyCard, error) {
	return nil, nil
}

// InferredCardInHand gets mocked.
func (mockGame mockGameState) InferredCardInHand(
	holdingPlayerName string,
	indexInHand int) (game.InferredCard, error) {
	return game.InferredCard{}, nil
}

type mockGameCollection struct {
	FunctionsAndArgumentsReceived []functionNameAndArgument
	ErrorToReturn                 error
	ReturnForViewAllWithPlayer    []*game.PlayerView
	ReturnForViewState            *game.PlayerView
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
	return mockCollection.ReturnForViewState, mockCollection.ErrorToReturn
}

// ViewAllWithPlayer gets mocked.
func (mockCollection *mockGameCollection) ViewAllWithPlayer(
	playerName string) ([]*game.PlayerView, error) {
	mockCollection.recordFunctionAndArgument(
		"ViewAllWithPlayer",
		playerName)
	return mockCollection.ReturnForViewAllWithPlayer, mockCollection.ErrorToReturn
}

// RecordChatMessage gets mocked.
func (mockCollection *mockGameCollection) RecordChatMessage(
	gameName string,
	playerName string,
	chatMessage string) error {
	mockCollection.recordFunctionAndArgument(
		"RecordChatMessage",
		stringTriple{first: gameName, second: playerName, third: chatMessage})
	return mockCollection.ErrorToReturn
}

// AddNew gets mocked.
func (mockCollection *mockGameCollection) AddNew(
	gameName string,
	gameRuleset game.Ruleset,
	playerNames []string) error {
	functionArgument := mockGameDefinition{
		GameName:           gameName,
		RulesetDescription: gameRuleset.FrontendDescription(),
	}

	numberOfPLayers := len(playerNames)

	if numberOfPLayers > 0 {
		functionArgument.FirstPlayerName = playerNames[0]
	}

	if numberOfPLayers > 1 {
		functionArgument.SecondPlayerName = playerNames[1]
	}
	if numberOfPLayers > 2 {
		functionArgument.ThirdPlayerName = playerNames[2]
	}

	if numberOfPLayers > 3 {
		functionArgument.FourthPlayerName = playerNames[3]
	}

	if numberOfPLayers > 4 {
		functionArgument.FifthPlayerName = playerNames[4]
	}

	mockCollection.recordFunctionAndArgument(
		"AddNew",
		functionArgument)
	return mockCollection.ErrorToReturn
}

// newGameCollectionAndServer prepares a mock game collection and uses it to
// prepare a server.State with the default endpoint segment translator for
// the tests, in a consistent way for the tests of the player endpoints,
// returning the mock collection and the server state.
func newGameCollectionAndServer() (*mockGameCollection, *server.State) {
	return newGameCollectionAndServerForTranslator(segmentTranslatorForTest())
}

// newGameCollectionAndServerForTranslator prepares a mock game collection and
// uses it to prepare a server.State with the given endpoint segment translator
// in a consistent way for the tests of the game endpoints, returning the mock
// collection and the server state.
func newGameCollectionAndServerForTranslator(
	segmentTranslator parsing.SegmentTranslator) (*mockGameCollection, *server.State) {
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

func TestGetAllGamesWithPlayerRejectedIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier := "GET all-games-with-player rejected by collection"
	mockCollection, testServer := newGameCollectionAndServer()

	mockCollection.ErrorToReturn = errors.New("error")

	mockPlayerName := "Mock Mock"
	mockPlayerIdentifier := segmentTranslatorForTest().ToSegment(mockPlayerName)

	getResponse := mockGet(testServer, "/backend/game/all-games-with-player/"+mockPlayerIdentifier)

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		getResponse,
		nil,
		http.StatusBadRequest)

	functionRecord :=
		mockCollection.getFirstAndEnsureOnly(
			unitTest,
			testIdentifier)

	assertFunctionRecordIsCorrect(
		unitTest,
		functionRecord,
		functionNameAndArgument{
			FunctionName:     "ViewAllWithPlayer",
			FunctionArgument: mockPlayerName,
		},
		testIdentifier)
}

func TestGetAllGamesWithPlayerWhenEmptyList(unitTest *testing.T) {
	testIdentifier := "GET all-games-with-player with invalid identifier"
	mockCollection, testServer := newGameCollectionAndServer()

	mockCollection.ReturnForViewAllWithPlayer = make([]*game.PlayerView, 0)

	mockPlayerName := "Mock Mock"
	mockPlayerIdentifier := segmentTranslatorForTest().ToSegment(mockPlayerName)

	getResponse := mockGet(testServer, "/backend/game/all-games-with-player/"+mockPlayerIdentifier)

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		getResponse,
		nil,
		http.StatusOK)

	bodyDecoder := json.NewDecoder(getResponse.Body)

	var responseTurnSummaryList endpoint.TurnSummaryList
	parsingError := bodyDecoder.Decode(&responseTurnSummaryList)
	if parsingError != nil {
		unitTest.Fatalf(
			testIdentifier+"/error parsing JSON from HTTP response body: %v",
			parsingError)
	}

	if len(responseTurnSummaryList.TurnSummaries) != 0 {
		unitTest.Fatalf(
			testIdentifier+"/empty game view list did not produce empty turn summary list: %v",
			responseTurnSummaryList)
	}

	functionRecord :=
		mockCollection.getFirstAndEnsureOnly(
			unitTest,
			testIdentifier)

	assertFunctionRecordIsCorrect(
		unitTest,
		functionRecord,
		functionNameAndArgument{
			FunctionName:     "ViewAllWithPlayer",
			FunctionArgument: mockPlayerName,
		},
		testIdentifier)
}

func TestGetAllGamesWithPlayerWhenThreeGames(unitTest *testing.T) {
	testIdentifier := "GET all-games-with-player with invalid identifier"
	mockCollection, testServer := newGameCollectionAndServer()

	firstTestGame := &mockGameState{
		mockName:    "first test game",
		mockPlayers: testPlayerStates,
		mockTurn:    1,
	}

	firstTestView, errorForFirstView :=
		game.ViewForPlayer(firstTestGame, testPlayerStates[0].Name())

	if errorForFirstView != nil {
		unitTest.Fatalf(
			testIdentifier+"/error when creating view on first game: %v",
			errorForFirstView)
	}

	secondTestGame := &mockGameState{
		mockName:    "second test game",
		mockPlayers: testPlayerStates,
		mockTurn:    2,
	}

	secondTestView, errorForSecondView :=
		game.ViewForPlayer(secondTestGame, testPlayerStates[1].Name())

	if errorForSecondView != nil {
		unitTest.Fatalf(
			testIdentifier+"/error when creating view on second game: %v",
			errorForSecondView)
	}

	thirdTestGame := &mockGameState{
		mockName:    "third test game",
		mockPlayers: testPlayerStates,
		mockTurn:    3,
	}

	thirdTestView, errorForThirdView :=
		game.ViewForPlayer(thirdTestGame, testPlayerStates[2].Name())

	if errorForThirdView != nil {
		unitTest.Fatalf(
			testIdentifier+"/error when creating view on third game: %v",
			errorForThirdView)
	}

	expectedViews := []*game.PlayerView{
		firstTestView,
		secondTestView,
		thirdTestView,
	}

	mockCollection.ReturnForViewAllWithPlayer = expectedViews

	mockPlayerName := "Mock Player"
	mockPlayerIdentifier := segmentTranslatorForTest().ToSegment(mockPlayerName)

	getResponse := mockGet(testServer, "/backend/game/all-games-with-player/"+mockPlayerIdentifier)

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
			FunctionName:     "ViewAllWithPlayer",
			FunctionArgument: mockPlayerName,
		},
		testIdentifier)

	bodyDecoder := json.NewDecoder(getResponse.Body)

	var responseTurnSummaryList endpoint.TurnSummaryList
	parsingError := bodyDecoder.Decode(&responseTurnSummaryList)
	if parsingError != nil {
		unitTest.Fatalf(
			testIdentifier+"/error parsing JSON from HTTP response body: %v",
			parsingError)
	}

	if len(responseTurnSummaryList.TurnSummaries) != len(expectedViews) {
		unitTest.Fatalf(
			testIdentifier+"/game view list %v did not produce turn summary list %v with same number of elements",
			expectedViews,
			responseTurnSummaryList)
	}

	// The list of expected players contains no duplicates, so it suffices to compare lengths
	// and that every expected players is found.
	for _, expectedView := range expectedViews {
		foundGame := false
		for _, actualTurnSummary := range responseTurnSummaryList.TurnSummaries {
			expectedIdentifier := segmentTranslatorForTest().ToSegment(expectedView.GameName())
			_, expectedIsPlayerTurn := expectedView.CurrentTurnOrder()
			if (actualTurnSummary.GameIdentifier == expectedIdentifier) &&
				(actualTurnSummary.GameName == expectedView.GameName()) &&
				(actualTurnSummary.IsPlayerTurn == expectedIsPlayerTurn) {
				foundGame = true
			}
		}

		if !foundGame {
			unitTest.Fatalf(
				testIdentifier+
					"/returned %v which does not match the expected list of games %v"+
					" (did not find %v).",
				responseTurnSummaryList.TurnSummaries,
				expectedViews,
				expectedView)
		}
	}
}

func TestGetGameForPlayerNoFurtherSegmentBadRequest(unitTest *testing.T) {
	testIdentifier := "GET with no segments after game-as-seen-by-player"
	mockCollection, testServer := newGameCollectionAndServer()

	getResponse := mockGet(testServer, "/backend/game/game-as-seen-by-player")

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

func TestGetGameForPlayerOnlyGameSegmentBadRequest(unitTest *testing.T) {
	testIdentifier := "GET with only one segment after game-as-seen-by-player"
	mockCollection, testServer := newGameCollectionAndServer()

	mockGameName := "Mock game"
	mockGameIdentifier := segmentTranslatorForTest().ToSegment(mockGameName)

	getResponse := mockGet(testServer, "/backend/game/game-as-seen-by-player/"+mockGameIdentifier)

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

func TestGetGameForPlayerInvalidGameIdentifierBadRequest(unitTest *testing.T) {
	testIdentifier := "GET game-as-seen-by-player with invalid game identifier"
	mockCollection, testServer := newGameCollectionAndServer()

	mockPlayerName := "Mock Player"
	mockPlayerIdentifier := segmentTranslatorForTest().ToSegment(mockPlayerName)

	// The character '+' is not a valid base-32 character.
	mockGameIdentifier := "+++"

	getResponse := mockGet(testServer, "/backend/game/game-as-seen-by-player/"+mockGameIdentifier+"/"+mockPlayerIdentifier)

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

func TestGetGameForPlayerInvalidPlayerIdentifierBadRequest(unitTest *testing.T) {
	testIdentifier := "GET game-as-seen-by-player with invalid game identifier"
	mockCollection, testServer := newGameCollectionAndServer()

	// The character '+' is not a valid base-32 character.
	mockPlayerIdentifier := "+++"

	mockGameName := "Mock game"
	mockGameIdentifier := segmentTranslatorForTest().ToSegment(mockGameName)

	getResponse := mockGet(testServer, "/backend/game/game-as-seen-by-player/"+mockGameIdentifier+"/"+mockPlayerIdentifier)

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

func TestGetGameForPlayerRejectedIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier := "GET game-as-seen-by-player rejected by collection"
	mockCollection, testServer := newGameCollectionAndServer()

	mockCollection.ErrorToReturn = errors.New("error")

	mockPlayerName := "Mock Player"
	mockPlayerIdentifier := segmentTranslatorForTest().ToSegment(mockPlayerName)
	mockGameName := "Mock game"
	mockGameIdentifier := segmentTranslatorForTest().ToSegment(mockGameName)

	getResponse := mockGet(testServer, "/backend/game/game-as-seen-by-player/"+mockGameIdentifier+"/"+mockPlayerIdentifier)

	assertResponseIsCorrect(
		unitTest,
		testIdentifier,
		getResponse,
		nil,
		http.StatusBadRequest)

	functionRecord :=
		mockCollection.getFirstAndEnsureOnly(
			unitTest,
			testIdentifier)

	assertFunctionRecordIsCorrect(
		unitTest,
		functionRecord,
		functionNameAndArgument{
			FunctionName:     "ViewState",
			FunctionArgument: stringPair{first: mockGameName, second: mockPlayerName},
		},
		testIdentifier)
}

func TestGetGameForPlayer(unitTest *testing.T) {
	testIdentifier := "GET game-as-seen-by-player"
	mockCollection, testServer := newGameCollectionAndServer()

	chattingPlayer := testPlayerStates[0]
	playerName := chattingPlayer.Name()

	expectedChatLog := chat.NewLog()
	expectedChatLog.AppendNewMessage(playerName, chattingPlayer.Color(), "first message")
	expectedChatLog.AppendNewMessage(playerName, chattingPlayer.Color(), "second message")

	testGame := &mockGameState{
		mockName:    "test game",
		mockPlayers: testPlayerStates,
		mockTurn:    1,
		mockChatLog: expectedChatLog,
	}

	testView, viewError :=
		game.ViewForPlayer(testGame, playerName)
	if viewError != nil {
		unitTest.Fatalf(
			testIdentifier+"/error when creating view on test game: %v",
			viewError)
	}

	mockCollection.ReturnForViewState = testView

	mockPlayerIdentifier := segmentTranslatorForTest().ToSegment(playerName)
	mockGameName := "Mock game"
	mockGameIdentifier := segmentTranslatorForTest().ToSegment(mockGameName)

	getResponse := mockGet(testServer, "/backend/game/game-as-seen-by-player/"+mockGameIdentifier+"/"+mockPlayerIdentifier)

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
			FunctionName:     "ViewState",
			FunctionArgument: stringPair{first: mockGameName, second: playerName},
		},
		testIdentifier)

	bodyDecoder := json.NewDecoder(getResponse.Body)

	var responseGameView endpoint.GameView
	parsingError := bodyDecoder.Decode(&responseGameView)
	if parsingError != nil {
		unitTest.Fatalf(
			testIdentifier+"/error parsing JSON from HTTP response body: %v",
			parsingError)
	}

	expectedMessages := expectedChatLog.Sorted()
	numberOfExpectedMessages := len(expectedMessages)

	if len(responseGameView.ChatLog) != numberOfExpectedMessages {
		unitTest.Fatalf(
			testIdentifier+"/game view chat log %v did not have same length %v as expected chat log %v",
			responseGameView.ChatLog,
			numberOfExpectedMessages,
			expectedMessages)
	}

	for messageIndex := 0; messageIndex < numberOfExpectedMessages; messageIndex++ {
		expectedMessage := expectedMessages[messageIndex]
		actualMessage := responseGameView.ChatLog[messageIndex]
		if (actualMessage.TimestampInSeconds != expectedMessage.CreationTime.Unix()) ||
			(actualMessage.ChatColor != expectedMessage.ChatColor) ||
			(actualMessage.PlayerName != expectedMessage.PlayerName) ||
			(actualMessage.MessageText != expectedMessage.MessageText) {
			unitTest.Fatalf(
				testIdentifier+"/game view chat log %v was not same as expected chat log %v, differed in element %v",
				responseGameView.ChatLog,
				expectedMessages,
				messageIndex)
		}
	}

	if (responseGameView.ScoreSoFar != testView.Score()) ||
		(responseGameView.NumberOfReadyHints != testView.NumberOfReadyHints()) ||
		(responseGameView.NumberOfSpentHints != testView.NumberOfSpentHints()) ||
		(responseGameView.NumberOfMistakesStillAllowed != testView.NumberOfMistakesStillAllowed()) ||
		(responseGameView.NumberOfMistakesMade != testView.NumberOfMistakesMade()) {
		unitTest.Fatalf(
			testIdentifier+"/game view %v was not same as expected view %v",
			responseGameView,
			testView)
	}
}

func TestRejectInvalidChatWithMalformedRequest(unitTest *testing.T) {
	testIdentifier := "Reject invalid POST record-chat-message with malformed JSON body"

	// There is no point testing with valid JSON objects which do not correspond
	// to the expected JSON object, as the JSON will just be parsed with empty
	// strings for the missing attributes and extra attributes will just be
	// ignored. The tests of the game state collection can cover the cases of
	// empty attributes.
	bodyString := "{\"PlayerName\" :\"Something\", \"GameName\":}"

	mockCollection, testServer := newGameCollectionAndServer()

	mockCollection.ErrorToReturn = errors.New("error")

	postResponse :=
		mockPostWithDirectBody(
			testServer,
			"/backend/game/record-chat-message",
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

func TestRejectChatIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier := "Reject POST record-chat-message if collection rejects it"
	mockCollection, testServer := newGameCollectionAndServer()

	mockCollection.ErrorToReturn = errors.New("error")

	bodyObject := endpoint.PlayerChatMessage{
		GameName:    "Test game",
		PlayerName:  "A. Player Name",
		ChatMessage: "Blah blah blah",
	}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/game/record-chat-message", bodyObject)

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
			FunctionName:     "RecordChatMessage",
			FunctionArgument: stringTriple{first: bodyObject.GameName, second: bodyObject.PlayerName, third: bodyObject.ChatMessage},
		},
		testIdentifier)
}

func TestAcceptValidChat(unitTest *testing.T) {
	testIdentifier := "POST record-chat-message"
	mockCollection, testServer := newGameCollectionAndServer()

	bodyObject := endpoint.PlayerChatMessage{
		GameName:    "Test game",
		PlayerName:  "A. Player Name",
		ChatMessage: "Blah blah blah",
	}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/game/record-chat-message", bodyObject)

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

	functionRecord :=
		mockCollection.getFirstAndEnsureOnly(
			unitTest,
			testIdentifier)

	assertFunctionRecordIsCorrect(
		unitTest,
		functionRecord,
		functionNameAndArgument{
			FunctionName:     "RecordChatMessage",
			FunctionArgument: stringTriple{first: bodyObject.GameName, second: bodyObject.PlayerName, third: bodyObject.ChatMessage},
		},
		testIdentifier)
}

func TestRejectInvalidNewGameWithMalformedRequest(unitTest *testing.T) {
	testIdentifier := "Reject invalid POST create-new-game with malformed JSON body"

	// There is no point testing with valid JSON objects which do not correspond
	// to the expected JSON object, as the JSON will just be parsed with empty
	// strings for the missing attributes and extra attributes will just be
	// ignored. The tests of the game state collection can cover the cases of
	// empty attributes.
	bodyString := "{\"GameName\" :\"Something\", \"PlayerIdentifiers\":}"

	mockCollection, testServer := newGameCollectionAndServer()

	mockCollection.ErrorToReturn = errors.New("error")

	postResponse :=
		mockPostWithDirectBody(
			testServer,
			"/backend/game/create-new-game",
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

func TestRejectNewGameWithInvalidRulesetIdentifier(unitTest *testing.T) {
	testIdentifier := "Reject POST create-new-game with invalid ruleset identifier"
	mockCollection, testServer := newGameCollectionAndServer()

	// All the valid ruleset identifiers should be > 0.
	bodyObject := endpoint.GameDefinition{
		GameName:          "test game",
		RulesetIdentifier: -1,
		PlayerNames:       []string{"Player One", "Player Two"},
	}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/game/create-new-game", bodyObject)

	unitTest.Logf(
		testIdentifier+"/object %v generated encoding error %v.",
		bodyObject,
		encodingError)

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

func TestRejectNewGameIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier := "Reject POST create-new-game if collection rejects it"
	mockCollection, testServer := newGameCollectionAndServer()

	mockCollection.ErrorToReturn = errors.New("error")

	bodyObject := endpoint.GameDefinition{
		GameName:          "test game",
		RulesetIdentifier: game.ValidRulesetIdentifiers()[0],
		PlayerNames:       []string{"Player One", "Player Two"},
	}

	expectedRuleset, rulesetError :=
		game.RulesetFromIdentifier(bodyObject.RulesetIdentifier)

	if rulesetError != nil {
		unitTest.Fatalf(
			testIdentifier+"/error when getting valid expected ruleset: %v",
			rulesetError)
	}

	expectedFunctionArgument :=
		mockGameDefinition{
			GameName:           bodyObject.GameName,
			RulesetDescription: expectedRuleset.FrontendDescription(),
			FirstPlayerName:    bodyObject.PlayerNames[0],
			SecondPlayerName:   bodyObject.PlayerNames[1],
		}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/game/create-new-game", bodyObject)

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
			FunctionName:     "AddNew",
			FunctionArgument: expectedFunctionArgument,
		},
		testIdentifier)
}

func TestRejectNewGameIfIdentifierIncludesSegmentDelimiter(unitTest *testing.T) {
	testIdentifier := "Reject POST create-new-game if identifier includes segment delimiter"
	mockCollection, testServer :=
		newGameCollectionAndServerForTranslator(&parsing.NoOperationTranslator{})

	bodyObject := endpoint.GameDefinition{
		GameName:          "name/which/cannot/work/as/identifier",
		RulesetIdentifier: game.ValidRulesetIdentifiers()[0],
		PlayerNames:       []string{"Player One", "Player Two"},
	}

	expectedRuleset, rulesetError :=
		game.RulesetFromIdentifier(bodyObject.RulesetIdentifier)

	if rulesetError != nil {
		unitTest.Fatalf(
			testIdentifier+"/error when getting valid expected ruleset: %v",
			rulesetError)
	}

	expectedFunctionArgument :=
		mockGameDefinition{
			GameName:           bodyObject.GameName,
			RulesetDescription: expectedRuleset.FrontendDescription(),
			FirstPlayerName:    bodyObject.PlayerNames[0],
			SecondPlayerName:   bodyObject.PlayerNames[1],
		}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/game/create-new-game", bodyObject)

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
			FunctionName:     "AddNew",
			FunctionArgument: expectedFunctionArgument,
		},
		testIdentifier)
}

func TestAcceptValidNewGame(unitTest *testing.T) {
	testIdentifier := "POST record-chat-message"
	mockCollection, testServer := newGameCollectionAndServer()

	bodyObject := endpoint.GameDefinition{
		GameName:          "test game",
		RulesetIdentifier: game.ValidRulesetIdentifiers()[0],
		PlayerNames:       []string{"Player One", "Player Two"},
	}

	expectedRuleset, rulesetError :=
		game.RulesetFromIdentifier(bodyObject.RulesetIdentifier)

	if rulesetError != nil {
		unitTest.Fatalf(
			testIdentifier+"/error when getting valid expected ruleset: %v",
			rulesetError)
	}

	expectedFunctionArgument :=
		mockGameDefinition{
			GameName:           bodyObject.GameName,
			RulesetDescription: expectedRuleset.FrontendDescription(),
			FirstPlayerName:    bodyObject.PlayerNames[0],
			SecondPlayerName:   bodyObject.PlayerNames[1],
		}

	postResponse, encodingError :=
		mockPost(testServer, "/backend/game/create-new-game", bodyObject)

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

	functionRecord :=
		mockCollection.getFirstAndEnsureOnly(
			unitTest,
			testIdentifier)

	assertFunctionRecordIsCorrect(
		unitTest,
		functionRecord,
		functionNameAndArgument{
			FunctionName:     "AddNew",
			FunctionArgument: expectedFunctionArgument,
		},
		testIdentifier)
}
