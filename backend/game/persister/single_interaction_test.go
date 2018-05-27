package persister_test

import (
	"testing"
	"time"
)

func TestErrorFromInvalidPlayerVisibleHand(unitTest *testing.T) {
	initialDeck := defaultTestRuleset.CopyOfFullCardset()

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
			"visible hand for invalid player/" + gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			invalidPlayer := "Invalid Player"
			soughtIndex := 0
			visibleCard, errorFromGet :=
				gameAndDescription.GameState.Read().VisibleCardInHand(invalidPlayer, soughtIndex)

			if errorFromGet == nil {
				unitTest.Fatalf(
					"VisibleCardInHand(%v, %v) %v did not produce expected error",
					invalidPlayer,
					soughtIndex,
					visibleCard)
			}
		})
	}
}

func TestErrorFromInvalidPlayerInferredHand(unitTest *testing.T) {
	initialDeck := defaultTestRuleset.CopyOfFullCardset()

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
			"inferred hand for invalid player/" + gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			invalidPlayer := "Invalid Player"
			soughtIndex := 0
			inferredCard, errorFromGet :=
				gameAndDescription.GameState.Read().InferredCardInHand(invalidPlayer, soughtIndex)

			if errorFromGet == nil {
				unitTest.Fatalf(
					"InferredCardInHand(%v, %v) %v did not produce expected error",
					invalidPlayer,
					soughtIndex,
					inferredCard)
			}
		})
	}
}

func TestRecordAndRetrieveSingleChatMessage(unitTest *testing.T) {
	testStartTime := time.Now()
	initialDeck := defaultTestRuleset.CopyOfFullCardset()

	// A nil initial action log should not be a problem for this test.
	gamesAndDescriptions :=
		prepareGameStates(
			unitTest,
			defaultTestRuleset,
			threePlayersWithHands,
			initialDeck,
			nil)

	chattingPlayer := &mockPlayerState{
		threePlayersWithHands[0].PlayerName,
		"test color",
	}

	testMessage := "test message!"

	for _, gameAndDescription := range gamesAndDescriptions {
		testIdentifier :=
			"single chat message/" + gameAndDescription.PersisterDescription

		unitTest.Run(testIdentifier, func(unitTest *testing.T) {
			errorFromRecord :=
				gameAndDescription.GameState.RecordChatMessage(chattingPlayer, testMessage)

			if errorFromRecord != nil {
				unitTest.Fatalf(
					"RecordChatMessage(%v, %v) produced error %v",
					chattingPlayer,
					testMessage,
					errorFromRecord)
			}

			retrievedChatLog := gameAndDescription.GameState.Read().ChatLog()

			if len(retrievedChatLog) != logLengthForTest {
				unitTest.Fatalf(
					"ChatLog() had wrong number of messages %v, expected %v",
					retrievedChatLog,
					logLengthForTest)
			}

			// The first message starts at the end of the log, since there
			// have been no other messages.
			firstMessage := retrievedChatLog[logLengthForTest-1]
			if (firstMessage.PlayerName() != chattingPlayer.Name()) ||
				(firstMessage.TextColor() != chattingPlayer.Color()) ||
				(firstMessage.MessageText() != testMessage) {
				unitTest.Fatalf(
					"first message %+v did not have expected player %+v",
					firstMessage,
					chattingPlayer)
			}

			recordingTime := firstMessage.CreationTime()
			currentTime := time.Now()
			if (recordingTime.Before(testStartTime)) ||
				(recordingTime.After(currentTime)) {
				unitTest.Fatalf(
					"first message %v was not between %v and %v",
					firstMessage,
					testStartTime,
					currentTime)
			}

			for messageIndex := 0; messageIndex < logLengthForTest-1; messageIndex++ {
				if (retrievedChatLog[messageIndex].PlayerName() != "") ||
					(retrievedChatLog[messageIndex].TextColor() != "") ||
					(retrievedChatLog[messageIndex].MessageText() != "") {
					unitTest.Errorf(
						"ChatLog() %+v had non-empty message",
						retrievedChatLog)
				}
			}
		})
	}
}
