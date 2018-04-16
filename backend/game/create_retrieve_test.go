package game_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/chat"
	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/player"
)

var playerNamesAvailableInTest []string = []string{"A", "B", "C", "D", "E", "F", "G"}
var testRuleset game.Ruleset = &game.StandardWithoutRainbowRuleset{}

type mockGameState struct {
	MockGameName     string
	MockCreationTime time.Time
}

// Read gets mocked.
func (gameState *mockGameState) read() game.ReadonlyState {
	return gameState
}

// Name gets mocked.
func (gameState *mockGameState) Name() string {
	return gameState.MockGameName
}

// Ruleset gets mocked.
func (gameState *mockGameState) Ruleset() game.Ruleset {
	return testRuleset
}

// Players gets mocked.
func (gameState *mockGameState) Players() []player.ReadonlyState {
	return nil
}

// Turn gets mocked.
func (gameState *mockGameState) Turn() int {
	return -2
}

// CreationTime gets mocked.
func (gameState *mockGameState) CreationTime() time.Time {
	return gameState.MockCreationTime
}

// ChatLog gets mocked.
func (gameState *mockGameState) ChatLog() *chat.Log {
	return nil
}

// HasPlayerAsParticipant gets mocked.
func (gameState *mockGameState) HasPlayerAsParticipant(playerName string) bool {
	return false
}

// Score gets mocked.
func (gameState *mockGameState) Score() int {
	return -3
}

// NumberOfReadyHints gets mocked.
func (gameState *mockGameState) NumberOfReadyHints() int {
	return -4
}

// NumberOfMistakesMade gets mocked.
func (gameState *mockGameState) NumberOfMistakesMade() int {
	return -5
}

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
	mockPlayer, isInMap := mockPlayers[playerName]

	if !isInMap {
		return nil, fmt.Errorf("not in map")
	}

	return mockPlayer, nil
}

func GetAvailableRulesets(unitTest *testing.T) []endpoint.SelectableRuleset {
	availableRulesets := game.AvailableRulesets()

	if len(availableRulesets) < 1 {
		unitTest.Fatalf(
			"At least one ruleset must be available for tests: game.AvailableRulesets() returned %v",
			availableRulesets)
	}

	return availableRulesets
}

func DescriptionOfRuleset(unitTest *testing.T, rulesetIdentifier int) string {
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
				statePersister.GamePersister,
				mockProvider)
		stateCollections[persisterIndex] = collectionAndDescription{
			PlayerCollection:      stateCollection,
			CollectionDescription: "collection around " + statePersister.PersisterDescription,
		}
	}

	return stateCollections
}

func TestOrderByCreationTime(unitTest *testing.T) {
	mockGames := game.ByCreationTime([]game.ReadonlyState{
		&mockGameState{
			MockGameName:     "Far future",
			MockCreationTime: time.Now().Add(100 * time.Second),
		},
		&mockGameState{
			MockGameName:     "Far past",
			MockCreationTime: time.Now().Add(-100 * time.Second),
		},
		&mockGameState{
			MockGameName:     "Near future",
			MockCreationTime: time.Now().Add(1 * time.Second),
		},
		&mockGameState{
			MockGameName:     "Near past",
			MockCreationTime: time.Now().Add(-1 * time.Second),
		},
	})

	sort.Sort(mockGames)

	if (mockGames[0].Name() != mockGames[1].Name()) ||
		(mockGames[1].Name() != mockGames[3].Name()) ||
		(mockGames[2].Name() != mockGames[2].Name()) ||
		(mockGames[3].Name() != mockGames[0].Name()) {
		unitTest.Fatalf(
			"Game states were not sorted: expected names [%v, %v, %v, %v], instead had %v",
			mockGames[1].Name(),
			mockGames[3].Name(),
			mockGames[2].Name(),
			mockGames[0].Name(),
			mockGames)
	}
}

func TestRejectInvalidNewGame(unitTest *testing.T) {
	validGameName := "Test game"

	validPlayerNameList :=
		[]string{
			playerNamesAvailableInTest[0],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[2],
		}

	testCases := []struct {
		testName    string
		gameName    string
		playerNames []string
	}{
		{
			testName:    "Empty game name",
			gameName:    "",
			playerNames: validPlayerNameList,
		},
		{
			testName:    "Nil players",
			gameName:    validGameName,
			playerNames: nil,
		},
		{
			testName:    "No players",
			gameName:    validGameName,
			playerNames: []string{},
		},
		{
			testName: "Too few players",
			gameName: validGameName,
			playerNames: []string{
				playerNamesAvailableInTest[0],
			},
		},
		{
			testName:    "Too many players",
			gameName:    validGameName,
			playerNames: playerNamesAvailableInTest,
		},
		{
			testName: "Repeated player",
			gameName: validGameName,
			playerNames: []string{
				playerNamesAvailableInTest[2],
				playerNamesAvailableInTest[1],
				playerNamesAvailableInTest[1],
				playerNamesAvailableInTest[3],
			},
		},
		{
			testName: "Unregistered player",
			gameName: validGameName,
			playerNames: []string{
				playerNamesAvailableInTest[2],
				playerNamesAvailableInTest[1],
				"Not A. Registered Player",
				playerNamesAvailableInTest[3],
			},
		},
	}

	for _, testCase := range testCases {
		collectionTypes := prepareCollections(unitTest)

		for _, collectionType := range collectionTypes {
			testIdentifier := testCase.testName + "/" + collectionType.CollectionDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				errorFromAdd :=
					collectionType.GameCollection.AddNew(
						testCase.gameName,
						testRuleset,
						testCase.playerNames)

				if errorFromAdd == nil {
					unitTest.Fatalf(
						"AddNew(game name %v, standard ruleset, player names %v) did not return an error",
						testCase.gameName,
						testCase.playerNames)
				}
			})
		}
	}
}

