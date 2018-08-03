package game_test

// This file tests the github.com/benoleary/ilutulestikud/backend/server/endpoint/game package,
// but does not import it directly, as it is most convenient to use a server.State which contains
// a struct which came from that package. On the other hand, the
// github.com/benoleary/ilutulestikud/backend/game package must be imported for the purposes of
// comparisons within the tests.
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/defaults"
	game_state "github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
	game_endpoint "github.com/benoleary/ilutulestikud/backend/server/endpoint/game"
	"github.com/benoleary/ilutulestikud/backend/server/endpoint/parsing"
)

const logLengthForTest = 8

var colorsAvailableInTest []string = defaults.AvailableColors()

var testPlayers = []string{
	"Player One",
	"Player Two",
	"Player Three",
	"Player Four",
}

// segmentTranslatorForTest returns the standard base-32 translator.
func segmentTranslatorForTest() parsing.SegmentTranslator {
	return &parsing.Base32Translator{}
}

// newGameCollectionAndHandler prepares a mock game collection and uses it to
// prepare a server.State with the default endpoint segment translator for
// the tests, in a consistent way for the tests of the player endpoints,
// returning the mock collection and the server state.
func newGameCollectionAndHandler() (*mockGameCollection, *game_endpoint.Handler) {
	return newGameCollectionAndHandlerForTranslator(segmentTranslatorForTest())
}

// newGameCollectionAndHandlerForTranslator prepares a mock game collection and
// uses it to prepare a server.State with the given endpoint segment translator
// in a consistent way for the tests of the game endpoints, returning the mock
// collection and the server state.
func newGameCollectionAndHandlerForTranslator(
	segmentTranslator parsing.SegmentTranslator) (*mockGameCollection, *game_endpoint.Handler) {
	mockCollection := &mockGameCollection{}

	handlerForGame := game_endpoint.New(mockCollection, segmentTranslator)

	return mockCollection, handlerForGame
}

