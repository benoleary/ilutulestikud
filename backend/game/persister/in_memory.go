package persister

import (
	"context"
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
// It ignores all context structs passed to its functions.
type inMemoryPersister struct {
	mutualExclusion       sync.Mutex
	randomNumberGenerator *rand.Rand
	gameStates            map[string]*inMemoryState
}

// NewInMemory creates a game state persister around a map of games.
func NewInMemory() game.StatePersister {
	return &inMemoryPersister{
		mutualExclusion:       sync.Mutex{},
		randomNumberGenerator: rand.New(rand.NewSource(time.Now().Unix())),
		gameStates:            make(map[string]*inMemoryState, 1),
	}
}

// RandomSeed provides an int64 which can be used as a seed for the
// rand.NewSource(...) function.
func (gamePersister *inMemoryPersister) RandomSeed() int64 {
	gamePersister.mutualExclusion.Lock()
	defer gamePersister.mutualExclusion.Unlock()
	return gamePersister.randomNumberGenerator.Int63()
}

// ReadAndWriteGame returns the game.ReadAndWriteState corresponding to the given
// game name, or nil with an error if it does not exist. The context is ignored.
func (gamePersister *inMemoryPersister) ReadAndWriteGame(
	executionContext context.Context,
	gameName string) (game.ReadAndWriteState, error) {
	return gamePersister.GetInMemoryState(gameName)
}

// ReadAllWithPlayer returns a slice of all the game.ReadonlyState instances in the
// collection which have the given player as a participant. The context is ignored.
func (gamePersister *inMemoryPersister) ReadAllWithPlayer(
	executionContext context.Context,
	playerName string) ([]game.ReadonlyState, error) {
	gamesWithPlayer := make([]game.ReadonlyState, 0)

	for _, gameState := range gamePersister.gameStates {
		if gameState.HasCurrentParticipant(playerName) {
			gamesWithPlayer = append(gamesWithPlayer, gameState)
		}
	}

	return gamesWithPlayer, nil
}

// AddGame adds an element to the collection which is a new object implementing
// the ReadAndWriteState interface from the given arguments, and returns the
// identifier of the newly-created game, along with an error which of course is
// nil if there was no problem. It returns an error if a game with the given name
// already exists. The context is ignored.
func (gamePersister *inMemoryPersister) AddGame(
	executionContext context.Context,
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

	newGame := &inMemoryState{
		mutualExclusion:   sync.Mutex{},
		DeserializedState: CreateDeserializedState(serializableState, gameRuleset),
	}

	gamePersister.mutualExclusion.Lock()
	gamePersister.gameStates[gameName] = newGame
	gamePersister.mutualExclusion.Unlock()

	return nil
}

// RemoveGameFromListForPlayer removes the given player from the given game
// in the sense that the game will no longer show up in the result of
// ReadAllWithPlayer(playerName). It returns an error if the player is not a
// participant. The context is ignored.
func (gamePersister *inMemoryPersister) RemoveGameFromListForPlayer(
	executionContext context.Context,
	gameName string,
	playerName string) error {
	gameToUpdate, errorFromGet :=
		gamePersister.GetInMemoryState(gameName)

	if errorFromGet != nil {
		return errorFromGet
	}

	return gameToUpdate.RemovePlayerFromParticipantList(playerName)
}

// Delete deletes the given game from the collection. It returns no error.
// The context is ignored.
func (gamePersister *inMemoryPersister) Delete(
	executionContext context.Context,
	gameName string) error {
	gamePersister.mutualExclusion.Lock()
	delete(gamePersister.gameStates, gameName)
	gamePersister.mutualExclusion.Unlock()

	return nil
}

// GetInMemoryState returns a pointer to an inMemoryState struct with the given name.
func (gamePersister *inMemoryPersister) GetInMemoryState(
	gameName string) (*inMemoryState, error) {
	gameState, gameExists := gamePersister.gameStates[gameName]

	if !gameExists {
		return nil, fmt.Errorf("Game %v does not exist", gameName)
	}

	return gameState, nil
}

// inMemoryState is a struct meant to encapsulate all the state required for a
// single game to function. It ignores all context structs passed to its functions.
type inMemoryState struct {
	mutualExclusion sync.Mutex
	DeserializedState
}

// RecordChatMessage records a chat message from the given player. The context is ignored.
func (gameState *inMemoryState) RecordChatMessage(
	executionContext context.Context,
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
// counts of available hints and mistakes made respectively. The context is
// ignored.
func (gameState *inMemoryState) EnactTurnByDiscardingAndReplacing(
	executionContext context.Context,
	actionMessage string,
	actingPlayer player.ReadonlyState,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	numberOfReadyHintsToAdd int,
	numberOfMistakesMadeToAdd int) error {
	gameState.mutualExclusion.Lock()
	defer gameState.mutualExclusion.Unlock()

	return gameState.DeserializedState.EnactTurnByDiscardingAndReplacing(
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
// The context is ignored.
func (gameState *inMemoryState) EnactTurnByPlayingAndReplacing(
	executionContext context.Context,
	actionMessage string,
	actingPlayer player.ReadonlyState,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	numberOfReadyHintsToAdd int) error {
	gameState.mutualExclusion.Lock()
	defer gameState.mutualExclusion.Unlock()

	return gameState.DeserializedState.EnactTurnByPlayingAndReplacing(
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
// deck. The context is ignored.
func (gameState *inMemoryState) EnactTurnByUpdatingHandWithHint(
	executionContext context.Context,
	actionMessage string,
	actingPlayer player.ReadonlyState,
	receivingPlayerName string,
	updatedReceiverKnowledgeOfOwnHand []card.Inferred,
	numberOfReadyHintsToSubtract int) error {
	gameState.mutualExclusion.Lock()
	defer gameState.mutualExclusion.Unlock()

	return gameState.DeserializedState.EnactTurnByUpdatingHandWithHint(
		actionMessage,
		actingPlayer,
		receivingPlayerName,
		updatedReceiverKnowledgeOfOwnHand,
		numberOfReadyHintsToSubtract)
}
