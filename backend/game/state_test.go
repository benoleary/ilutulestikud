package game_test

import (
	"sort"
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/chat"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/player"
)

func prepareImplementations(
	unitTest *testing.T,
	gameName string,
	playerNames []string) []game.State {
	if len(playerNames) < game.MinimumNumberOfPlayers {
		unitTest.Fatalf(
			"Not enough players: %v",
			playerNames)
	}

	nameToIdentifier := &endpoint.Base64NameEncoder{}
	playerCollection := player.NewInMemoryCollection(nameToIdentifier, playerNames, []string{"red", "green", "blue"})
	gameCollections := []game.Collection{
		game.NewInMemoryCollection(nameToIdentifier),
	}

	playerIdentifiers := make([]string, len(playerNames))
	for playerIndex, playerName := range playerNames {
		playerIdentifiers[playerIndex] = nameToIdentifier.Identifier(playerName)
	}

	gameDefinition := endpoint.GameDefinition{
		Name:    gameName,
		Players: playerIdentifiers,
	}

	gameStates := make([]game.State, len(gameCollections))

	for collectionIndex, gameCollection := range gameCollections {
		addError := gameCollection.Add(gameDefinition, playerCollection)

		if addError != nil {
			unitTest.Fatalf(
				"Error when trying to add game for collection index %v: %v",
				collectionIndex,
				addError)
		}

		// We find the game identifier by looking for all the games for the first
		// player of the game definition, as there should only be one game in the
		// collection, and it should include this player.
		playerIdentifier := gameDefinition.Players[0]
		allGames := gameCollection.All(playerIdentifier)

		if (allGames == nil) || (len(allGames) != 1) {
			unitTest.Fatalf(
				"Error when trying to add find identifier for game for collection index %v: allGames = %v",
				collectionIndex,
				allGames)
		}

		gameStates[collectionIndex] = allGames[0]
	}

	return gameStates
}

func TestOrderByCreationTime(unitTest *testing.T) {
	playerNames := []string{"Player One", "Player Two", "Player Three"}
	earlyGameStates := prepareImplementations(
		unitTest,
		"Early game",
		playerNames)

	time.Sleep(100 * time.Millisecond)

	lateGameStates := prepareImplementations(
		unitTest,
		"Late game",
		playerNames)

	for stateIndex := 0; stateIndex < len(lateGameStates); stateIndex++ {
		gameList := game.ByCreationTime([]game.State{
			lateGameStates[stateIndex],
			earlyGameStates[stateIndex],
		})

		if !gameList[1].CreationTime().Before(gameList[0].CreationTime()) {
			unitTest.Fatalf(
				"Game states for state index %v were not differentiable by creation time: early at %v; late at %v",
				stateIndex,
				gameList[1].CreationTime(),
				gameList[0].CreationTime())
		}

		sort.Sort(gameList)

		if (gameList[0].Name() != earlyGameStates[stateIndex].Name()) ||
			(gameList[1].Name() != lateGameStates[stateIndex].Name()) {
			unitTest.Fatalf(
				"Game states were not sorted: expected names [%v, %v], instead had [%v, %v]",
				earlyGameStates[stateIndex].Name(),
				lateGameStates[stateIndex].Name(),
				gameList[0].Name(),
				gameList[1].Name())
		}
	}
}

func TestInitialState(unitTest *testing.T) {
	playerNames := []string{"Player One", "Player Two", "Player Three", "Player Four"}
	numberOfPlayers := len(playerNames)
	namesToFind := make(map[string]bool, numberOfPlayers)
	for _, playerName := range playerNames {
		namesToFind[playerName] = true
	}

	gameStates := prepareImplementations(
		unitTest,
		"Test game",
		playerNames)

	for stateIndex := 0; stateIndex < len(gameStates); stateIndex++ {
		gameState := gameStates[stateIndex]

		participatingPlayers := gameState.Players()

		if len(participatingPlayers) != numberOfPlayers {
			unitTest.Fatalf(
				"Expected %v participants %v but retrieved",
				numberOfPlayers,
				participatingPlayers)
		}

		for _, participatingPlayer := range participatingPlayers {
			if !namesToFind[participatingPlayer.Name()] {
				unitTest.Fatalf(
					"Input participants %v does not include retrieved participant %v",
					playerNames,
					participatingPlayer.Name())
			}

			namesToFind[participatingPlayer.Name()] = false
		}

		for playerName, nameIsMissing := range namesToFind {
			if nameIsMissing {
				unitTest.Fatalf(
					"Input participant %v was not found in retrieve participant list %v",
					playerName,
					participatingPlayers)
			}
		}

		viewingPlayer := participatingPlayers[0]

		gameView := game.ForPlayer(gameState, viewingPlayer.Identifier())

		chat.AssertLogCorrect(
			unitTest,
			gameView.ChatLog,
			[]endpoint.ChatLogMessage{},
		)
	}
}