func TestGetGameNilFutherSegmentSliceBadRequest(unitTest *testing.T) {
	testIdentifier := "GET with nil segment slice after game"
	mockCollection, testHandler := newGameCollectionAndHandler()

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

func TestGetGameEmptyFutherSegmentSliceBadRequest(unitTest *testing.T) {
	testIdentifier := "GET with empty segment slice after game"
	mockCollection, testHandler := newGameCollectionAndHandler()

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

func TestGetGameInvalidSegmentNotFound(unitTest *testing.T) {
	testIdentifier := "GET game/invalid-segment"
	mockCollection, testHandler := newGameCollectionAndHandler()

	_, responseCode := testHandler.HandleGet([]string{"invalid-segment"})

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

func TestPostGameNilFutherSegmentSliceBadRequest(unitTest *testing.T) {
	testIdentifier := "POST with nil segment slice after game"
	mockCollection, testHandler := newGameCollectionAndHandler()

	_, responseCode :=
		testHandler.HandlePost(
			DecoderAroundInterface(
				unitTest,
				testIdentifier,
				parsing.GameDefinition{GameName: "test"}),
			nil)

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

func TestPostGameEmptyFutherSegmentSliceBadRequest(unitTest *testing.T) {
	testIdentifier := "POST with empty segment slice after game"
	mockCollection, testHandler := newGameCollectionAndHandler()

	_, responseCode :=
		testHandler.HandlePost(
			DecoderAroundInterface(
				unitTest,
				testIdentifier,
				parsing.GameDefinition{GameName: "test"}),
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

func TestPostGameInvalidSegmentNotFound(unitTest *testing.T) {
	testIdentifier := "POST game/invalid-segment"
	mockCollection, testHandler := newGameCollectionAndHandler()

	_, responseCode :=
		testHandler.HandlePost(
			DecoderAroundInterface(
				unitTest,
				testIdentifier,
				parsing.GameDefinition{GameName: "test"}),
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

func TestAvailableRulesetsCorrectlyDelivered(unitTest *testing.T) {
	testIdentifier := "GET available-rulesets"
	mockCollection, testHandler := newGameCollectionAndHandler()

	expectedRulesetIdentifiers := game_state.ValidRulesetIdentifiers()

	returnedInterface, responseCode :=
		testHandler.HandleGet([]string{"available-rulesets"})

	if responseCode != http.StatusOK {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			responseCode)
	}

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)

	responseRulesetList, isInterfaceCorrect :=
		returnedInterface.(parsing.RulesetList)

	if !isInterfaceCorrect {
		unitTest.Fatalf(
			testIdentifier+"/received %v instead of expected parsing.RulesetList",
			returnedInterface)
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
		expectedRuleset, identificationError :=
			game_state.RulesetFromIdentifier(expectedRulesetIdentifier)
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
	mockCollection, testHandler := newGameCollectionAndHandler()

	_, responseCode := testHandler.HandleGet([]string{"all-games-with-player"})

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

func TestGetAllGamesWithPlayerInvalidPlayerIdentifierBadRequest(unitTest *testing.T) {
	testIdentifier := "GET all-games-with-player with invalid identifier"
	mockCollection, testHandler := newGameCollectionAndHandler()

	// The character '+' is not a valid base-32 character.
	_, responseCode := testHandler.HandleGet([]string{"all-games-with-player", "++++"})

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

func TestGetAllGamesWithPlayerRejectedIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier := "GET all-games-with-player rejected by collection"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	mockPlayerName := "Mock MacMock"
	mockPlayerIdentifier := segmentTranslatorForTest().ToSegment(mockPlayerName)

	_, responseCode :=
		testHandler.HandleGet([]string{"all-games-with-player", mockPlayerIdentifier})

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
			FunctionName:     "ViewAllWithPlayer",
			FunctionArgument: mockPlayerName,
		},
		testIdentifier)
}

func TestGetAllGamesWithPlayerWhenEmptyList(unitTest *testing.T) {
	testIdentifier := "GET all-games-with-player when empty list"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ReturnForViewAllWithPlayer = make([]game_state.ViewForPlayer, 0)

	mockPlayerName := "Mock MacMock"
	mockPlayerIdentifier := segmentTranslatorForTest().ToSegment(mockPlayerName)

	returnedInterface, responseCode :=
		testHandler.HandleGet([]string{"all-games-with-player", mockPlayerIdentifier})

	if responseCode != http.StatusOK {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			responseCode)
	}

	responseTurnSummaryList, isInterfaceCorrect :=
		returnedInterface.(parsing.TurnSummaryList)

	if !isInterfaceCorrect {
		unitTest.Fatalf(
			testIdentifier+"/received %v instead of expected parsing.TurnSummaryList",
			returnedInterface)
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
	testIdentifier := "GET all-games-with-player when three games"
	mockCollection, testHandler := newGameCollectionAndHandler()

	firstTestView := NewMockView()
	firstTestView.MockPlayers = testPlayers
	firstTestView.MockPlayerTurnIndex = 0

	secondTestView := NewMockView()
	secondTestView.MockPlayers =
		[]string{
			testPlayers[1],
			testPlayers[0],
		}
	secondTestView.MockPlayerTurnIndex = 1

	thirdTestView := NewMockView()
	thirdTestView.MockPlayers =
		[]string{
			testPlayers[0],
			testPlayers[1],
		}
	thirdTestView.MockPlayerTurnIndex = 2

	expectedViews :=
		[]game_state.ViewForPlayer{
			firstTestView,
			secondTestView,
			thirdTestView,
		}

	mockCollection.ReturnForViewAllWithPlayer = expectedViews

	mockPlayerName := "Mock Player"
	mockPlayerIdentifier := segmentTranslatorForTest().ToSegment(mockPlayerName)

	returnedInterface, responseCode :=
		testHandler.HandleGet([]string{"all-games-with-player", mockPlayerIdentifier})

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
			FunctionName:     "ViewAllWithPlayer",
			FunctionArgument: mockPlayerName,
		},
		testIdentifier)

	responseTurnSummaryList, isInterfaceCorrect :=
		returnedInterface.(parsing.TurnSummaryList)

	if !isInterfaceCorrect {
		unitTest.Fatalf(
			testIdentifier+"/received %v instead of expected parsing.TurnSummaryList",
			returnedInterface)
	}

	if len(responseTurnSummaryList.TurnSummaries) != len(expectedViews) {
		unitTest.Fatalf(
			testIdentifier+
				"/game view list %v did not produce turn summary list %v with same number of elements",
			expectedViews,
			responseTurnSummaryList)
	}

	// The list of expected players contains no duplicates, so it suffices to compare lengths
	// and that every expected players is found.
	for _, expectedView := range expectedViews {
		foundGame := false
		for _, actualTurnSummary := range responseTurnSummaryList.TurnSummaries {
			expectedIdentifier :=
				segmentTranslatorForTest().ToSegment(expectedView.GameName())
			_, expectedPlayerTurnIndex, _ := expectedView.CurrentTurnOrder()
			if (actualTurnSummary.GameIdentifier == expectedIdentifier) &&
				(actualTurnSummary.GameName == expectedView.GameName()) &&
				(actualTurnSummary.IsPlayerTurn == (expectedPlayerTurnIndex == 0)) {
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
	mockCollection, testHandler := newGameCollectionAndHandler()

	_, responseCode :=
		testHandler.HandleGet([]string{"game-as-seen-by-player"})

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

func TestGetGameForPlayerOnlyGameSegmentBadRequest(unitTest *testing.T) {
	testIdentifier := "GET with only one segment after game-as-seen-by-player"
	mockCollection, testHandler := newGameCollectionAndHandler()

	mockGameName := "Mock game"
	mockGameIdentifier := segmentTranslatorForTest().ToSegment(mockGameName)

	_, responseCode :=
		testHandler.HandleGet([]string{"game-as-seen-by-player", mockGameIdentifier})

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

func TestGetGameForPlayerInvalidGameIdentifierBadRequest(unitTest *testing.T) {
	testIdentifier := "GET game-as-seen-by-player with invalid game identifier"
	mockCollection, testHandler := newGameCollectionAndHandler()

	mockPlayerName := "Mock Player"
	mockPlayerIdentifier := segmentTranslatorForTest().ToSegment(mockPlayerName)

	// The character '+' is not a valid base-32 character.
	mockGameIdentifier := "+++"

	segmentSlice :=
		[]string{"game-as-seen-by-player", mockGameIdentifier, mockPlayerIdentifier}
	_, responseCode := testHandler.HandleGet(segmentSlice)

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

func TestGetGameForPlayerInvalidPlayerIdentifierBadRequest(unitTest *testing.T) {
	testIdentifier := "GET game-as-seen-by-player with invalid game identifier"
	mockCollection, testHandler := newGameCollectionAndHandler()

	// The character '+' is not a valid base-32 character.
	mockPlayerIdentifier := "+++"

	mockGameName := "Mock game"
	mockGameIdentifier := segmentTranslatorForTest().ToSegment(mockGameName)

	segmentSlice :=
		[]string{"game-as-seen-by-player", mockGameIdentifier, mockPlayerIdentifier}
	_, responseCode := testHandler.HandleGet(segmentSlice)

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

func TestGetGameForPlayerRejectedIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier := "GET game-as-seen-by-player rejected by collection"
	mockCollection, testHandler := newGameCollectionAndHandler()

	mockCollection.ErrorToReturn = errors.New("expected error")

	mockPlayerName := "Mock Player"
	mockPlayerIdentifier := segmentTranslatorForTest().ToSegment(mockPlayerName)
	mockGameName := "Mock game"
	mockGameIdentifier := segmentTranslatorForTest().ToSegment(mockGameName)

	segmentSlice :=
		[]string{"game-as-seen-by-player", mockGameIdentifier, mockPlayerIdentifier}
	_, responseCode := testHandler.HandleGet(segmentSlice)

	if responseCode != http.StatusInternalServerError {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusInternalServerError,
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
			FunctionName:     "ViewState",
			FunctionArgument: stringPair{first: mockGameName, second: mockPlayerName},
		},
		testIdentifier)
}

func TestGetGameForPlayerRejectedIfViewStateYieldsError(unitTest *testing.T) {
	testCases :=
		[]struct {
			testName                         string
			errorForVisibleHand              error
			errorForViewerKnowledgeOfOwnHand error
			errorForOtherKnowledgeOfOwnHand  error
			errorForGameIsFinished           error
		}{
			{
				testName:                         "error from VisibleHand",
				errorForVisibleHand:              fmt.Errorf("mock error"),
				errorForViewerKnowledgeOfOwnHand: nil,
				errorForOtherKnowledgeOfOwnHand:  nil,
				errorForGameIsFinished:           nil,
			},
			{
				testName:                         "error from viewer KnowledgeOfOwnHand",
				errorForVisibleHand:              nil,
				errorForViewerKnowledgeOfOwnHand: fmt.Errorf("mock error"),
				errorForOtherKnowledgeOfOwnHand:  nil,
				errorForGameIsFinished:           nil,
			},
			{
				testName:                         "error from other KnowledgeOfOwnHand",
				errorForVisibleHand:              nil,
				errorForViewerKnowledgeOfOwnHand: nil,
				errorForOtherKnowledgeOfOwnHand:  fmt.Errorf("mock error"),
				errorForGameIsFinished:           nil,
			},
		}

	for _, testCase := range testCases {
		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			testIdentifier :=
				"GET game-as-seen-by-player getting error from ViewState/" +
					testCase.testName
			mockCollection, testHandler := newGameCollectionAndHandler()
			mockPlayerName := "Mock Player"
			mockPlayerIdentifier := segmentTranslatorForTest().ToSegment(mockPlayerName)
			mockGameName := "Mock game"
			mockGameIdentifier := segmentTranslatorForTest().ToSegment(mockGameName)

			mockView := NewMockView()
			mockView.MockPlayers = testPlayers
			mockView.MockPlayerTurnIndex = 1
			mockView.ErrorForVisibleHand = testCase.errorForVisibleHand
			mockView.ErrorMapForKnowledgeOfOwnHand[mockPlayerName] =
				testCase.errorForViewerKnowledgeOfOwnHand
			for _, participantName := range testPlayers {
				mockView.ErrorMapForKnowledgeOfOwnHand[participantName] =
					testCase.errorForOtherKnowledgeOfOwnHand
			}

			mockCollection.ReturnForViewState = mockView

			segmentSlice :=
				[]string{"game-as-seen-by-player", mockGameIdentifier, mockPlayerIdentifier}
			_, responseCode := testHandler.HandleGet(segmentSlice)

			if responseCode != http.StatusInternalServerError {
				unitTest.Fatalf(
					testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
					http.StatusInternalServerError,
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
					FunctionName:     "ViewState",
					FunctionArgument: stringPair{first: mockGameName, second: mockPlayerName},
				},
				testIdentifier)
		})
	}
}

func TestGetGameForPlayer(unitTest *testing.T) {
	testIdentifier := "GET game-as-seen-by-player"
	mockCollection, testHandler := newGameCollectionAndHandler()

	playerName := testPlayers[0]
	chatColor := "some valid color"

	expectedChatLog :=
		[]message.Readonly{
			message.NewReadonly(playerName, chatColor, "first message"),
			message.NewReadonly(playerName, chatColor, "second message"),
		}

	testView := NewMockView()
	testView.MockPlayerTurnIndex = 0
	testView.MockPlayers =
		[]string{
			testPlayers[2],
			playerName,
			testPlayers[3],
			testPlayers[1],
		}
	testView.MockChatLog = expectedChatLog
	testView.MockPlayerTurnIndex = 1
	testView.ReturnForVisibleHand =
		[]card.Defined{
			card.Defined{ColorSuit: "some color",
				SequenceIndex: 1,
			},
			card.Defined{ColorSuit: "some color",
				SequenceIndex: 2,
			},
			card.Defined{ColorSuit: "another color",
				SequenceIndex: 1,
			},
		}
	testView.ReturnForKnowledgeOfOwnHand =
		[]card.Inferred{
			card.Inferred{
				PossibleColors:  []string{"some color", "another color"},
				PossibleIndices: []int{1, 2, 3},
			},
			card.Inferred{
				PossibleColors:  []string{"some color", "yet another color"},
				PossibleIndices: []int{1, 2},
			},
			card.Inferred{
				PossibleColors:  []string{"some color"},
				PossibleIndices: []int{3},
			},
		}
	testView.ReturnForPlayedCards =
		[][]card.Defined{
			[]card.Defined{
				card.Defined{ColorSuit: "another color",
					SequenceIndex: 1,
				},
				card.Defined{ColorSuit: "another color",
					SequenceIndex: 2,
				},
				card.Defined{ColorSuit: "another color",
					SequenceIndex: 3,
				},
			},
			[]card.Defined{},
			[]card.Defined{
				card.Defined{ColorSuit: "some color",
					SequenceIndex: 1,
				},
				card.Defined{ColorSuit: "some color",
					SequenceIndex: 2,
				},
			},
		}

	mockCollection.ReturnForViewState = testView

	mockPlayerIdentifier := segmentTranslatorForTest().ToSegment(playerName)
	mockGameName := "Mock game"
	mockGameIdentifier := segmentTranslatorForTest().ToSegment(mockGameName)

	segmentSlice :=
		[]string{"game-as-seen-by-player", mockGameIdentifier, mockPlayerIdentifier}
	returnedInterface, responseCode := testHandler.HandleGet(segmentSlice)

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
			FunctionName:     "ViewState",
			FunctionArgument: stringPair{first: mockGameName, second: playerName},
		},
		testIdentifier)

	responseGameView, isInterfaceCorrect := returnedInterface.(parsing.GameView)

	if !isInterfaceCorrect {
		unitTest.Fatalf(
			testIdentifier+"/received %+v instead of expected parsing.GameView",
			returnedInterface)
	}

	numberOfExpectedMessages := len(expectedChatLog)

	if len(responseGameView.ChatLog) != numberOfExpectedMessages {
		unitTest.Fatalf(
			testIdentifier+
				"/game view chat log %+v did not have same length %v as expected chat log %+v",
			responseGameView.ChatLog,
			numberOfExpectedMessages,
			expectedChatLog)
	}

	// We only test score because it's not worth setting up all the functionality
	// of a view in the mock implementation, though we do check that the hands and
	// played cards get propagated correctly too.
	if responseGameView.ScoreSoFar != testView.Score() {
		unitTest.Fatalf(
			testIdentifier+"/game view %+v was not same as expected view %v",
			responseGameView,
			testView)
	}

	if len(responseGameView.HandsBeforeThisPlayer) != 1 {
		unitTest.Fatalf(
			testIdentifier+
				"/game view %+v had wrong number of hands before viewing player, expected %v",
			responseGameView,
			1)
	}

	assertVisibleHandCorrect(
		testIdentifier,
		unitTest,
		responseGameView.HandsBeforeThisPlayer[0],
		testView.MockPlayers[0],
		testView.ReturnForVisibleHand,
		false)

	assertInferredCardSlicesCorrect(
		testIdentifier,
		unitTest,
		responseGameView.HandOfThisPlayer,
		testView.ReturnForKnowledgeOfOwnHand)

	if len(responseGameView.HandsAfterThisPlayer) != 2 {
		unitTest.Fatalf(
			testIdentifier+
				"/game view %+v had wrong number of hands after viewing player",
			responseGameView,
			2)
	}

	assertVisibleHandCorrect(
		testIdentifier,
		unitTest,
		responseGameView.HandsAfterThisPlayer[0],
		testView.MockPlayers[2],
		testView.ReturnForVisibleHand,
		false)

	assertVisibleHandCorrect(
		testIdentifier,
		unitTest,
		responseGameView.HandsAfterThisPlayer[1],
		testView.MockPlayers[3],
		testView.ReturnForVisibleHand,
		false)

	numberOfExpectedPiles := len(testView.ReturnForPlayedCards)
	if len(responseGameView.PlayedCards) != numberOfExpectedPiles {
		unitTest.Fatalf(
			testIdentifier+
				"/game view %+v did not have expected piles of played cards %v",
			responseGameView,
			testView.ReturnForPlayedCards)
	}

	for pileIndex := 0; pileIndex < numberOfExpectedPiles; pileIndex++ {
		assertVisibleCardSlicesCorrect(
			testIdentifier,
			unitTest,
			responseGameView.PlayedCards[pileIndex],
			testView.ReturnForPlayedCards[pileIndex])
	}
}

func TestRejectInvalidNewGameWithMalformedRequest(unitTest *testing.T) {
	testIdentifier := "Reject invalid POST create-new-game with malformed JSON body"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	// There is no point testing with valid JSON objects which do not correspond
	// to the expected JSON object, as the JSON will just be parsed with empty
	// strings for the missing attributes and extra attributes will just be
	// ignored. The tests of the game state collection can cover the cases of
	// empty attributes.
	bodyString := "{\"GameName\" :\"Something\", \"PlayerIdentifiers\":}"

	bodyDecoder := json.NewDecoder(bytes.NewReader(bytes.NewBufferString(bodyString).Bytes()))

	_, responseCode := testHandler.HandlePost(bodyDecoder, []string{"create-new-game"})

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

func TestRejectNewGameWithInvalidRulesetIdentifier(unitTest *testing.T) {
	testIdentifier := "Reject POST create-new-game with invalid ruleset identifier"
	mockCollection, testHandler := newGameCollectionAndHandler()

	// All the valid ruleset identifiers should be > 0.
	bodyObject :=
		parsing.GameDefinition{
			GameName:          "test game",
			RulesetIdentifier: -1,
			PlayerNames:       []string{"Player One", "Player Two"},
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"create-new-game"})

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

func TestRejectNewGameIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier := "Reject POST create-new-game if collection rejects it"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	bodyObject :=
		parsing.GameDefinition{
			GameName:          "test game",
			RulesetIdentifier: game_state.ValidRulesetIdentifiers()[0],
			PlayerNames:       []string{"Player One", "Player Two"},
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"create-new-game"})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			responseCode)
	}

	expectedRuleset, rulesetError :=
		game_state.RulesetFromIdentifier(bodyObject.RulesetIdentifier)

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
	testIdentifier :=
		"Reject POST create-new-game if identifier includes segment delimiter"
	mockCollection, testHandler :=
		newGameCollectionAndHandlerForTranslator(&parsing.NoOperationTranslator{})

	bodyObject :=
		parsing.GameDefinition{
			GameName:          "name/which/cannot/work/as/identifier",
			RulesetIdentifier: game_state.ValidRulesetIdentifiers()[0],
			PlayerNames:       []string{"Player One", "Player Two"},
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"create-new-game"})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			responseCode)
	}

	expectedRuleset, rulesetError :=
		game_state.RulesetFromIdentifier(bodyObject.RulesetIdentifier)

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
	testIdentifier := "POST create-new-game"
	mockCollection, testHandler := newGameCollectionAndHandler()

	bodyObject :=
		parsing.GameDefinition{
			GameName:          "test game",
			RulesetIdentifier: game_state.ValidRulesetIdentifiers()[0],
			PlayerNames:       []string{"Player One", "Player Two"},
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"create-new-game"})

	if responseCode != http.StatusOK {
		unitTest.Fatalf(
			testIdentifier+
				"/did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			responseCode)
	}

	expectedRuleset, rulesetError :=
		game_state.RulesetFromIdentifier(bodyObject.RulesetIdentifier)

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

func TestRejectInvalidLeaveGameWithMalformedRequest(unitTest *testing.T) {
	testIdentifier := "Reject invalid POST leave-game with malformed JSON body"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	// There is no point testing with valid JSON objects which do not correspond
	// to the expected JSON object, as the JSON will just be parsed with empty
	// strings for the missing attributes and extra attributes will just be
	// ignored. The tests of the game state collection can cover the cases of
	// empty attributes.
	bodyString := "{\"GameName\" :\"Something\", \"PlayerName\":}"

	bodyDecoder := json.NewDecoder(bytes.NewReader(bytes.NewBufferString(bodyString).Bytes()))

	_, responseCode := testHandler.HandlePost(bodyDecoder, []string{"leave-game"})

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

func TestRejectLeaveGameIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier := "Reject POST leave-game if collection rejects it"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	bodyObject :=
		parsing.PlayerInGameIndication{
			GameName:   "test game",
			PlayerName: "Test Player",
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"leave-game"})

	if responseCode != http.StatusInternalServerError {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusInternalServerError,
			responseCode)
	}

	expectedFunctionArgument :=
		stringPair{
			first:  bodyObject.GameName,
			second: bodyObject.PlayerName,
		}

	functionRecord :=
		mockCollection.getFirstAndEnsureOnly(
			unitTest,
			testIdentifier)

	assertFunctionRecordIsCorrect(
		unitTest,
		functionRecord,
		functionNameAndArgument{
			FunctionName:     "RemoveGameFromListForPlayer",
			FunctionArgument: expectedFunctionArgument,
		},
		testIdentifier)
}

