package game

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/chat"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// InMemoryPersister stores inMemoryState objects as State objects. The games are
// mapped to by their name, as are the players.
type InMemoryPersister struct {
	mutualExclusion       sync.Mutex
	randomNumberGenerator *rand.Rand
	gameStates            map[string]readAndWriteState
	gamesWithPlayers      map[string][]ReadonlyState
}

// NewInMemoryPersister creates a game state persister around a map of games.
func NewInMemoryPersister() *InMemoryPersister {
	return &InMemoryPersister{
		mutualExclusion:       sync.Mutex{},
		randomNumberGenerator: rand.New(rand.NewSource(time.Now().Unix())),
		gameStates:            make(map[string]readAndWriteState, 1),
		gamesWithPlayers:      make(map[string][]ReadonlyState, 0),
	}
}

// randomSeed provides an int64 which can be used as a seed for the
// rand.NewSource(...) function.
func (inMemoryPersister *InMemoryPersister) randomSeed() int64 {
	return inMemoryPersister.randomNumberGenerator.Int63()
}

// addGame adds an element to the collection which is a new object implementing
// the readAndWriteState interface from the given arguments, and returns the
// identifier of the newly-created game, along with an error which of course is
// nil if there was no problem. It returns an error if a game with the given name
// already exists.
func (inMemoryPersister *InMemoryPersister) addGame(
	gameName string,
	gameRuleset Ruleset,
	playerStates []player.ReadonlyState,
	initialShuffle []card.Readonly) error {
	if gameName == "" {
		return fmt.Errorf("Game must have a name")
	}

	_, gameExists := inMemoryPersister.gameStates[gameName]

	if gameExists {
		return fmt.Errorf("Game %v already exists", gameName)
	}

	newGame :=
		newInMemoryState(
			gameName,
			gameRuleset,
			playerStates,
			initialShuffle)

	inMemoryPersister.mutualExclusion.Lock()

	inMemoryPersister.gameStates[gameName] = newGame

	for _, playerState := range playerStates {
		playerName := playerState.Name()
		existingGamesWithPlayer := inMemoryPersister.gamesWithPlayers[playerName]
		inMemoryPersister.gamesWithPlayers[playerName] =
			append(existingGamesWithPlayer, newGame.read())
	}

	inMemoryPersister.mutualExclusion.Unlock()
	return nil
}

// readAllWithPlayer returns a slice of all the ReadonlyState instances in the collection
// which have the given player as a participant. The order is not consistent with repeated
// calls, as it is defined by the entry set of a standard Golang map.
func (inMemoryPersister *InMemoryPersister) readAllWithPlayer(
	playerIdentifier string) []ReadonlyState {
	return inMemoryPersister.gamesWithPlayers[playerIdentifier]
}

// ReadGame returns the ReadonlyState corresponding to the given game identifier if
// it exists already (or else nil) along with whether the game exists, analogously
// to a standard Golang map.
func (inMemoryPersister *InMemoryPersister) readAndWriteGame(
	gameIdentifier string) (readAndWriteState, bool) {
	gameState, gameExists := inMemoryPersister.gameStates[gameIdentifier]
	return gameState, gameExists
}

// inMemoryState is a struct meant to encapsulate all the state required for a single game to function.
type inMemoryState struct {
	mutualExclusion        sync.Mutex
	gameName               string
	gameRuleset            Ruleset
	creationTime           time.Time
	participatingPlayers   []player.ReadonlyState
	chatLog                *chat.Log
	turnNumber             int
	currentScore           int
	numberOfReadyHints     int
	numberOfMistakesMade   int
	undrawnDeck            []card.Readonly
	lastPlayedCardForColor map[string]card.Readonly
	discardedCards         map[card.Readonly]int
	playerHands            map[string][]card.Readonly
}

