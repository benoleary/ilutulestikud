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

// Identifier encodes the name as itself.
func (nameToIdentifier *nameToName) Identifier(name string) string {
	return name
}

// newCollectionAndHandler prepares a game.Collection and a game.GetAndPostHandler
// in a consistent way for the tests. The player.Collection is created with a simple
// name-to-indentifier encoder which just uses the name as its identifier.
func newCollectionAndHandlerWithIdentifier() (game.Collection, *game.GetAndPostHandler, endpoint.NameToIdentifier) {
	nameToIdentifier := &nameToName{}
	playerCollection :=
		player.NewInMemoryCollection(
			nameToIdentifier,
			defaults.InitialPlayerNames(),
			defaults.AvailableColors())
	gameCollection := game.NewInMemoryCollection(&endpoint.Base64NameEncoder{})
	gameHandler := game.NewGetAndPostHandler(playerCollection, gameCollection)
	return gameCollection, gameHandler, nameToIdentifier
}

func TestGetNoSegmentBadRequest(unitTest *testing.T) {
	_, gameHandler, _ := newCollectionAndHandlerWithIdentifier()
	_, actualCode := gameHandler.HandleGet(make([]string, 0))

	if actualCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"GET with empty list of relevant segments did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			actualCode)
	}
}

func TestGetInvalidSegmentNotFound(unitTest *testing.T) {
	_, gameHandler, _ := newCollectionAndHandlerWithIdentifier()
	_, actualCode := gameHandler.HandleGet([]string{"invalid-segment"})

	if actualCode != http.StatusNotFound {
		unitTest.Fatalf(
			"GET invalid-segment did not return expected HTTP code %v, instead was %v.",
			http.StatusNotFound,
			actualCode)
	}
}

func TestPostNoSegmentBadRequest(unitTest *testing.T) {
	_, gameHandler, _ := newCollectionAndHandlerWithIdentifier()
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
	_, gameHandler, _ := newCollectionAndHandlerWithIdentifier()
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
					Colors: []string{"Player 1", "Player 2"},
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
					Name:    "Test game",
					Players: []string{"identifier_one"},
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
					Name:    "Test game",
					Players: []string{"1", "2", "3", "4", "5", "6"},
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
					Name:    "Test game",
					Players: []string{"1", "2", "2", "3"},
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

			_, gameHandler, _ := newCollectionAndHandlerWithIdentifier()
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
	gameName := "Test game"
	firstBodyObject := &endpoint.GameDefinition{
		Name:    gameName,
		Players: []string{"1", "2", "3"},
	}

	firstBytesBuffer := new(bytes.Buffer)
	json.NewEncoder(firstBytesBuffer).Encode(firstBodyObject)

	_, gameHandler, _ := newCollectionAndHandlerWithIdentifier()
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
		Name:    gameName,
		Players: []string{"1", "10", "100"},
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
			gameCollection, gameHandler, nameToIdentifier := newCollectionAndHandlerWithIdentifier()

			bytesBuffer := new(bytes.Buffer)
			json.NewEncoder(bytesBuffer).Encode(endpoint.GameDefinition{
				Name:    testCase.arguments.gameName,
				Players: playerList,
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
				playerList,
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
			// We can just compare names as the player name encoding in the test uses the
			// original player names as identifiers for players.
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
