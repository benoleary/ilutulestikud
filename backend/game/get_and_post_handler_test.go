package game_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/defaults"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/player"
)

type nameToName struct {
}

func testPlayerNames() []string {
	return []string{"a", "b", "c", "d", "e", "f", "g"}
}

func testPlayerIdentifier(playerIndex int) string {
	// Terribly inefficient, but it is the easiest way to be consistent in the tests.
	nameToIdentifier := &endpoint.Base64NameEncoder{}
	return nameToIdentifier.Identifier(testPlayerNames()[playerIndex])
}

// newCollectionAndHandler prepares a game.Collection and a game.GetAndPostHandler
// in a consistent way for the tests. The player.Collection is created with a simple
// name-to-indentifier encoder which just uses the name as its identifier.
func setUpHandlerAndRequirements(registeredPlayers []string) (
	endpoint.NameToIdentifier, game.Collection, *game.GetAndPostHandler) {
	nameToIdentifier := &endpoint.Base64NameEncoder{}
	playerCollection :=
		player.NewInMemoryCollection(
			nameToIdentifier,
			registeredPlayers,
			defaults.AvailableColors())
	gameCollection := game.NewInMemoryCollection(nameToIdentifier)
	gameHandler := game.NewGetAndPostHandler(playerCollection, gameCollection)
	return nameToIdentifier, gameCollection, gameHandler
}

func TestGetNoSegmentBadRequest(unitTest *testing.T) {
	_, _, gameHandler := setUpHandlerAndRequirements(testPlayerNames())
	_, actualCode := gameHandler.HandleGet(make([]string, 0))

	if actualCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"GET with empty list of relevant segments did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			actualCode)
	}
}

func TestGetInvalidSegmentNotFound(unitTest *testing.T) {
	_, _, gameHandler := setUpHandlerAndRequirements(testPlayerNames())
	_, actualCode := gameHandler.HandleGet([]string{"invalid-segment"})

	if actualCode != http.StatusNotFound {
		unitTest.Fatalf(
			"GET invalid-segment did not return expected HTTP code %v, instead was %v.",
			http.StatusNotFound,
			actualCode)
	}
}

func TestPostNoSegmentBadRequest(unitTest *testing.T) {
	_, _, gameHandler := setUpHandlerAndRequirements(testPlayerNames())
	bytesBuffer := new(bytes.Buffer)
	json.NewEncoder(bytesBuffer).Encode(endpoint.GameDefinition{
		Name:    "Game name",
		Players: []string{"Player One", "Player Two"},
	})

	_, actualCode := gameHandler.HandlePost(json.NewDecoder(bytesBuffer), make([]string, 0))

	if actualCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"POST with empty list of relevant segments did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			actualCode)
	}
}

func TestPostInvalidSegmentNotFound(unitTest *testing.T) {
	_, _, gameHandler := setUpHandlerAndRequirements(testPlayerNames())
	bytesBuffer := new(bytes.Buffer)
	json.NewEncoder(bytesBuffer).Encode(endpoint.GameDefinition{
		Name:    "Game name",
		Players: []string{"Player One", "Player Two"},
	})

	_, actualCode := gameHandler.HandlePost(json.NewDecoder(bytesBuffer), []string{"invalid-segment"})

	if actualCode != http.StatusNotFound {
		unitTest.Fatalf(
			"POST invalid-segment did not return expected HTTP code %v, instead was %v.",
			http.StatusNotFound,
			actualCode)
	}
}