func TestRejectNewGameWithExistingName(unitTest *testing.T) {
	collectionTypes := prepareCollections(unitTest)

	gameName := "Test game"

	for _, collectionType := range collectionTypes {
		testIdentifier := "Reject new game with existing name/" + collectionType.CollectionDescription
		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			initialGamePlayerNames := []string{
				playerNamesAvailableInTest[0],
				playerNamesAvailableInTest[1],
				playerNamesAvailableInTest[2],
			}

			errorFromInitialAdd := collectionType.GameCollection.AddNew(
				gameName,
				testRuleset,
				initialGamePlayerNames)

			if errorFromInitialAdd != nil {
				unitTest.Fatalf(
					"First AddNew(game name %v, standard ruleset, player names %v) produced an error: %v",
					gameName,
					initialGamePlayerNames,
					errorFromInitialAdd)
			}

			invalidGamePlayerNames := []string{
				playerNamesAvailableInTest[3],
				playerNamesAvailableInTest[2],
				playerNamesAvailableInTest[4],
			}

			errorFromInvalidAdd := collectionType.GameCollection.AddNew(
				gameName,
				testRuleset,
				invalidGamePlayerNames)

			if errorFromInvalidAdd == nil {
				unitTest.Fatalf(
					"Second AddNew(same game name %v, standard ruleset, player names %v) did not return an error",
					gameName,
					invalidGamePlayerNames)
			}
		})
	}
}

func TestRegisterAndRetrieveNewGame(unitTest *testing.T) {
	collectionTypes := prepareCollections(unitTest)

	gameName := "Test game"

	for _, collectionType := range collectionTypes {
		testIdentifier := "Add new game and retrieve it by name/" + collectionType.CollectionDescription
		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			playerNames := []string{
				playerNamesAvailableInTest[0],
				playerNamesAvailableInTest[1],
				playerNamesAvailableInTest[2],
			}

			errorFromInitialAdd := collectionType.GameCollection.AddNew(
				gameName,
				testRuleset,
				initialGamePlayerNames)

			if errorFromInitialAdd != nil {
				unitTest.Fatalf(
					"AddNew(game name %v, standard ruleset, player names %v) produced an error: %v",
					gameName,
					playerNames,
					errorFromInitialAdd)
			}

			viewingPlayer := playerNames[0]
			playerView, errorFromView :=
				collectionType.GameCollection.ViewState(
					gameName,
					viewingPlayer)

			if errorFromView != nil {
				unitTest.Fatalf(
					"ViewState(same game name %v, player name %v) produced an error: %v",
					gameName,
					viewingPlayer,
					errorFromView)
			}

			if playerView.GameName() != gameName {
				unitTest.Fatalf(
					"ViewState(same game name %v, player name %v) produced an incorrect view %v",
					gameName,
					viewingPlayer,
					playerView)
			}
		})
	}
}

func assertStateSummaryFunctionsAreCorrect(
	unitTest *testing.T,
	expectedGameName string,
	expectedPlayers []string,
	actualGame game.ReadonlyState,
	testIdentifier string) {
	if actualGame.Name() != expectedGameName {
		unitTest.Fatalf(
			testIdentifier+": game %v was found but had name %v.",
			expectedGameName,
			actualGame.Name())
	}

	actualPlayers := actualGame.Players()
	playerSlicesMatch := (len(actualPlayers) == len(expectedPlayers))

	if playerSlicesMatch {
		for playerIndex := 0; playerIndex < len(actualPlayers); playerIndex++ {
			playerSlicesMatch =
				(actualPlayers[playerIndex].Name() == expectedPlayers[playerIndex])
			if !playerSlicesMatch {
				break
			}
		}
	}

	if !playerSlicesMatch {
		unitTest.Fatalf(
			testIdentifier+": game %v was found but had players %v instead of expected %v.",
			expectedGameName,
			actualPlayers,
			expectedPlayers)
	}
}
