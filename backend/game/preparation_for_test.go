package game_test

import (
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/persister"
)

var playerNamesAvailableInTest []string = []string{"A", "B", "C", "D", "E", "F", "G"}
var testRuleset game.Ruleset = game.NewStandardWithoutRainbow()

var mockChatColor string = defaults.AvailableColors()[0]

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
