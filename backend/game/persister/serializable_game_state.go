package persister

import (
	"fmt"
	"time"

	"github.com/benoleary/ilutulestikud/backend/game"
	"github.com/benoleary/ilutulestikud/backend/game/card"
	"github.com/benoleary/ilutulestikud/backend/game/message"
	"github.com/benoleary/ilutulestikud/backend/player"
)

// PlayerNameAndHand is a struct to contain a player's name and current
// hand, mainly to get around Google Cloud Datastore not allowing lists
// of lists directly.
type PlayerNameAndHand struct {
	Name string
	Hand []card.InHand
}

// SerializableState is a struct meant to encapsulate all the state required
// for a single game to function, in a form which is simple to serialize. It
// implements almost all of the ReadAndWriteState interface, but for the
// Ruleset() function, which should be taken care of by some wrapper around
// an instance of this struct. The discarded cards are simply the cards in
// order in which they were discarded, the played cards are simply the cards
// in the order in which they were played, and player hands are the hands of
// the players in the same order as the players appear in the list of
// participant names in turn order.
type SerializableState struct {
	GameName                        string
	RulesetIdentifier               int
	TimeOfCreation                  time.Time
	ParticipantNamesInTurnOrder     []string
	ParticipantsWhoHaveLeft         []string
	ChatMessageLog                  []message.FromPlayer
	ActionMessageLog                []message.FromPlayer
	TurnNumber                      int
	NumberOfTurnsTakenWithEmptyDeck int
	NumberOfHintsAvailable          int
	NumberOfMistakesMadeSoFar       int
	UndrawnDeck                     []card.Defined
	PlayedCards                     []card.Defined
	DiscardedCards                  []card.Defined
	PlayerHandsInTurnOrder          []PlayerNameAndHand
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
	playerHands := make([]PlayerNameAndHand, numberOfParticipants)
	for playerIndex := 0; playerIndex < numberOfParticipants; playerIndex++ {
		playerName := playersInTurnOrderWithInitialHands[playerIndex].PlayerName
		participantNamesInTurnOrder[playerIndex] = playerName
		playerHands[playerIndex] =
			PlayerNameAndHand{
				Name: playerName,
				Hand: playersInTurnOrderWithInitialHands[playerIndex].InitialHand,
			}
	}

	initialChatLog := make([]message.FromPlayer, chatLogLength)

	for messageIndex := 0; messageIndex < chatLogLength; messageIndex++ {
		initialChatLog[messageIndex] = message.NewFromPlayer("", "", "")
	}

	// We could already set up the capacity for the maps by getting slices from
	// the ruleset and counting, but that is a lot of effort for very little gain.
	return SerializableState{
		GameName:                        gameName,
		RulesetIdentifier:               gameRuleset.BackendIdentifier(),
		TimeOfCreation:                  time.Now(),
		ParticipantNamesInTurnOrder:     participantNamesInTurnOrder,
		ParticipantsWhoHaveLeft:         []string{},
		ChatMessageLog:                  initialChatLog,
		ActionMessageLog:                initialActionLog,
		TurnNumber:                      1,
		NumberOfTurnsTakenWithEmptyDeck: 0,
		NumberOfHintsAvailable:          gameRuleset.MaximumNumberOfHints(),
		NumberOfMistakesMadeSoFar:       0,
		UndrawnDeck:                     shuffledDeck,
		PlayedCards:                     []card.Defined{},
		DiscardedCards:                  []card.Defined{},
		PlayerHandsInTurnOrder:          playerHands,
	}
}

// Name returns the value of the private gameName string.
func (serializableState *SerializableState) Name() string {
	return serializableState.GameName
}

// PlayerNames returns a slice of the private participantNames array.
func (serializableState *SerializableState) PlayerNames() []string {
	return serializableState.ParticipantNamesInTurnOrder
}

// CreationTime returns the value of the private time object describing the time at
// which the state was created.
func (serializableState *SerializableState) CreationTime() time.Time {
	return serializableState.TimeOfCreation
}

