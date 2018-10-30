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

	handStart := gameState.PlayerHandStartIndicesInTurnOrder[playerIndex]
	flattenedCards := gameState.FlattenedDefinedCardsInHands

	handEnd, _ := gameState.indexOfHandEnd(playerIndex)
	playerHand := flattenedCards[handStart:handEnd]

	visibleHand := make([]card.Defined, handEnd-handStart)

	copy(visibleHand, playerHand)

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

	handStart := gameState.PlayerHandStartIndicesInTurnOrder[playerIndex]
	flattenedCards := gameState.FlattenedInferredCardsInHands

	handEnd, indexOfLastPlayer :=
		gameState.indexOfHandEnd(playerIndex)
	playerHand := flattenedCards[handStart:handEnd]

	handSize := len(playerHand)

	inferredHand := make([]card.Inferred, handSize)

	for indexInHand := 0; indexInHand < handSize; indexInHand++ {
		colorStart := playerHand[indexInHand].StartIndexOfColors
		indexStart := playerHand[indexInHand].StartIndexOfIndices

		var colorEnd int
		var indexEnd int

		// If this player is not the last in turn order, then there definitely
		// is another card in the flattened array. If this player is the last
		// player though, then for the last card in the hand, there is no next
		// card, so the end of the array marks the end of the inferred knowledge
		// for that card.
		if (playerIndex < indexOfLastPlayer) || (indexInHand < (handSize - 1)) {
			nextCard := gameState.FlattenedInferredCardsInHands[handStart+indexInHand+1]
			colorEnd = nextCard.StartIndexOfColors
			indexEnd = nextCard.StartIndexOfIndices
		} else {
			colorEnd = len(gameState.FlattenedInferredColors)
			indexEnd = len(gameState.FlattenedInferredIndices)
		}

		inferredHand[indexInHand].PossibleColors =
			make([]string, colorEnd-colorStart)
		copy(
			inferredHand[indexInHand].PossibleColors,
			gameState.FlattenedInferredColors[colorStart:colorEnd])
		inferredHand[indexInHand].PossibleIndices =
			make([]int, indexEnd-indexStart)
		copy(
			inferredHand[indexInHand].PossibleIndices,
			gameState.FlattenedInferredIndices[indexStart:indexEnd])
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
	receiverIndex, hasHand := gameState.PlayerNamesToIndices[receivingPlayerName]

	if !hasHand {
		return fmt.Errorf("Player %v has no hand", receivingPlayerName)
	}

	handStart := gameState.PlayerHandStartIndicesInTurnOrder[receiverIndex]
	flattenedCards := gameState.FlattenedInferredCardsInHands

	handEnd, indexOfLastPlayer :=
		gameState.indexOfHandEnd(receiverIndex)

	handSize := handEnd - handStart

	if len(updatedReceiverKnowledgeOfOwnHand) != handSize {
		return fmt.Errorf(
			"Updated hand knowledge %+v does not match hand size %v",
			updatedReceiverKnowledgeOfOwnHand,
			handSize)
	}

	startOfColorChange := flattenedCards[handStart].StartIndexOfColors
	startOfIndexChange := flattenedCards[handStart].StartIndexOfIndices

	endOfUpdatedColors, endOfUpdatedIndices :=
		gameState.updateInferredIndicesOfHintedHand(
			updatedReceiverKnowledgeOfOwnHand,
			handSize,
			handStart,
			startOfColorChange,
			startOfIndexChange)

	// If we are updating a hand which is not the last in the flattened array,
	// we have to shunt the indices of the subsequent hands to account for the
	// changes in this hand. We make a note of where the hands began (the default
	// values of zero are not used if there are no following hands).
	startOfFollowingColors, startOfFollowingIndices, fullColorsLength, fullIndicesLength :=
		gameState.updateIndicesOfSubsequentHands(
			receiverIndex,
			indexOfLastPlayer,
			endOfUpdatedColors,
			endOfUpdatedIndices,
			handStart,
			handEnd)

	gameState.updateFlattenedInferredArrays(
		fullColorsLength,
		startOfColorChange,
		fullIndicesLength,
		startOfIndexChange,
		handSize,
		updatedReceiverKnowledgeOfOwnHand,
		receiverIndex,
		indexOfLastPlayer,
		startOfFollowingColors,
		startOfFollowingIndices)

	gameState.NumberOfHintsAvailable -= numberOfReadyHintsToSubtract

	// It is not a problem to take the deck size now, as giving a hint does
	// not involve drawing from the deck.
	gameState.incrementTurnNumbers(gameState.DeckSize() <= 0)

	gameState.recordActionMessage(
		actingPlayer,
		actionMessage)

	return nil
}

