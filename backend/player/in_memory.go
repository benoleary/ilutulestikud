package player

import (
	"fmt"
	"sync"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// InMemoryPersister stores inMemoryState objects as instances
// of the implementation of the readAndWriteState interface. The
// players are mapped to by their identifiers.
type InMemoryPersister struct {
	mutualExclusion    sync.Mutex
	playerStates       map[string]*inMemoryState
	nameToIdentifier   endpoint.NameToIdentifier
	initialPlayerNames map[string]bool
}

// NewInMemoryPersister creates a stateCollection around a map of
// players created from the given initial player names, with colors
// according to the available chat colors.
func NewInMemoryPersister(
	nameToIdentifier endpoint.NameToIdentifier) *InMemoryPersister {
	return &InMemoryPersister{
		mutualExclusion:    sync.Mutex{},
		playerStates:       make(map[string]*inMemoryState, 0),
		nameToIdentifier:   nameToIdentifier,
		initialPlayerNames: make(map[string]bool, 0),
	}
}

// add creates a new inMemoryState object with the given name and
// given color, and adds a reference to it into the collection.
func (inMemoryPersister *InMemoryPersister) add(
	playerDefinition endpoint.PlayerState) (string, error) {
	playerName := playerDefinition.Name
	playerIdentifier :=
		inMemoryPersister.nameToIdentifier.Identifier(playerName)

	_, playerExists := inMemoryPersister.playerStates[playerIdentifier]

	if playerExists {
		return "", fmt.Errorf("Player %v already exists", playerName)
	}

	inMemoryPersister.mutualExclusion.Lock()

	inMemoryPersister.playerStates[playerIdentifier] = &inMemoryState{
		mutualExclusion: sync.Mutex{},
		identifier:      playerIdentifier,
		name:            playerName,
		color:           playerDefinition.Color,
	}

	inMemoryPersister.mutualExclusion.Unlock()

	return playerIdentifier, nil
}

// updateFromPresentAttributes updates the player identified by the endpoint.PlayerState
// by over-writing all non-name string attributes with those from updaterReference, except
// for strings in updaterReference which are empty strings. It uses a mutex to ensure thread
// safety.
func (inMemoryPersister *InMemoryPersister) updateFromPresentAttributes(
	updaterReference endpoint.PlayerState) error {
	playerToUpdate, playerExists := inMemoryPersister.playerStates[updaterReference.Identifier]

	if !playerExists {
		return fmt.Errorf(
			"No player with identifier %v is registered",
			updaterReference.Identifier)
	}

	// It would be more efficient to only lock if we go into an if statement,
	// but then multiple if statements would be less efficient, and there would
	// be a mutex in each if statement.
	inMemoryPersister.mutualExclusion.Lock()

	if updaterReference.Color != "" {
		playerToUpdate.color = updaterReference.Color
	}

	inMemoryPersister.mutualExclusion.Unlock()

	return nil
}

// get returns the ReadOnly corresponding to the given player identifier if it exists
// already along with an error which is nil if there was no problem. If the player does
// not exist, a non-nil error is returned along with a nil ReadOnly.
func (inMemoryPersister *InMemoryPersister) get(playerIdentifier string) (ReadonlyState, error) {
	playerState, playerExists := inMemoryPersister.playerStates[playerIdentifier]
	if !playerExists {
		return nil, fmt.Errorf(
			"No player with identifier %v is registered",
			playerIdentifier)
	}

	return playerState, nil
}

// all returns a slice of all the players in the collection as
// ReadonlyState instances, ordered in the random way the iteration
// over the entries of a Golang map normally is.
func (inMemoryPersister *InMemoryPersister) all() []ReadonlyState {
	playerList := make([]ReadonlyState, 0, len(inMemoryPersister.playerStates))
	for _, playerState := range inMemoryPersister.playerStates {
		playerList = append(playerList, playerState)
	}

	return playerList
}

// reset removes all players which are not among the initial players.
// It does not restore any initial players who have been removed as
// there is no possibility to remove them anyway.
func (inMemoryPersister *InMemoryPersister) reset() {
	playersToRemove := make([]string, 0)
	for _, playerState := range inMemoryPersister.playerStates {
		if !inMemoryPersister.initialPlayerNames[playerState.Name()] {
			playersToRemove = append(playersToRemove, playerState.Identifier())
		}
	}

	for _, playerToRemove := range playersToRemove {
		delete(inMemoryPersister.playerStates, playerToRemove)
	}
}

// inMemoryState encapsulates all the state that the backend needs to know about a player,
// using a mutex to ensure that updates are thread-safe.
type inMemoryState struct {
	mutualExclusion sync.Mutex
	identifier      string
	name            string
	color           string
}

// Identifier returns the private identifier field.
func (playerState *inMemoryState) Identifier() string {
	return playerState.identifier
}

// Name returns the private name field.
func (playerState *inMemoryState) Name() string {
	return playerState.name
}

// Color returns the private color field.
func (playerState *inMemoryState) Color() string {
	return playerState.color
}
