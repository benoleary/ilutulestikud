package game_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/chat"
	"github.com/benoleary/ilutulestikud/backend/chat/assertchat"
	"github.com/benoleary/ilutulestikud/backend/game"
)

func TestThreePlayersChatting(unitTest *testing.T) {
	collectionTypes := prepareCollections(unitTest)

	gameName := "test game"

	playerNames :=
		[]string{
			playerNamesAvailableInTest[2],
			playerNamesAvailableInTest[1],
			playerNamesAvailableInTest[3],
		}

	viewingPlayerName := playerNames[0]

	chatMessages := []chat.Message{
		chat.Message{
			PlayerName:  viewingPlayerName,
			ChatColor:   "red",
			MessageText: "hello",
		},
		chat.Message{
			PlayerName:  playerNames[1],
			ChatColor:   "green",
			MessageText: "Hi!",
		},
		chat.Message{
			PlayerName:  playerNames[2],
			ChatColor:   "blue",
			MessageText: "o/",
		},
		chat.Message{
			PlayerName:  viewingPlayerName,
			ChatColor:   "white",
			MessageText: ":)",
		},
	}

	for _, collectionType := range collectionTypes {
		testIdentifier := "Three players chatting test/" + collectionType.CollectionDescription

		errorFromInitialAdd := collectionType.GameCollection.AddNew(
			gameName,
			testRuleset,
			playerNames)

		if errorFromInitialAdd != nil {
			unitTest.Fatalf(
				"AddNew(game name %v, standard ruleset, player names %v) produced an error: %v",
				gameName,
				playerNames,
				errorFromInitialAdd)
		}

		// At first, there should be no chat.
		assertGetChatLogIsCorrect(
			unitTest,
			testIdentifier,
			collectionType.GameCollection,
			gameName,
			viewingPlayerName,
			[]chat.Message{})

		for messageCount := 0; messageCount < len(chatMessages); messageCount++ {
			chatMessage := chatMessages[messageCount]
			errorFromChat :=
				collectionType.GameCollection.RecordChatMessage(
					gameName,
					chatMessage.PlayerName,
					chatMessage.MessageText)

			if errorFromChat != nil {
				unitTest.Fatalf(
					"RecordChatMessage(game name %v, player name %v, chat message %v) produced an error: %v",
					gameName,
					chatMessage.PlayerName,
					chatMessage.MessageText,
					errorFromChat)
			}

			assertGetChatLogIsCorrect(
				unitTest,
				testIdentifier,
				collectionType.GameCollection,
				gameName,
				viewingPlayerName,
				chatMessages[:messageCount+1])
		}
	}
}

func assertGetChatLogIsCorrect(
	unitTest *testing.T,
	testIdentifier string,
	gameCollection *game.StateCollection,
	gameName string,
	viewingPlayerName string,
	expectedMessages []chat.Message) {
	playerKnowledge, errorFromViewState := gameCollection.ViewState(gameName, viewingPlayerName)

	if errorFromViewState != nil {
		unitTest.Fatalf(
			testIdentifier+"/ViewState(game name %v, player name %v) produced an error: %v",
			gameName,
			viewingPlayerName,
			errorFromViewState)
	}

	assertchat.LogIsCorrect(
		unitTest,
		testIdentifier,
		expectedMessages,
		playerKnowledge.SortedChatLog())
}
