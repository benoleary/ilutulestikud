package game_test

import (
	"testing"

	"github.com/benoleary/ilutulestikud/backend/chat/assertchat"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/player"
)

func prepareImplementations(
	unitTest *testing.T,
	gameName string,
	rulesetIdentifier int,
	playerNames []string) ([]game.ReadonlyState, game.Ruleset) {
	gameRuleset, identifierError := game.RulesetFromIdentifier(rulesetIdentifier)

	if identifierError != nil {
		unitTest.Fatalf(
			"Unable to get valid ruleset for identifier %v: error is %v",
			rulesetIdentifier,
			identifierError)
	}

	if len(playerNames) < gameRuleset.MinimumNumberOfPlayers() {
		unitTest.Fatalf(
			"Not enough players: %v",
			playerNames)
	}

	if len(playerNames) > gameRuleset.MaximumNumberOfPlayers() {
		unitTest.Fatalf(
			"Too many players: %v",
			playerNames)
	}

	nameToIdentifier := &endpoint.Base32NameEncoder{}
	playerCollection := player.NewInMemoryPersister(nameToIdentifier, playerNames, []string{"red", "green", "blue"})
	gameCollections := []game.StateCollection{
		game.NewInMemoryCollection(nameToIdentifier),
	}

	playerIdentifiers := make([]string, len(playerNames))
	for playerIndex, playerName := range playerNames {
		playerIdentifiers[playerIndex] = nameToIdentifier.Identifier(playerName)
	}

	gameDefinition := endpoint.GameDefinition{
		GameName:          gameName,
		RulesetIdentifier: rulesetIdentifier,
		PlayerIdentifiers: playerIdentifiers,
	}

	gameStates := make([]game.ReadonlyState, len(gameCollections))

	for collectionIndex, gameCollection := range gameCollections {
		gameIdentifier, addError :=
			game.AddNew(gameDefinition, gameCollection, playerCollection)

		if addError != nil {
			unitTest.Fatalf(
				"Error when trying to add game for collection index %v: %v",
				collectionIndex,
				addError)
		}

		addedGame, gameExists := game.ReadState(gameCollection, gameIdentifier)
		if !gameExists {
			unitTest.Fatalf(
				"Error when trying to find identifier %v for game for collection index %v: gameCollection = %v",
				gameIdentifier,
				collectionIndex,
				gameCollection)
		}

		gameStates[collectionIndex] = addedGame
	}

	return gameStates, gameRuleset
}

func TestInitialState(unitTest *testing.T) {
	type testArguments struct {
		initialPlayerNames []string
		rulesetIdentifier  int
	}

	testCases := []struct {
		name      string
		arguments testArguments
	}{
		{
			name: "Two players, no rainbow",
			arguments: testArguments{
				initialPlayerNames: []string{"Player One", "Player Two"},
				rulesetIdentifier:  game.StandardWithoutRainbowIdentifier,
			},
		},
		{
			name: "Three players, no rainbow",
			arguments: testArguments{
				initialPlayerNames: []string{"Player One", "Player Two", "Player Three"},
				rulesetIdentifier:  game.StandardWithoutRainbowIdentifier,
			},
		},
		{
			name: "Four players, no rainbow",
			arguments: testArguments{
				initialPlayerNames: []string{"Player One", "Player Two", "Player Three", "Player Four"},
				rulesetIdentifier:  game.StandardWithoutRainbowIdentifier,
			},
		},
		{
			name: "Five players, no rainbow",
			arguments: testArguments{
				initialPlayerNames: []string{"Player One", "Player Two", "Player Three", "Player Four", "Player Five"},
				rulesetIdentifier:  game.StandardWithoutRainbowIdentifier,
			},
		},
		{
			name: "Two players, with rainbow (as separate, but doesn't matter for initial state)",
			arguments: testArguments{
				initialPlayerNames: []string{"Player One", "Player Two"},
				rulesetIdentifier:  game.WithRainbowAsSeparateIdentifier,
			},
		},
		{
			name: "Five players, with rainbow (as compound, but doesn't matter for initial state)",
			arguments: testArguments{
				initialPlayerNames: []string{"Player One", "Player Two", "Player Three", "Player Four", "Player Five"},
				rulesetIdentifier:  game.WithRainbowAsCompoundIdentifier,
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			gameStates, gameRuleset := prepareImplementations(
				unitTest,
				"Test game",
				testCase.arguments.rulesetIdentifier,
				testCase.arguments.initialPlayerNames)

			numberOfPlayers := len(testCase.arguments.initialPlayerNames)

			for stateIndex := 0; stateIndex < len(gameStates); stateIndex++ {
				gameState := gameStates[stateIndex]
				participatingPlayers := gameState.Players()

				AssertThatParticipantsAreCorrect(
					unitTest,
					testCase.arguments.initialPlayerNames,
					participatingPlayers)

				viewingPlayer := participatingPlayers[0]

				gameView := game.ForPlayer(gameState, viewingPlayer.Identifier())

				assertchat.LogIsCorrect(
					unitTest,
					testCase.name,
					[]endpoint.ChatLogMessage{},
					gameView.ChatLog,
				)

				numberOfCardsPerHand := gameRuleset.NumberOfCardsInPlayerHand(numberOfPlayers)
				expectedNumberOfCardsInPlayerHands :=
					numberOfPlayers * numberOfCardsPerHand
				colorSuits := gameRuleset.ColorSuits()
				sequenceIndices := gameRuleset.SequenceIndices()
				numberOfCardsInTotal := len(colorSuits) * len(sequenceIndices)
				expectedNumberOfCardsInDeck := numberOfCardsInTotal - expectedNumberOfCardsInPlayerHands

				expectedVisibleHands := make([]endpoint.VisibleHand, numberOfPlayers)

				// We start from index 1 as player 0 is the viewing player.
				for playerIndex := 1; playerIndex < numberOfPlayers; playerIndex++ {
					expectedVisibleHands[playerIndex] = endpoint.VisibleHand{
						PlayerIdentifier: participatingPlayers[playerIndex].Identifier(),
						PlayerName:       participatingPlayers[playerIndex].Name(),
						HandCards:        make([]endpoint.VisibleCard, numberOfCardsPerHand),
					}
				}

				AssertThatMechanicalGameStateIsCorrect(
					"Initial state",
					unitTest,
					len(testCase.arguments.initialPlayerNames),
					gameRuleset,
					endpoint.GameView{
						ScoreSoFar:                   0,
						NumberOfReadyHints:           game.MaximumNumberOfHints,
						NumberOfSpentHints:           0,
						NumberOfMistakesStillAllowed: game.MaximumNumberOfMistakesAllowed,
						NumberOfMistakesMade:         0,
						NumberOfCardsLeftInDeck:      expectedNumberOfCardsInDeck,
						PlayedCards:                  [][]endpoint.VisibleCard{},
						DiscardedCards:               [][]endpoint.VisibleCard{},
						ThisPlayerHand:               make([]endpoint.CardFromBehind, numberOfCardsPerHand),
						OtherPlayerHands:             expectedVisibleHands,
					},
					gameView)
			}
		})
	}
}

func AssertThatParticipantsAreCorrect(
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

func AssertThatMechanicalGameStateIsCorrect(
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
