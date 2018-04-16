package game_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/game"
)

// just dumps of old endpoint handler tests below here.

func TestRegisterAndRetrieveNewGame(unitTest *testing.T) {
	playerList := []string{"a", "b", "c"}

	type testArguments struct {
		gameName string
	}

	testCases := []struct {
		name      string
		arguments testArguments
	}{
		{
			name: "Ascii only, ",
			arguments: testArguments{
				gameName: "Easy Test Name",
			},
		},
		{
			name: "Punctuation and non-standard characters",
			arguments: testArguments{
				gameName: "?ß@äô#\"'\"",
			},
		},
		{
			name: "Breaks base64",
			arguments: testArguments{
				gameName: breaksBase64,
			},
		},
		{
			name: "URI segment delimiter",
			arguments: testArguments{
				gameName: "/Slashes/are/reserved/for/parsing/URI/segments/",
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			nameToIdentifier, _, gameCollection, gameHandler := setUpHandlerAndRequirements(playerList)

			// First we get the list of rulesets which are valid for creating a game.
			getRulesetsInterface, getRulesetsCode := gameHandler.HandleGet([]string{"available-rulesets"})

			if getRulesetsCode != http.StatusOK {
				unitTest.Fatalf(
					"GET available-rulesets did not return expected HTTP code %v, instead was %v.",
					http.StatusOK,
					getRulesetsCode)
			}

			availableRulesetList, isTypeCorrect := getRulesetsInterface.(endpoint.RulesetList)

			if !isTypeCorrect {
				unitTest.Fatalf(
					"GET available-rulesets did not return expected endpoint.RulesetList, instead was %v.",
					getRulesetsInterface)
			}

			if len(availableRulesetList.Rulesets) < 1 {
				unitTest.Fatalf(
					"GET available-rulesets returned a nil or empty list of rulesets: %v.",
					getRulesetsInterface)
			}

			playerIdentifiers := make([]string, len(playerList))
			for playerIndex, playerName := range playerList {
				playerIdentifiers[playerIndex] = nameToIdentifier.Identifier(playerName)
			}

			// We prepare the definition of the game, choosing the first available ruleset.
			bytesBuffer := new(bytes.Buffer)
			json.NewEncoder(bytesBuffer).Encode(endpoint.GameDefinition{
				GameName:          testCase.arguments.gameName,
				RulesetIdentifier: availableRulesetList.Rulesets[0].Identifier,
				PlayerIdentifiers: playerIdentifiers,
			})

			// Now we add the new game.
			_, postCode :=
				gameHandler.HandlePost(json.NewDecoder(bytesBuffer), []string{"create-new-game"})

			// Then we check that the POST returned a valid response.
			if postCode != http.StatusOK {
				unitTest.Fatalf(
					"POST create-new-game did not return expected HTTP code %v, instead was %v.",
					http.StatusOK,
					postCode)
			}

			// We fetch the game directly.
			gameIdentifier := nameToIdentifier.Identifier(testCase.arguments.gameName)
			actualGame, gameExists := game.ReadState(gameCollection, gameIdentifier)
			if !gameExists {
				unitTest.Fatalf(
					"POST create-new-game did not create a game that can be accessed internally with identifier %v",
					gameIdentifier)
			}

			// Finally we check that the game was registered properly.
			assertgame.StateIsCorrect(
				unitTest,
				testCase.arguments.gameName,
				playerIdentifiers,
				actualGame,
				"Register new player")
		})
	}
}

func TestRejectGetTurnSummariesWithInvalidPlayer(unitTest *testing.T) {
	nameToIdentifier, _, _, gameHandler := setUpHandlerAndRequirements(testPlayerNames())
	playerIdentifier := nameToIdentifier.Identifier("Unregistered Player")
	_, actualCode :=
		gameHandler.HandleGet([]string{"all-games-with-player", playerIdentifier})

	if actualCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"GET all-games-with-player with invalid player did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			actualCode)
	}
}