// newInMemoryState creates a new game given the required information,
// using the given seed for the random number generator used to shuffle the deck
// initially.
func newInMemoryState(
	gameName string,
	gameRuleset Ruleset,
	playerStates []player.ReadonlyState,
	shuffledDeck []card.Readonly) readAndWriteState {
	return &inMemoryState{
		mutualExclusion:        sync.Mutex{},
		gameName:               gameName,
		gameRuleset:            gameRuleset,
		creationTime:           time.Now(),
		participatingPlayers:   playerStates,
		chatLog:                chat.NewLog(),
		turnNumber:             1,
		numberOfReadyHints:     gameRuleset.MaximumNumberOfHints(),
		numberOfMistakesMade:   0,
		undrawnDeck:            shuffledDeck,
		lastPlayedCardForColor: make(map[string]card.Readonly, 0),
		discardedCards:         make(map[card.Readonly]int, 0),
		playerHands:            make(map[string][]card.Readonly, 0),
	}
}

// Name returns the value of the private gameName string.
func (gameState *inMemoryState) Name() string {
	return gameState.gameName
}

// Ruleset returns the ruleset for the game.
func (gameState *inMemoryState) Ruleset() Ruleset {
	return gameState.gameRuleset
}

// Players returns a slice of the private participatingPlayers array.
func (gameState *inMemoryState) Players() []player.ReadonlyState {
	return gameState.participatingPlayers
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
func (gameState *inMemoryState) LastPlayedForColor(colorSuit string) (card.Readonly, bool) {
	lastPlayedCard, hasCardBeenPlayedForColor := gameState.lastPlayedCardForColor[colorSuit]

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

	return playerHand[indexInHand], nil
}

// InferredCardInHand returns the inferred information about the card held by the given
// player in the given position.
func (gameState *inMemoryState) InferredCardInHand(
	holdingPlayerName string,
	indexInHand int) (card.Inferred, error) {
	return card.Inferred{}, fmt.Errorf("not implemented yet")
}

// read returns the gameState itself as a read-only object for the purposes of reading
// properties.
func (gameState *inMemoryState) read() ReadonlyState {
	return gameState
}

// recordChatMessage records a chat message from the given player.
func (gameState *inMemoryState) recordChatMessage(
	actingPlayer player.ReadonlyState,
	chatMessage string) error {
	gameState.chatLog.AppendNewMessage(
		actingPlayer.Name(),
		actingPlayer.Color(),
		chatMessage)
	return nil
}

// drawCard returns the top-most card of the deck, or a card representing an error
// along with an actual Go error if there are no cards left.
func (gameState *inMemoryState) drawCard() (card.Readonly, error) {
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

// replaceCardInHand replaces the card at the given index in the hand of the given
// player with the given replacement card, and returns the card which has just been
// replaced.
func (gameState *inMemoryState) replaceCardInHand(
	holdingPlayerName string,
	indexInHand int,
	replacementCard card.Readonly) (card.Readonly, error) {
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

	return cardBeingReplaced, nil
}

// addCardToPlayedSequence adds the given card to the appropriate sequence of played
// cards (by just over-writing what was the top-most card of the sequence).
func (gameState *inMemoryState) addCardToPlayedSequence(playedCard card.Readonly) error {
	gameState.lastPlayedCardForColor[playedCard.ColorSuit()] = playedCard
	return nil
}

// addCardToDiscardPile adds the given card to the pile of discarded cards (by just
// incrementing the number of copies of that card marked as discarded). This assumes
// that the given card was emitted by an instance of inMemoryState and so is a
// simpleCard instance, as each key of the map so far has been. If it is not, no error
// is returned, and the number of copies of that implementation gets incremented.
func (gameState *inMemoryState) addCardToDiscardPile(discardedCard card.Readonly) error {
	discardedCopiesUntilNow, _ := gameState.discardedCards[discardedCard]
	gameState.discardedCards[discardedCard] = discardedCopiesUntilNow + 1
	return nil
}