func TestAcceptValidLeaveGame(unitTest *testing.T) {
	testIdentifier := "POST leave-game"
	mockCollection, testHandler := newGameCollectionAndHandler()

	bodyObject :=
		parsing.PlayerInGameIndication{
			GameName:   "test game",
			PlayerName: "Test Player",
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"leave-game"})

	if responseCode != http.StatusOK {
		unitTest.Fatalf(
			testIdentifier+
				"/did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			responseCode)
	}

	expectedFunctionArgument :=
		stringPair{
			first:  bodyObject.GameName,
			second: bodyObject.PlayerName,
		}

	functionRecord :=
		mockCollection.getFirstAndEnsureOnly(
			unitTest,
			testIdentifier)

	assertFunctionRecordIsCorrect(
		unitTest,
		functionRecord,
		functionNameAndArgument{
			FunctionName:     "RemoveGameFromListForPlayer",
			FunctionArgument: expectedFunctionArgument,
		},
		testIdentifier)
}

func TestRejectInvalidDeleteGameWithMalformedRequest(unitTest *testing.T) {
	testIdentifier := "Reject invalid POST delete-game with malformed JSON body"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	// There is no point testing with valid JSON objects which do not correspond
	// to the expected JSON object, as the JSON will just be parsed with empty
	// strings for the missing attributes and extra attributes will just be
	// ignored. The tests of the game state collection can cover the cases of
	// empty attributes.
	bodyString := "{\"GameName\" :\"Something\", \"PlayerIdentifiers\":}"

	bodyDecoder := json.NewDecoder(bytes.NewReader(bytes.NewBufferString(bodyString).Bytes()))

	_, responseCode := testHandler.HandlePost(bodyDecoder, []string{"delete-game"})

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

func TestRejectDeleteGameIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier := "Reject POST delete-game if collection rejects it"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	bodyObject :=
		parsing.GameDefinition{
			GameName: "test game",
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"delete-game"})

	if responseCode != http.StatusInternalServerError {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			http.StatusInternalServerError,
			responseCode)
	}

	expectedFunctionArgument := bodyObject.GameName

	functionRecord :=
		mockCollection.getFirstAndEnsureOnly(
			unitTest,
			testIdentifier)

	assertFunctionRecordIsCorrect(
		unitTest,
		functionRecord,
		functionNameAndArgument{
			FunctionName:     "Delete",
			FunctionArgument: expectedFunctionArgument,
		},
		testIdentifier)
}

