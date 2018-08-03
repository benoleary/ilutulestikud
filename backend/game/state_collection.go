package game

import (
	"fmt"
	"sort"

	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
)

// StateCollection wraps around a game.StatePersister to encapsulate logic acting on
// the functions of the interface.
type StateCollection struct {
	statePersister StatePersister
	chatLogLength  int
	playerProvider ReadonlyPlayerProvider
}

// NewCollection creates a new StateCollection around the given StatePersister and list
// of rulesets.
func NewCollection(
	statePersister StatePersister,
	chatLogLength int,
	playerProvider ReadonlyPlayerProvider) *StateCollection {
	return &StateCollection{
		statePersister: statePersister,
		chatLogLength:  chatLogLength,
		playerProvider: playerProvider,
	}
}

// ViewState returns a view around the read-only game state corresponding to the
// given name as seen by the given player. If the game does not exist or the
// player is not a participant, it returns an error.
func (gameCollection *StateCollection) ViewState(
	gameName string,
	playerName string) (ViewForPlayer, error) {
	gameState, errorFromGet :=
		gameCollection.statePersister.ReadAndWriteGame(gameName)

	if errorFromGet != nil {
		gameDoesNotExistError :=
			fmt.Errorf(
				"Could not find game %v (%v), cannot be viewed by player %v",
				gameName,
				errorFromGet,
				playerName)
		return nil, gameDoesNotExistError
	}

	return ViewOnStateForPlayer(
		gameState.Read(),
		gameCollection.playerProvider,
		playerName)
}

// ViewAllWithPlayer wraps every read-only state given by the persister for the given player
// in a view. It returns an error if there is an error in creating any of the player views.
// The views are ordered by creation timestamp, oldest first.
func (gameCollection *StateCollection) ViewAllWithPlayer(
	playerName string) ([]ViewForPlayer, error) {
	gameStates := gameCollection.statePersister.ReadAllWithPlayer(playerName)
	numberOfGames := len(gameStates)

	sort.Sort(ByCreationTime(gameStates))

	playerViews := make([]ViewForPlayer, numberOfGames)

	for gameIndex := 0; gameIndex < numberOfGames; gameIndex++ {
		playerView, participantError :=
			ViewOnStateForPlayer(
				gameStates[gameIndex],
				gameCollection.playerProvider,
				playerName)

		if participantError != nil {
			overallError :=
				fmt.Errorf(
					"When trying to wrap views around read-only game states, encountered error %v",
					participantError)
			return nil, overallError
		}

		playerViews[gameIndex] = playerView
	}

	return playerViews, nil
}

// AddNew prepares a new shuffled deck using a random seed taken from the given
// collection, and uses it to create a new game in the given collection from the
// given definition. It returns an error if a game with the given name already
// exists, or if the definition includes invalid players.
func (gameCollection *StateCollection) AddNew(
	gameName string,
	gameRuleset Ruleset,
	playerNames []string) error {
	initialDeck := gameRuleset.CopyOfFullCardset()

	card.ShuffleInPlace(initialDeck, gameCollection.statePersister.RandomSeed())

	return gameCollection.AddNewWithGivenDeck(
		gameName,
		gameRuleset,
		playerNames,
		initialDeck)
}

// AddNewWithGivenDeck creates a new game in the given collection from the given
// definition and the given deck. It returns an error if a game with the given name
// already exists, or if the definition includes invalid players.
func (gameCollection *StateCollection) AddNewWithGivenDeck(
	gameName string,
	gameRuleset Ruleset,
	playerNames []string,
	initialDeck []card.Defined) error {
	if gameName == "" {
		return fmt.Errorf("Game must have a name")
	}

	namesWithHands, initialDeck, initialActionLog, errorFromHands :=
		gameCollection.createPlayerHands(
			playerNames,
			gameRuleset,
			initialDeck)

	if errorFromHands != nil {
		return errorFromHands
	}

	return gameCollection.statePersister.AddGame(
		gameName,
		gameCollection.chatLogLength,
		initialActionLog,
		gameRuleset,
		namesWithHands,
		initialDeck)
}

// ExecuteAction finds the given game and wraps it in an executor for the given
// player, or returns an error.
func (gameCollection *StateCollection) ExecuteAction(
	gameName string,
	playerName string) (ExecutorForPlayer, error) {
	actingPlayer, playerIdentificationError :=
		gameCollection.playerProvider.Get(playerName)

	if playerIdentificationError != nil {
		return nil, playerIdentificationError
	}

	gameState, errorFromGet :=
		gameCollection.statePersister.ReadAndWriteGame(gameName)

	if errorFromGet != nil {
		errorWrappingErrorFromGet :=
			fmt.Errorf(
				"Could not find game %v (%v), cannot execute action for player %v",
				gameName,
				errorFromGet,
				playerName)

		return nil, errorWrappingErrorFromGet
	}

	return ExecutorOfActionsForPlayer(gameState, actingPlayer)
}

