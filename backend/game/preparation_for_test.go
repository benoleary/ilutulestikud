package game_test

import (
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
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
	testReference     *testing.T
	mockName          string
	mockRuleset       game.Ruleset
	mockNamesAndHands []game.PlayerNameWithHand
	mockDeck          []card.Readonly
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
		mockImplementation.TestReference.Errorf(
			"RandomSeed(): %v",
			mockImplementation.TestErrorForRandomSeed)
	}

	return mockImplementation.ReturnForRandomSeed
}

func (mockImplementation *mockGamePersister) ReadAndWriteGame(
	gameName string) (game.ReadAndWriteState, error) {
	if mockImplementation.TestErrorForReadAndWriteGame != nil {
		mockImplementation.TestReference.Errorf(
			"ReadAndWriteGame(%v): %v",
			gameName,
			mockImplementation.TestErrorForReadAndWriteGame)
	}

	return mockImplementation.ReturnForReadAndWriteGame, mockImplementation.ReturnForNontestError
}

func (mockImplementation *mockGamePersister) ReadAllWithPlayer(playerName string) []game.ReadonlyState {
	if mockImplementation.TestErrorForReadAllWithPlayer != nil {
		mockImplementation.TestReference.Errorf(
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
		mockImplementation.TestReference.Errorf(
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
			mockName:          gameName,
			mockRuleset:       gameRuleset,
			mockNamesAndHands: playersInTurnOrderWithInitialHands,
			mockDeck:          initialDeck,
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
