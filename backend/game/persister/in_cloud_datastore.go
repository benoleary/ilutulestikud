package persister

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/benoleary/ilutulestikud/backend/cloud"
	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
	"github.com/benoleary/ilutulestikud/backend/player"
	"google.golang.org/api/iterator"
)

// CloudDatastoreKeyKind denotes the kind for the entities which will store
// games in the Google Cloud Datastore.
const CloudDatastoreKeyKind = "Game"

// inCloudDatastorePersister stores game states by creating
// inCloudDatastoreStates and saving them as game.ReadAndWriteStates
// in Google Cloud Datastore.
type inCloudDatastorePersister struct {
	randomNumberGenerator *rand.Rand
	clientProvider        cloud.DatastoreClientProvider
	datastoreClient       cloud.LimitedClient
}

// NewInCloudDatastore creates a game state persister.
func NewInCloudDatastore(
	clientProvider cloud.DatastoreClientProvider) game.StatePersister {
	return &inCloudDatastorePersister{
		randomNumberGenerator: rand.New(rand.NewSource(time.Now().Unix())),
		clientProvider:        clientProvider,
		datastoreClient:       nil,
	}
}

// RandomSeed provides an int64 which can be used as a seed for the
// rand.NewSource(...) function.
func (gamePersister *inCloudDatastorePersister) RandomSeed() int64 {
	return gamePersister.randomNumberGenerator.Int63()
}

// ReadAndWriteGame returns the game.ReadAndWriteState corresponding to the given
// game name, or nil with an error if it does not exist.
func (gamePersister *inCloudDatastorePersister) ReadAndWriteGame(
	executionContext context.Context,
	gameName string) (game.ReadAndWriteState, error) {
	return gamePersister.GetInCloudDatastoreState(executionContext, gameName)
}

// ReadAllWithPlayer returns a slice of all the game.ReadonlyState instances in the
// collection which have the given player as a participant.
func (gamePersister *inCloudDatastorePersister) ReadAllWithPlayer(
	executionContext context.Context,
	playerName string) ([]game.ReadonlyState, error) {
	initializedClient, errorFromAcquiral :=
		gamePersister.acquireClient(executionContext)
	if errorFromAcquiral != nil {
		return nil, errorFromAcquiral
	}

	// https://cloud.google.com/datastore/docs/concepts/queries
	// #properties_with_array_values_can_behave_in_surprising_ways
	// => we just search for the player's name by equality with
	// ParticipantNamesInTurnOrder, as the equality filter on arrays selects
	// the entity if any of the elements match the sought value.
	resultIterator :=
		initializedClient.AllMatching(
			executionContext,
			"ParticipantNamesInTurnOrder =",
			playerName)

	gameStates := []game.ReadonlyState{}

	for {
		var matchedGame SerializableState
		errorFromNext := resultIterator.DeserializeNext(&matchedGame)

		if errorFromNext == iterator.Done {
			break
		}

		if errorFromNext != nil {
			return nil, errorFromNext
		}

		hasLeftGame := false
		for _, playerWhoHasLeft := range matchedGame.ParticipantsWhoHaveLeft {
			if playerWhoHasLeft == playerName {
				hasLeftGame = true
				break
			}
		}

		if !hasLeftGame {
			deserializedState, errorFromDeserialization :=
				newInCloudDatastoreState(
					initializedClient,
					matchedGame.GameName,
					matchedGame)

			if errorFromDeserialization != nil {
				return nil, errorFromDeserialization
			}

			gameStates = append(gameStates, deserializedState.Read())
		}
	}

	return gameStates, nil
}

// AddGame adds an element to the collection which is a new object implementing
// the ReadAndWriteState interface from the given arguments, and returns the
// identifier of the newly-created game, along with an error which of course is
// nil if there was no problem. It returns an error if a game with the given name
// already exists.
func (gamePersister *inCloudDatastorePersister) AddGame(
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

	initializedClient, errorFromAcquiral :=
		gamePersister.acquireClient(executionContext)

	if errorFromAcquiral != nil {
		return errorFromAcquiral
	}

	isAlreadyInDatastore, errorFromCheck :=
		cloud.DoesNameExist(
			executionContext,
			initializedClient,
			gameName)

	if errorFromCheck != nil {
		return errorFromCheck
	}

	if isAlreadyInDatastore {
		return fmt.Errorf("Game with name %v already exists", gameName)
	}

	serializableState :=
		NewSerializableState(
			gameName,
			chatLogLength,
			initialActionLog,
			gameRuleset,
			playersInTurnOrderWithInitialHands,
			initialDeck)

	errorFromPut :=
		initializedClient.Put(
			executionContext,
			gameName,
			&serializableState)

	return errorFromPut
}

