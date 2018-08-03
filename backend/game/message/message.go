package message

import (
	"time"
)

// FromPlayer encapsulates the read-only state of a single message
// from a player, with the color which should be used to display it
// as defined at the time of the creation of the message.
type FromPlayer struct {
	CreationTime time.Time
	PlayerName   string
	TextColor    string
	MessageText  string
}

// NewFromPlayer returns a new FromPlayer message.
func NewFromPlayer(
	playerName string,
	textColor string,
	messageText string) FromPlayer {
	return FromPlayer{
		CreationTime: time.Now(),
		PlayerName:   playerName,
		TextColor:    textColor,
		MessageText:  messageText,
	}
}
