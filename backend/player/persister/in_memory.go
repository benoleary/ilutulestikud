package persister

import (
	"fmt"
	"sync"

	"github.com/benoleary/ilutulestikud/backend/player"
)

// inMemoryPersister stores inMemoryState objects as instances of the
// implementation of the ReadonlyState interface. The players are
// mapped to by their names.
type inMemoryPersister struct {
	mutualExclusion sync.Mutex
	playerStates    map[string]*player.ReadAndWriteState
}

// NewInMemory creates a player state persister around a map of players
// created from the given initial player names, with colors according to
// the available chat colors.
func NewInMemory() player.StatePersister {
	return &inMemoryPersister{
		mutualExclusion: sync.Mutex{},
		playerStates:    make(map[string]*player.ReadAndWriteState, 0),
	}
}

// Add creates a new inMemoryState object with the given name and given
// color, and adds a reference to it into the collection. It returns an
// error if the player already exists.
func (playerPersister *inMemoryPersister) Add(
	playerName string,
	chatColor string) error {
	_, playerExists := playerPersister.playerStates[playerName]

	if playerExists {
		return fmt.Errorf("Player %v already exists", playerName)
	}

	playerPersister.mutualExclusion.Lock()

	playerPersister.playerStates[playerName] =
		&player.ReadAndWriteState{
			PlayerName: playerName,
			ChatColor:  chatColor,
		}

	playerPersister.mutualExclusion.Unlock()

	return nil
}

// UpdateColor updates the given player to have the given chat color. It
// uses a mutex to ensure thread safety.
func (playerPersister *inMemoryPersister) UpdateColor(
	playerName string,
	chatColor string) error {
	playerToUpdate, playerExists :=
		playerPersister.playerStates[playerName]

	if !playerExists {
		return fmt.Errorf(
			"No player with name %v is registered",
			playerName)
	}

	playerPersister.mutualExclusion.Lock()
	playerToUpdate.ChatColor = chatColor
	playerPersister.mutualExclusion.Unlock()

	return nil
}

// Get returns the ReadOnly corresponding to the given player identifier if
// it exists already along with an error which is nil if there was no problem.
// If the player does not exist, a non-nil error is returned along with a nil
// ReadonlyState.
func (playerPersister *inMemoryPersister) Get(
	playerName string) (player.ReadonlyState, error) {
	playerState, playerExists := playerPersister.playerStates[playerName]
	if !playerExists {
		errorToReturn :=
			fmt.Errorf(
				"No player with name %v is registered",
				playerName)
		return nil, errorToReturn
	}

	return playerState, nil
}

// All returns a slice of all the players in the collection as ReadonlyState
// instances, ordered in the random way the iteration over the entries of a
// Golang map normally is.
func (playerPersister *inMemoryPersister) All() ([]player.ReadonlyState, error) {
	playerList :=
		make([]player.ReadonlyState, 0, len(playerPersister.playerStates))
	for _, playerState := range playerPersister.playerStates {
		playerList = append(playerList, playerState)
	}

	return playerList, nil
}

// Delete deletes the given player from the collection. It returns an error
// if the player does not exist before the deletion attempt.
func (playerPersister *inMemoryPersister) Delete(playerName string) error {
	_, playerExists := playerPersister.playerStates[playerName]

	if !playerExists {
		return fmt.Errorf("No player %v exists to delete", playerName)
	}

	delete(playerPersister.playerStates, playerName)

	return nil
}
