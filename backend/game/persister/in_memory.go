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
	chatLogLength int,
	initialActionLog []message.Readonly,
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
			chatLogLength,
			initialActionLog,
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
				gamePersister.gamesWithPlayers[playerName] =
					append(reducedListOfGames, gameStates[gameIndex+1:]...)

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

	for _, participantName := range gameToDelete.Read().PlayerNames() {
		errorFromRemovalFromListForPlayer :=
			gamePersister.RemoveGameFromListForPlayer(gameName, participantName)
		if errorFromRemovalFromListForPlayer != nil {
			errorAroundRemovalError :=
				fmt.Errorf(
					"error %v while removing game %v from player lists, game not deleted",
					errorFromRemovalFromListForPlayer,
					gameName)

			return errorAroundRemovalError
		}
	}

	delete(gamePersister.gameStates, gameName)

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
	chatLog                     *rollingMessageAppender
	actionLog                   *rollingMessageAppender
	turnNumber                  int
	numberOfReadyHints          int
	numberOfMistakesMade        int
	undrawnDeck                 []card.Readonly
	playedCardsForColor         map[string][]card.Readonly
	discardedCards              map[card.Readonly]int
	playerHands                 map[string][]card.InHand
}

// newInMemoryState creates a new game given the required information, using the
// given shuffled deck.
func newInMemoryState(
	gameName string,
	chatLogLength int,
	initialActionLog []message.Readonly,
	gameRuleset game.Ruleset,
	playersInTurnOrderWithInitialHands []game.PlayerNameWithHand,
	shuffledDeck []card.Readonly) game.ReadAndWriteState {
	numberOfParticipants := len(playersInTurnOrderWithInitialHands)
	participantNamesInTurnOrder := make([]string, numberOfParticipants)
	playerHands := make(map[string][]card.InHand, numberOfParticipants)
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
		chatLog:                     newEmptyRollingMessageAppender(chatLogLength),
		actionLog:                   newRollingMessageAppender(initialActionLog),
		turnNumber:                  1,
		numberOfReadyHints:          gameRuleset.MaximumNumberOfHints(),
		numberOfMistakesMade:        0,
		undrawnDeck:                 shuffledDeck,
		playedCardsForColor:         make(map[string][]card.Readonly, 0),
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
func (gameState *inMemoryState) ChatLog() []message.Readonly {
	return gameState.chatLog.sortedCopyOfMessages()
}

// ActionLog returns the action log of the game at the current moment.
func (gameState *inMemoryState) ActionLog() []message.Readonly {
	return gameState.actionLog.sortedCopyOfMessages()
}

// Turn returns the value of the private turnNumber int.
func (gameState *inMemoryState) Turn() int {
	return gameState.turnNumber
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

// PlayedForColor returns the cards, in order, which have been played
// correctly for the given color suit.
func (gameState *inMemoryState) PlayedForColor(
	colorSuit string) []card.Readonly {
	playedCards, _ :=
		gameState.playedCardsForColor[colorSuit]

	if playedCards == nil {
		return []card.Readonly{}
	}

	return playedCards
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
func (gameState *inMemoryState) VisibleHand(holdingPlayerName string) ([]card.Readonly, error) {
	playerHand, hasHand := gameState.playerHands[holdingPlayerName]

	if !hasHand {
		return nil, fmt.Errorf("Player has no hand")
	}

	handSize := len(playerHand)

	visibleHand := make([]card.Readonly, handSize)

	for indexInHand := 0; indexInHand < handSize; indexInHand++ {
		visibleHand[indexInHand] = playerHand[indexInHand].Readonly
	}

	return visibleHand, nil
}

// InferredCardInHand returns the inferred information about the card held by the given
// player in the given position.
func (gameState *inMemoryState) InferredHand(holdingPlayerName string) ([]card.Inferred, error) {
	playerHand, hasHand := gameState.playerHands[holdingPlayerName]

	if !hasHand {
		return nil, fmt.Errorf("Player has no hand")
	}

	handSize := len(playerHand)

	inferredHand := make([]card.Inferred, handSize)

	for indexInHand := 0; indexInHand < handSize; indexInHand++ {
		inferredHand[indexInHand] = playerHand[indexInHand].Inferred
	}

	return inferredHand, nil
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
	gameState.chatLog.appendNewMessage(
		actingPlayer.Name(),
		actingPlayer.Color(),
		chatMessage)
	return nil
}

// EnactTurnByDiscardingAndReplacing increments the turn number and moves the
// card in the acting player's hand at the given index into the discard pile,
// and replaces it in the player's hand with the next card from the deck,
// bundled with the given knowledge about the new card from the deck which the
// player should have (which should always be that any color suit is possible
// and any sequence index is possible). It also adds the given numbers to the
// counts of available hints and mistakes made respectively.
func (gameState *inMemoryState) EnactTurnByDiscardingAndReplacing(
	actionMessage string,
	actingPlayer player.ReadonlyState,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	numberOfReadyHintsToAdd int,
	numberOfMistakesMadeToAdd int) error {
	discardedCard, errorFromTakingCard :=
		gameState.takeCardFromHandReplacingIfPossible(
			actingPlayer.Name(),
			indexInHand,
			knowledgeOfDrawnCard)

	if errorFromTakingCard != nil {
		gameState.recordActionMessage(
			actingPlayer,
			errorFromTakingCard.Error())

		return errorFromTakingCard
	}

	discardedCopiesUntilNow, _ := gameState.discardedCards[discardedCard]
	gameState.discardedCards[discardedCard] = discardedCopiesUntilNow + 1

	gameState.numberOfReadyHints += numberOfReadyHintsToAdd
	gameState.numberOfMistakesMade += numberOfMistakesMadeToAdd
	gameState.turnNumber++

	gameState.recordActionMessage(
		actingPlayer,
		actionMessage)

	return nil
}

// EnactTurnByPlayingAndReplacing increments the turn number and moves the card
// in the acting player's hand at the given index into the appropriate color
// sequence, and replaces it in the player's hand with the next card from the
// deck, bundled with the given knowledge about the new card from the deck which
// the player should have (which should always be that any color suit is possible
// and any sequence index is possible). It also adds the given number of hints to
// the count of ready hints available (such as when playing the end of sequence
// gives a bonus hint).
func (gameState *inMemoryState) EnactTurnByPlayingAndReplacing(
	actionMessage string,
	actingPlayer player.ReadonlyState,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	numberOfReadyHintsToAdd int) error {
	playedCard, errorFromTakingCard :=
		gameState.takeCardFromHandReplacingIfPossible(
			actingPlayer.Name(),
			indexInHand,
			knowledgeOfDrawnCard)

	if errorFromTakingCard != nil {
		gameState.recordActionMessage(
			actingPlayer,
			errorFromTakingCard.Error())

		return errorFromTakingCard
	}

	playedSuit := playedCard.ColorSuit()
	sequenceBeforeNow := gameState.playedCardsForColor[playedSuit]
	gameState.playedCardsForColor[playedSuit] = append(sequenceBeforeNow, playedCard)

	gameState.numberOfReadyHints += numberOfReadyHintsToAdd
	gameState.turnNumber++

	gameState.recordActionMessage(
		actingPlayer,
		actionMessage)

	return nil
}

func (gameState *inMemoryState) recordActionMessage(
	actingPlayer player.ReadonlyState,
	actionMessage string) {
	gameState.actionLog.appendNewMessage(
		actingPlayer.Name(),
		actingPlayer.Color(),
		actionMessage)
}

// ReplaceCardInHand replaces the card at the given index in the hand of the given
// player with the given replacement card, and returns the card which has just been
// replaced.
func (gameState *inMemoryState) takeCardFromHandReplacingIfPossible(
	holdingPlayerName string,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred) (card.Readonly, error) {
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

	gameState.updatePlayerHand(
		holdingPlayerName,
		indexInHand,
		knowledgeOfDrawnCard,
		playerHand)

	return cardBeingReplaced.Readonly, nil
}

func (gameState *inMemoryState) updatePlayerHand(
	holdingPlayerName string,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	playerHand []card.InHand) {
	if len(gameState.undrawnDeck) <= 0 {
		// If we have run out of replacement cards, we just reduce the size of the
		// player's hand. We could do this in a slightly faster way as we do not
		// strictly need to preserve order, but it is probably less confusing for
		// the player to see the order of the cards in the hand stay unchanged.
		// We also do not worry about the card at the end of the array which is no
		// longer visible to the slice, as it can only ever be one card per player
		// before the game ends.
		gameState.playerHands[holdingPlayerName] =
			append(playerHand[:indexInHand], playerHand[indexInHand+1:]...)
	} else {
		// If we have a replacement card, we bundle it with the information about it
		// which the player should have.
		playerHand[indexInHand] =
			card.InHand{
				Readonly: gameState.undrawnDeck[0],
				Inferred: knowledgeOfDrawnCard,
			}

		// We should not ever re-visit this card, but in case we do somehow, we ensure
		// that this element represents an error.
		gameState.undrawnDeck[0] = card.ErrorReadonly()

		gameState.undrawnDeck = gameState.undrawnDeck[1:]
	}
}

// rollingMessageAppender holds a fixed number of messages, discarding the
// oldest when appending a new one.
type rollingMessageAppender struct {
	listLength      int
	messageList     []message.Readonly
	indexOfOldest   int
	mutualExclusion sync.Mutex
}

// newEmptyRollingMessageAppender makes a new rollingMessageAppender with a
// fixed message capacity, initially filled with empty messages
func newEmptyRollingMessageAppender(listLength int) *rollingMessageAppender {
	messageList := make([]message.Readonly, listLength)

	for messageIndex := 0; messageIndex < listLength; messageIndex++ {
		messageList[messageIndex] = message.NewReadonly("", "", "")
	}

	return &rollingMessageAppender{
		listLength:    listLength,
		messageList:   messageList,
		indexOfOldest: 0,
	}
}

// newRollingMessageAppender makes a new ollingMessageAppender with a fixed
// message capacity equal to the length of the given list of messages, with
// those messages as its initial messages.
func newRollingMessageAppender(
	initialMessages []message.Readonly) *rollingMessageAppender {
	listLength := len(initialMessages)
	messageList := make([]message.Readonly, listLength)

	for messageIndex := 0; messageIndex < listLength; messageIndex++ {
		messageList[messageIndex] = initialMessages[messageIndex]
	}

	return &rollingMessageAppender{
		listLength:    listLength,
		messageList:   messageList,
		indexOfOldest: 0,
	}
}

// sortedCopyOfMessages returns the messages in the log starting with the
// oldest in a simple array, in order by timestamp.
func (rollingAppender *rollingMessageAppender) sortedCopyOfMessages() []message.Readonly {
	logLength := rollingAppender.listLength
	sortedMessages := make([]message.Readonly, logLength)
	for messageIndex := 0; messageIndex < logLength; messageIndex++ {
		// We take the relevant message indexed with the oldest message at 0,
		// wrapping around if newer messages occupy earlier spots in the
		// actual array.
		adjustedIndex :=
			(messageIndex + rollingAppender.indexOfOldest) % logLength
		sortedMessages[messageIndex] = rollingAppender.messageList[adjustedIndex]
	}

	return sortedMessages
}

// appendNewMessage adds the given message as the newest message, over-writing
// the oldest message and increasing the offset of the index to the oldest
// message.
func (rollingAppender *rollingMessageAppender) appendNewMessage(
	playerName string,
	textColor string,
	messageText string) {
	rollingAppender.mutualExclusion.Lock()

	// We over-write the oldest message.
	rollingAppender.messageList[rollingAppender.indexOfOldest] =
		message.NewReadonly(playerName, textColor, messageText)

	// Now we mark the next-oldest message as the oldest, thus implicitly
	// marking the updated message as the newest message.
	rollingAppender.indexOfOldest =
		(rollingAppender.indexOfOldest + 1) % rollingAppender.listLength

	rollingAppender.mutualExclusion.Unlock()
}
