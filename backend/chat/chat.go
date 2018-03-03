package chat

import (
	"sync"
	"time"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// logSize gives the size of the list of the last chat messages.
const logSize = 8

var availableColors = [...]string{
	"pink",
	"red",
	"orange",
	"yellow",
	"green",
	"blue",
	"purple",
	"white",
}

var numberOfAvailableColors = len(availableColors)

// message is a struct to hold the details of a single chat message.
type message struct {
	CreationTime time.Time
	PlayerName   string
	ChatColor    string
	MessageText  string
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
	messageList     []message
	indexOfOldest   int
	mutualExclusion sync.Mutex
}

// NewLog makes a new Log with a fixed number of messages.
func NewLog() *Log {
	return &Log{messageList: make([]message, logSize), indexOfOldest: 0}
}

// ForFrontend creates a JSON object to represent the Log for the front-end.
func (log *Log) ForFrontend() []endpoint.ChatLogMessage {
	messagesForFrontend := make([]endpoint.ChatLogMessage, logSize)
	for messageIndex := 0; messageIndex < logSize; messageIndex++ {
		// We take the relevant message indexed with the oldest message at 0, wrapping
		// around if newer messages occupy earlier spots in the actual array.
		logMessage := log.messageList[(messageIndex+log.indexOfOldest)%logSize]
		messageForFrontend := &messagesForFrontend[messageIndex]
		messageForFrontend.TimestampInSeconds = logMessage.CreationTime.Unix()
		messageForFrontend.PlayerName = logMessage.PlayerName
		messageForFrontend.ChatColor = logMessage.ChatColor
		messageForFrontend.MessageText = logMessage.MessageText
	}

	return messagesForFrontend
}

// AppendNewMessage adds the given message as the newest message, over-writing
// the oldest message and increasing the offset of the index to the oldest
// message.
func (log *Log) AppendNewMessage(playerName string, chatColor string, messageText string) {
	log.mutualExclusion.Lock()

	// We over-write the oldest message.
	logMessage := &log.messageList[log.indexOfOldest]
	logMessage.CreationTime = time.Now()
	logMessage.PlayerName = playerName
	logMessage.ChatColor = chatColor
	logMessage.MessageText = messageText

	// Now we mark the next-oldest message as the oldest, thus implicitly
	// marking the updated message as the newest message.
	log.indexOfOldest = (log.indexOfOldest + 1) % logSize

	log.mutualExclusion.Unlock()
}