// RemoveGameFromListForPlayer removes the given player from the given game
// in the sense that the game will no longer show up in the result of
// ReadAllWithPlayer(playerName). It returns an error if the player is not a
// participant, or if the player has already left, or if there is an error
// reading the game state from the store.
func (gamePersister *inCloudDatastorePersister) RemoveGameFromListForPlayer(
	executionContext context.Context,
	gameName string,
	playerName string) error {
	gameToUpdate, errorFromGet :=
		gamePersister.GetInCloudDatastoreState(executionContext, gameName)

	if errorFromGet != nil {
		return errorFromGet
	}

	return gameToUpdate.RemovePlayerFromParticipantList(
		executionContext,
		playerName)
}

// Delete deletes the given game from the collection. It returns an error
// if the Cloud Datastore API returns an error.
func (gamePersister *inCloudDatastorePersister) Delete(
	executionContext context.Context,
	gameName string) error {
	initializedClient, errorFromAcquiral :=
		gamePersister.acquireClient(executionContext)

	if errorFromAcquiral != nil {
		return errorFromAcquiral
	}

	return initializedClient.Delete(
		executionContext,
		gameName)
}

// GetInCloudDatastoreState returns a pointer to an inCloudDatastoreState
// struct de-serialized from the Google Cloud Datastore with the given name.
func (gamePersister *inCloudDatastorePersister) GetInCloudDatastoreState(
	executionContext context.Context,
	gameName string) (*inCloudDatastoreState, error) {
	initializedClient, errorFromAcquiral :=
		gamePersister.acquireClient(executionContext)

	if errorFromAcquiral != nil {
		return nil, errorFromAcquiral
	}

	serializablePart := SerializableState{}

	errorFromGet :=
		initializedClient.Get(
			executionContext,
			gameName,
			&serializablePart)

	if errorFromGet != nil {
		return nil, errorFromGet
	}

	return newInCloudDatastoreState(
		initializedClient,
		gameName,
		serializablePart)
}

// acquireClient returns the connection to the Cloud Datastore,
// initializing it if it has not already been initialized.
func (gamePersister *inCloudDatastorePersister) acquireClient(
	executionContext context.Context) (cloud.LimitedClient, error) {
	if gamePersister.datastoreClient == nil {
		cloudDatastoreClient, errorFromCloudDatastore :=
			gamePersister.clientProvider.NewClient(executionContext)
		if errorFromCloudDatastore != nil {
			return nil, errorFromCloudDatastore
		}

		gamePersister.datastoreClient = cloudDatastoreClient
	}

	return gamePersister.datastoreClient, nil
}

// inCloudDatastoreState is a struct meant to encapsulate all the state
// required for a single game to function, and also to persist itself in
// Google Cloud Datastore.
type inCloudDatastoreState struct {
	mutualExclusion sync.Mutex
	datastoreClient cloud.LimitedClient
	keyName         string
	DeserializedState
}

// newInCloudDatastoreState creates a new game given the required information,
// using the given shuffled deck.
func newInCloudDatastoreState(
	datastoreClient cloud.LimitedClient,
	keyName string,
	serializablePart SerializableState) (*inCloudDatastoreState, error) {
	deserializedRuleset, errorFromRuleset :=
		game.RulesetFromIdentifier(serializablePart.RulesetIdentifier)
	if errorFromRuleset != nil {
		return nil, errorFromRuleset
	}

	newState := &inCloudDatastoreState{
		mutualExclusion:   sync.Mutex{},
		datastoreClient:   datastoreClient,
		keyName:           keyName,
		DeserializedState: CreateDeserializedState(serializablePart, deserializedRuleset),
	}

	return newState, nil
}

