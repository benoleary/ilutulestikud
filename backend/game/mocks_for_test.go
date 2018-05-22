package game_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/log"
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

type stringTriple struct {
	FirstString  string
	SecondString string
	ThirdString  string
}

type mockGameState struct {
	testReference                       *testing.T
	ReturnForNontestError               error
	MockName                            string
	MockRuleset                         game.Ruleset
	MockNamesAndHands                   []game.PlayerNameWithHand
	MockDeck                            []card.Readonly
	MockTurn                            int
	MockChatLog                         *log.RollingAppender
	TestErrorForName                    error
	TestErrorForRuleset                 error
	ReturnForRuleset                    game.Ruleset
	TestErrorForPlayerNames             error
	ReturnForPlayerNames                []string
	TestErrorForTurn                    error
	TestErrorForCreationTime            error
	ReturnForCreationTime               time.Time
	TestErrorForChatLog                 error
	TestErrorForActionLog               error
	TestErrorForScore                   error
	TestErrorForNumberOfReadyHints      error
	TestErrorForNumberOfMistakesMade    error
	TestErrorForDeckSize                error
	TestErrorForPlayedForColor          error
	TestErrorForNumberOfDiscardedCards  error
	TestErrorForVisibleCardInHand       error
	TestErrorForInferredCardInHand      error
	TestErrorForRecordChatMessage       error
	ArgumentsFromRecordChatMessage      []stringTriple
	TestErrorForDrawCard                error
	TestErrorForReplaceCardInHand       error
	TestErrorForAddCardToPlayedSequence error
	TestErrorForAddCardToDiscardPile    error
}

func NewMockGameState(
	testReference *testing.T,
	testError error) *mockGameState {
	return &mockGameState{
		testReference:                       testReference,
		ReturnForNontestError:               nil,
		MockName:                            "",
		MockRuleset:                         nil,
		MockNamesAndHands:                   nil,
		MockDeck:                            nil,
		MockTurn:                            -1,
		MockChatLog:                         nil,
		TestErrorForName:                    testError,
		TestErrorForRuleset:                 testError,
		ReturnForRuleset:                    nil,
		TestErrorForPlayerNames:             testError,
		ReturnForPlayerNames:                nil,
		TestErrorForTurn:                    testError,
		TestErrorForCreationTime:            testError,
		ReturnForCreationTime:               time.Now(),
		TestErrorForChatLog:                 testError,
		TestErrorForActionLog:               testError,
		TestErrorForScore:                   testError,
		TestErrorForNumberOfReadyHints:      testError,
		TestErrorForNumberOfMistakesMade:    testError,
		TestErrorForDeckSize:                testError,
		TestErrorForPlayedForColor:          testError,
		TestErrorForNumberOfDiscardedCards:  testError,
		TestErrorForVisibleCardInHand:       testError,
		TestErrorForInferredCardInHand:      testError,
		TestErrorForRecordChatMessage:       testError,
		ArgumentsFromRecordChatMessage:      make([]stringTriple, 0),
		TestErrorForDrawCard:                testError,
		TestErrorForReplaceCardInHand:       testError,
		TestErrorForAddCardToPlayedSequence: testError,
		TestErrorForAddCardToDiscardPile:    testError,
	}
}

// Name gets mocked.
func (mockGame *mockGameState) Name() string {
	if mockGame.TestErrorForName != nil {
		mockGame.testReference.Fatalf(
			"Name(): %v",
			mockGame.TestErrorForName)
	}

	return mockGame.MockName
}

// Ruleset gets mocked.
func (mockGame *mockGameState) Ruleset() game.Ruleset {
	if mockGame.TestErrorForRuleset != nil {
		mockGame.testReference.Fatalf(
			"Ruleset(): %v",
			mockGame.TestErrorForRuleset)
	}

	return mockGame.ReturnForRuleset
}

// PlayerNames gets mocked.
func (mockGame *mockGameState) PlayerNames() []string {
	if mockGame.TestErrorForPlayerNames != nil {
		mockGame.testReference.Fatalf(
			"PlayerNames(): %v",
			mockGame.TestErrorForPlayerNames)
	}

	return mockGame.ReturnForPlayerNames
}

// Turn gets mocked.
func (mockGame *mockGameState) Turn() int {
	if mockGame.TestErrorForTurn != nil {
		mockGame.testReference.Fatalf(
			"Turn(): %v",
			mockGame.TestErrorForTurn)
	}

	return mockGame.MockTurn
}

// CreationTime gets mocked.
func (mockGame *mockGameState) CreationTime() time.Time {
	if mockGame.TestErrorForCreationTime != nil {
		mockGame.testReference.Fatalf(
			"CreationTime(): %v",
			mockGame.TestErrorForCreationTime)
	}

	return mockGame.ReturnForCreationTime
}

// ChatLog gets mocked.
func (mockGame *mockGameState) ChatLog() *log.RollingAppender {
	if mockGame.TestErrorForChatLog != nil {
		mockGame.testReference.Fatalf(
			"ChatLog(): %v",
			mockGame.TestErrorForChatLog)
	}

	return nil
}

// ActionLog gets mocked.
func (mockGame *mockGameState) ActionLog() *log.RollingAppender {
	if mockGame.TestErrorForActionLog != nil {
		mockGame.testReference.Fatalf(
			"ActionLog(): %v",
			mockGame.TestErrorForActionLog)
	}

	return nil
}

// Score gets mocked.
func (mockGame *mockGameState) Score() int {
	if mockGame.TestErrorForScore != nil {
		mockGame.testReference.Fatalf(
			"Score(): %v",
			mockGame.TestErrorForScore)
	}

	return -1
}

