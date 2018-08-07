package persister

import (
	"fmt"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// SerializableState is a struct meant to encapsulate all the state required
// for a single game to function, in a form which is simple to serialize.
type SerializableState struct {
	GameName                    string
	GameRuleset                 game.Ruleset
	CreationTime                time.Time
	ParticipantNamesInTurnOrder []string
	ChatLog                     []message.FromPlayer
	ActionLog                   []message.FromPlayer
	TurnNumber                  int
	TurnsTakenWithEmptyDeck     int
	NumberOfReadyHints          int
	NumberOfMistakesMade        int
	UndrawnDeck                 []card.Defined
	PlayedCardsForColor         map[string][]card.Defined
	DiscardedCards              map[card.Defined]int
	PlayerHands                 map[string][]card.InHand
}

// NewSerializableState creates a new game given the required information, using the
// given shuffled deck.
func NewSerializableState(
	gameName string,
	chatLogLength int,
	initialActionLog []message.FromPlayer,
	gameRuleset game.Ruleset,
	playersInTurnOrderWithInitialHands []game.PlayerNameWithHand,
	shuffledDeck []card.Defined) SerializableState {
	numberOfParticipants := len(playersInTurnOrderWithInitialHands)
	participantNamesInTurnOrder := make([]string, numberOfParticipants)
	playerHands := make(map[string][]card.InHand, numberOfParticipants)
	for playerIndex := 0; playerIndex < numberOfParticipants; playerIndex++ {
		playerName := playersInTurnOrderWithInitialHands[playerIndex].PlayerName
		participantNamesInTurnOrder[playerIndex] = playerName
		playerHands[playerName] =
			playersInTurnOrderWithInitialHands[playerIndex].InitialHand
	}

	initialChatLog := make([]message.FromPlayer, chatLogLength)

	for messageIndex := 0; messageIndex < chatLogLength; messageIndex++ {
		initialChatLog[messageIndex] = message.NewFromPlayer("", "", "")
	}

	// We could already set up the capacity for the maps by getting slices from
	// the ruleset and counting, but that is a lot of effort for very little gain.
	return SerializableState{
		GameName:                    gameName,
		GameRuleset:                 gameRuleset,
		CreationTime:                time.Now(),
		ParticipantNamesInTurnOrder: participantNamesInTurnOrder,
		ChatLog:                     initialChatLog,
		ActionLog:                   initialActionLog,
		TurnNumber:                  1,
		TurnsTakenWithEmptyDeck:     0,
		NumberOfReadyHints:          gameRuleset.MaximumNumberOfHints(),
		NumberOfMistakesMade:        0,
		UndrawnDeck:                 shuffledDeck,
		PlayedCardsForColor:         make(map[string][]card.Defined, 0),
		DiscardedCards:              make(map[card.Defined]int, 0),
		PlayerHands:                 playerHands,
	}
}

// DeckSize returns the number of cards left to draw from the deck.
func (serializableState *SerializableState) DeckSize() int {
	return len(serializableState.UndrawnDeck)
}

// PlayedForColor returns the cards, in order, which have been played
// correctly for the given color suit.
func (serializableState *SerializableState) PlayedForColor(
	colorSuit string) []card.Defined {
	playedCards, _ :=
		serializableState.PlayedCardsForColor[colorSuit]

	if playedCards == nil {
		return []card.Defined{}
	}

	return playedCards
}

// NumberOfDiscardedCards returns the number of cards with the given suit and index
// which were discarded or played incorrectly.
func (serializableState *SerializableState) NumberOfDiscardedCards(
	colorSuit string,
	sequenceIndex int) int {
	mapKey :=
		card.Defined{
			ColorSuit:     colorSuit,
			SequenceIndex: sequenceIndex,
		}

	// We ignore the bool about whether it was found, as the default 0 for an int in
	// Go is the correct value to return.
	numberOfCopies, _ := serializableState.DiscardedCards[mapKey]

	return numberOfCopies
}

// VisibleHand returns the card held by the given player in the given position.
func (serializableState *SerializableState) VisibleHand(
	holdingPlayerName string) ([]card.Defined, error) {
	playerHand, hasHand := serializableState.PlayerHands[holdingPlayerName]

	if !hasHand {
		return nil, fmt.Errorf("Player has no hand")
	}

	handSize := len(playerHand)

	visibleHand := make([]card.Defined, handSize)

	for indexInHand := 0; indexInHand < handSize; indexInHand++ {
		visibleHand[indexInHand] = playerHand[indexInHand].Defined
	}

	return visibleHand, nil
}

// InferredHand returns the inferred information about the card held by the given
// player in the given position.
func (serializableState *SerializableState) InferredHand(
	holdingPlayerName string) ([]card.Inferred, error) {
	playerHand, hasHand := serializableState.PlayerHands[holdingPlayerName]

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

// RecordChatMessage records a chat message from the given player.
func (serializableState *SerializableState) RecordChatMessage(
	actingPlayer player.ReadonlyState,
	chatMessage string) error {
	appendNewMessageInPlaceDiscardingFirst(
		serializableState.ChatLog,
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
// and any sequence index is possible). If there is no card to draw from the
// deck, it increments the number of turns taken with an empty deck of
// replacing the card in the hand. It also adds the given numbers to the
// counts of available hints and mistakes made respectively.
func (serializableState *SerializableState) EnactTurnByDiscardingAndReplacing(
	actionMessage string,
	actingPlayer player.ReadonlyState,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	numberOfReadyHintsToAdd int,
	numberOfMistakesMadeToAdd int) error {
	// We need to check if the deck was empty at the start of the turn so that
	// we do not mistakenly increment the number of turns with an empty deck
	// on a turn which empties the deck, but we cannot increment the turn counts
	// until we have checked that we generate no errors.
	deckAlreadyEmptyAtStartOfTurn := serializableState.DeckSize() <= 0

	discardedCard, errorFromTakingCard :=
		serializableState.takeCardFromHandReplacingIfPossible(
			actingPlayer.Name(),
			indexInHand,
			knowledgeOfDrawnCard)

	if errorFromTakingCard != nil {
		return errorFromTakingCard
	}

	discardedCopiesUntilNow, _ := serializableState.DiscardedCards[discardedCard]
	serializableState.DiscardedCards[discardedCard] = discardedCopiesUntilNow + 1

	serializableState.NumberOfReadyHints += numberOfReadyHintsToAdd
	serializableState.NumberOfMistakesMade += numberOfMistakesMadeToAdd
	serializableState.incrementTurnNumbers(deckAlreadyEmptyAtStartOfTurn)

	serializableState.recordActionMessage(
		actingPlayer,
		actionMessage)

	return nil
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
func (serializableState *SerializableState) EnactTurnByPlayingAndReplacing(
	actionMessage string,
	actingPlayer player.ReadonlyState,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	numberOfReadyHintsToAdd int) error {
	// We need to check if the deck was empty at the start of the turn so that
	// we do not mistakenly increment the number of turns with an empty deck
	// on a turn which empties the deck, but we cannot increment the turn counts
	// until we have checked that we generate no errors.
	deckAlreadyEmptyAtStartOfTurn := serializableState.DeckSize() <= 0

	playedCard, errorFromTakingCard :=
		serializableState.takeCardFromHandReplacingIfPossible(
			actingPlayer.Name(),
			indexInHand,
			knowledgeOfDrawnCard)

	if errorFromTakingCard != nil {
		return errorFromTakingCard
	}

	playedSuit := playedCard.ColorSuit
	sequenceBeforeNow := serializableState.PlayedCardsForColor[playedSuit]
	serializableState.PlayedCardsForColor[playedSuit] =
		append(sequenceBeforeNow, playedCard)

	serializableState.NumberOfReadyHints += numberOfReadyHintsToAdd
	serializableState.incrementTurnNumbers(deckAlreadyEmptyAtStartOfTurn)

	serializableState.recordActionMessage(
		actingPlayer,
		actionMessage)

	return nil
}

// EnactTurnByUpdatingHandWithHint increments the turn number and replaces the
// given player's inferred hand with the given inferred hand, while also
// decrementing the number of available hints appropriately. If the deck is
// empty, this function also increments the number of turns taken with an empty
// deck.
func (serializableState *SerializableState) EnactTurnByUpdatingHandWithHint(
	actionMessage string,
	actingPlayer player.ReadonlyState,
	receivingPlayerName string,
	updatedReceiverKnowledgeOfOwnHand []card.Inferred,
	numberOfReadyHintsToSubtract int) error {
	receiverHand, hasHand := serializableState.PlayerHands[receivingPlayerName]

	if !hasHand {
		return fmt.Errorf("Player %v has no hand", receivingPlayerName)
	}

	handSize := len(receiverHand)

	if len(updatedReceiverKnowledgeOfOwnHand) != handSize {
		return fmt.Errorf(
			"Updated hand knowledge %+v does not match hand size %v",
			updatedReceiverKnowledgeOfOwnHand,
			handSize)
	}

	for indexInHand := 0; indexInHand < handSize; indexInHand++ {
		receiverHand[indexInHand].Inferred =
			updatedReceiverKnowledgeOfOwnHand[indexInHand]
	}

	serializableState.NumberOfReadyHints -= numberOfReadyHintsToSubtract

	// It is not a problem to take the deck size now as giving a hint does
	// not involve drawing from the deck.
	serializableState.incrementTurnNumbers(serializableState.DeckSize() <= 0)

	serializableState.recordActionMessage(
		actingPlayer,
		actionMessage)

	return nil
}

func (serializableState *SerializableState) incrementTurnNumbers(
	deckAlreadyEmptyAtStartOfTurn bool) {
	serializableState.TurnNumber++

	if deckAlreadyEmptyAtStartOfTurn {
		serializableState.TurnsTakenWithEmptyDeck++
	}
}

func (serializableState *SerializableState) recordActionMessage(
	actingPlayer player.ReadonlyState,
	actionMessage string) {
	appendNewMessageInPlaceDiscardingFirst(
		serializableState.ActionLog,
		actingPlayer.Name(),
		actingPlayer.Color(),
		actionMessage)
}

// ReplaceCardInHand replaces the card at the given index in the hand of the given
// player with the given replacement card, and returns the card which has just been
// replaced.
func (serializableState *SerializableState) takeCardFromHandReplacingIfPossible(
	holdingPlayerName string,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred) (card.Defined, error) {
	if indexInHand < 0 {
		return invalidCardAndErrorFromOutOfRange(indexInHand)
	}

	playerHand, hasHand := serializableState.PlayerHands[holdingPlayerName]

	if !hasHand {
		errorFromNoHand := fmt.Errorf("Player has no hand")
		invalidCard :=
			card.Defined{
				ColorSuit:     errorFromNoHand.Error(),
				SequenceIndex: -1,
			}

		return invalidCard, errorFromNoHand
	}

	if indexInHand >= len(playerHand) {
		return invalidCardAndErrorFromOutOfRange(indexInHand)
	}

	cardBeingReplaced := playerHand[indexInHand]

	serializableState.updatePlayerHand(
		holdingPlayerName,
		indexInHand,
		knowledgeOfDrawnCard,
		playerHand)

	return cardBeingReplaced.Defined, nil
}

func (serializableState *SerializableState) updatePlayerHand(
	holdingPlayerName string,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	playerHand []card.InHand) {
	if len(serializableState.UndrawnDeck) <= 0 {
		// If we have run out of replacement cards, we just reduce the size of the
		// player's hand. We could do this in a slightly faster way as we do not
		// strictly need to preserve order, but it is probably less confusing for
		// the player to see the order of the cards in the hand stay unchanged.
		// We also do not worry about the card at the end of the array which is no
		// longer visible to the slice, as it can only ever be one card per player
		// before the game ends.
		serializableState.PlayerHands[holdingPlayerName] =
			append(playerHand[:indexInHand], playerHand[indexInHand+1:]...)
	} else {
		// If we have a replacement card, we bundle it with the information about it
		// which the player should have.
		playerHand[indexInHand] =
			card.InHand{
				Defined:  serializableState.UndrawnDeck[0],
				Inferred: knowledgeOfDrawnCard,
			}

		// We should not ever re-visit this card, but in case we do somehow, we ensure
		// that this element represents an error.
		serializableState.UndrawnDeck[0] =
			card.Defined{
				ColorSuit:     "error: already removed from deck",
				SequenceIndex: -1,
			}

		serializableState.UndrawnDeck = serializableState.UndrawnDeck[1:]
	}
}

func invalidCardAndErrorFromOutOfRange(indexOutOfRange int) (card.Defined, error) {
	errorFromOutOfRange := fmt.Errorf("Index %v is out of allowed range", indexOutOfRange)
	invalidCard :=
		card.Defined{
			ColorSuit:     errorFromOutOfRange.Error(),
			SequenceIndex: -1,
		}

	return invalidCard, errorFromOutOfRange
}

// appendNewMessageInPlaceDiscardingFirst copies each message after the
// first in the given array into the position before it (thus eliminating
// the original first message of the array) and over-writes the last
// message in the array with the given message.
// I originally had a different structure which would only over-write the
// oldest message, and kept track of which was the oldest with an index,
// and rolled over from the end of the array back to the start when
// reading the log, but that would have been too much faff to serialize
// for totally negligible performance differences.
func appendNewMessageInPlaceDiscardingFirst(
	messageSlice []message.FromPlayer,
	playerName string,
	textColor string,
	messageText string) {
	sliceLength := len(messageSlice)

	for messageIndex := 1; messageIndex < sliceLength; messageIndex++ {
		messageSlice[messageIndex-1] = messageSlice[messageIndex]
	}

	messageSlice[sliceLength-1] =
		message.NewFromPlayer(playerName, textColor, messageText)
}