func TestRejectInvalidNewGame(unitTest *testing.T) {
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
			name: "Nil players",
			arguments: testArguments{
				bodyObject: &endpoint.GameDefinition{
					Name: "Test game",
				},
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
		{
			name: "No players",
			arguments: testArguments{
				bodyObject: &endpoint.GameDefinition{
					Name:    "Test game",
					Players: make([]string, 0),
				},
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
		{
			name: "Too few players",
			arguments: testArguments{
				bodyObject: &endpoint.GameDefinition{
					Name: "Test game",
					// We use the same set of player names here as used to set up the game.Collection
					// as well as the name encoding.
					Players: []string{testPlayerIdentifier(1)},
				},
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
		{
			name: "Too many players",
			arguments: testArguments{
				bodyObject: &endpoint.GameDefinition{
					Name: "Test game",
					// We use the same set of player names here as used to set up the game.Collection
					// as well as the name encoding.
					Players: []string{
						testPlayerIdentifier(0),
						testPlayerIdentifier(1),
						testPlayerIdentifier(2),
						testPlayerIdentifier(3),
						testPlayerIdentifier(4),
						testPlayerIdentifier(5),
					},
				},
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
		{
			name: "Repeated player",
			arguments: testArguments{
				bodyObject: &endpoint.GameDefinition{
					Name: "Test game",
					Players: []string{
						// We use the same set of player names here as used to set up the game.Collection
						// as well as the name encoding.
						testPlayerIdentifier(0),
						testPlayerIdentifier(1),
						testPlayerIdentifier(1),
						testPlayerIdentifier(2),
					},
				},
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
		{
			name: "Unregistered player",
			arguments: testArguments{
				bodyObject: &endpoint.GameDefinition{
					Name: "Test game",
					Players: []string{
						// We use the same set of player names here as used to set up the game.Collection.
						testPlayerIdentifier(0),
						testPlayerIdentifier(1),
						"I am not registered!",
						testPlayerIdentifier(2),
					},
				},
			},
			expected: expectedReturns{
				codeFromPost: http.StatusBadRequest,
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			bytesBuffer := new(bytes.Buffer)
			if testCase.arguments.bodyObject != nil {
				json.NewEncoder(bytesBuffer).Encode(testCase.arguments.bodyObject)
			}

			_, _, gameHandler := setUpHandlerAndRequirements(testPlayerNames())
			_, postCode :=
				gameHandler.HandlePost(json.NewDecoder(bytesBuffer), []string{"create-new-game"})

			if postCode != http.StatusBadRequest {
				unitTest.Fatalf(
					"POST create-new-game with invalid JSON %v did not return expected HTTP code %v, instead was %v.",
					testCase.arguments.bodyObject,
					http.StatusBadRequest,
					postCode)
			}
		})
	}
}

func TestRejectNewGameWithExistingName(unitTest *testing.T) {
	playerNames := testPlayerNames()
	nameToIdentifier, _, gameHandler := setUpHandlerAndRequirements(playerNames)

	gameName := "Test game"
	firstBodyObject := &endpoint.GameDefinition{
		Name: gameName,
		Players: []string{
			nameToIdentifier.Identifier(playerNames[1]),
			nameToIdentifier.Identifier(playerNames[2]),
			nameToIdentifier.Identifier(playerNames[3]),
		},
	}

	firstBytesBuffer := new(bytes.Buffer)
	json.NewEncoder(firstBytesBuffer).Encode(firstBodyObject)

	_, validRegistrationCode :=
		gameHandler.HandlePost(json.NewDecoder(firstBytesBuffer), []string{"create-new-game"})

	if validRegistrationCode != http.StatusOK {
		unitTest.Fatalf(
			"POST create-new-game with valid JSON %v did not return expected HTTP code %v, instead was %v.",
			firstBodyObject,
			http.StatusOK,
			validRegistrationCode)
	}

	secondBodyObject := endpoint.GameDefinition{
		Name: gameName,
		Players: []string{
			nameToIdentifier.Identifier(playerNames[1]),
			nameToIdentifier.Identifier(playerNames[3]),
			nameToIdentifier.Identifier(playerNames[4]),
		},
	}

	secondBytesBuffer := new(bytes.Buffer)
	json.NewEncoder(secondBytesBuffer).Encode(secondBodyObject)

	_, invalidRegistrationCode :=
		gameHandler.HandlePost(json.NewDecoder(secondBytesBuffer), []string{"create-new-game"})

	if invalidRegistrationCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"POST new-player with valid JSON %v but second request for same player name %v"+
				" did not return expected HTTP code %v, instead was %v.",
			gameName,
			secondBodyObject,
			http.StatusBadRequest,
			invalidRegistrationCode)
	}
}

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
			name: "URI segment delimiter",
			arguments: testArguments{
				gameName: "/Slashes/are/reserved/for/parsing/URI/segments/",
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			nameToIdentifier, gameCollection, gameHandler := setUpHandlerAndRequirements(playerList)

			playerIdentifiers := make([]string, len(playerList))
			for playerIndex, playerName := range playerList {
				playerIdentifiers[playerIndex] = nameToIdentifier.Identifier(playerName)
			}

			bytesBuffer := new(bytes.Buffer)
			json.NewEncoder(bytesBuffer).Encode(endpoint.GameDefinition{
				Name:    testCase.arguments.gameName,
				Players: playerIdentifiers,
			})

			// First we add the new game.
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
			actualGame, gameExists := gameCollection.Get(gameIdentifier)
			if !gameExists {
				unitTest.Fatalf(
					"POST create-new-game did not create a game that can be accessed internally with identifier %v",
					gameIdentifier)
			}

			// Finally we check that the game was registered properly.
			assertGameIsCorrect(
				unitTest,
				testCase.arguments.gameName,
				playerIdentifiers,
				actualGame,
				"Register new player")
		})
	}
}

func assertGameIsCorrect(
	unitTest *testing.T,
	expectedGameName string,
	expectedPlayers []string,
	actualGame game.State,
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
				(actualPlayers[playerIndex].Identifier() == expectedPlayers[playerIndex])
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
