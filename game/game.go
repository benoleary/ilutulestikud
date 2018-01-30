package game

import (
	"github.com/benoleary/ilutulestikud/player"
	"sync"
)

// State is a struct meant to encapsulate all the state required for a game to function.
type State struct {
	participatingPlayers []player.State
	mutualExclusion      sync.Mutex
}
