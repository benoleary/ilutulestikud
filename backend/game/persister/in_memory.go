package persister

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// inMemoryPersister stores game states by creating inMemoryStates and
// saving them as game.ReadAndWriteStates, mapped to by their names.
// It also maintains a map of player names to slices of game states,
// where each game state in the slice mapped to by a player includes
// that player as a participant.
type inMemoryPersister struct {
	mutualExclusion       sync.Mutex
	randomNumberGenerator *rand.Rand
	gameStates            map[string]game.ReadAndWriteState
	gamesWithPlayers      map[string][]game.ReadonlyState
}

// NewInMemory creates a game state persister around a map of games.
func NewInMemory() game.StatePersister {
	return &inMemoryPersister{
		mutualExclusion:       sync.Mutex{},
		randomNumberGenerator: rand.New(rand.NewSource(time.Now().Unix())),
		gameStates:            make(map[string]game.ReadAndWriteState, 1),
		gamesWithPlayers:      make(map[string][]game.ReadonlyState, 0),
	}
}

// RandomSeed provides an int64 which can be used as a seed for the
// rand.NewSource(...) function.
func (gamePersister *inMemoryPersister) RandomSeed() int64 {
	return gamePersister.randomNumberGenerator.Int63()
}

// ReadAndWriteGame returns the game.ReadAndWriteState corresponding to the given
// game name, or nil with an error if it does not exist.
func (gamePersister *inMemoryPersister) ReadAndWriteGame(
	gameName string) (game.ReadAndWriteState, error) {
	gameState, gameExists := gamePersister.gameStates[gameName]

	if !gameExists {
		return nil, fmt.Errorf("Game %v does not exist", gameName)
	}

	return gameState, nil
}

// ReadAllWithPlayer returns a slice of all the game.ReadonlyState instances in the
// collection which have the given player as a participant.
func (gamePersister *inMemoryPersister) ReadAllWithPlayer(
	playerName string) ([]game.ReadonlyState, error) {
	// We do not care if there was no entry for the player, as the default in this
	// case is nil, and we are going to explicitly check for nil to ensure that we
	// return an empty list instead anyway (in case the player was mapped to nil
	// somehow).
	gameStates, _ := gamePersister.gamesWithPlayers[playerName]

	if gameStates == nil {
		return []game.ReadonlyState{}, nil
	}

	return gameStates, nil
}

// AddGame adds an element to the collection which is a new object implementing
// the ReadAndWriteState interface from the given arguments, and returns the
// identifier of the newly-created game, along with an error which of course is
// nil if there was no problem. It returns an error if a game with the given name
// already exists.
func (gamePersister *inMemoryPersister) AddGame(
	gameName string,
	chatLogLength int,
	initialActionLog []message.FromPlayer,
	gameRuleset game.Ruleset,
	playersInTurnOrderWithInitialHands []game.PlayerNameWithHand,
	initialDeck []card.Defined) error {
	if gameName == "" {
		return fmt.Errorf("Game must have a name")
	}

	_, gameExists := gamePersister.gameStates[gameName]

	if gameExists {
		return fmt.Errorf("Game %v already exists", gameName)
	}

	serializableState :=
		NewSerializableState(gameName,
			chatLogLength,
			initialActionLog,
			gameRuleset,
			playersInTurnOrderWithInitialHands,
			initialDeck)

	newGame, errorFromRuleset :=
		newInMemoryState(serializableState)

	if errorFromRuleset != nil {
		return fmt.Errorf(
			"Error when trying to create game with name %v: %v",
			gameName,
			errorFromRuleset)
	}

	gamePersister.mutualExclusion.Lock()

	gamePersister.gameStates[gameName] = newGame

	for _, nameWithHand := range playersInTurnOrderWithInitialHands {
		playerName := nameWithHand.PlayerName
		existingGamesWithPlayer := gamePersister.gamesWithPlayers[playerName]
		gamePersister.gamesWithPlayers[playerName] =
			append(existingGamesWithPlayer, newGame.Read())
	}

	gamePersister.mutualExclusion.Unlock()
	return nil
}

// RemoveGameFromListForPlayer removes the given player from the given game
// in the sense that the game will no longer show up in the result of
// ReadAllWithPlayer(playerName). It returns an error if the player is not a
// participant.
func (gamePersister *inMemoryPersister) RemoveGameFromListForPlayer(
	gameName string,
	playerName string) error {
	// We only remove the player from the look-up map used for
	// ReadAllWithPlayer(...) rather than changing the internal state of
	// the game.
	gameStates, playerHasGames := gamePersister.gamesWithPlayers[playerName]

	if playerHasGames {
		for gameIndex, gameState := range gameStates {
			if gameName != gameState.Name() {
				continue
			}

			for _, participantName := range gameState.PlayerNames() {
				if participantName != playerName {
					continue
				}

				// We make a new array and copy in the elements of the original
				// list except for the given game, just to let the whole old array
				// qualify for garbage collection.
				originalListOfGames := gamePersister.gamesWithPlayers[playerName]
				reducedListOfGames := make([]game.ReadonlyState, gameIndex)
				copy(reducedListOfGames, originalListOfGames[:gameIndex])

				// We don't have to worry about gameIndex+1 being out of bounds as
				// a slice can start at index == length, and in this case just
				// produces an empty slice. (For gameIndex < length - 1, there is
				// obviously no problem.)
				gamePersister.gamesWithPlayers[playerName] =
					append(reducedListOfGames, originalListOfGames[gameIndex+1:]...)

				return nil
			}
		}
	}

	return fmt.Errorf(
		"Player %v is not a participant of game %v",
		playerName,
		gameName)
}

