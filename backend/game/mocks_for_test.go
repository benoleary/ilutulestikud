package game_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game/message"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// This file defines mock implementations of interfaces.
// I could have used a package to avoid all this, but I learned about
// github.com/golang/mock a bit too late, re-inventing the wheel is a
// good way to learn about stuff, and I want to avoid 3rd-party
// dependencies as much as I can in the backend.

type mockPlayerState struct {
	MockName  string
	MockColor string
}

// Name gets mocked.
func (mockPlayer *mockPlayerState) Name() string {
	return mockPlayer.MockName
}

// Color gets mocked.
func (mockPlayer *mockPlayerState) Color() string {
	return mockPlayer.MockColor
}

type mockPlayerProvider struct {
	MockPlayers map[string]*mockPlayerState
}

func NewMockPlayerProvider(initialPlayers []string) *mockPlayerProvider {
	mockProvider := &mockPlayerProvider{
		MockPlayers: make(map[string]*mockPlayerState, 0),
	}

	for _, initialPlayer := range initialPlayers {
		mockProvider.MockPlayers[initialPlayer] = &mockPlayerState{
			MockName:  initialPlayer,
			MockColor: mockChatColor,
		}
	}

	return mockProvider
}

func (mockProvider *mockPlayerProvider) Get(
	playerName string) (player.ReadonlyState, error) {
	mockPlayer, isInMap := mockProvider.MockPlayers[playerName]

	if !isInMap {
		return nil, fmt.Errorf("not in map")
	}

	return mockPlayer, nil
}

type argumentsForRecordChatMessage struct {
	NameString    string
	ColorString   string
	MessageString string
}

type argumentsForEnactTurn struct {
	MessageString string
	PlayerState   player.ReadonlyState
	IndexInt      int
	DrawnInferred card.Inferred
	HintsInt      int
	MistakesInt   int
}

// mockGameState mocks the game.ReadAndWriteState, causing test failures if
// writing functions have not been explicitly allowed, but not imposing such
// a restriction on read-only functions.
type mockGameState struct {
	testReference                                  *testing.T
	ReturnForNontestError                          error
	ReturnForName                                  string
	ReturnForRuleset                               game.Ruleset
	ReturnForPlayerNames                           []string
	ReturnForCreationTime                          time.Time
	ReturnForChatLog                               []message.Readonly
	ReturnForActionLog                             []message.Readonly
	ReturnForTurn                                  int
	ReturnForScore                                 int
	ReturnForNumberOfReadyHints                    int
	ReturnForNumberOfMistakesMade                  int
	ReturnForDeckSize                              int
	ReturnForPlayedForColor                        map[string][]card.Readonly
	ReturnForNumberOfDiscardedCards                map[card.Readonly]int
	ReturnForVisibleHand                           map[string][]card.Readonly
	ReturnForInferredHand                          map[string][]card.Inferred
	TestErrorForRecordChatMessage                  error
	ArgumentsFromRecordChatMessage                 []argumentsForRecordChatMessage
	TestErrorForEnactTurnByDiscardingAndReplacing  error
	ArgumentsFromEnactTurnByDiscardingAndReplacing []argumentsForEnactTurn
	TestErrorForEnactTurnByPlayingAndReplacing     error
	ArgumentsFromEnactTurnByPlayingAndReplacing    []argumentsForEnactTurn
}

func NewMockGameState(testReference *testing.T) *mockGameState {
	testError := fmt.Errorf("No write function should be called")
	return &mockGameState{
		testReference:                                  testReference,
		ReturnForNontestError:                          nil,
		ReturnForName:                                  "",
		ReturnForRuleset:                               nil,
		ReturnForPlayerNames:                           nil,
		ReturnForCreationTime:                          time.Now(),
		ReturnForChatLog:                               nil,
		ReturnForActionLog:                             nil,
		ReturnForTurn:                                  -1,
		ReturnForScore:                                 -1,
		ReturnForNumberOfReadyHints:                    -1,
		ReturnForNumberOfMistakesMade:                  -1,
		ReturnForDeckSize:                              -1,
		ReturnForPlayedForColor:                        make(map[string][]card.Readonly, 0),
		ReturnForNumberOfDiscardedCards:                make(map[card.Readonly]int, 0),
		ReturnForVisibleHand:                           make(map[string][]card.Readonly, 0),
		ReturnForInferredHand:                          make(map[string][]card.Inferred, 0),
		TestErrorForRecordChatMessage:                  testError,
		ArgumentsFromRecordChatMessage:                 make([]argumentsForRecordChatMessage, 0),
		TestErrorForEnactTurnByDiscardingAndReplacing:  testError,
		ArgumentsFromEnactTurnByDiscardingAndReplacing: make([]argumentsForEnactTurn, 0),
		TestErrorForEnactTurnByPlayingAndReplacing:     testError,
		ArgumentsFromEnactTurnByPlayingAndReplacing:    make([]argumentsForEnactTurn, 0),
	}
}

// Name gets mocked.
func (mockGame *mockGameState) Name() string {
	return mockGame.ReturnForName
}

