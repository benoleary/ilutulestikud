package game_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
)

// This file defines mock implementations of interfaces.
// I could have used a package to avoid all this, but I learned about
// github.com/golang/mock a bit too late, re-inventing the wheel is a
// good way to learn about stuff, and I want to avoid 3rd-party
// dependencies as much as I can in the backend.

// mockViewForPlayer does not cause any problems for functions which
// should not be called, because the point of the view interface is
// that it is read-only, so there should never be a problem with any
// tested code ever reading various properties.
type mockViewForPlayer struct {
	MockGameName                  string
	MockPlayers                   []string
	MockChatLog                   []message.FromPlayer
	MockPlayerTurnIndex           int
	MockScore                     int
	ErrorForVisibleHand           error
	ReturnForVisibleHand          []card.Defined
	ErrorMapForKnowledgeOfOwnHand map[string]error
	ReturnForKnowledgeOfOwnHand   []card.Inferred
	ReturnForPlayedCards          [][]card.Defined
}

func NewMockView() *mockViewForPlayer {
	return &mockViewForPlayer{
		MockGameName:                  "",
		MockPlayers:                   nil,
		MockChatLog:                   nil,
		MockPlayerTurnIndex:           -1,
		MockScore:                     -1,
		ErrorForVisibleHand:           nil,
		ReturnForVisibleHand:          nil,
		ErrorMapForKnowledgeOfOwnHand: make(map[string]error, 0),
		ReturnForKnowledgeOfOwnHand:   nil,
		ReturnForPlayedCards:          nil,
	}
}

// GameName gets mocked.
func (mockView *mockViewForPlayer) GameName() string {
	return mockView.MockGameName
}

// RulesetDescription gets mocked.
func (mockView *mockViewForPlayer) RulesetDescription() string {
	return ""
}

// ChatLog gets mocked.
func (mockView *mockViewForPlayer) ChatLog() []message.FromPlayer {
	return mockView.MockChatLog
}

// ActionLog gets mocked.
func (mockView *mockViewForPlayer) ActionLog() []message.FromPlayer {
	return make([]message.FromPlayer, logLengthForTest)
}

// GameIsFinished gets mocked.
func (mockView *mockViewForPlayer) GameIsFinished() bool {
	return false
}

// CurrentTurnOrder gets mocked.
func (mockView *mockViewForPlayer) CurrentTurnOrder() ([]string, int, int) {
	return mockView.MockPlayers, mockView.MockPlayerTurnIndex, 0
}

// Turn gets mocked.
func (mockView *mockViewForPlayer) Turn() int {
	return -1
}

// Score gets mocked.
func (mockView *mockViewForPlayer) Score() int {
	return mockView.MockScore
}

// NumberOfReadyHints gets mocked.
func (mockView *mockViewForPlayer) NumberOfReadyHints() int {
	return -1
}

// MaximumNumberOfHints gets mocked.
func (mockView *mockViewForPlayer) MaximumNumberOfHints() int {
	return -1
}

// ColorsAvailableAsHint gets mocked.
func (mockView *mockViewForPlayer) ColorsAvailableAsHint() []string {
	return nil
}

// IndicesAvailableAsHint gets mocked.
func (mockView *mockViewForPlayer) IndicesAvailableAsHint() []int {
	return nil
}

// NumberOfMistakesMade gets mocked.
func (mockView *mockViewForPlayer) NumberOfMistakesMade() int {
	return -1
}

// NumberOfMistakesIndicatingGameOver gets mocked.
func (mockView *mockViewForPlayer) NumberOfMistakesIndicatingGameOver() int {
	return -1
}

// DeckSize gets mocked.
func (mockView *mockViewForPlayer) DeckSize() int {
	return -1
}

// PlayedCards gets mocked.
func (mockView *mockViewForPlayer) PlayedCards() [][]card.Defined {
	return mockView.ReturnForPlayedCards
}

// DiscardedCards gets mocked.
func (mockView *mockViewForPlayer) DiscardedCards() []card.Defined {
	return []card.Defined{}
}

// VisibleHand gets mocked.
func (mockView *mockViewForPlayer) VisibleHand(
	playerName string) ([]card.Defined, string, error) {
	return mockView.ReturnForVisibleHand, "not tested", mockView.ErrorForVisibleHand
}