// Returns the "slice end index" (i.e. the index which is used as the parameter
// after the colon to define a slice) of the hand of the given player (using the
// flattened array of defined cards, but since the inferred knowledge structs are
// in one-to-one correspondence with their cards, it should be fine). If this is
// the last hand in the flattened array (e.g. with player index 3 in a 4-player
// game), the hand ends at the end of the flattened array. Otherwise, it ends at
// the start of the next hand. It also returns the index of the last player.
func (gameState *DeserializedState) indexOfHandEnd(
	playerIndex int) (handEndIndex int, lastPlayerIndex int) {
	lastPlayerIndex = len(gameState.ParticipantNamesInTurnOrder) - 1
	if playerIndex < lastPlayerIndex {
		return gameState.PlayerHandStartIndicesInTurnOrder[playerIndex+1], lastPlayerIndex
	}

	return len(gameState.FlattenedDefinedCardsInHands), lastPlayerIndex
}

func (gameState *DeserializedState) updateInferredIndicesOfHintedHand(
	updatedReceiverKnowledgeOfOwnHand []card.Inferred,
	handSize int,
	handStart int,
	startOfColorChange int,
	startOfIndexChange int) (int, int) {
	updatedColorSliceLength :=
		len(updatedReceiverKnowledgeOfOwnHand[0].PossibleColors)
	updatedIndexSliceLength :=
		len(updatedReceiverKnowledgeOfOwnHand[0].PossibleIndices)

	for indexInHand := 1; indexInHand < handSize; indexInHand++ {
		gameState.FlattenedInferredCardsInHands[handStart+indexInHand].StartIndexOfColors =
			startOfColorChange + updatedColorSliceLength
		gameState.FlattenedInferredCardsInHands[handStart+indexInHand].StartIndexOfIndices =
			startOfIndexChange + updatedIndexSliceLength

		updatedColorSliceLength +=
			len(updatedReceiverKnowledgeOfOwnHand[indexInHand].PossibleColors)
		updatedIndexSliceLength +=
			len(updatedReceiverKnowledgeOfOwnHand[indexInHand].PossibleIndices)
	}

	updatedFlattenedColorsLength := startOfColorChange + updatedColorSliceLength
	updatedFlattenedIndicesLength := startOfIndexChange + updatedIndexSliceLength

	return updatedFlattenedColorsLength, updatedFlattenedIndicesLength
}

// updateIndicesOfSubsequentHands updates the indices of hands in the flattened
// arrays which come after a hand which has just been updated by a hint.
func (gameState *DeserializedState) updateIndicesOfSubsequentHands(
	receiverIndex int,
	indexOfLastPlayer int,
	updatedColorsEnd int,
	updatedIndicesEnd int,
	handStart int,
	handEnd int) (int, int, int, int) {
	if receiverIndex >= indexOfLastPlayer {
		return 0, 0, updatedColorsEnd, updatedIndicesEnd
	}

	// The next hand should start where the updated hand ends.
	// Right now, the updated hand has indeed been updated, but
	// the following hands have not yet been updated. This means
	// that we can get an offset by comparing where the following
	// hard are currently pointing (outdated) to where they should
	// be pointing.
	flattenedCards := gameState.FlattenedInferredCardsInHands
	firstFollowingCard := flattenedCards[handEnd]
	followingColorStart := firstFollowingCard.StartIndexOfColors
	followingIndexStart := firstFollowingCard.StartIndexOfIndices

	colorsOffset := updatedColorsEnd - followingColorStart
	colorsFullLength := updatedColorsEnd + colorsOffset

	indicesOffset := updatedIndicesEnd - followingIndexStart
	indicesFullLength := updatedIndicesEnd + indicesOffset

	// The first card of the player following the updated player in the
	// flattened array is the first we update, and we update all cards
	// until the end of the array.
	indexOfLastCard := len(flattenedCards) - 1
	for cardIndex := handEnd; cardIndex <= indexOfLastCard; cardIndex++ {
		flattenedCards[cardIndex].StartIndexOfColors += colorsOffset
		flattenedCards[cardIndex].StartIndexOfIndices += indicesOffset
	}

	return followingColorStart, followingIndexStart, colorsFullLength, indicesFullLength
}