// Ruleset gets mocked.
func (mockGame *mockGameState) Ruleset() game.Ruleset {
	return mockGame.ReturnForRuleset
}

// PlayerNames gets mocked.
func (mockGame *mockGameState) PlayerNames() []string {
	return mockGame.ReturnForPlayerNames
}

// CreationTime gets mocked.
func (mockGame *mockGameState) CreationTime() time.Time {
	return mockGame.ReturnForCreationTime
}

// ActionLog gets mocked.
func (mockGame *mockGameState) ActionLog() []message.Readonly {
	return mockGame.ReturnForActionLog
}

// ChatLog gets mocked.
func (mockGame *mockGameState) ChatLog() []message.Readonly {
	return mockGame.ReturnForChatLog
}

// Turn gets mocked.
func (mockGame *mockGameState) Turn() int {
	return mockGame.ReturnForTurn
}

// Score gets mocked.
func (mockGame *mockGameState) Score() int {
	return mockGame.ReturnForScore
}

// NumberOfReadyHints gets mocked.
func (mockGame *mockGameState) NumberOfReadyHints() int {
	return mockGame.ReturnForNumberOfReadyHints
}

// NumberOfMistakesMade gets mocked.
func (mockGame *mockGameState) NumberOfMistakesMade() int {
	return mockGame.ReturnForNumberOfMistakesMade
}

// DeckSize gets mocked.
func (mockGame *mockGameState) DeckSize() int {
	return mockGame.ReturnForDeckSize
}

// PlayedForColor gets mocked.
func (mockGame *mockGameState) PlayedForColor(
	colorSuit string) []card.Readonly {
	return mockGame.ReturnForPlayedForColor[colorSuit]
}

// NumberOfDiscardedCards gets mocked.
func (mockGame *mockGameState) NumberOfDiscardedCards(
	colorSuit string,
	sequenceIndex int) int {
	cardAsKey := card.NewReadonly(colorSuit, sequenceIndex)
	return mockGame.ReturnForNumberOfDiscardedCards[cardAsKey]
}

// VisibleHand gets mocked.
func (mockGame *mockGameState) VisibleHand(holdingPlayerName string) ([]card.Readonly, error) {
	visibleCard :=
		mockGame.ReturnForVisibleHand[holdingPlayerName]
	return visibleCard, mockGame.ReturnForNontestError
}

// InferredHand gets mocked.
func (mockGame *mockGameState) InferredHand(holdingPlayerName string) ([]card.Inferred, error) {
	inferredCard :=
		mockGame.ReturnForInferredHand[holdingPlayerName]
	return inferredCard, mockGame.ReturnForNontestError
}

// Read actually does what it is supposed to.
func (mockGame *mockGameState) Read() game.ReadonlyState {
	return mockGame
}

// RecordChatMessage gets mocked.
func (mockGame *mockGameState) RecordChatMessage(
	actingPlayer player.ReadonlyState, chatMessage string) error {
	if mockGame.TestErrorForRecordChatMessage != nil {
		mockGame.testReference.Fatalf(
			"RecordChatMessage(%v, %v): %v",
			actingPlayer,
			chatMessage,
			mockGame.TestErrorForRecordChatMessage)
	}

	mockGame.ArgumentsFromRecordChatMessage =
		append(
			mockGame.ArgumentsFromRecordChatMessage,
			argumentsForRecordChatMessage{
				NameString:    actingPlayer.Name(),
				ColorString:   actingPlayer.Color(),
				MessageString: chatMessage,
			})

	return mockGame.ReturnForNontestError
}

// EnactTurnByDiscardingAndReplacing gets mocked.
func (mockGame *mockGameState) EnactTurnByDiscardingAndReplacing(
	actionMessage string,
	actingPlayer player.ReadonlyState,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	numberOfReadyHintsToAdd int,
	numberOfMistakesMadeToAdd int) error {
	if mockGame.TestErrorForEnactTurnByDiscardingAndReplacing != nil {
		mockGame.testReference.Fatalf(
			"EnactTurnByDiscardingAndReplacing(%v, %v, %v, %v, %v, %v): %v",
			actionMessage,
			actingPlayer,
			indexInHand,
			knowledgeOfDrawnCard,
			numberOfReadyHintsToAdd,
			numberOfMistakesMadeToAdd,
			mockGame.TestErrorForEnactTurnByDiscardingAndReplacing)
	}

	mockGame.ArgumentsFromEnactTurnByDiscardingAndReplacing =
		append(
			mockGame.ArgumentsFromEnactTurnByDiscardingAndReplacing,
			argumentsForEnactTurn{
				MessageString: actionMessage,
				PlayerState:   actingPlayer,
				IndexInt:      indexInHand,
				DrawnInferred: knowledgeOfDrawnCard,
				HintsInt:      numberOfReadyHintsToAdd,
				MistakesInt:   numberOfMistakesMadeToAdd,
			})

	return mockGame.ReturnForNontestError
}

