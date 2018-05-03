package game_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/player"
)

func TestInitialState(unitTest *testing.T) {
	gameName := "test game"

	testCases := []struct {
		testName          string
		playerNames       []string
		rulesetIdentifier int
	}{
		{
			nametestNameOfTest: "Two players, no rainbow",
			playerNames:        []string{"Player One", "Player Two"},
			rulesetIdentifier:  game.StandardWithoutRainbowIdentifier,
		},
		{
			testName:           "Three players, no rainbow",
			initialPlayerNames: []string{"Player One", "Player Two", "Player Three"},
			rulesetIdentifier:  game.StandardWithoutRainbowIdentifier,
		},
		{
			testName:          "Four players, no rainbow",
			playerNames:       []string{"Player One", "Player Two", "Player Three", "Player Four"},
			rulesetIdentifier: game.StandardWithoutRainbowIdentifier,
		},
		{
			testName:           "Five players, no rainbow",
			initialPlayerNames: []string{"Player One", "Player Two", "Player Three", "Player Four", "Player Five"},
			rulesetIdentifier:  game.StandardWithoutRainbowIdentifier,
		},
		{
			testName:          "Two players, with rainbow (as separate, but doesn't matter for initial state)",
			playerNames:       []string{"Player One", "Player Two"},
			rulesetIdentifier: game.WithRainbowAsSeparateIdentifier,
		},
		{
			testName:          "Five players, with rainbow (as compound, but doesn't matter for initial state)",
			playerNames:       []string{"Player One", "Player Two", "Player Three", "Player Four", "Player Five"},
			rulesetIdentifier: game.WithRainbowAsCompoundIdentifier,
		},
	}

	for _, testCase := range testCases {
		collectionTypes := prepareCollections(unitTest)

		for _, collectionType := range collectionTypes {
			testIdentifier := testCase.testName + "/" + collectionType.CollectionDescription

			unitTest.Run(testIdentifier, func(unitTest *testing.T) {
				testRuleset, errorFromRuleset := game.RulesetFromIdentifier(testCase.rulesetIdentifier)

				if errorFromRuleset != nil {
					unitTest.Fatalf(
						"game.RulesetFromIdentifier(ruleset identifier %v) produced an error: %v",
						testCase.rulesetIdentifier,
						errorFromRuleset)
				}

				gameDeck := gameRuleset.CopyOfFullCardset()

				game.ShuffleInPlace(gameDeck, gameCollection.statePersister.randomSeed())

				errorFromAdd :=
					collectionType.GameCollection.AddNewWithGivenDeck(
						gameName,
						testRuleset,
						testCase.playerNames,
						gameDeck)

				if errorFromAdd != nil {
					unitTest.Fatalf(
						"AddNew(game name %v, ruleset %v, player names %v) produced an error: %v",
						gameName,
						testRuleset,
						testCase.playerNames)
				}

				firstPlayerView :=
					collectionType.GameCollection.ViewState(gameName, testCase.playerNames[0])
				secondPlayerView :=
					collectionType.GameCollection.ViewState(gameName, testCase.playerNames[1])

				assertHandsAreCorrect(
					unitTest,
					testCase.playerNames,
					firstPlayerView,
					secondPlayerView)

				// deck
				// sequnces
				// discards
				// score
				// turn
				// hints
				// mistakes
				// player order
			})
		}
	}
}

func assertHandsAreCorrect(
	unitTest *testing.T,
	playerNames []string,
	firstPlayerView game.PlayerView,
	secondPlayerView game.PlayerView) {
	// We use the second player's view to directly check the hand of the first
	// player, then use the first player's view to check all the other hands.

}