// ChatLog returns the chat log of the game at the current moment.
func (serializableState *SerializableState) ChatLog() []message.FromPlayer {
	return serializableState.ChatMessageLog
}

// ActionLog returns the action log of the game at the current moment.
func (serializableState *SerializableState) ActionLog() []message.FromPlayer {
	return serializableState.ActionMessageLog
}

// Turn returns the value of the private turnNumber int.
func (serializableState *SerializableState) Turn() int {
	return serializableState.TurnNumber
}

// TurnsTakenWithEmptyDeck returns the number of turns which have been taken
// since the turn which drew the last card from the deck.
func (serializableState *SerializableState) TurnsTakenWithEmptyDeck() int {
	return serializableState.NumberOfTurnsTakenWithEmptyDeck
}

// NumberOfReadyHints returns the total number of hints which are available to be
// played.
func (serializableState *SerializableState) NumberOfReadyHints() int {
	return serializableState.NumberOfHintsAvailable
}

// NumberOfMistakesMade returns the total number of cards which have been played
// incorrectly.
func (serializableState *SerializableState) NumberOfMistakesMade() int {
	return serializableState.NumberOfMistakesMadeSoFar
}

// DeckSize returns the number of cards left to draw from the deck.
func (serializableState *SerializableState) DeckSize() int {
	return len(serializableState.UndrawnDeck)
}

// RecordChatMessage records a chat message from the given player.
func (serializableState *SerializableState) RecordChatMessage(
	actingPlayer player.ReadonlyState,
	chatMessage string) error {
	appendNewMessageInPlaceDiscardingFirst(
		serializableState.ChatMessageLog,
		actingPlayer.Name(),
		actingPlayer.Color(),
		chatMessage)
	return nil
}

// HasOriginalParticipant returns true if the given player was an original
// participant regardless of who has left the game.
func (serializableState *SerializableState) HasOriginalParticipant(
	playerName string) bool {
	for _, originalParticipant := range serializableState.ParticipantNamesInTurnOrder {
		if originalParticipant == playerName {
			return true
		}
	}

	return false
}

// HasParticipantWhoLeft returns true if the given player has left the game.
func (serializableState *SerializableState) HasParticipantWhoLeft(
	playerName string) bool {
	for _, participantWhoHaveLeft := range serializableState.ParticipantsWhoHaveLeft {
		if participantWhoHaveLeft == playerName {
			return true
		}
	}

	return false
}

// HasCurrentParticipant returns true if the given player was an original
// participant who has not yet left the game.
func (serializableState *SerializableState) HasCurrentParticipant(
	playerName string) bool {
	if !serializableState.HasOriginalParticipant(playerName) {
		return false
	}

	return !serializableState.HasParticipantWhoLeft(playerName)
}

// RemovePlayerFromParticipantList marks the player as no longer being a
// participant of the given game.
func (serializableState *SerializableState) RemovePlayerFromParticipantList(
	playerName string) error {
	if !serializableState.HasOriginalParticipant(playerName) {
		return fmt.Errorf(
			"Player %v is not a participant of game %v",
			playerName,
			serializableState.GameName)
	}

	if serializableState.HasParticipantWhoLeft(playerName) {
		return fmt.Errorf(
			"Player %v has already left game %v",
			playerName,
			serializableState.GameName)
	}

	serializableState.ParticipantsWhoHaveLeft =
		append(serializableState.ParticipantsWhoHaveLeft, playerName)

	return nil
}

func (serializableState *SerializableState) incrementTurnNumbers(
	deckAlreadyEmptyAtStartOfTurn bool) {
	serializableState.TurnNumber++

	if deckAlreadyEmptyAtStartOfTurn {
		serializableState.NumberOfTurnsTakenWithEmptyDeck++
	}
}

func (serializableState *SerializableState) recordActionMessage(
	actingPlayer player.ReadonlyState,
	actionMessage string) {
	appendNewMessageInPlaceDiscardingFirst(
		serializableState.ActionMessageLog,
		actingPlayer.Name(),
		actingPlayer.Color(),
		actionMessage)
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
