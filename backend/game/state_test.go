package game_test

import (
	"sort"
	"testing"
	"time"

	"github.com/benoleary/ilutulestikud/backend/chat"
	"github.com/benoleary/ilutulestikud/backend/endpoint"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/player"
)

func prepareImplementations(
	unitTest *testing.T,
	gameName string,
	playerNames []string) []game.State {
	if len(playerNames) < game.MinimumNumberOfPlayers {
		unitTest.Fatalf(
			"Not enough players: %v",
			playerNames)
	}

	nameToIdentifier := &endpoint.Base64NameEncoder{}
	playerCollection := player.NewInMemoryCollection(nameToIdentifier, playerNames, []string{"red", "green", "blue"})
	gameCollections := []game.Collection{
		game.NewInMemoryCollection(nameToIdentifier),
	}

	playerIdentifiers := make([]string, len(playerNames))
	for playerIndex, playerName := range playerNames {
		playerIdentifiers[playerIndex] = nameToIdentifier.Identifier(playerName)
	}

	gameDefinition := endpoint.GameDefinition{
		Name:    gameName,
		Players: playerIdentifiers,
	}

	gameStates := make([]game.State, len(gameCollections))

	for collectionIndex, gameCollection := range gameCollections {
		addError := gameCollection.Add(gameDefinition, playerCollection)

		if addError != nil {
			unitTest.Fatalf(
				"Error when trying to add game for collection index %v: %v",
				collectionIndex,
				addError)
		}

		// We find the game identifier by looking for all the games for the first
		// player of the game definition, as there should only be one game in the
		// collection, and it should include this player.
		playerIdentifier := gameDefinition.Players[0]
		allGames := gameCollection.All(playerIdentifier)

		if (allGames == nil) || (len(allGames) != 1) {
			unitTest.Fatalf(
				"Error when trying to add find identifier for game for collection index %v: allGames = %v",
				collectionIndex,
				allGames)
		}

		gameStates[collectionIndex] = allGames[0]
	}

	return gameStates
}

func TestOrderByCreationTime(unitTest *testing.T) {
	playerNames := []string{"Player One", "Player Two", "Player Three"}
	earlyGameStates := prepareImplementations(
		unitTest,
		"Early game",
		playerNames)

	time.Sleep(100 * time.Millisecond)

	lateGameStates := prepareImplementations(
		unitTest,
		"Late game",
		playerNames)

	for stateIndex := 0; stateIndex < len(lateGameStates); stateIndex++ {
		gameList := game.ByCreationTime([]game.State{
			lateGameStates[stateIndex],
			earlyGameStates[stateIndex],
		})

		if !gameList[1].CreationTime().Before(gameList[0].CreationTime()) {
			unitTest.Fatalf(
				"Game states for state index %v were not differentiable by creation time: early at %v; late at %v",
				stateIndex,
				gameList[1].CreationTime(),
				gameList[0].CreationTime())
		}

		sort.Sort(gameList)

		if (gameList[0].Name() != earlyGameStates[stateIndex].Name()) ||
			(gameList[1].Name() != lateGameStates[stateIndex].Name()) {
			unitTest.Fatalf(
				"Game states were not sorted: expected names [%v, %v], instead had [%v, %v]",
				earlyGameStates[stateIndex].Name(),
				lateGameStates[stateIndex].Name(),
				gameList[0].Name(),
				gameList[1].Name())
		}
	}
}

func TestInitialState(unitTest *testing.T) {
	standardRuleset := &game.StandardWithoutRainbowRuleset{}

	separateRainbowRuleset := &game.RainbowAsSeparateSuitRuleset{
		BasisRules: standardRuleset,
	}

	compoundRainbowRuleset := &game.RainbowAsCompoundSuitRuleset{
		BasisRainbow: separateRainbowRuleset,
	}

	type testArguments struct {
		initialPlayerNames []string
		gameRuleset        game.Ruleset
	}

	testCases := []struct {
		name      string
		arguments testArguments
	}{
		{
			name: "Two players, no rainbow",
			arguments: testArguments{
				initialPlayerNames: []string{"Player One", "Player Two"},
				gameRuleset:        standardRuleset,
			},
		},
		{
			name: "Three players, no rainbow",
			arguments: testArguments{
				initialPlayerNames: []string{"Player One", "Player Two", "Player Three"},
				gameRuleset:        standardRuleset,
			},
		},
		{
			name: "Four players, no rainbow",
			arguments: testArguments{
				initialPlayerNames: []string{"Player One", "Player Two", "Player Three", "Player Four"},
				gameRuleset:        standardRuleset,
			},
		},
		{
			name: "Five players, no rainbow",
			arguments: testArguments{
				initialPlayerNames: []string{"Player One", "Player Two", "Player Three", "Player Four", "Player Five"},
				gameRuleset:        standardRuleset,
			},
		},
		{
			name: "Two players, with rainbow (as separate, but doesn't matter for initial state)",
			arguments: testArguments{
				initialPlayerNames: []string{"Player One", "Player Two"},
				gameRuleset:        separateRainbowRuleset,
			},
		},
		{
			name: "Five players, with rainbow (as compound, but doesn't matter for initial state)",
			arguments: testArguments{
				initialPlayerNames: []string{"Player One", "Player Two", "Player Three", "Player Four", "Player Five"},
				gameRuleset:        compoundRainbowRuleset,
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			gameStates := prepareImplementations(
				unitTest,
				"Test game",
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

				chat.AssertLogCorrect(
					unitTest,
					[]endpoint.ChatLogMessage{},
					gameView.ChatLog,
				)

				numberOfCardsPerHand := testCase.arguments.gameRuleset.NumberOfCardsInPlayerHand(numberOfPlayers)
				expectedNumberOfCardsInPlayerHands :=
					numberOfPlayers * numberOfCardsPerHand
				colorSuits := testCase.arguments.gameRuleset.ColorSuits()
				sequenceIndices := testCase.arguments.gameRuleset.SequenceIndices()
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
					testCase.arguments.gameRuleset,
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
	participatingPlayers []player.State) {
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
