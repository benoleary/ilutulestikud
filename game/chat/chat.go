package chat

import (
	"time"

	"github.com/benoleary/ilutulestikud/backendjson"
)

// LogSize gives the size of the list of the last chat messages.
const LogSize = 20

var availableColors = [...]string{
	"pink",
	"red",
	"orange",
	"yellow",
	"green",
	"blue",
	"purple",
	"white"}

var numberOfAvailableColors = len(availableColors)

// Message is a struct to hold the details of a single chat message.
type Message struct {
	CreationTime time.Time
	PlayerName   string
	ChatColor    string
	MessageText  string
}

// ForFrontend creates a JSON object to represent the Message for the front-end.
func (message *Message) ForFrontend() backendjson.ChatLogMessage {
	return backendjson.ChatLogMessage{
		TimestampInSeconds: message.CreationTime.Unix(),
		PlayerName:         message.PlayerName,
		ChatColor:          message.ChatColor,
		MessageText:        message.MessageText}
}

// DefaultColor provides a default color for a given index, cycling round
// to the first color again if there are not enough.
func DefaultColor(playerIndex int) string {
	return availableColors[playerIndex%numberOfAvailableColors]
}

// AvailableColors returns a copy of the list of colors available for chat messages.
func AvailableColors() []string {
	return availableColors[:]
}

// Log implements sort.Interface for []*State based on the creationTime field.
type Log struct {
	Messages []Message
}

// NewLog makes a new Log with a fixed number of messages.
func NewLog() *Log {
	return &Log{Messages: make([]Message, LogSize)}
}

// ForFrontend creates a JSON object to represent the Log for the front-end.
func (log *Log) ForFrontend() []backendjson.ChatLogMessage {
	messageList := make([]backendjson.ChatLogMessage, LogSize)
	for messageIndex := 0; messageIndex < LogSize; messageIndex++ {
		messageList[messageIndex] = log.Messages[messageIndex].ForFrontend()
	}

	return messageList
}

// Append makes a new Log with all the messages shifted one back (discarding the
// oldest), with the given message as the newest.
func (log *Log) Append(message Message) *Log {
	messageList := make([]Message, LogSize)
	// This could probably be more efficient, but is unlikely to be a performance
	// bottleneck...
	for messageIndex := 1; messageIndex < LogSize; messageIndex++ {
		messageList[messageIndex-1] = log.Messages[messageIndex]
	}

	messageList[LogSize-1] = message

	return &Log{Messages: messageList}
}
