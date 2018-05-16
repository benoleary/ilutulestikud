package game_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game"
)

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

func TestRegisterAndRetrieveNewGames(unitTest *testing.T) {
	collectionTypes := prepareCollections(unitTest)

	gamesToAddInSequence := []struct {
		gameName    string
		playerNames []string
	}{
		{
			gameName: "Test game 01",
			playerNames: []string{
				playerNamesAvailableInTest[2],
				playerNamesAvailableInTest[1],
				playerNamesAvailableInTest[3],
			},
		},
		{
			gameName: "Test game 02",
			playerNames: []string{
				playerNamesAvailableInTest[1],
				playerNamesAvailableInTest[3],
			},
		},
		{
			gameName: "Test game 03",
			playerNames: []string{
				playerNamesAvailableInTest[0],
				playerNamesAvailableInTest[1],
				playerNamesAvailableInTest[3],
			},
		},
		{
			gameName: "Test game 04",
			playerNames: []string{
				playerNamesAvailableInTest[2],
				playerNamesAvailableInTest[4],
				playerNamesAvailableInTest[0],
				playerNamesAvailableInTest[1],
				playerNamesAvailableInTest[3],
			},
		},
	}

	for _, collectionType := range collectionTypes {
		testIdentifier := "Add new games and retrieve them by name/" + collectionType.CollectionDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			gamesForPlayerMap := make(map[string][]string, 0)

			for _, gameToAdd := range gamesToAddInSequence {
				for _, playerName := range gameToAdd.playerNames {
					gamesForPlayerMap[playerName] =
						append(gamesForPlayerMap[playerName], gameToAdd.gameName)
				}

				errorFromInitialAdd := collectionType.GameCollection.AddNew(
					gameToAdd.gameName,
					testRuleset,
					gameToAdd.playerNames)

				if errorFromInitialAdd != nil {
					unitTest.Fatalf(
						"AddNew(game name %v, standard ruleset, player names %v) produced an error: %v",
						gameToAdd.gameName,
						gameToAdd.playerNames,
						errorFromInitialAdd)
				}

				viewingPlayer := gameToAdd.playerNames[0]
				playerView, errorFromView :=
					collectionType.GameCollection.ViewState(
						gameToAdd.gameName,
						viewingPlayer)

				if errorFromView != nil {
					unitTest.Fatalf(
						"ViewState(same game name %v, player name %v) produced an error: %v",
						gameToAdd.gameName,
						viewingPlayer,
						errorFromView)
				}

				assertStateSummaryFunctionsAreCorrect(
					unitTest,
					viewingPlayer,
					gameToAdd.gameName,
					gameToAdd.playerNames,
					playerView,
					"ViewState(game name "+gameToAdd.gameName+", player name "+viewingPlayer+")")

				// We check that an unknown player causes an error when trying to view games.
				unknownPlayerName := "Not A. Player"
				gamesForUnknownPlayer, errorFromUnknownViewAll :=
					collectionType.GameCollection.ViewAllWithPlayer(unknownPlayerName)

				if errorFromUnknownViewAll == nil {
					unitTest.Fatalf(
						"ViewAllWithPlayer(player name %v) did not produce an error as expected, instead gave %v",
						unknownPlayerName,
						gamesForUnknownPlayer)
				}

				// Now we check that all games for each player can be seen by that player.
				for _, playerName := range playerNamesAvailableInTest {
					gamesForPlayer, errorFromViewAll :=
						collectionType.GameCollection.ViewAllWithPlayer(playerName)

					if errorFromViewAll != nil {
						unitTest.Fatalf(
							"ViewAllWithPlayer(player name %v) produced an error: %v",
							playerName,
							errorFromViewAll)
					}

					expectedGameNames, _ := gamesForPlayerMap[playerName]
					expectedNumberOfGames := len(expectedGameNames)
					if len(gamesForPlayer) != expectedNumberOfGames {
						unitTest.Fatalf(
							"Expected game names %v, but ViewAllWithPlayer(player name %v) returned %v",
							expectedGameNames,
							playerName,
							gamesForPlayer)
					}

					// Since the games should be ordered by creation time, the slices should match
					// element by element.
					for gameIndex := 0; gameIndex < expectedNumberOfGames; gameIndex++ {
						if gamesForPlayer[gameIndex].GameName() != expectedGameNames[gameIndex] {
							unitTest.Fatalf(
								"Expected game names %v, but ViewAllWithPlayer(player name %v) returned %v",
								expectedGameNames,
								playerName,
								gamesForPlayer)
						}
					}
				}
			}
		})
	}
}

func assertStateSummaryFunctionsAreCorrect(
	unitTest *testing.T,
	viewingPlayer string,
	expectedGameName string,
	expectedPlayers []string,
	actualGameView game.ViewForPlayer,
	testIdentifier string) {
	if actualGameView.GameName() != expectedGameName {
		unitTest.Fatalf(
			testIdentifier+": game %v was found but had name %v.",
			expectedGameName,
			actualGameView.GameName())
	}

	actualPlayers, viewingPlayerGoesNext := actualGameView.CurrentTurnOrder()

	playerSlicesMatch := (len(actualPlayers) == len(expectedPlayers))

	if playerSlicesMatch {
		for playerIndex := 0; playerIndex < len(actualPlayers); playerIndex++ {
			playerSlicesMatch =
				(actualPlayers[playerIndex] == expectedPlayers[playerIndex])
			if !playerSlicesMatch {
				break
			}
		}
	}

	if !playerSlicesMatch {
		unitTest.Fatalf(
			testIdentifier+"/game %v was found but had players %v instead of expected %v.",
			expectedGameName,
			actualPlayers,
			expectedPlayers)
	}

	if viewingPlayerGoesNext != (actualPlayers[0] == viewingPlayer) {
		unitTest.Fatalf(
			testIdentifier+"/game %v for viewing player %v had wrong flag for next turn %v instead of expected %v.",
			expectedGameName,
			viewingPlayer,
			viewingPlayerGoesNext,
			actualPlayers[0] == viewingPlayer)
	}
}