// NumberOfReadyHints gets mocked.
func (mockGame *mockGameState) NumberOfReadyHints() int {
	if mockGame.TestErrorForNumberOfReadyHints != nil {
		mockGame.testReference.Fatalf(
			"NumberOfReadyHints(): %v",
			mockGame.TestErrorForNumberOfReadyHints)
	}

	return -1
}

// NumberOfMistakesMade gets mocked.
func (mockGame *mockGameState) NumberOfMistakesMade() int {
	if mockGame.TestErrorForNumberOfMistakesMade != nil {
		mockGame.testReference.Fatalf(
			"NumberOfMistakesMade(): %v",
			mockGame.TestErrorForNumberOfMistakesMade)
	}

	return -1
}

// DeckSize gets mocked.
func (mockGame *mockGameState) DeckSize() int {
	if mockGame.TestErrorForDeckSize != nil {
		mockGame.testReference.Fatalf(
			"DeckSize(): %v",
			mockGame.TestErrorForDeckSize)
	}

	return -1
}

// PlayedForColor gets mocked.
func (mockGame *mockGameState) PlayedForColor(
	colorSuit string) []card.Readonly {
	if mockGame.TestErrorForPlayedForColor != nil {
		mockGame.testReference.Fatalf(
			"PlayedForColor(%v): %v",
			colorSuit,
			mockGame.TestErrorForPlayedForColor)
	}

	return []card.Readonly{}
}

// NumberOfDiscardedCards gets mocked.
func (mockGame *mockGameState) NumberOfDiscardedCards(
	colorSuit string,
	sequenceIndex int) int {
	if mockGame.TestErrorForNumberOfDiscardedCards != nil {
		mockGame.testReference.Fatalf(
			"NumberOfDiscardedCards(%v, %v): %v",
			colorSuit,
			sequenceIndex,
			mockGame.TestErrorForNumberOfDiscardedCards)
	}

	return -1
}

// VisibleCardInHand gets mocked.
func (mockGame *mockGameState) VisibleCardInHand(
	holdingPlayerName string,
	indexInHand int) (card.Readonly, error) {
	if mockGame.TestErrorForVisibleCardInHand != nil {
		mockGame.testReference.Fatalf(
			"VisibleCardInHand(%v, %v): %v",
			holdingPlayerName,
			indexInHand,
			mockGame.TestErrorForVisibleCardInHand)
	}

	return card.ErrorReadonly(), nil
}

// InferredCardInHand gets mocked.
func (mockGame *mockGameState) InferredCardInHand(
	holdingPlayerName string,
	indexInHand int) (card.Inferred, error) {
	if mockGame.TestErrorForInferredCardInHand != nil {
		mockGame.testReference.Fatalf(
			"InferredCardInHand(%v, %v): %v",
			holdingPlayerName,
			indexInHand,
			mockGame.TestErrorForInferredCardInHand)
	}

	return card.ErrorInferred(), nil
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
			mockGame.TestErrorForInferredCardInHand)
	}

	mockGame.ArgumentsFromRecordChatMessage = append(
		mockGame.ArgumentsFromRecordChatMessage,
		stringTriple{
			FirstString:  actingPlayer.Name(),
			SecondString: actingPlayer.Color(),
			ThirdString:  chatMessage,
		})

	return mockGame.ReturnForNontestError
}

// DrawCard gets mocked.
func (mockGame *mockGameState) DrawCard() (card.Readonly, error) {
	if mockGame.TestErrorForDrawCard != nil {
		mockGame.testReference.Fatalf(
			"DrawCard(): %v",
			mockGame.TestErrorForDrawCard)
	}

	return card.ErrorReadonly(), nil
}

// ReplaceCardInHand gets mocked.
func (mockGame *mockGameState) ReplaceCardInHand(
	holdingPlayerName string,
	indexInHand int,
	replacementCard card.Inferred) (card.Readonly, error) {
	if mockGame.TestErrorForReplaceCardInHand != nil {
		mockGame.testReference.Fatalf(
			"ReplaceCardInHand(%v, %v, %v): %v",
			holdingPlayerName,
			indexInHand,
			replacementCard,
			mockGame.TestErrorForReplaceCardInHand)
	}

	return card.ErrorReadonly(), nil
}

// AddCardToPlayedSequence gets mocked.
func (mockGame *mockGameState) AddCardToPlayedSequence(
	playedCard card.Readonly) error {
	if mockGame.TestErrorForAddCardToPlayedSequence != nil {
		mockGame.testReference.Fatalf(
			"AddCardToPlayedSequence(%v): %v",
			playedCard,
			mockGame.TestErrorForAddCardToPlayedSequence)
	}

	return nil
}

// AddCardToDiscardPile gets mocked.
func (mockGame *mockGameState) AddCardToDiscardPile(
	discardedCard card.Readonly) error {
	if mockGame.TestErrorForAddCardToDiscardPile != nil {
		mockGame.testReference.Fatalf(
			"AddCardToDiscardPile(%v): %v",
			discardedCard,
			mockGame.TestErrorForAddCardToDiscardPile)
	}

	return nil
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
	gameRuleset game.Ruleset,
	playersInTurnOrderWithInitialHands []game.PlayerNameWithHand,
	initialDeck []card.Readonly) error {
	if mockImplementation.TestErrorForAddGame != nil {
		mockImplementation.TestReference.Fatalf(
			"AddGame(%v, %v, %v, %v, %v): %v",
			gameName,
			chatLogLength,
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
