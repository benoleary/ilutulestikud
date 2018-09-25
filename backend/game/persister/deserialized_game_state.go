package persister

import (
	"fmt"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// DeserializedState is a struct meant to encapsulate all the state
// required for a single game to function, along with having de-serialized
// the ruleset from its identifier.
type DeserializedState struct {
	SerializableState
	deserializedRuleset     game.Ruleset
	PlayedCardsForColor     map[string][]card.Defined
	NumbersOfDiscardedCards map[card.Defined]int
	PlayerNamesToIndices    map[string]int
}

// CreateDeserializedState wraps a DeserializedState around a given ruleset
// and creates maps (which cannot be trivially serialized in Google Cloud
// Datastore, for example).
func CreateDeserializedState(
	serializableState SerializableState,
	deserializedRuleset game.Ruleset) DeserializedState {
	playedCardsForColor := make(map[string][]card.Defined)

	for _, playedCard := range serializableState.PlayedCards {
		cardColor := playedCard.ColorSuit

		// We can ignore whether or not there is already an array for
		// this color, as append(...) works even on a nil slice.
		playedBeforeThisCard, _ := playedCardsForColor[cardColor]
		playedCardsForColor[cardColor] = append(playedBeforeThisCard, playedCard)
	}

	numbersOfDiscardedCards := make(map[card.Defined]int)
	for _, discardedCard := range serializableState.DiscardedCards {
		// We can ignore whether or not there is already an count for
		// this card, as the default of 0 is correct in this case.
		numberOfCopiesBeforeThisCard, _ := numbersOfDiscardedCards[discardedCard]
		numbersOfDiscardedCards[discardedCard] = numberOfCopiesBeforeThisCard + 1
	}

	playerNames := serializableState.ParticipantNamesInTurnOrder
	numberOfPlayers := len(playerNames)
	playerNamesToIndices := make(map[string]int, numberOfPlayers)

	for playerIndex := 0; playerIndex < numberOfPlayers; playerIndex++ {
		playerNamesToIndices[playerNames[playerIndex]] = playerIndex
	}

	return DeserializedState{
		SerializableState:       serializableState,
		deserializedRuleset:     deserializedRuleset,
		PlayedCardsForColor:     playedCardsForColor,
		NumbersOfDiscardedCards: numbersOfDiscardedCards,
		PlayerNamesToIndices:    playerNamesToIndices,
	}
}

// Ruleset returns the ruleset for the game.
func (gameState *DeserializedState) Ruleset() game.Ruleset {
	return gameState.deserializedRuleset
}

// PlayedForColor returns the cards, in order, which have been played
// correctly for the given color suit.
func (gameState *DeserializedState) PlayedForColor(
	colorSuit string) []card.Defined {
	playedCards, _ := gameState.PlayedCardsForColor[colorSuit]

	if playedCards == nil {
		return []card.Defined{}
	}

	return playedCards
}

// NumberOfDiscardedCards returns the number of cards with the given suit and index
// which were discarded or played incorrectly.
func (gameState *DeserializedState) NumberOfDiscardedCards(
	colorSuit string,
	sequenceIndex int) int {
	mapKey :=
		card.Defined{
			ColorSuit:     colorSuit,
			SequenceIndex: sequenceIndex,
		}

	// We ignore the bool about whether it was found, as the default 0 for an int in
	// Go is the correct value to return.
	numberOfCopies, _ := gameState.NumbersOfDiscardedCards[mapKey]

	return numberOfCopies
}

// VisibleHand returns the card held by the given player in the given position.
func (gameState *DeserializedState) VisibleHand(
	holdingPlayerName string) ([]card.Defined, error) {
	playerIndex, hasHand := gameState.PlayerNamesToIndices[holdingPlayerName]

	if !hasHand {
		return nil, fmt.Errorf("Player %v has no hand", holdingPlayerName)
	}

	playerHand := gameState.PlayerHandsInTurnOrder[playerIndex]

	handSize := len(playerHand)

	visibleHand := make([]card.Defined, handSize)

	for indexInHand := 0; indexInHand < handSize; indexInHand++ {
		visibleHand[indexInHand] = playerHand[indexInHand].Defined
	}

	return visibleHand, nil
}

// InferredHand returns the inferred information about the card held by the given
// player in the given position.
func (gameState *DeserializedState) InferredHand(
	holdingPlayerName string) ([]card.Inferred, error) {
	playerIndex, hasHand := gameState.PlayerNamesToIndices[holdingPlayerName]

	if !hasHand {
		return nil, fmt.Errorf("Player %v has no hand", holdingPlayerName)
	}

	playerHand := gameState.PlayerHandsInTurnOrder[playerIndex]

	handSize := len(playerHand)

	inferredHand := make([]card.Inferred, handSize)

	for indexInHand := 0; indexInHand < handSize; indexInHand++ {
		inferredHand[indexInHand] = playerHand[indexInHand].Inferred
	}

	return inferredHand, nil
}

// Read returns the game state itself as a read-only object for the
// purposes of reading properties.
func (gameState *DeserializedState) Read() game.ReadonlyState {
	return gameState
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
func (gameState *DeserializedState) EnactTurnByDiscardingAndReplacing(
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
	deckAlreadyEmptyAtStartOfTurn := gameState.DeckSize() <= 0

	discardedCard, errorFromTakingCard :=
		gameState.takeCardFromHandReplacingIfPossible(
			actingPlayer.Name(),
			indexInHand,
			knowledgeOfDrawnCard)

	if errorFromTakingCard != nil {
		return errorFromTakingCard
	}

	// We can ignore whether or not there is already an count for
	// this card, as the default of 0 is correct in this case.
	numberOfCopiesBeforeThisCard, _ := gameState.NumbersOfDiscardedCards[discardedCard]
	gameState.NumbersOfDiscardedCards[discardedCard] = numberOfCopiesBeforeThisCard + 1

	// We also have to update the array which we serialize,
	// as the map is only part of the de-serialized state.
	gameState.DiscardedCards = append(gameState.DiscardedCards, discardedCard)

	gameState.NumberOfHintsAvailable += numberOfReadyHintsToAdd
	gameState.NumberOfMistakesMadeSoFar += numberOfMistakesMadeToAdd
	gameState.incrementTurnNumbers(deckAlreadyEmptyAtStartOfTurn)

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
// and any sequence index is possible). If there is no card to draw from the deck,
// it increments the number of turns taken with an empty deck of replacing the
// card in the hand. It also adds the given number of hints to the count of ready
// hints available (such as when playing the end of sequence gives a bonus hint).
func (gameState *DeserializedState) EnactTurnByPlayingAndReplacing(
	actionMessage string,
	actingPlayer player.ReadonlyState,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	numberOfReadyHintsToAdd int) error {
	// We need to check if the deck was empty at the start of the turn so that
	// we do not mistakenly increment the number of turns with an empty deck
	// on a turn which empties the deck, but we cannot increment the turn counts
	// until we have checked that we generate no errors.
	deckAlreadyEmptyAtStartOfTurn := gameState.DeckSize() <= 0

	playedCard, errorFromTakingCard :=
		gameState.takeCardFromHandReplacingIfPossible(
			actingPlayer.Name(),
			indexInHand,
			knowledgeOfDrawnCard)

	if errorFromTakingCard != nil {
		return errorFromTakingCard
	}

	playedSuit := playedCard.ColorSuit
	sequenceBeforeNow := gameState.PlayedCardsForColor[playedSuit]
	gameState.PlayedCardsForColor[playedSuit] =
		append(sequenceBeforeNow, playedCard)

	// We also have to update the array which we serialize,
	// as the map is only part of the de-serialized state.
	gameState.PlayedCards = append(gameState.PlayedCards, playedCard)

	gameState.NumberOfHintsAvailable += numberOfReadyHintsToAdd
	gameState.incrementTurnNumbers(deckAlreadyEmptyAtStartOfTurn)

	gameState.recordActionMessage(
		actingPlayer,
		actionMessage)

	return nil
}

// EnactTurnByUpdatingHandWithHint increments the turn number and replaces the
// given player's inferred hand with the given inferred hand, while also
// decrementing the number of available hints appropriately. If the deck is
// empty, this function also increments the number of turns taken with an empty
// deck.
func (gameState *DeserializedState) EnactTurnByUpdatingHandWithHint(
	actionMessage string,
	actingPlayer player.ReadonlyState,
	receivingPlayerName string,
	updatedReceiverKnowledgeOfOwnHand []card.Inferred,
	numberOfReadyHintsToSubtract int) error {
	recieverIndex, hasHand := gameState.PlayerNamesToIndices[receivingPlayerName]

	if !hasHand {
		return fmt.Errorf("Player %v has no hand", receivingPlayerName)
	}

	receiverHand := gameState.PlayerHandsInTurnOrder[recieverIndex]

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

	gameState.NumberOfHintsAvailable -= numberOfReadyHintsToSubtract

	// It is not a problem to take the deck size now as giving a hint does
	// not involve drawing from the deck.
	gameState.incrementTurnNumbers(gameState.DeckSize() <= 0)

	gameState.recordActionMessage(
		actingPlayer,
		actionMessage)

	return nil
}

// ReplaceCardInHand replaces the card at the given index in the hand of the given
// player with the given replacement card, and returns the card which has just been
// replaced.
func (gameState *DeserializedState) takeCardFromHandReplacingIfPossible(
	holdingPlayerName string,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred) (card.Defined, error) {
	if indexInHand < 0 {
		return invalidCardAndErrorFromOutOfRange(indexInHand)
	}

	holdingPlayerIndex, hasHand := gameState.PlayerNamesToIndices[holdingPlayerName]

	if !hasHand {
		errorFromNoHand := fmt.Errorf("Player %v has no hand", holdingPlayerName)
		invalidCard :=
			card.Defined{
				ColorSuit:     errorFromNoHand.Error(),
				SequenceIndex: -1,
			}

		return invalidCard, errorFromNoHand
	}

	playerHand := gameState.PlayerHandsInTurnOrder[holdingPlayerIndex]

	if indexInHand >= len(playerHand) {
		return invalidCardAndErrorFromOutOfRange(indexInHand)
	}

	cardBeingReplaced := playerHand[indexInHand]

	gameState.updatePlayerHand(
		holdingPlayerIndex,
		indexInHand,
		knowledgeOfDrawnCard,
		playerHand)

	return cardBeingReplaced.Defined, nil
}

func (gameState *DeserializedState) updatePlayerHand(
	holdingPlayerIndex int,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred,
	playerHand []card.InHand) {
	if len(gameState.UndrawnDeck) <= 0 {
		// If we have run out of replacement cards, we just reduce the size of the
		// player's hand. We could do this in a slightly faster way as we do not
		// strictly need to preserve order, but it is probably less confusing for
		// the player to see the order of the cards in the hand stay unchanged.
		// We also do not worry about the card at the end of the array which is no
		// longer visible to the slice, as it can only ever be one card per player
		// before the game ends.
		gameState.PlayerHandsInTurnOrder[holdingPlayerIndex] =
			append(playerHand[:indexInHand], playerHand[indexInHand+1:]...)
	} else {
		// If we have a replacement card, we bundle it with the information about it
		// which the player should have.
		playerHand[indexInHand] =
			card.InHand{
				Defined:  gameState.UndrawnDeck[0],
				Inferred: knowledgeOfDrawnCard,
			}

		// We should not ever re-visit this card, but in case we do somehow, we ensure
		// that this element represents an error.
		gameState.UndrawnDeck[0] =
			card.Defined{
				ColorSuit:     "error: already removed from deck",
				SequenceIndex: -1,
			}

		gameState.UndrawnDeck = gameState.UndrawnDeck[1:]
	}
}