func (gameState *DeserializedState) updateFlattenedInferredArrays(
	updatedFlattenedColorsLength int,
	startOfColorChange int,
	updatedFlattenedIndicesLength int,
	startOfIndexChange int,
	handSize int,
	updatedReceiverKnowledgeOfOwnHand []card.Inferred,
	receiverIndex int,
	indexOfLastPlayer int,
	startOfFollowingColors int,
	startOfFollowingIndices int) {
	updatedFlattenedColors :=
		make([]string, 0, updatedFlattenedColorsLength)
	updatedFlattenedColors =
		append(
			updatedFlattenedColors,
			gameState.FlattenedInferredColors[:startOfColorChange]...)
	updatedFlattenedIndices :=
		make([]int, 0, updatedFlattenedIndicesLength)
	updatedFlattenedIndices =
		append(
			updatedFlattenedIndices,
			gameState.FlattenedInferredIndices[:startOfIndexChange]...)

	for indexInHand := 0; indexInHand < handSize; indexInHand++ {
		updatedFlattenedColors =
			append(
				updatedFlattenedColors,
				updatedReceiverKnowledgeOfOwnHand[indexInHand].PossibleColors...)
		updatedFlattenedIndices =
			append(
				updatedFlattenedIndices,
				updatedReceiverKnowledgeOfOwnHand[indexInHand].PossibleIndices...)
	}

	// If we are updating a hand which is not the last in the flattened array,
	// we can just take the remainder of the original flattened array.
	if receiverIndex < indexOfLastPlayer {
		updatedFlattenedColors =
			append(
				updatedFlattenedColors,
				gameState.FlattenedInferredColors[startOfFollowingColors:]...)
		updatedFlattenedIndices =
			append(
				updatedFlattenedIndices,
				gameState.FlattenedInferredIndices[startOfFollowingIndices:]...)
	}

	gameState.FlattenedInferredColors = updatedFlattenedColors
	gameState.FlattenedInferredIndices = updatedFlattenedIndices
}

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

	handEnd, indexOfLastPlayer :=
		gameState.indexOfHandEnd(holdingPlayerIndex)

	if indexInHand >= handEnd {
		return invalidCardAndErrorFromOutOfRange(indexInHand)
	}

	handStart :=
		gameState.PlayerHandStartIndicesInTurnOrder[holdingPlayerIndex]
	cardBeingReplaced :=
		gameState.FlattenedDefinedCardsInHands[handStart+indexInHand]

	gameState.updatePlayerHand(
		holdingPlayerIndex,
		indexOfLastPlayer,
		handStart,
		indexInHand,
		knowledgeOfDrawnCard)

	return cardBeingReplaced, nil
}

