package game_test

import (
	"sort"
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/chat"
	"github.com/benoleary/ilutulestikud/backend/player"
)

type mockGameState struct {
	MockGameName     string
	MockCreationTime time.Time
}

// Read gets mocked.
func (gameState *mockGameState) Read() game.ReadonlyState {
	return gameState
}

// Name gets mocked.
func (gameState *mockGameState) Name() string {
	return gameState.MockGameName
}

// Ruleset gets mocked.
func (gameState *mockGameState) Ruleset() game.Ruleset {
	return testRuleset
}

// Players gets mocked.
func (gameState *mockGameState) Players() []player.ReadonlyState {
	return nil
}

// Turn gets mocked.
func (gameState *mockGameState) Turn() int {
	return -2
}

// CreationTime gets mocked.
func (gameState *mockGameState) CreationTime() time.Time {
	return gameState.MockCreationTime
}

// ChatLog gets mocked.
func (gameState *mockGameState) ChatLog() *chat.Log {
	return nil
}

// HasPlayerAsParticipant gets mocked.
func (gameState *mockGameState) HasPlayerAsParticipant(playerName string) bool {
	return false
}

// Score gets mocked.
func (gameState *mockGameState) Score() int {
	return -3
}

// NumberOfReadyHints gets mocked.
func (gameState *mockGameState) NumberOfReadyHints() int {
	return -4
}

// NumberOfMistakesMade gets mocked.
func (gameState *mockGameState) NumberOfMistakesMade() int {
	return -5
}

func TestOrderByCreationTime(unitTest *testing.T) {
	mockGames := game.ByCreationTime([]game.ReadonlyState{
		&mockGameState{
			MockGameName:     "Far future",
			MockCreationTime: time.Now().Add(100 * time.Second),
		},
		&mockGameState{
			MockGameName:     "Far past",
			MockCreationTime: time.Now().Add(-100 * time.Second),
		},
		&mockGameState{
			MockGameName:     "Near future",
			MockCreationTime: time.Now().Add(1 * time.Second),
		},
		&mockGameState{
			MockGameName:     "Near past",
			MockCreationTime: time.Now().Add(-1 * time.Second),
		},
	})

	sort.Sort(mockGames)

	if (mockGames[0].Name() != mockGames[1].Name()) ||
		(mockGames[1].Name() != mockGames[3].Name()) ||
		(mockGames[2].Name() != mockGames[2].Name()) ||
		(mockGames[3].Name() != mockGames[0].Name()) {
		unitTest.Fatalf(
			"Game states were not sorted: expected names [%v, %v, %v, %v], instead had %v",
			mockGames[1].Name(),
			mockGames[3].Name(),
			mockGames[2].Name(),
			mockGames[0].Name(),
			mockGames)
	}
}
