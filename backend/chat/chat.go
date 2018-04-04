package chat

import (
	"sync"
	"time"
)

// LogSize gives the size of the list of the last chat messages.
const LogSize = 8

// Message is a struct to hold the details of a single chat message.
type Message struct {
	CreationTime time.Time
	PlayerName   string
	ChatColor    string
	MessageText  string
}

// Log implements sort.Interface for []*State based on the creationTime field.
type Log struct {
	messageList     []Message
	indexOfOldest   int
	mutualExclusion sync.Mutex
}

// NewLog makes a new Log with a fixed number of messages.
func NewLog() *Log {
	return &Log{messageList: make([]Message, LogSize), indexOfOldest: 0}
}

// Sorted returns the messages in the log starting with the oldest in a simple
// array, in order by timestamp.
func (log *Log) Sorted() []Message {
	sortedMessages := make([]Message, LogSize)
	for messageIndex := 0; messageIndex < LogSize; messageIndex++ {
		// We take the relevant message indexed with the oldest message at 0, wrapping
		// around if newer messages occupy earlier spots in the actual array.
		sortedMessages[messageIndex] = log.messageList[(messageIndex+log.indexOfOldest)%LogSize]
	}

	return sortedMessages
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