// RemoveGameFromListForPlayer calls the RemoveGameFromListForPlayer of the
// internal persistence store.
func (gameCollection *StateCollection) RemoveGameFromListForPlayer(
	gameName string,
	playerName string) error {
	return gameCollection.statePersister.RemoveGameFromListForPlayer(
		gameName,
		playerName)
}

// Delete calls the Delete of the internal persistence store.
func (gameCollection *StateCollection) Delete(gameName string) error {
	return gameCollection.statePersister.Delete(gameName)
}

// createPlayerHands deals out each player's hand (a full hand per player rather
// than one card each time to each player) and then returns a list of player names
// paired with their initial hands, the remaining deck, the initial action log, and
// a possible error.
func (gameCollection *StateCollection) createPlayerHands(
	playerNames []string,
	gameRuleset Ruleset,
	initialDeck []card.Defined) (
	[]PlayerNameWithHand,
	[]card.Defined,
	[]message.Readonly,
	error) {
	// A nil slice still has a length of 0, so this is OK.
	numberOfPlayers := len(playerNames)

	if numberOfPlayers < gameRuleset.MinimumNumberOfPlayers() {
		tooFewPlayersError :=
			fmt.Errorf(
				"Game must have at least %v players",
				gameRuleset.MinimumNumberOfPlayers())
		return nil, nil, nil, tooFewPlayersError
	}

	if numberOfPlayers > gameRuleset.MaximumNumberOfPlayers() {
		tooManyPlayersError :=
			fmt.Errorf(
				"Game must have no more than %v players",
				gameRuleset.MaximumNumberOfPlayers())
		return nil, nil, nil, tooManyPlayersError
	}

	handSize := gameRuleset.NumberOfCardsInPlayerHand(numberOfPlayers)
	minimumNumberOfCardsRequired := handSize * numberOfPlayers

	if len(initialDeck) < minimumNumberOfCardsRequired {
		tooFewCardsError :=
			fmt.Errorf(
				"Game must have at least %v cards",
				minimumNumberOfCardsRequired)
		return nil, nil, nil, tooFewCardsError
	}

	namesWithHands := make([]PlayerNameWithHand, numberOfPlayers)
	actionLog := make([]message.Readonly, numberOfPlayers)
	uniquePlayerNames := make(map[string]bool, numberOfPlayers)

	for playerIndex := 0; playerIndex < numberOfPlayers; playerIndex++ {
		playerName := playerNames[playerIndex]

		playerState, errorFromPlayerProvider := gameCollection.playerProvider.Get(playerName)

		if errorFromPlayerProvider != nil {
			return nil, nil, nil, errorFromPlayerProvider
		}

		if uniquePlayerNames[playerName] {
			degenerateNameError :=
				fmt.Errorf(
					"Player with name %v appears more than once in the list of players",
					playerName)
			return nil, nil, nil, degenerateNameError
		}

		uniquePlayerNames[playerName] = true

		playerHand := make([]card.InHand, handSize)

		for cardsInHand := 0; cardsInHand < handSize; cardsInHand++ {
			playerHand[cardsInHand] =
				card.InHand{
					Defined: initialDeck[cardsInHand],
					Inferred: card.Inferred{
						PossibleColors:  gameRuleset.ColorSuits(),
						PossibleIndices: gameRuleset.DistinctPossibleIndices(),
					},
				}

			// We should not ever re-visit these cards, but we set them to
			// represent an error just in case.
			initialDeck[cardsInHand] =
				card.Defined{
					ColorSuit:     "error: already removed from deck",
					SequenceIndex: -1,
				}
		}

		actionLog[playerIndex] =
			message.NewReadonly(playerName, playerState.Color(), "receieved initial hand")

		// Now we ensure that the cards just dealt out are no longer part of the deck.
		initialDeck = initialDeck[handSize:]

		namesWithHands[playerIndex] =
			PlayerNameWithHand{
				PlayerName:  playerName,
				InitialHand: playerHand,
			}
	}

	return namesWithHands, initialDeck, actionLog, nil
}

// ByCreationTime implements sort interface for []ReadonlyState based on the return
// from its CreationTime(). It is exported for ease of testing.
type ByCreationTime []ReadonlyState

// Len implements part of the sort interface for ByCreationTime.
func (byCreationTime ByCreationTime) Len() int {
	return len(byCreationTime)
}

// Swap implements part of the sort interface for ByCreationTime.
func (byCreationTime ByCreationTime) Swap(firstIndex int, secondIndex int) {
	byCreationTime[firstIndex], byCreationTime[secondIndex] =
		byCreationTime[secondIndex], byCreationTime[firstIndex]
}

// Less implements part of the sort interface for ByCreationTime.
func (byCreationTime ByCreationTime) Less(firstIndex int, secondIndex int) bool {
	return byCreationTime[firstIndex].CreationTime().Before(
		byCreationTime[secondIndex].CreationTime())
}
