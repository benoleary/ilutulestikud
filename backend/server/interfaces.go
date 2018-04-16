package server

import (
	"encoding/json"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/player"
)

type httpGetAndPostHandler interface {
	HandleGet(relevantSegments []string) (interface{}, int)

	HandlePost(httpBodyDecoder *json.Decoder, relevantSegments []string) (interface{}, int)
}

type playerCollection interface {
	// All should return a slice of all the players in the collection. The order is not
	// mandated, and may even change with repeated calls to the same unchanged collection
	// (analogously to the entry set of a standard Golang map, for example), though of
	// course an implementation may order the slice consistently.
	All() []player.ReadonlyState

	// Get should return a read-only state for the identified player.
	Get(playerIdentifier string) (player.ReadonlyState, error)

	// AvailableChatColors should return the chat colors available to the collection.
	AvailableChatColors() []string

	// Add should add a new player to the collection, defined by the given arguments.
	Add(playerName string, chatColor string) error

	// UpdateColor should update the given player with the given chat color.
	UpdateColor(playerName string, chatColor string) error

	// Reset should reset the players to the initial set.
	Reset()
}

type gameCollection interface {
	// ViewState should return a view around the read-only game state corresponding
	// to the given name as seen by the given player. If the game does not exist or
	// the player is not a participant, it should return an error.
	ViewState(gameName string, playerName string) (*game.PlayerView, error)

	// ViewAllWithPlayer should return a slice of read-only views on all the games in the
	// collection which have the given player as a participant. It should return an
	// error if there is a problem wrapping any of the read-only game states in a view.
	// The order is not mandated, and may even change with repeated calls to the same
	// unchanged collection (analogously to the entry set of a standard Golang map, for
	// example), though of course an implementation may order the slice consistently.
	ViewAllWithPlayer(playerName string) ([]*game.PlayerView, error)

	// RecordChatMessage should find the given game and record the given chat message
	// from the given player, or return an error.
	RecordChatMessage(
		gameName string,
		playerName string,
		chatMessage string) error

	// AddNew should add a new game to the collection based on the given arguments.
	AddNew(
		gameName string,
		gameRuleset game.Ruleset,
		playerNames []string) error
}
