package persister

import (
	"fmt"
	"sync"

	"github.com/benoleary/ilutulestikud/backend/player"
)

// InMemoryPersister stores inMemoryState objects as instances
// of the implementation of the readAndWriteState interface. The
// players are mapped to by their identifiers.
type InMemoryPersister struct {
	mutualExclusion sync.Mutex
	playerStates    map[string]*inMemoryState
}

// NewInMemoryPersister creates a stateCollection around a map of
// players created from the given initial player names, with colors
// according to the available chat colors.
func NewInMemoryPersister() *InMemoryPersister {
	return &InMemoryPersister{
		mutualExclusion: sync.Mutex{},
		playerStates:    make(map[string]*inMemoryState, 0),
	}
}

// Add creates a new inMemoryState object with the given name and given color,
// and adds a reference to it into the collection. It returns an error if the
// player already exists.
func (inMemoryPersister *InMemoryPersister) Add(playerName string, chatColor string) error {
	_, playerExists := inMemoryPersister.playerStates[playerName]

	if playerExists {
		return fmt.Errorf("Player %v already exists", playerName)
	}

	inMemoryPersister.mutualExclusion.Lock()

	inMemoryPersister.playerStates[playerName] =
		&inMemoryState{
			name:  playerName,
			color: chatColor,
		}

	inMemoryPersister.mutualExclusion.Unlock()

	return nil
}

// UpdateColor updates the given player to have the given chat color. It uses
// a mutex to ensure thread safety.
func (inMemoryPersister *InMemoryPersister) UpdateColor(
	playerName string,
	chatColor string) error {
	playerToUpdate, playerExists := inMemoryPersister.playerStates[playerName]

	if !playerExists {
		return fmt.Errorf(
			"No player with name %v is registered",
			playerName)
	}

	if chatColor != "" {
		inMemoryPersister.mutualExclusion.Lock()
		playerToUpdate.color = chatColor
		inMemoryPersister.mutualExclusion.Unlock()
	}

	return nil
}

// Get returns the ReadOnly corresponding to the given player identifier if it exists
// already along with an error which is nil if there was no problem. If the player does
// not exist, a non-nil error is returned along with a nil ReadOnly.
func (inMemoryPersister *InMemoryPersister) Get(playerName string) (player.ReadonlyState, error) {
	playerState, playerExists := inMemoryPersister.playerStates[playerName]
	if !playerExists {
		return nil, fmt.Errorf(
			"No player with name %v is registered",
			playerName)
	}

	return playerState, nil
}

// All returns a slice of all the players in the collection as
// ReadonlyState instances, ordered in the random way the iteration
// over the entries of a Golang map normally is.
func (inMemoryPersister *InMemoryPersister) All() []player.ReadonlyState {
	playerList := make([]player.ReadonlyState, 0, len(inMemoryPersister.playerStates))
	for _, playerState := range inMemoryPersister.playerStates {
		playerList = append(playerList, playerState)
	}

	return playerList
}

// Reset removes all players.
func (inMemoryPersister *InMemoryPersister) Reset() {
	for playerName := range inMemoryPersister.playerStates {
		delete(inMemoryPersister.playerStates, playerName)
	}
}

// inMemoryState encapsulates all the state that the backend needs to know about a player.
type inMemoryState struct {
	name  string
	color string
}

// Name returns the private name field.
func (playerState *inMemoryState) Name() string {
	return playerState.name
}

// Color returns the private color field.
func (playerState *inMemoryState) Color() string {
	return playerState.color
}
