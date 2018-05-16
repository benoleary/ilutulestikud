package defaults_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/defaults"
)

func TestAvailableColorsNotEmpty(unitTest *testing.T) {
	initialPlayerNames := defaults.InitialPlayerNames()

	if len(initialPlayerNames) < 2 {
		unitTest.Fatalf(
			"defaults.InitialPlayerNames() %v had less than 2 elements",
			initialPlayerNames)
	}
}

func TestInitialPlayerNamesNotEmpty(unitTest *testing.T) {
	availableColors := defaults.AvailableColors()

	if len(availableColors) < 2 {
		unitTest.Fatalf(
			"defaults.AvailableColors() %v had less than 2 elements",
			availableColors)
	}
}
