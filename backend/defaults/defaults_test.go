package defaults_test

import (
	"github.com/benoleary/ilutulestikud/backend/defaults"
	"testing"
)

func TestInitialPlayerNames(unitTest *testing.T) {
	initialPlayerNames := defaults.InitialPlayerNames()
	if initialPlayerNames == nil {
		unitTest.Fatalf("InitialPlayerNames() returned nil slice")
	}

	if len(initialPlayerNames) < 1 {
		unitTest.Fatalf("InitialPlayerNames() returned empty slice")
	}
}

func TestAvailableColors(unitTest *testing.T) {
	availableColors := defaults.AvailableColors()
	if availableColors == nil {
		unitTest.Fatalf("AvailableColors() returned nil slice")
	}

	if len(availableColors) < 1 {
		unitTest.Fatalf("AvailableColors() returned empty slice")
	}
}
