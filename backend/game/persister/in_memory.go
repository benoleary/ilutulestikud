package persister

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/chat"
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
	playerName string) []game.ReadonlyState {
	// We do not care if there was no entry for the player, as the default in this
	// case is nil, and we are going to explicitly check for nil to ensure that we
	// return an empty list instead anyway (in case the player was mapped to nil
	// somehow).
	gameStates, _ := gamePersister.gamesWithPlayers[playerName]

	if gameStates == nil {
		return []game.ReadonlyState{}
	}

	return gameStates
}

// AddGame adds an element to the collection which is a new object implementing
// the ReadAndWriteState interface from the given arguments, and returns the
// identifier of the newly-created game, along with an error which of course is
// nil if there was no problem. It returns an error if a game with the given name
// already exists.
func (gamePersister *inMemoryPersister) AddGame(
	gameName string,
	gameRuleset game.Ruleset,
	playersInTurnOrderWithInitialHands []game.PlayerNameWithHand,
	initialDeck []card.Readonly) error {
	if gameName == "" {
		return fmt.Errorf("Game must have a name")
	}

	_, gameExists := gamePersister.gameStates[gameName]

	if gameExists {
		return fmt.Errorf("Game %v already exists", gameName)
	}

	newGame :=
		newInMemoryState(
			gameName,
			gameRuleset,
			playersInTurnOrderWithInitialHands,
			initialDeck)

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

// inMemoryState is a struct meant to encapsulate all the state required for a
// single game to function.
type inMemoryState struct {
	mutualExclusion             sync.Mutex
	gameName                    string
	gameRuleset                 game.Ruleset
	creationTime                time.Time
	participantNamesInTurnOrder []string
	chatLog                     *chat.Log
	turnNumber                  int
	currentScore                int
	numberOfReadyHints          int
	numberOfMistakesMade        int
	undrawnDeck                 []card.Readonly
	lastPlayedCardForColor      map[string]card.Readonly
	discardedCards              map[card.Readonly]int
	playerHands                 map[string][]card.Inferred
}

// newInMemoryState creates a new game given the required information, using the
// given shuffled deck.
func newInMemoryState(
	gameName string,
	gameRuleset game.Ruleset,
	playersInTurnOrderWithInitialHands []game.PlayerNameWithHand,
	shuffledDeck []card.Readonly) game.ReadAndWriteState {
	numberOfParticipants := len(playersInTurnOrderWithInitialHands)
	participantNamesInTurnOrder := make([]string, numberOfParticipants)
	playerHands := make(map[string][]card.Inferred, numberOfParticipants)
	for playerIndex := 0; playerIndex < numberOfParticipants; playerIndex++ {
		playerName := playersInTurnOrderWithInitialHands[playerIndex].PlayerName
		participantNamesInTurnOrder[playerIndex] = playerName
		playerHands[playerName] =
			playersInTurnOrderWithInitialHands[playerIndex].InitialHand
	}

	// We could already set up the capacity for the maps by getting slices from
	// the ruleset and counting, but that is a lot of effort for very little gain.
	return &inMemoryState{
		mutualExclusion:             sync.Mutex{},
		gameName:                    gameName,
		gameRuleset:                 gameRuleset,
		creationTime:                time.Now(),
		participantNamesInTurnOrder: participantNamesInTurnOrder,
		chatLog:                     chat.NewLog(),
		turnNumber:                  1,
		numberOfReadyHints:          gameRuleset.MaximumNumberOfHints(),
		numberOfMistakesMade:        0,
		undrawnDeck:                 shuffledDeck,
		lastPlayedCardForColor:      make(map[string]card.Readonly, 0),
		discardedCards:              make(map[card.Readonly]int, 0),
		playerHands:                 playerHands,
	}
}

// Name returns the value of the private gameName string.
func (gameState *inMemoryState) Name() string {
	return gameState.gameName
}

// Ruleset returns the ruleset for the game.
func (gameState *inMemoryState) Ruleset() game.Ruleset {
	return gameState.gameRuleset
}

// Players returns a slice of the private participantNames array.
func (gameState *inMemoryState) PlayerNames() []string {
	return gameState.participantNamesInTurnOrder
}

// CreationTime returns the value of the private time object describing the time at
// which the state was created.
func (gameState *inMemoryState) CreationTime() time.Time {
	return gameState.creationTime
}

// ChatLog returns the chat log of the game at the current moment.
func (gameState *inMemoryState) ChatLog() *chat.Log {
	return gameState.chatLog
}

// Turn returns the value of the private turnNumber int.
func (gameState *inMemoryState) Turn() int {
	return gameState.turnNumber
}

// Score returns the total score of the cards which have been correctly played in the
// game so far.
func (gameState *inMemoryState) Score() int {
	return gameState.currentScore
}

// NumberOfReadyHints returns the total number of hints which are available to be
// played.
func (gameState *inMemoryState) NumberOfReadyHints() int {
	return gameState.numberOfReadyHints
}

// NumberOfMistakesMade returns the total number of cards which have been played
// incorrectly.
func (gameState *inMemoryState) NumberOfMistakesMade() int {
	return gameState.numberOfMistakesMade
}

// DeckSize returns the number of cards left to draw from the deck.
func (gameState *inMemoryState) DeckSize() int {
	return len(gameState.undrawnDeck)
}

// LastPlayedForColor returns the last card which has been played correctly for the
// given color suit along with whether any card has been played in that suit so far,
// analogously to how a Go map works.
func (gameState *inMemoryState) LastPlayedForColor(
	colorSuit string) (card.Readonly, bool) {
	lastPlayedCard, hasCardBeenPlayedForColor :=
		gameState.lastPlayedCardForColor[colorSuit]

	return lastPlayedCard, hasCardBeenPlayedForColor
}

// NumberOfDiscardedCards returns the number of cards with the given suit and index
// which were discarded or played incorrectly.
func (gameState *inMemoryState) NumberOfDiscardedCards(
	colorSuit string,
	sequenceIndex int) int {
	mapKey := card.NewReadonly(colorSuit, sequenceIndex)

	// We ignore the bool about whether it was found, as the default 0 for an int in
	// Go is the correct value to return.
	numberOfCopies, _ := gameState.discardedCards[mapKey]

	return numberOfCopies
}

// VisibleCardInHand returns the card held by the given player in the given position.
func (gameState *inMemoryState) VisibleCardInHand(
	holdingPlayerName string,
	indexInHand int) (card.Readonly, error) {
	playerHand, hasHand := gameState.playerHands[holdingPlayerName]

	if !hasHand {
		return card.ErrorReadonly(), fmt.Errorf("Player has no hand")
	}

	return playerHand[indexInHand].UnderlyingCard(), nil
}

// InferredCardInHand returns the inferred information about the card held by the given
// player in the given position.
func (gameState *inMemoryState) InferredCardInHand(
	holdingPlayerName string,
	indexInHand int) (card.Inferred, error) {
	playerHand, hasHand := gameState.playerHands[holdingPlayerName]

	if !hasHand {
		return card.ErrorInferred(), fmt.Errorf("Player has no hand")
	}

	return playerHand[indexInHand], nil
}

// Read returns the gameState itself as a read-only object for the purposes of reading
// properties.
func (gameState *inMemoryState) Read() game.ReadonlyState {
	return gameState
}

// RecordChatMessage records a chat message from the given player.
func (gameState *inMemoryState) RecordChatMessage(
	actingPlayer player.ReadonlyState,
	chatMessage string) error {
	gameState.chatLog.AppendNewMessage(
		actingPlayer.Name(),
		actingPlayer.Color(),
		chatMessage)
	return nil
}

// DrawCard returns the top-most card of the deck, or a card representing an error
// along with an actual Golang error if there are no cards left.
func (gameState *inMemoryState) DrawCard() (card.Readonly, error) {
	if len(gameState.undrawnDeck) <= 0 {
		return card.ErrorReadonly(), fmt.Errorf("No cards left to draw")
	}

	drawnCard := gameState.undrawnDeck[0]

	// We should not ever re-visit this card, but in case we do somehow, we ensure
	// that this element represents an error.
	gameState.undrawnDeck[0] = card.ErrorReadonly()

	gameState.undrawnDeck = gameState.undrawnDeck[1:]

	return drawnCard, nil
}

// ReplaceCardInHand replaces the card at the given index in the hand of the given
// player with the given replacement card, and returns the card which has just been
// replaced.
func (gameState *inMemoryState) ReplaceCardInHand(
	holdingPlayerName string,
	indexInHand int,
	replacementCard card.Inferred) (card.Readonly, error) {
	if indexInHand < 0 {
		return card.ErrorReadonly(), fmt.Errorf("Index %v is out of allowed range", indexInHand)
	}

	playerHand, hasHand := gameState.playerHands[holdingPlayerName]

	if !hasHand {
		return card.ErrorReadonly(), fmt.Errorf("Player has no hand")
	}

	if indexInHand >= len(playerHand) {
		return card.ErrorReadonly(), fmt.Errorf("Index %v is out of allowed range", indexInHand)
	}

	cardBeingReplaced := playerHand[indexInHand]
	playerHand[indexInHand] = replacementCard

	return cardBeingReplaced.UnderlyingCard(), nil
}

// AddCardToPlayedSequence adds the given card to the appropriate sequence of played
// cards (by just over-writing what was the top-most card of the sequence).
func (gameState *inMemoryState) AddCardToPlayedSequence(playedCard card.Readonly) error {
	gameState.lastPlayedCardForColor[playedCard.ColorSuit()] = playedCard
	return nil
}

// AddCardToDiscardPile adds the given card to the pile of discarded cards (by just
// incrementing the number of copies of that card marked as discarded). This assumes
// that the given card was emitted by an instance of inMemoryState and so is a
// simpleCard instance, as each key of the map so far has been. If it is not, no error
// is returned, and the number of copies of that implementation gets incremented.
func (gameState *inMemoryState) AddCardToDiscardPile(discardedCard card.Readonly) error {
	discardedCopiesUntilNow, _ := gameState.discardedCards[discardedCard]
	gameState.discardedCards[discardedCard] = discardedCopiesUntilNow + 1
	return nil
}
