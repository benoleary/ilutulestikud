package log_test

import (
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game/log"
)

const logSizeForTest = 8

func PrepareMessages(
	numberOfMessages int,
	playerNames []string,
	textColors []string) []log.Message {
	numberOfPlayers := len(playerNames)
	numberOfColors := len(textColors)
	preparedMessages := make([]log.Message, numberOfMessages)
	for messageIndex := 0; messageIndex < numberOfMessages; messageIndex++ {
		preparedMessages[messageIndex] =
			log.Message{
				PlayerName:  playerNames[messageIndex%numberOfPlayers],
				TextColor:   textColors[messageIndex%numberOfColors],
				MessageText: fmt.Sprintf("Test message %v", messageIndex),
			}
	}

	return preparedMessages
}

func TestSortedLogAfterAppending(unitTest *testing.T) {
	playerNames := []string{"Player One", "Player Two"}
	textColors := []string{"red", "green", "blue"}
	type testArguments struct {
		messagesToAppend []log.Message
	}

	testCases := []struct {
		name      string
		arguments testArguments
	}{
		{
			name: "No messages",
			arguments: testArguments{
				messagesToAppend: []log.Message{},
			},
		},
		{
			name: "One message",
			arguments: testArguments{
				messagesToAppend: PrepareMessages(1, playerNames, textColors),
			},
		},
		{
			name: "Two messages",
			arguments: testArguments{
				messagesToAppend: PrepareMessages(2, playerNames, textColors),
			},
		},
		{
			name: "Full log",
			arguments: testArguments{
				messagesToAppend: PrepareMessages(logSizeForTest, playerNames, textColors),
			},
		},
		{
			name: "Overfull log",
			arguments: testArguments{
				messagesToAppend: PrepareMessages(logSizeForTest+1, playerNames, textColors),
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			chatLog := log.NewRollingAppender(logSizeForTest)

			for _, chatMessage := range testCase.arguments.messagesToAppend {
				chatLog.AppendNewMessage(
					chatMessage.PlayerName,
					chatMessage.TextColor,
					chatMessage.MessageText)
			}

			loggedMessages := chatLog.SortedCopyOfMessages()

			assertLogIsCorrect(
				unitTest,
				testCase.name,
				testCase.arguments.messagesToAppend,
				loggedMessages)
		})
	}
}

// assertLogIsCorrect fails the given unit test if the given slices do not match,
// with the logical equivalent of first padding out the expected messages list
// by prepending empty messages until it is the correct length, and truncating
// the actual messages from the front so that it is also the correct length.
func assertLogIsCorrect(
	unitTest *testing.T,
	testIdentifier string,
	expectedWithoutEmpties []log.Message,
	actualWithEmpties []log.Message) {
	if len(actualWithEmpties) != logSizeForTest {
		unitTest.Fatalf(
			testIdentifier+" - log did not have correct size: expected %v messages, but log was %v",
			logSizeForTest,
			actualWithEmpties)
	}

	numberOfSentMessages := len(expectedWithoutEmpties)

	// We work our way backwards from the latest message.
	for reverseIndex := 1; reverseIndex <= logSizeForTest; reverseIndex++ {
		loggedMessage := actualWithEmpties[logSizeForTest-reverseIndex]

		if reverseIndex > numberOfSentMessages {
			if (loggedMessage.PlayerName != "") ||
				(loggedMessage.TextColor != "") ||
				(loggedMessage.MessageText != "") {
				unitTest.Errorf(
					"Expected empty message with index %v, instead was %v",
					logSizeForTest-reverseIndex,
					loggedMessage)
			}
		} else {
			expectedMessage :=
				expectedWithoutEmpties[numberOfSentMessages-reverseIndex]
			if (loggedMessage.PlayerName != expectedMessage.PlayerName) ||
				(loggedMessage.TextColor != expectedMessage.TextColor) ||
				(loggedMessage.MessageText != expectedMessage.MessageText) {
				unitTest.Errorf(
					"For log index %v, expected %v, instead was %v",
					logSizeForTest-reverseIndex,
					expectedMessage,
					loggedMessage)
			}
		}
	}
}