func (gameState *DeserializedState) updatePlayerHand(
	holdingPlayerIndex int,
	indexOfLastPlayer int,
	handStartIndex int,
	indexInHand int,
	knowledgeOfDrawnCard card.Inferred) {
	isEmptyDeck := len(gameState.UndrawnDeck) <= 0
	indexOfCardToReplaceOrRemove := handStartIndex + indexInHand
	inferredToRemoveOrReplace :=
		gameState.FlattenedInferredCardsInHands[indexOfCardToReplaceOrRemove]

	fmt.Printf("\n\nupdatePlayerHand(holdingPlayerIndex = %v,"+
		" indexOfLastPlayer = %v, handStartIndex = %v, indexInHand = %v,"+
		" knowledgeOfDrawnCard = %+v):\n isEmptyDeck = %v, indexOfCardToReplaceOrRemove = %v, inferredToRemoveOrReplace = %+v\n\n",
		holdingPlayerIndex,
		indexOfLastPlayer,
		handStartIndex,
		indexInHand,
		knowledgeOfDrawnCard,
		isEmptyDeck,
		indexOfCardToReplaceOrRemove,
		inferredToRemoveOrReplace)

	originalLastIndex, isLastCard, firstToUpdate :=
		gameState.determineLastCardAndWhereToUpdate(
			indexOfCardToReplaceOrRemove,
			isEmptyDeck)

	colorsIncrease, replacementColors, indicesIncrease, replacementIndices :=
		gameState.replacementInferred(
			indexOfCardToReplaceOrRemove,
			inferredToRemoveOrReplace,
			firstToUpdate,
			isLastCard,
			knowledgeOfDrawnCard)

	// We update to the original last index because we have not even
	// determined if we will remove any card at this point.
	gameState.updateFlattenedInferred(
		inferredToRemoveOrReplace,
		colorsIncrease,
		replacementColors,
		indicesIncrease,
		replacementIndices,
		firstToUpdate,
		originalLastIndex)

	if !isEmptyDeck {
		// If the deck is not empty, we simply replace the card with the top of the deck.
		gameState.replaceCardFromDeck(indexOfCardToReplaceOrRemove)
	} else {
		// If the deck is now empty, we have to remove the card from the flattened arrays.
		gameState.removeCardFromFlattenedArrays(
			indexOfCardToReplaceOrRemove,
			originalLastIndex,
			isLastCard,
			firstToUpdate)

		// We also have to note that the hands after this player now
		// start one card earlier in the flattened arrays.
		for playerIndex := holdingPlayerIndex + 1; playerIndex <= indexOfLastPlayer; playerIndex++ {
			gameState.PlayerHandStartIndicesInTurnOrder[playerIndex]--
		}
	}
}

func (gameState *DeserializedState) determineLastCardAndWhereToUpdate(
	indexOfCardToReplaceOrRemove int,
	isEmptyDeck bool) (int, bool, int) {
	indexOfLastOriginalCard :=
		len(gameState.FlattenedDefinedCardsInHands) - 1

	isLastCard := indexOfCardToReplaceOrRemove == indexOfLastOriginalCard
	indexOfFirstCardToUpdate := 0
	if !isLastCard {
		indexOfFirstCardToUpdate = indexOfCardToReplaceOrRemove + 1
	}

	return indexOfLastOriginalCard, isLastCard, indexOfFirstCardToUpdate
}

func (gameState *DeserializedState) replacementInferred(
	indexOfCard int,
	inferredToRemoveOrReplace InferredCardFromFlattenedIndices,
	indexOfFirstCardToUpdate int,
	isLastCard bool,
	knowledgeOfDrawnCard card.Inferred) (int, []string, int, []int) {
	endOfColorsToRemove := len(gameState.FlattenedInferredColors)
	endOfIndicesToRemove := len(gameState.FlattenedInferredIndices)
	if !isLastCard {
		nextInferred :=
			gameState.FlattenedInferredCardsInHands[indexOfFirstCardToUpdate]
		endOfColorsToRemove = nextInferred.StartIndexOfColors
		endOfIndicesToRemove = nextInferred.StartIndexOfIndices
	}

	// First we assume that the existing inferred knowledge will be removed
	// (hence a negative length, as it should never have gotten to the state
	// where there is no inferred color or index for a card).
	colorsIncrease :=
		inferredToRemoveOrReplace.StartIndexOfColors - endOfColorsToRemove
	replacementColors := []string{}
	indicesIncrease :=
		inferredToRemoveOrReplace.StartIndexOfIndices - endOfIndicesToRemove
	replacementIndices := []int{}

	isEmptyDeck := len(gameState.UndrawnDeck) <= 0

	// If there is no card to draw from the deck as a replacement, we will be
	// simply deleting the card from the hand, and thus removing all the inferred
	// knowledge about it, so the length increase will remain the negative value
	// it was assigned when it was initialized. Otherwise, it will be the length
	// of the given knowledge minus the original length, which is to say the
	// increase variable is increased by the length of the replacement knowledge.
	if !isEmptyDeck {
		colorsIncrease += len(knowledgeOfDrawnCard.PossibleColors)
		indicesIncrease += len(knowledgeOfDrawnCard.PossibleIndices)

		replacementColors = knowledgeOfDrawnCard.PossibleColors
		replacementIndices = knowledgeOfDrawnCard.PossibleIndices
	}

	return colorsIncrease, replacementColors, indicesIncrease, replacementIndices
}