func TestAcceptValidDeleteGame(unitTest *testing.T) {
	testIdentifier := "POST delete-game"
	mockCollection, testHandler := newGameCollectionAndHandler()

	bodyObject :=
		parsing.GameDefinition{
			GameName: "test game",
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"delete-game"})

	if responseCode != http.StatusOK {
		unitTest.Fatalf(
			testIdentifier+
				"/did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			responseCode)
	}

	expectedFunctionArgument := bodyObject.GameName

	functionRecord :=
		mockCollection.getFirstAndEnsureOnly(
			unitTest,
			testIdentifier)

	assertFunctionRecordIsCorrect(
		unitTest,
		functionRecord,
		functionNameAndArgument{
			FunctionName:     "Delete",
			FunctionArgument: expectedFunctionArgument,
		},
		testIdentifier)
}

func TestRejectInvalidChatWithMalformedRequest(unitTest *testing.T) {
	testIdentifier := "Reject invalid POST record-chat-message with malformed JSON body"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	// There is no point testing with valid JSON objects which do not correspond
	// to the expected JSON object, as the JSON will just be parsed with empty
	// strings for the missing attributes and extra attributes will just be
	// ignored. The tests of the game state collection can cover the cases of
	// empty attributes.
	bodyString := "{\"PlayerName\" :\"Something\", \"GameName\":}"

	bodyDecoder :=
		json.NewDecoder(bytes.NewReader(bytes.NewBufferString(bodyString).Bytes()))

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"record-chat-message"})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+
				"/did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			responseCode)
	}

	assertNoFunctionWasCalled(
		unitTest,
		mockCollection.FunctionsAndArgumentsReceived,
		testIdentifier)
}

func TestRejectChatIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier := "Reject POST record-chat-message if collection rejects it"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	bodyObject :=
		parsing.PlayerChatMessage{
			PlayerInGameIndication: parsing.PlayerInGameIndication{
				GameName:   "Test game",
				PlayerName: "A. Player Name",
			},
			ChatMessage: "Blah blah blah",
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"record-chat-message"})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+
				"/did not return expected HTTP code %v, instead was %v.",
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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}

func TestPropagateErrorFromChat(unitTest *testing.T) {
	testIdentifier := "Reject POST record-chat-message if collection rejects it"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockExecutor := &mockActionExecutor{}
	mockExecutor.ErrorToReturn = fmt.Errorf("expected error")
	mockCollection.ReturnForExecuteAction = mockExecutor

	bodyObject :=
		parsing.PlayerChatMessage{
			PlayerInGameIndication: parsing.PlayerInGameIndication{
				GameName:   "Test game",
				PlayerName: "A. Player Name",
			},
			ChatMessage: "Blah blah blah",
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"record-chat-message"})

	if responseCode != http.StatusInternalServerError {
		unitTest.Fatalf(
			testIdentifier+
				"/did not return expected HTTP code %v, instead was %v.",
			http.StatusInternalServerError,
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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}

func TestAcceptValidChat(unitTest *testing.T) {
	testIdentifier := "POST record-chat-message"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ReturnForExecuteAction = &mockActionExecutor{}

	bodyObject :=
		parsing.PlayerChatMessage{
			PlayerInGameIndication: parsing.PlayerInGameIndication{
				GameName:   "Test game",
				PlayerName: "A. Player Name",
			},
			ChatMessage: "Blah blah blah",
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"record-chat-message"})

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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}

