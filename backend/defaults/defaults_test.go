package defaults_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/defaults"
)

func TestInitialPlayerNamesNotEmpty(unitTest *testing.T) {
	availableColors := defaults.AvailableColors()

	if len(availableColors) < 2 {
		unitTest.Fatalf(
			"defaults.AvailableColors() %v had less than 2 elements",
			availableColors)
	}
}
