package game

import (
	"fmt"
	"sort"

	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/player"
)

type readonlyPlayerProvider interface {
	Get(playerName string) (player.ReadonlyState, error)
}

// StateCollection wraps around a game.StatePersister to encapsulate logic acting on
// the functions of the interface.
type StateCollection struct {
	statePersister StatePersister
	playerProvider readonlyPlayerProvider
}

// NewCollection creates a new StateCollection around the given StatePersister and list
// of rulesets.
func NewCollection(
	statePersister StatePersister,
	playerProvider readonlyPlayerProvider) *StateCollection {
	return &StateCollection{
		statePersister: statePersister,
		playerProvider: playerProvider,
	}
}

// ViewState returns a view around the read-only game state corresponding to the
// given name as seen by the given player. If the game does not exist or the
// player is not a participant, it returns an error.
func (gameCollection *StateCollection) ViewState(
	gameName string,
	playerName string) (*PlayerView, error) {
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

	return ViewForPlayer(gameState.Read(), playerName)
}

// ViewAllWithPlayer wraps every read-only state given by the persister for the given player
// in a view. It returns an error if there is an error in creating any of the player views.
// The views are ordered by creation timestamp, oldest first.
func (gameCollection *StateCollection) ViewAllWithPlayer(
	playerName string) ([]*PlayerView, error) {
	gameStates := gameCollection.statePersister.ReadAllWithPlayer(playerName)
	numberOfGames := len(gameStates)

	sort.Sort(ByCreationTime(gameStates))

	playerViews := make([]*PlayerView, numberOfGames)

	for gameIndex := 0; gameIndex < numberOfGames; gameIndex++ {
		playerView, participantError :=
			ViewForPlayer(gameStates[gameIndex], playerName)

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
	initialDeck []card.Readonly) error {
	if gameName == "" {
		return fmt.Errorf("Game must have a name")
	}

	namesWithHands, initialDeck, errorFromHands :=
		gameCollection.createPlayerHands(
			playerNames,
			gameRuleset,
			initialDeck)

	if errorFromHands != nil {
		return errorFromHands
	}

	return gameCollection.statePersister.AddGame(
		gameName,
		gameRuleset,
		namesWithHands,
		initialDeck)
}

// RecordChatMessage finds the given game and records the given chat message from the
// given player, or returns an error.
func (gameCollection *StateCollection) RecordChatMessage(
	gameName string,
	playerName string,
	chatMessage string) error {
	chattingPlayer, playerIdentificationError :=
		gameCollection.playerProvider.Get(playerName)

	if playerIdentificationError != nil {
		return playerIdentificationError
	}

	gameState, errorFromGet :=
		gameCollection.statePersister.ReadAndWriteGame(gameName)

	if errorFromGet != nil {
		return fmt.Errorf(
			"Could not find game %v (%v), cannot record chat message from player %v",
			gameName,
			errorFromGet,
			playerName)
	}

	_, participantError := ViewForPlayer(gameState.Read(), playerName)

	if participantError != nil {
		return participantError
	}

	// No error is returned when recording a chat message.
	gameState.RecordChatMessage(chattingPlayer, chatMessage)
	return nil
}

// createPlayerHands deals out each player's hand (a full hand per player rather
// than one card each time to each player) and then returns a map of player names
// to hands along with the remaining deck.
func (gameCollection *StateCollection) createPlayerHands(
	playerNames []string,
	gameRuleset Ruleset,
	initialDeck []card.Readonly) ([]PlayerNameWithHand, []card.Readonly, error) {
	// A nil slice still has a length of 0, so this is OK.
	numberOfPlayers := len(playerNames)

	if numberOfPlayers < gameRuleset.MinimumNumberOfPlayers() {
		tooFewError :=
			fmt.Errorf(
				"Game must have at least %v players",
				gameRuleset.MinimumNumberOfPlayers())
		return nil, nil, tooFewError
	}

	if numberOfPlayers > gameRuleset.MaximumNumberOfPlayers() {
		tooManyError :=
			fmt.Errorf(
				"Game must have no more than %v players",
				gameRuleset.MaximumNumberOfPlayers())
		return nil, nil, tooManyError
	}

	handSize := gameRuleset.NumberOfCardsInPlayerHand(numberOfPlayers)

	namesWithHands := make([]PlayerNameWithHand, numberOfPlayers)
	uniquePlayerNames := make(map[string]bool, numberOfPlayers)

	for playerIndex := 0; playerIndex < numberOfPlayers; playerIndex++ {
		playerName := playerNames[playerIndex]

		if uniquePlayerNames[playerName] {
			degenerateNameError :=
				fmt.Errorf(
					"Player with name %v appears more than once in the list of players",
					playerName)
			return nil, nil, degenerateNameError
		}

		uniquePlayerNames[playerName] = true

		playerHand := make([]card.Inferred, handSize)

		for cardsInHand := 0; cardsInHand < handSize; cardsInHand++ {
			playerHand[cardsInHand] =
				card.NewInferred(
					initialDeck[cardsInHand],
					gameRuleset.ColorSuits(),
					gameRuleset.SequenceIndices())

			// We should not ever re-visit these cards, but we set them to
			// represent an error just in case.
			initialDeck[cardsInHand] = card.ErrorReadonly()
		}

		// Now we ensure that the cards just dealt out are no longer part of the deck.
		initialDeck = initialDeck[handSize:]

		namesWithHands[playerIndex] =
			PlayerNameWithHand{
				PlayerName:  playerName,
				InitialHand: playerHand,
			}
	}

	return namesWithHands, initialDeck, nil
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