func (gameState *DeserializedState) updateFlattenedInferred(
	inferredToRemoveOrReplace InferredCardFromFlattenedIndices,
	colorOffset int,
	replacementColors []string,
	indexOffset int,
	replacementIndices []int,
	indexOfFirstCardToUpdate int,
	indexOfLastCardToUpdate int) {
	// We update the inferred knowledge arrays before we update the arrays of inferred cards
	// as the indices can be updated without knowing what is going on in the knowledge arrays,
	// but updating the knowledge arrays requires knowing the index of the end of the card to
	// update.
	updatedColors :=
		make([]string, 0, len(gameState.FlattenedInferredColors)+colorOffset)
	updatedColors =
		append(
			updatedColors,
			gameState.FlattenedInferredColors[:inferredToRemoveOrReplace.StartIndexOfColors]...)
	updatedColors =
		append(
			updatedColors,
			replacementColors...)

	updatedIndices :=
		make([]int, 0, len(gameState.FlattenedInferredIndices)+indexOffset)
	updatedIndices =
		append(
			updatedIndices,
			gameState.FlattenedInferredIndices[:inferredToRemoveOrReplace.StartIndexOfIndices]...)
	updatedIndices =
		append(
			updatedIndices,
			replacementIndices...)

	// If the card is the last card and it also is being removed without replacement,
	// lastCardIndex is already less than firstCardToUpdate, so we do not update any
	// inferred card and we do not append the non-existent remainder of the array.
	if indexOfFirstCardToUpdate <= indexOfLastCardToUpdate {
		firstCardToUpdate :=
			gameState.FlattenedInferredCardsInHands[indexOfFirstCardToUpdate]

		updatedColors =
			append(
				updatedColors,
				gameState.FlattenedInferredColors[firstCardToUpdate.StartIndexOfColors:]...)

		updatedIndices =
			append(
				updatedIndices,
				gameState.FlattenedInferredIndices[firstCardToUpdate.StartIndexOfIndices:]...)

		flattenedCards := gameState.FlattenedInferredCardsInHands

		for cardIndex := indexOfFirstCardToUpdate; cardIndex <= indexOfLastCardToUpdate; cardIndex++ {
			flattenedCards[cardIndex].StartIndexOfColors += colorOffset
			flattenedCards[cardIndex].StartIndexOfIndices += indexOffset
		}
	}

	gameState.FlattenedInferredColors = updatedColors
	gameState.FlattenedInferredIndices = updatedIndices
}

func (gameState *DeserializedState) removeCardFromFlattenedArrays(
	indexOfCardToReplaceOrRemove int,
	originalLastIndex int,
	isLastCard bool,
	firstToUpdate int) {
	// There should be exactly as many inferred knowledge objects as defined cards,
	// and this number is equal to the index of the last card before the removal.
	updatedInferreds :=
		make([]InferredCardFromFlattenedIndices, 0, originalLastIndex)
	updatedInferreds =
		append(updatedInferreds,
			gameState.FlattenedInferredCardsInHands[:indexOfCardToReplaceOrRemove]...)

	updatedDefineds :=
		make([]card.Defined, 0, originalLastIndex)
	updatedDefineds =
		append(updatedDefineds,
			gameState.FlattenedDefinedCardsInHands[:indexOfCardToReplaceOrRemove]...)

	if !isLastCard {
		updatedInferreds =
			append(updatedInferreds,
				gameState.FlattenedInferredCardsInHands[firstToUpdate:]...)
		updatedDefineds =
			append(updatedDefineds,
				gameState.FlattenedDefinedCardsInHands[firstToUpdate:]...)

	}

	gameState.FlattenedDefinedCardsInHands = updatedDefineds
	gameState.FlattenedInferredCardsInHands = updatedInferreds
}

func (gameState *DeserializedState) replaceCardFromDeck(indexOfCard int) {
	// The inferred knowledge should have already been sorted out by
	// gameState.updateFlattenedInferred(...), so we only have to update
	// the defined card and the deck in this case.
	gameState.FlattenedDefinedCardsInHands[indexOfCard] =
		gameState.UndrawnDeck[0]

	// We should not ever re-visit this card, but in case we do somehow, we ensure
	// that this element represents an error.
	gameState.UndrawnDeck[0] =
		card.Defined{
			ColorSuit:     "error: already removed from deck",
			SequenceIndex: -1,
		}

	gameState.UndrawnDeck = gameState.UndrawnDeck[1:]
}
