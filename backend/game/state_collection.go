package game

import (
	"fmt"
	"math/rand"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
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
	gameState, gameExists := gameCollection.statePersister.readAndWriteGame(gameName)

	if !gameExists {
		gameDoesNotExistError :=
			fmt.Errorf(
				"Game %v does not exist, cannot be viewed by player %v",
				gameName,
				playerName)
		return nil, gameDoesNotExistError
	}

	return ViewForPlayer(gameState.read(), playerName)
}

// ViewAllWithPlayer wraps every read-only state given by the persister for the given player
// in a view. It returns an error if there is an error in creating any of the player views.
func (gameCollection *StateCollection) ViewAllWithPlayer(
	playerName string) ([]*PlayerView, error) {
	gameStates := gameCollection.statePersister.readAllWithPlayer(playerName)
	numberOfGames := len(gameStates)

	playerViews := make([]*PlayerView, numberOfGames)

	for gameIndex := 0; gameIndex < numberOfGames; gameIndex++ {
		playerView, participantError :=
			ViewForPlayer(gameStates[gameIndex], playerName)

		if participantError != nil {
			overallError :=
				fmt.Errorf(
					"When trying to wrap views around read-only game states, encountered errror %v",
					participantError)
			return nil, overallError
		}

		playerViews[gameIndex] = playerView
	}

	return playerViews, nil
}

// PerformAction finds the given game and performs the given action for its player,
// or returns an error.
func (gameCollection *StateCollection) PerformAction(
	playerAction endpoint.PlayerAction) error {
	actingPlayer, playeridentificationError :=
		gameCollection.playerProvider.Get(playerAction.PlayerIdentifier)

	if playeridentificationError != nil {
		return playeridentificationError
	}

	gameState, isFound :=
		gameCollection.statePersister.readAndWriteGame(playerAction.GameIdentifier)

	if !isFound {
		return fmt.Errorf(
			"Game %v does not exist, cannot perform action from player %v",
			playerAction.GameIdentifier,
			playerAction.PlayerIdentifier)
	}

	_, participantError :=
		ViewForPlayer(gameState.read(), playerAction.PlayerIdentifier)

	if participantError != nil {
		return participantError
	}

	return gameState.performAction(actingPlayer, playerAction)
}

// AddNew prepares a new shuffled deck using a random seed taken from the given
// collection, and uses it to create a new game in the given collection from the
// given definition. It returns an error if a game with the given name already
// exists, or if the definition includes invalid players.
func (gameCollection *StateCollection) AddNew(
	gameDefinition endpoint.GameDefinition) error {

	return gameCollection.AddNewWithGivenRandomSeed(
		gameDefinition,
		gameCollection.statePersister.randomSeed())
}

// AddNewWithGivenRandomSeed prepares a new shuffled deck using the given seed for
// a random number generator, and uses it to create a new game in the given collection
// from the given definition. It returns an error if a game with the given name already
// exists, or if the definition includes invalid players.
func (gameCollection *StateCollection) AddNewWithGivenRandomSeed(
	gameDefinition endpoint.GameDefinition,
	randomSeed int64) error {
	if gameDefinition.GameName == "" {
		return fmt.Errorf("Game must have a name")
	}

	gameRuleset, unknownRulesetError := RulesetFromIdentifier(gameDefinition.RulesetIdentifier)
	if unknownRulesetError != nil {
		return fmt.Errorf(
			"Problem identifying ruleset from identifier %v; error is: %v",
			gameDefinition.RulesetIdentifier,
			unknownRulesetError)
	}

	// A nil slice still has a length of 0, so this is OK.
	numberOfPlayers := len(gameDefinition.PlayerIdentifiers)

	if numberOfPlayers < gameRuleset.MinimumNumberOfPlayers() {
		return fmt.Errorf(
			"Game must have at least %v players",
			gameRuleset.MinimumNumberOfPlayers())
	}

	if numberOfPlayers > gameRuleset.MaximumNumberOfPlayers() {
		return fmt.Errorf(
			"Game must have no more than %v players",
			gameRuleset.MaximumNumberOfPlayers())
	}

	playerIdentifiers := make(map[string]bool, 0)

	playerStates := make([]player.ReadonlyState, numberOfPlayers)
	for playerIndex := 0; playerIndex < numberOfPlayers; playerIndex++ {
		playerIdentifier := gameDefinition.PlayerIdentifiers[playerIndex]
		playerState, identificationError := gameCollection.playerProvider.Get(playerIdentifier)

		if identificationError != nil {
			return identificationError
		}

		if playerIdentifiers[playerIdentifier] {
			return fmt.Errorf(
				"Player with identifier %v appears more than once in the list of players",
				playerIdentifier)
		}

		playerIdentifiers[playerIdentifier] = true

		playerStates[playerIndex] = playerState
	}

	randomNumberGenerator := rand.New(rand.NewSource(randomSeed))

	shuffledDeck := gameRuleset.FullCardset()

	numberOfCards := len(shuffledDeck)

	// This is probably excessive.
	numberOfShuffles := 8 * numberOfCards

	for shuffleCount := 0; shuffleCount < numberOfShuffles; shuffleCount++ {
		firstShuffleIndex := randomNumberGenerator.Intn(numberOfCards)
		secondShuffleIndex := randomNumberGenerator.Intn(numberOfCards)
		shuffledDeck[firstShuffleIndex], shuffledDeck[secondShuffleIndex] =
			shuffledDeck[secondShuffleIndex], shuffledDeck[firstShuffleIndex]
	}

	return gameCollection.statePersister.addGame(
		gameDefinition.GameName,
		gameRuleset,
		playerStates,
		shuffledDeck)
}
