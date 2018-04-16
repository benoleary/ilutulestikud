package chat_test

import (
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/chat"
	"github.com/benoleary/ilutulestikud/backend/chat/assertchat"
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

			assertchat.LogIsCorrect(
				unitTest,
				testCase.name,
				testCase.arguments.messagesToAppend,
				loggedMessages)
		})
	}
}
