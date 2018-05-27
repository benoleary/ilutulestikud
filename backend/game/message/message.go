package message

import (
	"time"
)

// Readonly encapsulates the read-only state of a single message.
type Readonly struct {
	creationTime time.Time
	playerName   string
	textColor    string
	messageText  string
}

// NewReadonly returns a new Readonly message.
func NewReadonly(
	playerName string,
	textColor string,
	messageText string) Readonly {
	return Readonly{
		creationTime: time.Now(),
		playerName:   playerName,
		textColor:    textColor,
		messageText:  messageText,
	}
}

// CreationTime returns the time when the message was created.
func (readonlyMessage *Readonly) CreationTime() time.Time {
	return readonlyMessage.creationTime
}

// PlayerName returns the name of the player associated with the message.
func (readonlyMessage *Readonly) PlayerName() string {
	return readonlyMessage.playerName
}

// TextColor returns the color to use when displaying the message.
func (readonlyMessage *Readonly) TextColor() string {
	return readonlyMessage.textColor
}

// MessageText returns the text of the message.
func (readonlyMessage *Readonly) MessageText() string {
	return readonlyMessage.messageText
}