func assertThatParticipantsAreCorrect(
	unitTest *testing.T,
	playerNames []string,
	participatingPlayers []player.ReadonlyState) {
	numberOfPlayers := len(playerNames)
	namesToFind := make(map[string]bool, numberOfPlayers)
	for _, playerName := range playerNames {
		namesToFind[playerName] = true
	}

	if len(participatingPlayers) != numberOfPlayers {
		unitTest.Fatalf(
			"Expected %v participants %v but retrieved",
			numberOfPlayers,
			participatingPlayers)
	}

	for _, participatingPlayer := range participatingPlayers {
		if !namesToFind[participatingPlayer.Name()] {
			unitTest.Errorf(
				"Input participants %v does not include retrieved participant %v",
				playerNames,
				participatingPlayer.Name())
		}

		namesToFind[participatingPlayer.Name()] = false
	}

	for playerName, nameIsMissing := range namesToFind {
		if nameIsMissing {
			unitTest.Errorf(
				"Input participant %v was not found in retrieve participant list %v",
				playerName,
				participatingPlayers)
		}
	}
}

func assertThatMechanicalGameStateIsCorrect(
	identifyingLabel string,
	unitTest *testing.T,
	numberOfPlayers int,
	gameRuleset game.Ruleset,
	expectedView endpoint.GameView,
	actualView endpoint.GameView) {
	if actualView.ScoreSoFar != expectedView.ScoreSoFar {
		unitTest.Errorf(
			identifyingLabel+": score was %v rather than expected %v",
			actualView.ScoreSoFar,
			expectedView.ScoreSoFar)
	}

	if actualView.NumberOfReadyHints != game.MaximumNumberOfHints {
		unitTest.Errorf(
			identifyingLabel+": number of hints was %v rather than expected %v",
			actualView.NumberOfReadyHints,
			expectedView.NumberOfReadyHints)
	}

	if actualView.NumberOfSpentHints != expectedView.NumberOfSpentHints {
		unitTest.Errorf(
			identifyingLabel+": number of spent hints was %v rather than expected %v",
			actualView.NumberOfSpentHints,
			expectedView.NumberOfSpentHints)
	}

	if actualView.NumberOfMistakesStillAllowed != game.MaximumNumberOfMistakesAllowed {
		unitTest.Errorf(
			identifyingLabel+": number of mistakes still allowed was %v rather than expected %v",
			actualView.NumberOfMistakesStillAllowed,
			expectedView.NumberOfMistakesStillAllowed)
	}

	if actualView.NumberOfMistakesMade != 0 {
		unitTest.Errorf(
			identifyingLabel+": number of mistakes made was %v rather than expected %v",
			actualView.NumberOfMistakesMade,
			0)
	}

	if actualView.NumberOfCardsLeftInDeck != expectedView.NumberOfCardsLeftInDeck {
		unitTest.Errorf(
			identifyingLabel+": number of cards in deck was %v rather than expected %v",
			actualView.NumberOfCardsLeftInDeck,
			expectedView.NumberOfCardsLeftInDeck)
	}

	if len(actualView.PlayedCards) != len(expectedView.PlayedCards) {
		unitTest.Errorf(
			identifyingLabel+": played cards set was %v rather than expected %v",
			actualView.PlayedCards,
			expectedView.PlayedCards)
	}

	unitTest.Errorf("Need to properly compare actualView.PlayedCards to expectedView.PlayedCards")

	if len(actualView.DiscardedCards) != len(expectedView.DiscardedCards) {
		unitTest.Errorf(
			identifyingLabel+": discarded cards set was %v rather than expected %v",
			actualView.DiscardedCards,
			expectedView.DiscardedCards)
	}

	unitTest.Errorf("Need to properly compare actualView.DiscardedCards to expectedView.DiscardedCards")

	if len(actualView.ThisPlayerHand) != len(expectedView.ThisPlayerHand) {
		unitTest.Errorf(
			identifyingLabel+": player hand card was %v rather than expected %v",
			actualView.ThisPlayerHand,
			expectedView.ThisPlayerHand)
	}

	unitTest.Errorf("Need to properly compare actualView.ThisPlayerHand to expectedView.ThisPlayerHand")

	expectedNumberOfVisibleHands := numberOfPlayers - 1
	if len(actualView.OtherPlayerHands) != expectedNumberOfVisibleHands {
		unitTest.Errorf(
			identifyingLabel+": visible player hands was %v rather than expected %v hands",
			actualView.ThisPlayerHand,
			expectedNumberOfVisibleHands)
	}
}
