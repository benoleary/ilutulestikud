package game_test

import (
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/player"
)

var playerNamesAvailableInTest []string = []string{"A", "B", "C", "D", "E", "F", "G"}
var testRuleset game.Ruleset = &game.StandardWithoutRainbowRuleset{}

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
	mockPlayers map[string]*mockPlayerState
}

func (mockProvider *mockPlayerProvider) Get(
	playerName string) (player.ReadonlyState, error) {
	mockPlayer, isInMap := mockProvider.mockPlayers[playerName]

	if !isInMap {
		return nil, fmt.Errorf("not in map")
	}

	return mockPlayer, nil
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
	chatColor := defaults.AvailableColors()[0]
	mockPlayerMap := make(map[string]*mockPlayerState, 0)
	for _, mockPlayerName := range playerNamesAvailableInTest {
		mockPlayerMap[mockPlayerName] = &mockPlayerState{
			MockName:  mockPlayerName,
			MockColor: chatColor,
		}
	}

	mockProvider := &mockPlayerProvider{
		mockPlayers: mockPlayerMap,
	}

	statePersisters := []persisterAndDescription{
		persisterAndDescription{
			GamePersister:        game.NewInMemoryPersister(),
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