// EnactTurnByPlayingAndReplacing gets mocked.
func (mockGame *mockGameState) EnactTurnByPlayingAndReplacing(
	actionMessage string,
	actingPlayer player.ReadonlyState,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	numberOfReadyHintsToAdd int) error {
	if mockGame.TestErrorForEnactTurnByPlayingAndReplacing != nil {
		mockGame.testReference.Fatalf(
			"EnactTurnByPlayingAndReplacing(%v, %v, %v, %v, %v): %v",
			actionMessage,
			actingPlayer,
			indexInHand,
			knowledgeOfDrawnCard,
			numberOfReadyHintsToAdd,
			mockGame.TestErrorForEnactTurnByPlayingAndReplacing)
	}

	mockGame.ArgumentsFromEnactTurnByPlayingAndReplacing =
		append(
			mockGame.ArgumentsFromEnactTurnByPlayingAndReplacing,
			argumentsForEnactTurn{
				MessageString: actionMessage,
				PlayerState:   actingPlayer,
				IndexInt:      indexInHand,
				DrawnInferred: knowledgeOfDrawnCard,
				HintsInt:      numberOfReadyHintsToAdd,
				MistakesInt:   0,
			})

	return mockGame.ReturnForNontestError
}

type mockGameDefinition struct {
	gameName                           string
	chatLogLength                      int
	gameRuleset                        game.Ruleset
	playersInTurnOrderWithInitialHands []game.PlayerNameWithHand
	initialDeck                        []card.Readonly
}

type mockGamePersister struct {
	TestReference                 *testing.T
	ReturnForRandomSeed           int64
	ReturnForReadAndWriteGame     game.ReadAndWriteState
	ReturnForReadAllWithPlayer    []game.ReadonlyState
	ReturnForNontestError         error
	TestErrorForRandomSeed        error
	TestErrorForReadAndWriteGame  error
	TestErrorForReadAllWithPlayer error
	TestErrorForAddGame           error
	ArgumentsForAddGame           []mockGameDefinition
}

func NewMockGamePersister(
	testReference *testing.T,
	testError error) *mockGamePersister {
	return &mockGamePersister{
		TestReference:                 testReference,
		ReturnForRandomSeed:           -1,
		ReturnForReadAndWriteGame:     nil,
		ReturnForReadAllWithPlayer:    nil,
		ReturnForNontestError:         nil,
		TestErrorForRandomSeed:        testError,
		TestErrorForReadAndWriteGame:  testError,
		TestErrorForReadAllWithPlayer: testError,
		TestErrorForAddGame:           testError,
		ArgumentsForAddGame:           make([]mockGameDefinition, 0),
	}
}

func (mockImplementation *mockGamePersister) RandomSeed() int64 {
	if mockImplementation.TestErrorForRandomSeed != nil {
		mockImplementation.TestReference.Fatalf(
			"RandomSeed(): %v",
			mockImplementation.TestErrorForRandomSeed)
	}

	return mockImplementation.ReturnForRandomSeed
}

func (mockImplementation *mockGamePersister) ReadAndWriteGame(
	gameName string) (game.ReadAndWriteState, error) {
	if mockImplementation.TestErrorForReadAndWriteGame != nil {
		mockImplementation.TestReference.Fatalf(
			"ReadAndWriteGame(%v): %v",
			gameName,
			mockImplementation.TestErrorForReadAndWriteGame)
	}

	gameState := mockImplementation.ReturnForReadAndWriteGame

	return gameState, mockImplementation.ReturnForNontestError
}

func (mockImplementation *mockGamePersister) ReadAllWithPlayer(
	playerName string) []game.ReadonlyState {
	if mockImplementation.TestErrorForReadAllWithPlayer != nil {
		mockImplementation.TestReference.Fatalf(
			"ReadAllWithPlayer(%v): %v",
			playerName,
			mockImplementation.TestErrorForReadAllWithPlayer)
	}

	return mockImplementation.ReturnForReadAllWithPlayer
}

func (mockImplementation *mockGamePersister) AddGame(
	gameName string,
	chatLogLength int,
	initialActionLog []message.Readonly,
	gameRuleset game.Ruleset,
	playersInTurnOrderWithInitialHands []game.PlayerNameWithHand,
	initialDeck []card.Readonly) error {
	if mockImplementation.TestErrorForAddGame != nil {
		mockImplementation.TestReference.Fatalf(
			"AddGame(%v, %v, %v, %v, %v, %v): %v",
			gameName,
			chatLogLength,
			initialActionLog,
			gameRuleset,
			playersInTurnOrderWithInitialHands,
			initialDeck,
			mockImplementation.TestErrorForAddGame)
	}

	addedGame :=
		mockGameDefinition{
			gameName:                           gameName,
			chatLogLength:                      chatLogLength,
			gameRuleset:                        gameRuleset,
			playersInTurnOrderWithInitialHands: playersInTurnOrderWithInitialHands,
			initialDeck:                        initialDeck,
		}

	mockImplementation.ArgumentsForAddGame =
		append(mockImplementation.ArgumentsForAddGame, addedGame)

	return mockImplementation.ReturnForNontestError
}
