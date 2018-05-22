package chat_test

import (
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game/chat"
)

func PrepareMessages(
	numberOfMessages int,
	playerNames []string,
	chatColors []string) []chat.Message {
	numberOfPlayers := len(playerNames)
	numberOfColors := len(chatColors)
	preparedMessages := make([]chat.Message, numberOfMessages)
	for messageIndex := 0; messageIndex < numberOfMessages; messageIndex++ {
		preparedMessages[messageIndex] =
			chat.Message{
				PlayerName:  playerNames[messageIndex%numberOfPlayers],
				ChatColor:   chatColors[messageIndex%numberOfColors],
				MessageText: fmt.Sprintf("Test message %v", messageIndex),
			}
	}

	return preparedMessages
}

func TestSortedLogAfterAppending(unitTest *testing.T) {
	playerNames := []string{"Player One", "Player Two"}
	chatColors := []string{"red", "green", "blue"}
	type testArguments struct {
		messagesToAppend []chat.Message
	}

	testCases := []struct {
		name      string
		arguments testArguments
	}{
		{
			name: "No messages",
			arguments: testArguments{
				messagesToAppend: []chat.Message{},
			},
		},
		{
			name: "One message",
			arguments: testArguments{
				messagesToAppend: PrepareMessages(1, playerNames, chatColors),
			},
		},
		{
			name: "Two messages",
			arguments: testArguments{
				messagesToAppend: PrepareMessages(2, playerNames, chatColors),
			},
		},
		{
			name: "Full log",
			arguments: testArguments{
				messagesToAppend: PrepareMessages(chat.LogSize, playerNames, chatColors),
			},
		},
		{
			name: "Overfull log",
			arguments: testArguments{
				messagesToAppend: PrepareMessages(chat.LogSize+1, playerNames, chatColors),
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			chatLog := chat.NewLog()

			for _, chatMessage := range testCase.arguments.messagesToAppend {
				chatLog.AppendNewMessage(
					chatMessage.PlayerName,
					chatMessage.ChatColor,
					chatMessage.MessageText)
			}

			loggedMessages := chatLog.Sorted()

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
	expectedWithoutEmpties []chat.Message,
	actualWithEmpties []chat.Message) {
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
