package game

import (
	"fmt"
	"math/rand"
	"sort"

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

// ReadState returns the read-only game state corresponding to the given name if it exists
// in the given collection already (or else nil) along with whether the game exists,
// analogously to a standard Golang map.
func (gameCollection *StateCollection) ReadState(gameName string) (ReadonlyState, bool) {
	gameState, gameExists := gameCollection.statePersister.readAndWriteGame(gameName)

	if gameState == nil {
		return nil, false
	}

	return gameState.read(), gameExists
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

	if !gameState.read().HasPlayerAsParticipant(playerAction.PlayerIdentifier) {
		return fmt.Errorf(
			"Player %v is not a participant in game %v",
			playerAction.PlayerIdentifier,
			playerAction.GameIdentifier)
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

// TurnSummariesForFrontend writes the turn summary information for each game which has
// the given player into the relevant JSON object for the frontend.
func (gameCollection *StateCollection) TurnSummariesForFrontend(playerName string) endpoint.TurnSummaryList {
	gameList := gameCollection.statePersister.readAllWithPlayer(playerName)

	sort.Sort(ByCreationTime(gameList))

	numberOfGamesWithPlayer := len(gameList)

	turnSummaries := make([]endpoint.TurnSummary, numberOfGamesWithPlayer)
	for gameIndex := 0; gameIndex < numberOfGamesWithPlayer; gameIndex++ {
		nameOfGame := gameList[gameIndex].Name()
		gameTurn := gameList[gameIndex].Turn()

		gameParticipants := gameList[gameIndex].Players()
		numberOfParticipants := len(gameParticipants)

		playerNamesInTurnOrder := make([]string, numberOfParticipants)

		turnsUntilPlayer := 0
		for playerIndex := 0; playerIndex < numberOfParticipants; playerIndex++ {
			// Game turns begin with 1 rather than 0, so this sets the player names in order,
			// wrapping index back to 0 when at the end of the list.
			// E.g. turn 3, 5 players: playerNamesInTurnOrder will start with
			// gameParticipants[2], then [3], then [4], then [0], then [1].
			playerInTurnOrder :=
				gameParticipants[(playerIndex+gameTurn-1)%numberOfParticipants]
			playerNamesInTurnOrder[playerIndex] =
				playerInTurnOrder.Name()

			if playerName == playerInTurnOrder.Name() {
				turnsUntilPlayer = playerIndex
			}
		}

		turnSummaries[gameIndex] = endpoint.TurnSummary{
			GameIdentifier:             gameList[gameIndex].Identifier(),
			GameName:                   nameOfGame,
			RulesetDescription:         gameList[gameIndex].Ruleset().FrontendDescription(),
			CreationTimestampInSeconds: gameList[gameIndex].CreationTime().Unix(),
			TurnNumber:                 gameTurn,
			PlayerNamesInNextTurnOrder: playerNamesInTurnOrder,
			IsPlayerTurn:               turnsUntilPlayer == 0,
		}
	}

	return endpoint.TurnSummaryList{TurnSummaries: turnSummaries}
}