func TestRejectInvalidDiscardWithMalformedRequest(unitTest *testing.T) {
	testIdentifier :=
		"Reject invalid POST take-turn-by-discarding with malformed JSON body"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	// There is no point testing with valid JSON objects which do not correspond
	// to the expected JSON object, as the JSON will just be parsed with empty
	// strings for the missing attributes and extra attributes will just be
	// ignored. The tests of the game state collection can cover the cases of
	// empty attributes.
	bodyString := "{\"PlayerName\" :\"Something\", \"GameName\":}"

	bodyDecoder :=
		json.NewDecoder(bytes.NewReader(bytes.NewBufferString(bodyString).Bytes()))

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-discarding"})

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

func TestRejectDiscardIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier :=
		"Reject POST take-turn-by-discarding if collection rejects it"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	bodyObject :=
		parsing.PlayerCardIndication{
			PlayerInGameIndication: parsing.PlayerInGameIndication{
				GameName:   "Test game",
				PlayerName: "A. Player Name",
			},
			CardIndex: 1,
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-discarding"})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+
				"/did not return expected HTTP code %v, instead was %v.",
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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}

func TestPropagateErrorFromDiscard(unitTest *testing.T) {
	testIdentifier := "Reject POST take-turn-by-discarding if collection rejects it"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockExecutor := &mockActionExecutor{}
	mockExecutor.ErrorToReturn = fmt.Errorf("expected error")
	mockCollection.ReturnForExecuteAction = mockExecutor

	bodyObject :=
		parsing.PlayerCardIndication{
			PlayerInGameIndication: parsing.PlayerInGameIndication{
				GameName:   "Test game",
				PlayerName: "A. Player Name",
			},
			CardIndex: 1,
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-discarding"})

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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}

func TestAcceptValidDiscard(unitTest *testing.T) {
	testIdentifier := "POST take-turn-by-discarding"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ReturnForExecuteAction = &mockActionExecutor{}

	bodyObject :=
		parsing.PlayerCardIndication{
			PlayerInGameIndication: parsing.PlayerInGameIndication{
				GameName:   "Test game",
				PlayerName: "A. Player Name",
			},
			CardIndex: 1,
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-discarding"})

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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}

func TestRejectInvalidPlayWithMalformedRequest(unitTest *testing.T) {
	testIdentifier :=
		"Reject invalid POST take-turn-by-attempting-to-play with malformed JSON body"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	// There is no point testing with valid JSON objects which do not correspond
	// to the expected JSON object, as the JSON will just be parsed with empty
	// strings for the missing attributes and extra attributes will just be
	// ignored. The tests of the game state collection can cover the cases of
	// empty attributes.
	bodyString := "{\"PlayerName\" :\"Something\", \"GameName\":}"

	bodyDecoder :=
		json.NewDecoder(bytes.NewReader(bytes.NewBufferString(bodyString).Bytes()))

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-attempting-to-play"})

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

func TestRejectPlayIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier :=
		"Reject POST take-turn-by-attempting-to-play if collection rejects it"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	bodyObject :=
		parsing.PlayerCardIndication{
			PlayerInGameIndication: parsing.PlayerInGameIndication{
				GameName:   "Test game",
				PlayerName: "A. Player Name",
			},
			CardIndex: 1,
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-attempting-to-play"})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+
				"/did not return expected HTTP code %v, instead was %v.",
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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}
func TestPropagateErrorFromPlay(unitTest *testing.T) {
	testIdentifier := "Reject POST take-turn-by-attempting-to-play if collection rejects it"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockExecutor := &mockActionExecutor{}
	mockExecutor.ErrorToReturn = fmt.Errorf("expected error")
	mockCollection.ReturnForExecuteAction = mockExecutor

	bodyObject :=
		parsing.PlayerCardIndication{
			PlayerInGameIndication: parsing.PlayerInGameIndication{
				GameName:   "Test game",
				PlayerName: "A. Player Name",
			},
			CardIndex: 1,
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-attempting-to-play"})

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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}

func TestAcceptValidPlay(unitTest *testing.T) {
	testIdentifier := "POST take-turn-by-attempting-to-play"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ReturnForExecuteAction = &mockActionExecutor{}

	bodyObject :=
		parsing.PlayerCardIndication{
			PlayerInGameIndication: parsing.PlayerInGameIndication{
				GameName:   "Test game",
				PlayerName: "A. Player Name",
			},
			CardIndex: 1,
		}

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-attempting-to-play"})

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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}

func TestRejectInvalidColorHintWithMalformedRequest(unitTest *testing.T) {
	testIdentifier :=
		"Reject invalid POST take-turn-by-hinting-color with malformed JSON body"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	// There is no point testing with valid JSON objects which do not correspond
	// to the expected JSON object, as the JSON will just be parsed with empty
	// strings for the missing attributes and extra attributes will just be
	// ignored. The tests of the game state collection can cover the cases of
	// empty attributes.
	bodyString := "{\"PlayerName\" :\"Something\", \"GameName\":}"

	bodyDecoder :=
		json.NewDecoder(bytes.NewReader(bytes.NewBufferString(bodyString).Bytes()))

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-hinting-color"})

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

func TestRejectColorHintIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier :=
		"Reject POST take-turn-by-hinting-color if collection rejects it"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	bodyObject := parsing.PlayerColorHint{}
	bodyObject.GameName = "Test game"
	bodyObject.PlayerName = "A. Player Name"
	bodyObject.ReceiverName = "Another Player"
	bodyObject.HintedColor = "test color"

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-hinting-color"})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+
				"/did not return expected HTTP code %v, instead was %v.",
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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}

func TestPropagateErrorFromColorHint(unitTest *testing.T) {
	testIdentifier := "Reject POST take-turn-by-hinting-color if collection rejects it"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockExecutor := &mockActionExecutor{}
	mockExecutor.ErrorToReturn = fmt.Errorf("expected error")
	mockCollection.ReturnForExecuteAction = mockExecutor

	bodyObject := parsing.PlayerColorHint{}
	bodyObject.GameName = "Test game"
	bodyObject.PlayerName = "A. Player Name"
	bodyObject.ReceiverName = "Another Player"
	bodyObject.HintedColor = "test color"

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-hinting-color"})

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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}

func TestAcceptValidColorHint(unitTest *testing.T) {
	testIdentifier := "POST take-turn-by-hinting-color"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ReturnForExecuteAction = &mockActionExecutor{}

	bodyObject := parsing.PlayerColorHint{}
	bodyObject.GameName = "Test game"
	bodyObject.PlayerName = "A. Player Name"
	bodyObject.ReceiverName = "Another Player"
	bodyObject.HintedColor = "test color"

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-hinting-color"})

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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}

func TestRejectInvalidIndexHintWithMalformedRequest(unitTest *testing.T) {
	testIdentifier :=
		"Reject invalid POST take-turn-by-hinting-number with malformed JSON body"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	// There is no point testing with valid JSON objects which do not correspond
	// to the expected JSON object, as the JSON will just be parsed with empty
	// strings for the missing attributes and extra attributes will just be
	// ignored. The tests of the game state collection can cover the cases of
	// empty attributes.
	bodyString := "{\"PlayerName\" :\"Something\", \"GameName\":}"

	bodyDecoder :=
		json.NewDecoder(bytes.NewReader(bytes.NewBufferString(bodyString).Bytes()))

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-hinting-number"})

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

func TestRejectIndexHintIfCollectionRejectsIt(unitTest *testing.T) {
	testIdentifier :=
		"Reject POST take-turn-by-hinting-number if collection rejects it"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ErrorToReturn = errors.New("expected error")

	bodyObject := parsing.PlayerIndexHint{}
	bodyObject.GameName = "Test game"
	bodyObject.PlayerName = "A. Player Name"
	bodyObject.ReceiverName = "Another Player"
	bodyObject.HintedNumber = 1

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-hinting-number"})

	if responseCode != http.StatusBadRequest {
		unitTest.Fatalf(
			testIdentifier+
				"/did not return expected HTTP code %v, instead was %v.",
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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}

func TestPropagateErrorFromIndexHint(unitTest *testing.T) {
	testIdentifier := "Reject POST take-turn-by-hinting-number if collection rejects it"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockExecutor := &mockActionExecutor{}
	mockExecutor.ErrorToReturn = fmt.Errorf("expected error")
	mockCollection.ReturnForExecuteAction = mockExecutor

	bodyObject := parsing.PlayerIndexHint{}
	bodyObject.GameName = "Test game"
	bodyObject.PlayerName = "A. Player Name"
	bodyObject.ReceiverName = "Another Player"
	bodyObject.HintedNumber = 1

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-hinting-number"})

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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}

func TestAcceptValidIndexHint(unitTest *testing.T) {
	testIdentifier := "POST take-turn-by-hinting-number"
	mockCollection, testHandler := newGameCollectionAndHandler()
	mockCollection.ReturnForExecuteAction = &mockActionExecutor{}

	bodyObject := parsing.PlayerIndexHint{}
	bodyObject.GameName = "Test game"
	bodyObject.PlayerName = "A. Player Name"
	bodyObject.ReceiverName = "Another Player"
	bodyObject.HintedNumber = 1

	bodyDecoder := DecoderAroundInterface(unitTest, testIdentifier, bodyObject)

	_, responseCode :=
		testHandler.HandlePost(bodyDecoder, []string{"take-turn-by-hinting-number"})

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
			FunctionName:     "ExecuteAction",
			FunctionArgument: stringPair{first: bodyObject.GameName, second: bodyObject.PlayerName},
		},
		testIdentifier)
}

func assertVisibleHandCorrect(
	testIdentifier string,
	unitTest *testing.T,
	actualHand parsing.VisibleHand,
	expectedPlayer string,
	expectedCards []card.Defined,
	expectedHasTakenLastTurn bool) {
	if actualHand.PlayerName != expectedPlayer {
		unitTest.Fatalf(
			testIdentifier+
				"/actual hand %v did not have expected player name %v",
			actualHand,
			expectedPlayer)
	}

	assertVisibleCardSlicesCorrect(
		testIdentifier,
		unitTest,
		actualHand.HandCards,
		expectedCards)

	if actualHand.PlayerHasTakenLastTurn != expectedHasTakenLastTurn {
		unitTest.Fatalf(
			testIdentifier+
				"/actual hand %v did not agree with expectedHasTakenLastTurn %v",
			actualHand,
			expectedHasTakenLastTurn)
	}
}

func assertVisibleCardSlicesCorrect(
	testIdentifier string,
	unitTest *testing.T,
	actualCards []parsing.VisibleCard,
	expectedCards []card.Defined) {
	numberOfExpectedCards := len(expectedCards)

	// We compare the possible colors and indices as sets. Then it is
	// sufficient to check that the lengths are the same and that every
	// actual value is found in the map of expected values.
	if len(actualCards) != numberOfExpectedCards {
		unitTest.Fatalf(
			testIdentifier+
				"/actual card %v did not match expected cards %v",
			actualCards,
			expectedCards)
	}

	for cardIndex := 0; cardIndex < numberOfExpectedCards; cardIndex++ {
		actualCard := actualCards[cardIndex]
		expectedCard := expectedCards[cardIndex]

		if (actualCard.ColorSuit != expectedCard.ColorSuit) ||
			(actualCard.SequenceIndex != expectedCard.SequenceIndex) {
			unitTest.Fatalf(
				testIdentifier+
					"/actual card %v did not match expected cards %v",
				actualCards,
				expectedCards)
		}
	}
}

func assertInferredCardSlicesCorrect(
	testIdentifier string,
	unitTest *testing.T,
	actualCards []parsing.CardFromBehind,
	expectedCards []card.Inferred) {
	numberOfExpectedCards := len(expectedCards)

	// We compare the possible colors and indices as sets. Then it is
	// sufficient to check that the lengths are the same and that every
	// actual value is found in the map of expected values.
	if len(actualCards) != numberOfExpectedCards {
		unitTest.Fatalf(
			testIdentifier+
				"/actual card %v did not match expected cards %v",
			actualCards,
			expectedCards)
	}

	for cardIndex := 0; cardIndex < numberOfExpectedCards; cardIndex++ {
		actualCard := actualCards[cardIndex]
		expectedCard := expectedCards[cardIndex]

		assertInferredCardPossibilitiesCorrect(
			testIdentifier,
			unitTest,
			actualCard,
			expectedCard.PossibleColors,
			expectedCard.PossibleIndices)
	}
}

func assertInferredCardPossibilitiesCorrect(
	testIdentifier string,
	unitTest *testing.T,
	actualCard parsing.CardFromBehind,
	expectedColors []string,
	expectedIndices []int) {

	// We compare the possible colors and indices as sets. Then it is
	// sufficient to check that the lengths are the same and that every
	// actual value is found in the map of expected values.
	if (len(actualCard.PossibleColorSuits) != len(expectedColors)) ||
		(len(actualCard.PossibleSequenceIndices) != len(expectedIndices)) {
		unitTest.Fatalf(
			testIdentifier+
				"/inferred card %v did not match expected colors %v and indices %v",
			actualCard,
			expectedColors,
			expectedIndices)
	}

	expectedColorMap := make(map[string]bool)
	for _, expectedColor := range expectedColors {
		if expectedColorMap[expectedColor] {
			unitTest.Fatalf(
				testIdentifier+
					"/expected colors %v had duplicate(s)",
				expectedColors)
		}

		expectedColorMap[expectedColor] = true
	}

	for _, actualColor := range actualCard.PossibleColorSuits {
		if !expectedColorMap[actualColor] {
			unitTest.Fatalf(
				testIdentifier+
					"/inferred card %v did not match expected colors %v and indices %v",
				actualCard,
				expectedColors,
				expectedIndices)
		}
	}

	expectedIndexMap := make(map[int]bool)
	for _, expectedIndex := range expectedIndices {
		if expectedIndexMap[expectedIndex] {
			unitTest.Fatalf(
				testIdentifier+
					"/expected indices %v had duplicate(s)",
				expectedIndices)
		}

		expectedIndexMap[expectedIndex] = true
	}

	for _, actualIndex := range actualCard.PossibleSequenceIndices {
		if !expectedIndexMap[actualIndex] {
			unitTest.Fatalf(
				testIdentifier+
					"/inferred card %v did not match expected colors %v and indices %v",
				actualCard,
				expectedColors,
				expectedIndices)
		}
	}
}
