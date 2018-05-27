package persister_test

import (
	"fmt"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/game/message"
)

const logSizeForTest = 8

func PrepareMessages(
	numberOfMessages int,
	playerNames []string,
	textColors []string) []message.Readonly {
	numberOfPlayers := len(playerNames)
	numberOfColors := len(textColors)
	preparedMessages := make([]message.Readonly, numberOfMessages)
	for messageIndex := 0; messageIndex < numberOfMessages; messageIndex++ {
		preparedMessages[messageIndex] =
			message.NewReadonly(
				playerNames[messageIndex%numberOfPlayers],
				textColors[messageIndex%numberOfColors],
				fmt.Sprintf("Test message %v", messageIndex))
	}

	return preparedMessages
}

func TestSortedLogAfterAppending(unitTest *testing.T) {
	initialDeck := defaultTestRuleset.CopyOfFullCardset()

	playerNames :=
		[]string{
			defaultTestPlayers[0],
			defaultTestPlayers[1],
		}

	textColors :=
		[]string{
			"red",
			"green",
			"blue",
		}

	testCases := []struct {
		testName         string
		messagesToAppend []message.Readonly
	}{
		{
			testName:         "No messages",
			messagesToAppend: []message.Readonly{},
		},
		{
			testName:         "One message",
			messagesToAppend: PrepareMessages(1, playerNames, textColors),
		},
		{
			testName:         "Two messages",
			messagesToAppend: PrepareMessages(2, playerNames, textColors),
		},
		{
			testName:         "Full log",
			messagesToAppend: PrepareMessages(logSizeForTest, playerNames, textColors),
		},
		{
			testName:         "Overfull log",
			messagesToAppend: PrepareMessages(logSizeForTest+1, playerNames, textColors),
		},
	}

	for _, testCase := range testCases {
		// A nil initial action log should not be a problem for this test.
		gamesAndDescriptions :=
			prepareGameStates(
				unitTest,
				defaultTestRuleset,
				threePlayersWithHands,
				initialDeck,
				nil)

		for _, gameAndDescription := range gamesAndDescriptions {
			testIdentifier :=
				testCase.testName + "/" + gameAndDescription.PersisterDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				for _, chatMessage := range testCase.messagesToAppend {
					playerState :=
						&mockPlayerState{
							mockName:  chatMessage.PlayerName(),
							mockColor: chatMessage.TextColor(),
						}

					gameAndDescription.GameState.RecordChatMessage(
						playerState,
						chatMessage.MessageText())
				}

				loggedMessages := gameAndDescription.GameState.Read().ChatLog()

				assertLogIsCorrect(
					unitTest,
					testIdentifier,
					testCase.messagesToAppend,
					loggedMessages)
			})
		}
	}
}

// assertLogIsCorrect fails the given unit test if the given slices do not match,
// with the logical equivalent of first padding out the expected messages list
// by prepending empty messages until it is the correct length, and truncating
// the actual messages from the front so that it is also the correct length.
func assertLogIsCorrect(
	unitTest *testing.T,
	testIdentifier string,
	expectedWithoutEmpties []message.Readonly,
	actualWithEmpties []message.Readonly) {
	if len(actualWithEmpties) != logSizeForTest {
		unitTest.Fatalf(
			testIdentifier+" - log did not have correct size: expected %v messages, but log was %+v",
			logSizeForTest,
			actualWithEmpties)
	}

	numberOfSentMessages := len(expectedWithoutEmpties)

	// We work our way backwards from the latest message.
	for reverseIndex := 1; reverseIndex <= logSizeForTest; reverseIndex++ {
		loggedMessage := actualWithEmpties[logSizeForTest-reverseIndex]

		if reverseIndex > numberOfSentMessages {
			if (loggedMessage.PlayerName() != "") ||
				(loggedMessage.TextColor() != "") ||
				(loggedMessage.MessageText() != "") {
				unitTest.Errorf(
					"Expected empty message with index %+v, but log was %+v",
					logSizeForTest-reverseIndex,
					actualWithEmpties)
			}
		} else {
			expectedMessage :=
				expectedWithoutEmpties[numberOfSentMessages-reverseIndex]
			if (loggedMessage.PlayerName() != expectedMessage.PlayerName()) ||
				(loggedMessage.TextColor() != expectedMessage.TextColor()) ||
				(loggedMessage.MessageText() != expectedMessage.MessageText()) {
				unitTest.Errorf(
					"For log index %v, expected %+v, but log was %+v",
					logSizeForTest-reverseIndex,
					expectedMessage,
					actualWithEmpties)
			}
		}
	}
}
