package game_test

import (
	"sort"
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
)

type mockReadonly struct {
	testReference    *testing.T
	mockName         string
	mockCreationTime time.Time
}

func NewMockReadonly(
	testReference *testing.T,
	mockName string,
	mockCreationTime time.Time) *mockReadonly {
	return &mockReadonly{
		testReference:    testReference,
		mockName:         mockName,
		mockCreationTime: mockCreationTime,
	}
}

// Name gets mocked.
func (mockGame *mockReadonly) Name() string {
	return mockGame.mockName
}

// Ruleset gets mocked.
func (mockGame *mockReadonly) Ruleset() game.Ruleset {
	mockGame.testReference.Fatalf("Ruleset() should not be called.")
	return nil
}

// Players gets mocked.
func (mockGame *mockReadonly) PlayerNames() []string {
	mockGame.testReference.Fatalf("Players() should not be called.")
	return nil
}

// Turn gets mocked.
func (mockGame *mockReadonly) Turn() int {
	mockGame.testReference.Fatalf("Turn() should not be called.")
	return -1
}

// CreationTime gets mocked.
func (mockGame *mockReadonly) CreationTime() time.Time {
	return mockGame.mockCreationTime
}

// ChatLog gets mocked.
func (mockGame *mockReadonly) ChatLog() []message.Readonly {
	mockGame.testReference.Fatalf("ChatLog() should not be called.")
	return nil
}

// ActionLog gets mocked.
func (mockGame *mockReadonly) ActionLog() []message.Readonly {
	mockGame.testReference.Fatalf("ActionLog() should not be called.")
	return nil
}

// HasPlayerAsParticipant gets mocked.
func (mockGame *mockReadonly) HasPlayerAsParticipant(playerName string) bool {
	mockGame.testReference.Fatalf(
		"HasPlayerAsParticipant(%v) should not be called.",
		playerName)
	return false
}

// Score gets mocked.
func (mockGame *mockReadonly) Score() int {
	mockGame.testReference.Fatalf("Score() should not be called.")
	return -1
}

// NumberOfReadyHints gets mocked.
func (mockGame *mockReadonly) NumberOfReadyHints() int {
	mockGame.testReference.Fatalf("NumberOfReadyHints() should not be called.")
	return -1
}

// NumberOfMistakesMade gets mocked.
func (mockGame *mockReadonly) NumberOfMistakesMade() int {
	mockGame.testReference.Fatalf("NumberOfMistakesMade() should not be called.")
	return -1
}

// DeckSize gets mocked.
func (mockGame *mockReadonly) DeckSize() int {
	mockGame.testReference.Fatalf("DeckSize() should not be called.")
	return -1
}

// PlayedForColor gets mocked.
func (mockGame *mockReadonly) PlayedForColor(colorSuit string) []card.Readonly {
	mockGame.testReference.Fatalf(
		"PlayedForColor(%v) should not be called.",
		colorSuit)
	return nil
}

// NumberOfDiscardedCards gets mocked.
func (mockGame *mockReadonly) NumberOfDiscardedCards(
	colorSuit string,
	sequenceIndex int) int {
	mockGame.testReference.Fatalf(
		"NumberOfDiscardedCards(%v, %v) should not be called.",
		colorSuit,
		sequenceIndex)
	return -1
}

// VisibleCardInHand gets mocked.
func (mockGame *mockReadonly) VisibleCardInHand(
	holdingPlayerName string,
	indexInHand int) (card.Readonly, error) {
	mockGame.testReference.Fatalf(
		"VisibleCardInHand(%v, %v) should not be called.",
		holdingPlayerName,
		indexInHand)
	return card.ErrorReadonly(), nil
}

// InferredCardInHand gets mocked.
func (mockGame *mockReadonly) InferredCardInHand(
	holdingPlayerName string,
	indexInHand int) (card.Inferred, error) {
	mockGame.testReference.Fatalf(
		"VisibleCardInHand(%v, %v) should not be called.",
		holdingPlayerName,
		indexInHand)
	return card.ErrorInferred(), nil
}

func TestOrderByCreationTime(unitTest *testing.T) {
	mockGames := game.ByCreationTime([]game.ReadonlyState{
		NewMockReadonly(unitTest, "Far future", time.Now().Add(100*time.Second)),
		NewMockReadonly(unitTest, "Far past", time.Now().Add(-100*time.Second)),
		NewMockReadonly(unitTest, "Near future", time.Now().Add(1*time.Second)),
		NewMockReadonly(unitTest, "Near past", time.Now().Add(-1*time.Second)),
	})

	expectedNames :=
		[]string{
			mockGames[1].Name(),
			mockGames[3].Name(),
			mockGames[2].Name(),
			mockGames[0].Name(),
		}

	sort.Sort(mockGames)

	for gameIndex := 0; gameIndex < len(expectedNames); gameIndex++ {
		if mockGames[gameIndex].Name() != expectedNames[gameIndex] {
			unitTest.Fatalf(
				"Game states were not sorted: expected names %v, instead had %v",
				expectedNames,
				mockGames)
		}
	}
}
