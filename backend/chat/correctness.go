package chat

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// AssertLogCorrect fails the given unit test if the given slices do not match,
// with the logical equivalent of first padding out the expected messages list
// by prepending empty messages until it is the correct length, and truncating
// the actual messages from the front so that it is also the correct length.
func AssertLogCorrect(
	unitTest *testing.T,
	expectedWithoutEmpties []endpoint.ChatLogMessage,
	actualWithEmpties []endpoint.ChatLogMessage) {
	if len(actualWithEmpties) != LogSize {
		unitTest.Fatalf(
			"Log did not have correct size: expected %v messages, but log was %v",
			LogSize,
			actualWithEmpties)
	}

	numberOfSentMessages := len(expectedWithoutEmpties)

	// We work our way backwards from the latest message.
	for reverseIndex := 1; reverseIndex <= LogSize; reverseIndex++ {
		loggedMessage := actualWithEmpties[LogSize-reverseIndex]

		if reverseIndex > numberOfSentMessages {
			if (loggedMessage.PlayerName != "") ||
				(loggedMessage.ChatColor != "") ||
				(loggedMessage.MessageText != "") {
				unitTest.Fatalf(
					"Expected empty message with index %v, instead was %v",
					LogSize-reverseIndex,
					loggedMessage)
			}
		} else {
			expectedMessage :=
				expectedWithoutEmpties[numberOfSentMessages-reverseIndex]
			if (loggedMessage.PlayerName != expectedMessage.PlayerName) ||
				(loggedMessage.ChatColor != expectedMessage.ChatColor) ||
				(loggedMessage.MessageText != expectedMessage.MessageText) {
				unitTest.Fatalf(
					"For log index %v, expected %v, instead was %v",
					LogSize-reverseIndex,
					expectedMessage,
					loggedMessage)
			}
		}
	}
}
