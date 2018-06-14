package game_test

import (
	"sort"
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game"
)

func TestOrderByCreationTime(unitTest *testing.T) {
	farFuture := NewMockGameState(unitTest)
	farFuture.ReturnForName = "Far future"
	farFuture.ReturnForCreationTime = time.Now().Add(100 * time.Second)
	farPast := NewMockGameState(unitTest)
	farPast.ReturnForName = "Far past"
	farPast.ReturnForCreationTime = time.Now().Add(-100 * time.Second)
	nearFuture := NewMockGameState(unitTest)
	nearFuture.ReturnForName = "Near future"
	nearFuture.ReturnForCreationTime = time.Now().Add(1 * time.Second)
	nearPast := NewMockGameState(unitTest)
	nearPast.ReturnForName = "Near past"
	nearPast.ReturnForCreationTime = time.Now().Add(-1 * time.Second)

	mockGames := game.ByCreationTime([]game.ReadonlyState{
		farFuture,
		farPast,
		nearFuture,
		nearPast,
	})

	expectedNames :=
		[]string{
			farPast.ReturnForName,
			nearPast.ReturnForName,
			nearFuture.ReturnForName,
			farFuture.ReturnForName,
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