func TestRejectGetGameForPlayerWithInvalidGame(unitTest *testing.T) {
	playerNames := testPlayerNames()
	nameToIdentifier, _, _, gameHandler := setUpHandlerAndRequirements(playerNames)
	gameName := "Test game"
	correctGameIdentifier := nameToIdentifier.Identifier(gameName)
	incorrectGameIdentifier := nameToIdentifier.Identifier("Invalid game")

	if incorrectGameIdentifier == correctGameIdentifier {
		unitTest.Fatalf(
			"Incorrect identifier %v should not have matched correct identifier %v.",
			incorrectGameIdentifier,
			correctGameIdentifier)
	}

	playerIdentifier := nameToIdentifier.Identifier(playerNames[1])

	bytesBuffer := new(bytes.Buffer)
	json.NewEncoder(bytesBuffer).Encode(endpoint.GameDefinition{
		GameName:          gameName,
		RulesetIdentifier: game.StandardWithoutRainbowIdentifier,
		PlayerIdentifiers: []string{
			playerIdentifier,
			nameToIdentifier.Identifier(playerNames[2]),
			nameToIdentifier.Identifier(playerNames[3]),
		},
	})

	_, postCode :=
		gameHandler.HandlePost(json.NewDecoder(bytesBuffer), []string{"create-new-game"})

	// We only check that the POST returned a valid response.
	if postCode != http.StatusOK {
		unitTest.Fatalf(
			"POST create-new-game did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			postCode)
	}

	_, getCode := gameHandler.HandleGet([]string{"game-as-seen-by-player", incorrectGameIdentifier, playerIdentifier})

	if getCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"GET game-as-seen-by-player/%v/%v without player did not return expected HTTP code %v, instead was %v.",
			incorrectGameIdentifier,
			playerIdentifier,
			http.StatusBadRequest,
			getCode)
	}
}

func TestRejectGetGameForPlayerWithNonparticipantPlayer(unitTest *testing.T) {
	playerNames := testPlayerNames()
	nameToIdentifier, _, _, gameHandler := setUpHandlerAndRequirements(playerNames)
	gameName := "Test game"
	gameIdentifier := nameToIdentifier.Identifier(gameName)

	bytesBuffer := new(bytes.Buffer)
	json.NewEncoder(bytesBuffer).Encode(endpoint.GameDefinition{
		GameName:          gameName,
		RulesetIdentifier: game.StandardWithoutRainbowIdentifier,
		PlayerIdentifiers: []string{
			nameToIdentifier.Identifier(playerNames[1]),
			nameToIdentifier.Identifier(playerNames[2]),
			nameToIdentifier.Identifier(playerNames[3]),
		},
	})

	playerIdentifier := nameToIdentifier.Identifier(playerNames[0])

	_, postCode :=
		gameHandler.HandlePost(json.NewDecoder(bytesBuffer), []string{"create-new-game"})

	// We only check that the POST returned a valid response.
	if postCode != http.StatusOK {
		unitTest.Fatalf(
			"POST create-new-game did not return expected HTTP code %v, instead was %v.",
			http.StatusOK,
			postCode)
	}

	_, getCode :=
		gameHandler.HandleGet([]string{"game-as-seen-by-player", gameIdentifier, playerIdentifier})

	if getCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"GET game-as-seen-by-player/%v with invalid player did not return expected HTTP code %v, instead was %v.",
			gameIdentifier,
			http.StatusBadRequest,
			getCode)
	}
}