// KnowledgeOfOwnHand gets mocked.
func (mockView *mockViewForPlayer) KnowledgeOfOwnHand(
	holdingPlayer string) ([]card.Inferred, error) {
	errorToReturn, _ := mockView.ErrorMapForKnowledgeOfOwnHand[holdingPlayer]
	return mockView.ReturnForKnowledgeOfOwnHand, errorToReturn
}

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

type mockActionExecutor struct {
	ErrorToReturn error
}

// RecordChatMessage gets mocked.
func (mockExecutor *mockActionExecutor) RecordChatMessage(chatMessage string) error {
	return mockExecutor.ErrorToReturn
}

// TakeTurnByDiscarding gets mocked.
func (mockExecutor *mockActionExecutor) TakeTurnByDiscarding(indexInHandToDiscard int) error {
	return mockExecutor.ErrorToReturn
}

// TakeTurnByPlaying gets mocked.
func (mockExecutor *mockActionExecutor) TakeTurnByPlaying(indexInHandToDiscard int) error {
	return mockExecutor.ErrorToReturn
}

// TakeTurnByHintingColor gets mocked.
func (mockExecutor *mockActionExecutor) TakeTurnByHintingColor(
	receivingPlayer string,
	hintedColor string) error {
	return mockExecutor.ErrorToReturn
}

// TakeTurnByHintingIndex gets mocked.
func (mockExecutor *mockActionExecutor) TakeTurnByHintingIndex(
	receivingPlayer string,
	hintedIndex int) error {
	return mockExecutor.ErrorToReturn
}

type mockGameCollection struct {
	FunctionsAndArgumentsReceived []functionNameAndArgument
	ErrorToReturn                 error
	ReturnForViewAllWithPlayer    []game.ViewForPlayer
	ReturnForViewState            game.ViewForPlayer
	ReturnForExecuteAction        game.ExecutorForPlayer
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
	playerName string) (game.ViewForPlayer, error) {
	mockCollection.recordFunctionAndArgument(
		"ViewState",
		stringPair{first: gameName, second: playerName})
	return mockCollection.ReturnForViewState, mockCollection.ErrorToReturn
}

// ViewAllWithPlayer gets mocked.
func (mockCollection *mockGameCollection) ViewAllWithPlayer(
	playerName string) ([]game.ViewForPlayer, error) {
	mockCollection.recordFunctionAndArgument(
		"ViewAllWithPlayer",
		playerName)
	return mockCollection.ReturnForViewAllWithPlayer, mockCollection.ErrorToReturn
}

// ExecuteAction gets mocked.
func (mockCollection *mockGameCollection) ExecuteAction(
	gameName string,
	playerName string) (game.ExecutorForPlayer, error) {
	mockCollection.recordFunctionAndArgument(
		"ExecuteAction",
		stringPair{first: gameName, second: playerName})
	return mockCollection.ReturnForExecuteAction, mockCollection.ErrorToReturn
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

	numberOfPlayers := len(playerNames)

	if numberOfPlayers > 0 {
		functionArgument.FirstPlayerName = playerNames[0]
	}

	if numberOfPlayers > 1 {
		functionArgument.SecondPlayerName = playerNames[1]
	}
	if numberOfPlayers > 2 {
		functionArgument.ThirdPlayerName = playerNames[2]
	}

	if numberOfPlayers > 3 {
		functionArgument.FourthPlayerName = playerNames[3]
	}

	if numberOfPlayers > 4 {
		functionArgument.FifthPlayerName = playerNames[4]
	}

	mockCollection.recordFunctionAndArgument(
		"AddNew",
		functionArgument)
	return mockCollection.ErrorToReturn
}

// RemoveGameFromListForPlayer gets mocked.
func (mockCollection *mockGameCollection) RemoveGameFromListForPlayer(
	gameName string,
	playerName string) error {
	mockCollection.recordFunctionAndArgument(
		"RemoveGameFromListForPlayer",
		stringPair{first: gameName, second: playerName})
	return mockCollection.ErrorToReturn
}

// Delete gets mocked.
func (mockCollection *mockGameCollection) Delete(gameName string) error {
	mockCollection.recordFunctionAndArgument(
		"Delete",
		gameName)
	return mockCollection.ErrorToReturn
}