// Delete deletes the given game from the collection. It returns an error
// if the game does not exist before the deletion attempt, or if there is
// an error while trying to remove the game from the list for any player.
func (gamePersister *inMemoryPersister) Delete(gameName string) error {
	gameToDelete, gameExists := gamePersister.gameStates[gameName]

	if !gameExists {
		return fmt.Errorf("No game %v exists to delete", gameName)
	}

	errorsFromLeaving := []error{}

	for _, participantName := range gameToDelete.Read().PlayerNames() {
		errorFromRemovalFromListForPlayer :=
			gamePersister.RemoveGameFromListForPlayer(gameName, participantName)
		if errorFromRemovalFromListForPlayer != nil {
			errorsFromLeaving =
				append(errorsFromLeaving, errorFromRemovalFromListForPlayer)
		}
	}

	delete(gamePersister.gameStates, gameName)

	if len(errorsFromLeaving) > 0 {
		errorAroundRemovalErrors :=
			fmt.Errorf(
				"errors %v while removing game %v from player lists, game still deleted",
				errorsFromLeaving,
				gameName)

		return errorAroundRemovalErrors
	}

	return nil
}

// inMemoryState is a struct meant to encapsulate all the state required for a
// single game to function.
type inMemoryState struct {
	mutualExclusion sync.Mutex
	DeserializedState
}

// newInMemoryState creates a new game given the required information, using the
// given shuffled deck.
func newInMemoryState(
	serializablePart SerializableState) (game.ReadAndWriteState, error) {
	deserializedState, errorFromDeserialization :=
		NewDeserializedState(serializablePart)

	if errorFromDeserialization != nil {
		return nil, errorFromDeserialization
	}

	newState := &inMemoryState{
		mutualExclusion:   sync.Mutex{},
		DeserializedState: deserializedState,
	}

	return newState, nil
}

// Ruleset returns the ruleset for the game.
func (gameState *inMemoryState) Ruleset() game.Ruleset {
	return gameState.deserializedRuleset
}

// Read returns the gameState itself as a read-only object for the
// purposes of reading properties.
func (gameState *inMemoryState) Read() game.ReadonlyState {
	return gameState
}

// RecordChatMessage records a chat message from the given player.
func (gameState *inMemoryState) RecordChatMessage(
	actingPlayer player.ReadonlyState,
	chatMessage string) error {
	gameState.mutualExclusion.Lock()
	defer gameState.mutualExclusion.Unlock()

	return gameState.SerializableState.RecordChatMessage(
		actingPlayer,
		chatMessage)
}

// EnactTurnByDiscardingAndReplacing increments the turn number and moves the
// card in the acting player's hand at the given index into the discard pile,
// and replaces it in the player's hand with the next card from the deck,
// bundled with the given knowledge about the new card from the deck which the
// player should have (which should always be that any color suit is possible
// and any sequence index is possible). If there is no card to draw from the
// deck, it increments the number of turns taken with an empty deck of
// replacing the card in the hand. It also adds the given numbers to the
// counts of available hints and mistakes made respectively.
func (gameState *inMemoryState) EnactTurnByDiscardingAndReplacing(
	actionMessage string,
	actingPlayer player.ReadonlyState,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	numberOfReadyHintsToAdd int,
	numberOfMistakesMadeToAdd int) error {
	gameState.mutualExclusion.Lock()
	defer gameState.mutualExclusion.Unlock()

	return gameState.SerializableState.EnactTurnByDiscardingAndReplacing(
		actionMessage,
		actingPlayer,
		indexInHand,
		knowledgeOfDrawnCard,
		numberOfReadyHintsToAdd,
		numberOfMistakesMadeToAdd)
}

// EnactTurnByPlayingAndReplacing increments the turn number and moves the card
// in the acting player's hand at the given index into the appropriate color
// sequence, and replaces it in the player's hand with the next card from the
// deck, bundled with the given knowledge about the new card from the deck which
// the player should have (which should always be that any color suit is possible
// and any sequence index is possible). If there is no card to draw from the deck,
// it increments the number of turns taken with an empty deck of replacing the
// card in the hand. It also adds the given number of hints to the count of ready
// hints available (such as when playing the end of sequence gives a bonus hint).
func (gameState *inMemoryState) EnactTurnByPlayingAndReplacing(
	actionMessage string,
	actingPlayer player.ReadonlyState,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	numberOfReadyHintsToAdd int) error {
	gameState.mutualExclusion.Lock()
	defer gameState.mutualExclusion.Unlock()

	return gameState.SerializableState.EnactTurnByPlayingAndReplacing(
		actionMessage,
		actingPlayer,
		indexInHand,
		knowledgeOfDrawnCard,
		numberOfReadyHintsToAdd)
}

// EnactTurnByUpdatingHandWithHint increments the turn number and replaces the
// given player's inferred hand with the given inferred hand, while also
// decrementing the number of available hints appropriately. If the deck is
// empty, this function also increments the number of turns taken with an empty
// deck.
func (gameState *inMemoryState) EnactTurnByUpdatingHandWithHint(
	actionMessage string,
	actingPlayer player.ReadonlyState,
	receivingPlayerName string,
	updatedReceiverKnowledgeOfOwnHand []card.Inferred,
	numberOfReadyHintsToSubtract int) error {
	gameState.mutualExclusion.Lock()
	defer gameState.mutualExclusion.Unlock()

	return gameState.SerializableState.EnactTurnByUpdatingHandWithHint(
		actionMessage,
		actingPlayer,
		receivingPlayerName,
		updatedReceiverKnowledgeOfOwnHand,
		numberOfReadyHintsToSubtract)
}