func TestGetTurnSummariesForValidPlayer(unitTest *testing.T) {
	nameToIdentifier := testNameToIdentifier()
	playerNames := []string{"a", "b", "c", "d", "e"}
	playerIdentifiers := make([]string, len(playerNames))
	for playerIndex, playerName := range playerNames {
		playerIdentifiers[playerIndex] = nameToIdentifier.Identifier(playerName)
	}

	gameNames := []string{"1", "2", "3", "4"}
	gameIdentifiers := make([]string, len(gameNames))
	availableRulesets := GetAvailableRulesets(unitTest)
	numberOfRulesets := len(availableRulesets)
	rulesetIdentifiers := make([]int, len(gameNames))
	for gameIndex, gameName := range gameNames {
		gameIdentifiers[gameIndex] = nameToIdentifier.Identifier(gameName)
		rulesetIdentifiers[gameIndex] = (gameIndex % numberOfRulesets) + game.StandardWithoutRainbowIdentifier
	}

	type testArguments struct {
		gameDefinitions []endpoint.GameDefinition
	}

	type expectedReturns struct {
		turnSummaries []endpoint.TurnSummary
	}

	testCases := []struct {
		name      string
		arguments testArguments
		expected  expectedReturns
	}{
		{
			name: "No games at all",
			arguments: testArguments{
				gameDefinitions: []endpoint.GameDefinition{},
			},
			expected: expectedReturns{
				turnSummaries: []endpoint.TurnSummary{},
			},
		},
		{
			name: "No games for player out of several",
			arguments: testArguments{
				gameDefinitions: []endpoint.GameDefinition{
					endpoint.GameDefinition{
						GameName:          gameNames[0],
						RulesetIdentifier: rulesetIdentifiers[0],
						PlayerIdentifiers: []string{
							playerIdentifiers[1],
							playerIdentifiers[2],
							playerIdentifiers[3],
						},
					},
					endpoint.GameDefinition{
						GameName:          gameNames[1],
						RulesetIdentifier: rulesetIdentifiers[1],
						PlayerIdentifiers: []string{
							playerIdentifiers[2],
							playerIdentifiers[3],
							playerIdentifiers[4],
						},
					},
					endpoint.GameDefinition{
						GameName:          gameNames[2],
						RulesetIdentifier: rulesetIdentifiers[2],
						PlayerIdentifiers: []string{
							playerIdentifiers[4],
							playerIdentifiers[3],
						},
					},
				},
			},
			expected: expectedReturns{
				turnSummaries: []endpoint.TurnSummary{},
			},
		},
		{
			name: "One game for player out of several",
			arguments: testArguments{
				gameDefinitions: []endpoint.GameDefinition{
					endpoint.GameDefinition{
						GameName:          gameNames[0],
						RulesetIdentifier: rulesetIdentifiers[0],
						PlayerIdentifiers: []string{
							playerIdentifiers[1],
							playerIdentifiers[2],
							playerIdentifiers[0],
						},
					},
					endpoint.GameDefinition{
						GameName:          gameNames[1],
						RulesetIdentifier: rulesetIdentifiers[1],
						PlayerIdentifiers: []string{
							playerIdentifiers[2],
							playerIdentifiers[3],
							playerIdentifiers[4],
						},
					},
					endpoint.GameDefinition{
						GameName:          gameNames[2],
						RulesetIdentifier: rulesetIdentifiers[2],
						PlayerIdentifiers: []string{
							playerIdentifiers[4],
							playerIdentifiers[3],
						},
					},
				},
			},
			expected: expectedReturns{
				turnSummaries: []endpoint.TurnSummary{
					endpoint.TurnSummary{
						GameIdentifier:     gameIdentifiers[0],
						GameName:           gameNames[0],
						RulesetDescription: DescriptionOfRuleset(unitTest, rulesetIdentifiers[0]),
						TurnNumber:         1,
						PlayerNamesInNextTurnOrder: []string{
							playerNames[1],
							playerNames[2],
							playerNames[0],
						},
						IsPlayerTurn: false,
					},
				},
			},
		},
		{
			name: "Several games for player out of many",
			arguments: testArguments{
				gameDefinitions: []endpoint.GameDefinition{
					endpoint.GameDefinition{
						GameName:          gameNames[0],
						RulesetIdentifier: rulesetIdentifiers[0],
						PlayerIdentifiers: []string{
							playerIdentifiers[1],
							playerIdentifiers[2],
							playerIdentifiers[0],
						},
					},
					endpoint.GameDefinition{
						GameName:          gameNames[1],
						RulesetIdentifier: rulesetIdentifiers[1],
						PlayerIdentifiers: []string{
							playerIdentifiers[2],
							playerIdentifiers[3],
							playerIdentifiers[4],
						},
					},
					endpoint.GameDefinition{
						GameName:          gameNames[2],
						RulesetIdentifier: rulesetIdentifiers[2],
						PlayerIdentifiers: []string{
							playerIdentifiers[4],
							playerIdentifiers[3],
						},
					},
					endpoint.GameDefinition{
						GameName:          gameNames[3],
						RulesetIdentifier: rulesetIdentifiers[3],
						PlayerIdentifiers: []string{
							playerIdentifiers[0],
							playerIdentifiers[4],
							playerIdentifiers[3],
						},
					},
				},
			},
			expected: expectedReturns{
				turnSummaries: []endpoint.TurnSummary{
					endpoint.TurnSummary{
						GameIdentifier:     gameIdentifiers[0],
						GameName:           gameNames[0],
						RulesetDescription: DescriptionOfRuleset(unitTest, rulesetIdentifiers[0]),
						TurnNumber:         1,
						PlayerNamesInNextTurnOrder: []string{
							playerNames[1],
							playerNames[2],
							playerNames[0],
						},
						IsPlayerTurn: false,
					},
					endpoint.TurnSummary{
						GameIdentifier:     gameIdentifiers[3],
						GameName:           gameNames[3],
						RulesetDescription: DescriptionOfRuleset(unitTest, rulesetIdentifiers[3]),
						TurnNumber:         1,
						PlayerNamesInNextTurnOrder: []string{
							playerIdentifiers[0],
							playerIdentifiers[4],
							playerIdentifiers[3],
						},
						IsPlayerTurn: true,
					},
				},
			},
		},
		{
			name: "All games for player out of several",
			arguments: testArguments{
				gameDefinitions: []endpoint.GameDefinition{
					endpoint.GameDefinition{
						GameName:          gameNames[0],
						RulesetIdentifier: rulesetIdentifiers[0],
						PlayerIdentifiers: []string{
							playerIdentifiers[1],
							playerIdentifiers[2],
							playerIdentifiers[0],
						},
					},
					endpoint.GameDefinition{
						GameName:          gameNames[1],
						RulesetIdentifier: rulesetIdentifiers[1],
						PlayerIdentifiers: []string{
							playerIdentifiers[4],
							playerIdentifiers[0],
						},
					},
					endpoint.GameDefinition{
						GameName:          gameNames[2],
						RulesetIdentifier: rulesetIdentifiers[2],
						PlayerIdentifiers: []string{
							playerIdentifiers[0],
							playerIdentifiers[4],
							playerIdentifiers[3],
						},
					},
				},
			},
			expected: expectedReturns{
				turnSummaries: []endpoint.TurnSummary{
					endpoint.TurnSummary{
						GameIdentifier:     gameIdentifiers[0],
						GameName:           gameNames[0],
						RulesetDescription: DescriptionOfRuleset(unitTest, rulesetIdentifiers[0]),
						TurnNumber:         1,
						PlayerNamesInNextTurnOrder: []string{
							playerNames[1],
							playerNames[2],
							playerNames[0],
						},
						IsPlayerTurn: false,
					},
					endpoint.TurnSummary{
						GameIdentifier:     gameIdentifiers[1],
						GameName:           gameNames[1],
						RulesetDescription: DescriptionOfRuleset(unitTest, rulesetIdentifiers[1]),
						TurnNumber:         1,
						PlayerNamesInNextTurnOrder: []string{
							playerIdentifiers[4],
							playerIdentifiers[0],
						},
						IsPlayerTurn: false,
					},
					endpoint.TurnSummary{
						GameIdentifier:     gameIdentifiers[2],
						GameName:           gameNames[2],
						RulesetDescription: DescriptionOfRuleset(unitTest, rulesetIdentifiers[2]),
						TurnNumber:         1,
						PlayerNamesInNextTurnOrder: []string{
							playerIdentifiers[0],
							playerIdentifiers[4],
							playerIdentifiers[3],
						},
						IsPlayerTurn: true,
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			_, _, _, gameHandler := setUpHandlerAndRequirements(playerNames)

			// First we add every required game.
			for _, gameDefinition := range testCase.arguments.gameDefinitions {
				bytesBuffer := new(bytes.Buffer)
				json.NewEncoder(bytesBuffer).Encode(gameDefinition)
				_, postCode :=
					gameHandler.HandlePost(json.NewDecoder(bytesBuffer), []string{"create-new-game"})

				// We only check that the POST returned a valid response.
				if postCode != http.StatusOK {
					unitTest.Fatalf(
						"In set-up of %v: POST create-new-game did not return expected HTTP code %v, instead was %v.",
						testCase.name,
						http.StatusOK,
						postCode)
				}
			}

			getInterface, getCode :=
				gameHandler.HandleGet([]string{"all-games-with-player", playerIdentifiers[0]})

			if getCode != http.StatusOK {
				unitTest.Fatalf(
					"GET all-games-with-player/%v did not return expected HTTP code %v, instead was %v.",
					playerIdentifiers[0],
					http.StatusOK,
					getCode)
			}

			actualTurnSummaryList, isTypeCorrect := getInterface.(endpoint.TurnSummaryList)

			if !isTypeCorrect {
				unitTest.Fatalf(
					"GET all-games-with-player/%v did not return expected endpoint.TurnSummaryList, instead was %v.",
					playerIdentifiers[0],
					getInterface)
			}

			if actualTurnSummaryList.TurnSummaries == nil {
				unitTest.Fatalf(
					"GET all-games-with-player/%v returned %v which has a nil list of turn summaries.",
					playerIdentifiers[0],
					getInterface)
			}

			if len(actualTurnSummaryList.TurnSummaries) != len(testCase.expected.turnSummaries) {
				unitTest.Fatalf(
					"GET all-games-with-player/%v returned %v which did not match expected %v.",
					playerIdentifiers[0],
					actualTurnSummaryList.TurnSummaries,
					testCase.expected.turnSummaries)
			}

			for summaryIndex := 0; summaryIndex < len(actualTurnSummaryList.TurnSummaries); summaryIndex++ {
				actualSummary := actualTurnSummaryList.TurnSummaries[summaryIndex]
				expectedSummary := testCase.expected.turnSummaries[summaryIndex]
				actualPlayerOrder := actualSummary.PlayerNamesInNextTurnOrder
				expectedPlayerOrder := actualSummary.PlayerNamesInNextTurnOrder

				// We do not bother checking the timestamps as that would be too much effort.
				if (actualSummary.GameIdentifier != expectedSummary.GameIdentifier) ||
					(actualSummary.GameName != expectedSummary.GameName) ||
					(actualSummary.RulesetDescription != expectedSummary.RulesetDescription) ||
					(actualSummary.TurnNumber != expectedSummary.TurnNumber) ||
					(actualSummary.IsPlayerTurn != expectedSummary.IsPlayerTurn) ||
					(len(actualPlayerOrder) != len(expectedPlayerOrder)) {
					unitTest.Fatalf(
						"GET all-games-with-player/%v returned %v which did not match expected %v.",
						playerIdentifiers[0],
						actualTurnSummaryList.TurnSummaries,
						testCase.expected.turnSummaries)
				}

				for playerIndex := 0; playerIndex < len(actualPlayerOrder); playerIndex++ {
					if actualPlayerOrder[playerIndex] != expectedPlayerOrder[playerIndex] {
						unitTest.Fatalf(
							"GET all-games-with-player/%v returned %v which did not match expected %v.",
							playerIdentifiers[0],
							actualTurnSummaryList.TurnSummaries,
							testCase.expected.turnSummaries)
					}
				}
			}
		})
	}
}

func TestRejectInvalidPlayerAction(unitTest *testing.T) {
	nameToIdentifier := testNameToIdentifier()
	gameName := "test game"
	gameIdentifier := nameToIdentifier.Identifier(gameName)
	playerNames := []string{"a", "b", "c", "d", "e"}
	playerIdentifiers := make([]string, len(playerNames))
	for playerIndex, playerName := range playerNames {
		playerIdentifiers[playerIndex] = nameToIdentifier.Identifier(playerName)
	}

	type testArguments struct {
		bodyObject interface{}
	}

	type expectedReturns struct {
		codeFromPost int
	}

	testCases := []struct {
		name      string
		arguments testArguments
		expected  expectedReturns
	}{
		{
			name: "Nil object",
			arguments: testArguments{
				bodyObject: nil,
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
		{
			name: "Wrong object",
			arguments: testArguments{
				bodyObject: &endpoint.ChatColorList{
					Colors: []string{"x", "y"},
				},
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
		{
			name: "Non-existent game",
			arguments: testArguments{
				bodyObject: &endpoint.PlayerAction{
					PlayerIdentifier: nameToIdentifier.Identifier(playerNames[0]),
					GameIdentifier:   nameToIdentifier.Identifier("Non-existent game"),
					ActionType:       "chat",
					ChatMessage:      "test message",
				},
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
		{
			name: "Non-existent player",
			arguments: testArguments{
				bodyObject: &endpoint.PlayerAction{
					PlayerIdentifier: nameToIdentifier.Identifier("Non-Existent Player"),
					GameIdentifier:   gameIdentifier,
					ActionType:       "chat",
					ChatMessage:      "test message",
				},
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
		{
			name: "Non-participant player",
			arguments: testArguments{
				bodyObject: &endpoint.PlayerAction{
					PlayerIdentifier: nameToIdentifier.Identifier(playerNames[4]),
					GameIdentifier:   gameIdentifier,
					ActionType:       "chat",
					ChatMessage:      "test message",
				},
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
		{
			name: "Nil action",
			arguments: testArguments{
				bodyObject: &endpoint.PlayerAction{
					PlayerIdentifier: nameToIdentifier.Identifier(playerNames[0]),
					GameIdentifier:   gameIdentifier,
					ChatMessage:      "test message",
				},
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
		{
			name: "Invalid action",
			arguments: testArguments{
				bodyObject: &endpoint.PlayerAction{
					PlayerIdentifier: nameToIdentifier.Identifier(playerNames[0]),
					GameIdentifier:   gameIdentifier,
					ActionType:       "invalid_action",
				},
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			creationBytesBuffer := new(bytes.Buffer)
			json.NewEncoder(creationBytesBuffer).Encode(
				endpoint.GameDefinition{
					GameName:          gameName,
					RulesetIdentifier: game.StandardWithoutRainbowIdentifier,
					PlayerIdentifiers: []string{
						playerIdentifiers[1],
						playerIdentifiers[2],
						playerIdentifiers[0],
					},
				})

			_, _, _, gameHandler := setUpHandlerAndRequirements(playerNames)
			creationResponse, creationCode :=
				gameHandler.HandlePost(json.NewDecoder(creationBytesBuffer), []string{"create-new-game"})

			unitTest.Logf("Response to POST create-new-game: %v", creationResponse)

			// We only check that the response code was OK, as other tests check that the game is correctly created.
			if creationCode != http.StatusOK {
				unitTest.Fatalf(
					"POST create-new-game setting up test game did not return expected HTTP code %v, instead was %v.",
					http.StatusOK,
					creationCode)
			}

			actionBytesBuffer := new(bytes.Buffer)
			if testCase.arguments.bodyObject != nil {
				json.NewEncoder(actionBytesBuffer).Encode(testCase.arguments.bodyObject)
			}

			actionResponse, actionCode :=
				gameHandler.HandlePost(json.NewDecoder(actionBytesBuffer), []string{"player-action"})

			unitTest.Logf("Response to POST player-action: %v", actionResponse)

			if actionCode != http.StatusBadRequest {
				unitTest.Fatalf(
					"POST player-action with body %v did not return expected HTTP code %v, instead was %v.",
					testCase.arguments.bodyObject,
					http.StatusBadRequest,
					actionCode)
			}
		})
	}
}
