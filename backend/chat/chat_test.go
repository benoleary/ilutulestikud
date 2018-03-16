package chat_test

import (
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/chat"
	"github.com/benoleary/ilutulestikud/backend/chat/assertchat"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

func PrepareMessages(
	numberOfMessages int,
	playerNames []string,
	chatColors []string) []endpoint.ChatLogMessage {
	numberOfPlayers := len(playerNames)
	numberOfColors := len(chatColors)
	preparedMessages := make([]endpoint.ChatLogMessage, numberOfMessages)
	for messageIndex := 0; messageIndex < numberOfMessages; messageIndex++ {
		preparedMessages[messageIndex] =
			endpoint.ChatLogMessage{
				PlayerName:  playerNames[messageIndex%numberOfPlayers],
				ChatColor:   chatColors[messageIndex%numberOfColors],
				MessageText: fmt.Sprintf("Test message %v", messageIndex),
			}
	}

	return preparedMessages
}

func TestLogForFrontendAfterAppending(unitTest *testing.T) {
	playerNames := []string{"Player One", "Player Two"}
	chatColors := []string{"red", "green", "blue"}
	type testArguments struct {
		messagesToAppend []endpoint.ChatLogMessage
	}

	testCases := []struct {
		name      string
		arguments testArguments
	}{
		{
			name: "No messages",
			arguments: testArguments{
				messagesToAppend: []endpoint.ChatLogMessage{},
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

			loggedMessages := chatLog.ForFrontend()

			assertchat.LogIsCorrect(
				unitTest,
				testCase.arguments.messagesToAppend,
				loggedMessages)
		})
	}
}