// Ruleset returns the ruleset for the game.
func (gameState *inCloudDatastoreState) Ruleset() game.Ruleset {
	return gameState.deserializedRuleset
}

// Read returns the gameState itself as a read-only object for the
// purposes of reading properties.
func (gameState *inCloudDatastoreState) Read() game.ReadonlyState {
	return gameState
}

// RecordChatMessage records a chat message from the given player.
func (gameState *inCloudDatastoreState) RecordChatMessage(
	executionContext context.Context,
	actingPlayer player.ReadonlyState,
	chatMessage string) error {
	gameState.mutualExclusion.Lock()
	defer gameState.mutualExclusion.Unlock()

	return gameState.uploadSerializablePartIfNoError(
		executionContext,
		gameState.SerializableState.RecordChatMessage(actingPlayer, chatMessage))
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
func (gameState *inCloudDatastoreState) EnactTurnByDiscardingAndReplacing(
	executionContext context.Context,
	actionMessage string,
	actingPlayer player.ReadonlyState,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	numberOfReadyHintsToAdd int,
	numberOfMistakesMadeToAdd int) error {
	gameState.mutualExclusion.Lock()
	defer gameState.mutualExclusion.Unlock()

	return gameState.uploadSerializablePartIfNoError(
		executionContext,
		gameState.DeserializedState.EnactTurnByDiscardingAndReplacing(
			actionMessage,
			actingPlayer,
			indexInHand,
			knowledgeOfDrawnCard,
			numberOfReadyHintsToAdd,
			numberOfMistakesMadeToAdd))
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
func (gameState *inCloudDatastoreState) EnactTurnByPlayingAndReplacing(
	executionContext context.Context,
	actionMessage string,
	actingPlayer player.ReadonlyState,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	numberOfReadyHintsToAdd int) error {
	gameState.mutualExclusion.Lock()
	defer gameState.mutualExclusion.Unlock()

	return gameState.uploadSerializablePartIfNoError(
		executionContext,
		gameState.DeserializedState.EnactTurnByPlayingAndReplacing(
			actionMessage,
			actingPlayer,
			indexInHand,
			knowledgeOfDrawnCard,
			numberOfReadyHintsToAdd))
}

// EnactTurnByUpdatingHandWithHint increments the turn number and replaces the
// given player's inferred hand with the given inferred hand, while also
// decrementing the number of available hints appropriately. If the deck is
// empty, this function also increments the number of turns taken with an empty
// deck.
func (gameState *inCloudDatastoreState) EnactTurnByUpdatingHandWithHint(
	executionContext context.Context,
	actionMessage string,
	actingPlayer player.ReadonlyState,
	receivingPlayerName string,
	updatedReceiverKnowledgeOfOwnHand []card.Inferred,
	numberOfReadyHintsToSubtract int) error {
	gameState.mutualExclusion.Lock()
	defer gameState.mutualExclusion.Unlock()

	return gameState.uploadSerializablePartIfNoError(
		executionContext,
		gameState.DeserializedState.EnactTurnByUpdatingHandWithHint(
			actionMessage,
			actingPlayer,
			receivingPlayerName,
			updatedReceiverKnowledgeOfOwnHand,
			numberOfReadyHintsToSubtract))
}

// RemovePlayerFromParticipantList marks the player as no longer being a
// participant of the given game.
func (gameState *inCloudDatastoreState) RemovePlayerFromParticipantList(
	executionContext context.Context,
	playerName string) error {
	errorUpdatingLocally :=
		gameState.SerializableState.RemovePlayerFromParticipantList(playerName)
	if errorUpdatingLocally != nil {
		return errorUpdatingLocally
	}

	return gameState.uploadSerializablePart(executionContext)
}

func (gameState *inCloudDatastoreState) uploadSerializablePartIfNoError(
	executionContext context.Context,
	errorFromUpdatingSerializablePart error) error {
	if errorFromUpdatingSerializablePart != nil {
		return errorFromUpdatingSerializablePart
	}

	return gameState.uploadSerializablePart(executionContext)
}

func (gameState *inCloudDatastoreState) uploadSerializablePart(
	executionContext context.Context) error {
	return gameState.datastoreClient.Put(
		executionContext,
		gameState.keyName,
		&gameState.SerializableState)
}
