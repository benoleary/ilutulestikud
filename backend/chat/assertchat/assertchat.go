package assertchat

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/chat"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// LogIsCorrect fails the given unit test if the given slices do not match,
// with the logical equivalent of first padding out the expected messages list
// by prepending empty messages until it is the correct length, and truncating
// the actual messages from the front so that it is also the correct length.
func LogIsCorrect(
	unitTest *testing.T,
	testIdentifier string,
	expectedWithoutEmpties []endpoint.ChatLogMessage,
	actualWithEmpties []endpoint.ChatLogMessage) {
	if len(actualWithEmpties) != chat.LogSize {
		unitTest.Fatalf(
			testIdentifier+" - log did not have correct size: expected %v messages, but log was %v",
			chat.LogSize,
			actualWithEmpties)
	}

	numberOfSentMessages := len(expectedWithoutEmpties)

	// We work our way backwards from the latest message.
	for reverseIndex := 1; reverseIndex <= chat.LogSize; reverseIndex++ {
		loggedMessage := actualWithEmpties[chat.LogSize-reverseIndex]

		if reverseIndex > numberOfSentMessages {
			if (loggedMessage.PlayerName != "") ||
				(loggedMessage.ChatColor != "") ||
				(loggedMessage.MessageText != "") {
				unitTest.Errorf(
					"Expected empty message with index %v, instead was %v",
					chat.LogSize-reverseIndex,
					loggedMessage)
			}
		} else {
			expectedMessage :=
				expectedWithoutEmpties[numberOfSentMessages-reverseIndex]
			if (loggedMessage.PlayerName != expectedMessage.PlayerName) ||
				(loggedMessage.ChatColor != expectedMessage.ChatColor) ||
				(loggedMessage.MessageText != expectedMessage.MessageText) {
				unitTest.Errorf(
					"For log index %v, expected %v, instead was %v",
					chat.LogSize-reverseIndex,
					expectedMessage,
					loggedMessage)
			}
		}
	}
}
