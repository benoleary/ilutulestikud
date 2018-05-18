package game_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/chat"
	"github.com/benoleary/ilutulestikud/backend/game/persister"
	"github.com/benoleary/ilutulestikud/backend/player"
)

var playerNamesAvailableInTest []string = []string{"A", "B", "C", "D", "E", "F", "G"}
var testRuleset game.Ruleset = &game.StandardWithoutRainbowRuleset{}

var mockChatColor string = defaults.AvailableColors()[0]

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

type mockGameState struct {
	testReference                       *testing.T
	MockName                            string
	MockRuleset                         game.Ruleset
	MockNamesAndHands                   []game.PlayerNameWithHand
	MockDeck                            []card.Readonly
	MockTurn                            int
	MockChatLog                         *chat.Log
	TestErrorForName                    error
	TestErrorForRuleset                 error
	ReturnForRuleset                    game.Ruleset
	TestErrorForPlayerNames             error
	ReturnForPlayerNames                []string
	TestErrorForTurn                    error
	TestErrorForCreationTime            error
	TestErrorForChatLog                 error
	TestErrorForScore                   error
	TestErrorForNumberOfReadyHints      error
	TestErrorForNumberOfMistakesMade    error
	TestErrorForDeckSize                error
	TestErrorForLastPlayedForColor      error
	TestErrorForNumberOfDiscardedCards  error
	TestErrorForVisibleCardInHand       error
	TestErrorForInferredCardInHand      error
	TestErrorForRecordChatMessage       error
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
		TestErrorForChatLog:                 testError,
		TestErrorForScore:                   testError,
		TestErrorForNumberOfReadyHints:      testError,
		TestErrorForNumberOfMistakesMade:    testError,
		TestErrorForDeckSize:                testError,
		TestErrorForLastPlayedForColor:      testError,
		TestErrorForNumberOfDiscardedCards:  testError,
		TestErrorForVisibleCardInHand:       testError,
		TestErrorForInferredCardInHand:      testError,
		TestErrorForRecordChatMessage:       testError,
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

	return time.Now()
}

// ChatLog gets mocked.
func (mockGame *mockGameState) ChatLog() *chat.Log {
	if mockGame.TestErrorForChatLog != nil {
		mockGame.testReference.Fatalf(
			"ChatLog(): %v",
			mockGame.TestErrorForChatLog)
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

// LastPlayedForColor gets mocked.
func (mockGame *mockGameState) LastPlayedForColor(
	colorSuit string) (card.Readonly, bool) {
	if mockGame.TestErrorForLastPlayedForColor != nil {
		mockGame.testReference.Fatalf(
			"LastPlayedForColor(%v): %v",
			colorSuit,
			mockGame.TestErrorForLastPlayedForColor)
	}

	return card.ErrorReadonly(), false
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

	return nil
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
	ArgumentsForAddGame           []mockGameState
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
		ArgumentsForAddGame:           make([]mockGameState, 0),
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
	gameRuleset game.Ruleset,
	playersInTurnOrderWithInitialHands []game.PlayerNameWithHand,
	initialDeck []card.Readonly) error {
	if mockImplementation.TestErrorForAddGame != nil {
		mockImplementation.TestReference.Fatalf(
			"AddGame(%v, %v, %v, %v): %v",
			gameName,
			gameRuleset,
			playersInTurnOrderWithInitialHands,
			initialDeck,
			mockImplementation.TestErrorForAddGame)
	}

	mockImplementation.ArgumentsForAddGame =
		append(mockImplementation.ArgumentsForAddGame, mockGameState{
			testReference:     mockImplementation.TestReference,
			MockName:          gameName,
			MockRuleset:       gameRuleset,
			MockNamesAndHands: playersInTurnOrderWithInitialHands,
			MockDeck:          initialDeck,
		})

	return mockImplementation.ReturnForNontestError
}

func prepareCollection(
	unitTest *testing.T,
	initialPlayers []string) (*game.StateCollection, *mockGamePersister, *mockPlayerProvider) {
	mockGamePersister :=
		NewMockGamePersister(unitTest, fmt.Errorf("initial error for every function"))
	mockPlayerProvider := NewMockPlayerProvider(initialPlayers)
	mockCollection := game.NewCollection(mockGamePersister, mockPlayerProvider)
	return mockCollection, mockGamePersister, mockPlayerProvider
}

func getAvailableRulesetIdentifiers(unitTest *testing.T) []int {
	availableRulesetIdentifiers := game.ValidRulesetIdentifiers()

	if len(availableRulesetIdentifiers) < 1 {
		unitTest.Fatalf(
			"At least one ruleset identifier must be available for tests: game.ValidRulesetIdentifiers() returned %v",
			availableRulesetIdentifiers)
	}

	return availableRulesetIdentifiers
}

func descriptionOfRuleset(unitTest *testing.T, rulesetIdentifier int) string {
	foundRuleset, identifierError := game.RulesetFromIdentifier(rulesetIdentifier)

	if identifierError != nil {
		unitTest.Fatalf(
			"Unable to find description of ruleset with identifier %v: error is %v",
			rulesetIdentifier,
			identifierError)
	}

	return foundRuleset.FrontendDescription()
}

type persisterAndDescription struct {
	GamePersister        game.StatePersister
	PersisterDescription string
}

type collectionAndDescription struct {
	GameCollection        *game.StateCollection
	CollectionDescription string
}

func prepareCollections(unitTest *testing.T) []collectionAndDescription {
	mockProvider := NewMockPlayerProvider(playerNamesAvailableInTest)

	statePersisters := []persisterAndDescription{
		persisterAndDescription{
			GamePersister:        persister.NewInMemory(),
			PersisterDescription: "in-memory persister",
		},
	}

	numberOfPersisters := len(statePersisters)

	stateCollections := make([]collectionAndDescription, numberOfPersisters)

	for persisterIndex := 0; persisterIndex < numberOfPersisters; persisterIndex++ {
		gamePersister := statePersisters[persisterIndex]
		stateCollection :=
			game.NewCollection(
				gamePersister.GamePersister,
				mockProvider)
		stateCollections[persisterIndex] = collectionAndDescription{
			GameCollection:        stateCollection,
			CollectionDescription: "collection around " + gamePersister.PersisterDescription,
		}
	}

	return stateCollections
}
