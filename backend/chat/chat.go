package chat

import (
	"sync"
	"time"

	"github.com/benoleary/ilutulestikud/backend/endpoint"
)

// logSize gives the size of the list of the last chat messages.
const LogSize = 8

// message is a struct to hold the details of a single chat message.
type message struct {
	CreationTime time.Time
	PlayerName   string
	ChatColor    string
	MessageText  string
}

// Log implements sort.Interface for []*State based on the creationTime field.
type Log struct {
	messageList     []message
	indexOfOldest   int
	mutualExclusion sync.Mutex
}

// NewLog makes a new Log with a fixed number of messages.
func NewLog() *Log {
	return &Log{messageList: make([]message, LogSize), indexOfOldest: 0}
}

// ForFrontend creates a JSON object to represent the Log for the front-end.
func (log *Log) ForFrontend() []endpoint.ChatLogMessage {
	messagesForFrontend := make([]endpoint.ChatLogMessage, LogSize)
	for messageIndex := 0; messageIndex < LogSize; messageIndex++ {
		// We take the relevant message indexed with the oldest message at 0, wrapping
		// around if newer messages occupy earlier spots in the actual array.
		logMessage := log.messageList[(messageIndex+log.indexOfOldest)%LogSize]
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
	log.indexOfOldest = (log.indexOfOldest + 1) % LogSize

	log.mutualExclusion.Unlock()
}
