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

// newCollectionAndHandler prepares a game.Collection and a game.GetAndPostHandler
// in a consistent way for the tests.
func newCollectionAndHandler() (game.Collection, *game.GetAndPostHandler) {
	playerCollection :=
		player.NewInMemoryCollection(defaults.InitialPlayerNames(), defaults.AvailableColors())
	gameCollection := game.NewInMemoryCollection()
	gameHandler := game.NewGetAndPostHandler(playerCollection, gameCollection)
	return gameCollection, gameHandler
}

func TestGetNoSegmentBadRequest(unitTest *testing.T) {
	_, gameHandler := newCollectionAndHandler()
	_, actualCode := gameHandler.HandleGet(make([]string, 0))

	if actualCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"GET with empty list of relevant segments did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			actualCode)
	}
}

func TestGetInvalidSegmentNotFound(unitTest *testing.T) {
	_, playerHandler := newCollectionAndHandler()
	_, actualCode := playerHandler.HandleGet([]string{"invalid-segment"})

	if actualCode != http.StatusNotFound {
		unitTest.Fatalf(
			"GET invalid-segment did not return expected HTTP code %v, instead was %v.",
			http.StatusNotFound,
			actualCode)
	}
}

func TestPostNoSegmentBadRequest(unitTest *testing.T) {
	_, playerHandler := newCollectionAndHandler()
	bytesBuffer := new(bytes.Buffer)
	json.NewEncoder(bytesBuffer).Encode(endpoint.GameDefinition{
		Name:    "Game name",
		Players: []string{"Player One", "Player Two"},
	})

	_, actualCode := playerHandler.HandlePost(json.NewDecoder(bytesBuffer), make([]string, 0))

	if actualCode != http.StatusBadRequest {
		unitTest.Fatalf(
			"POST with empty list of relevant segments did not return expected HTTP code %v, instead was %v.",
			http.StatusBadRequest,
			actualCode)
	}
}

func TestPostInvalidSegmentNotFound(unitTest *testing.T) {
	_, playerHandler := newCollectionAndHandler()
	bytesBuffer := new(bytes.Buffer)
	json.NewEncoder(bytesBuffer).Encode(endpoint.GameDefinition{
		Name:    "Game name",
		Players: []string{"Player One", "Player Two"},
	})

	_, actualCode := playerHandler.HandlePost(json.NewDecoder(bytesBuffer), []string{"invalid-segment"})

	if actualCode != http.StatusNotFound {
		unitTest.Fatalf(
			"POST invalid-segment did not return expected HTTP code %v, instead was %v.",
			http.StatusNotFound,
			actualCode)
	}
}
