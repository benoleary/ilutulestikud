package message_test

import (
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game/message"
)

func TestConstructor(unitTest *testing.T) {
	testStartTime := time.Now()
	testPlayerName := "Test Player"
	testTextColor := "test color"
	testMessageText := "test message"

	validReadonly :=
		message.NewFromPlayer(testPlayerName, testTextColor, testMessageText)

	if (validReadonly.PlayerName != testPlayerName) ||
		(validReadonly.TextColor != testTextColor) ||
		(validReadonly.MessageText != testMessageText) {
		unitTest.Fatalf(
			"NewReadonly(%v, %v, %v) produced unexpected %+v",
			testPlayerName,
			testTextColor,
			testMessageText,
			validReadonly)
	}

	if testStartTime.After(validReadonly.CreationTime) ||
		time.Now().Before(validReadonly.CreationTime) {
		unitTest.Fatalf(
			"NewReadonly(%v, %v, %v) produced %+v which has creation time outside valid range",
			testPlayerName,
			testTextColor,
			testMessageText,
			validReadonly)
	}
}
