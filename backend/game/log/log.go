package log

import (
	"sync"
	"time"
)

// Message is a struct to hold the details of a single message.
type Message struct {
	CreationTime time.Time
	PlayerName   string
	TextColor    string
	MessageText  string
}

// RollingAppender holds a fixed number of messages, discarding the
// oldest when appending a new one.
type RollingAppender struct {
	listLength      int
	messageList     []Message
	indexOfOldest   int
	mutualExclusion sync.Mutex
}

// NewRollingAppender makes a new RollingAppender with a fixed number
// of messages.
func NewRollingAppender(listLength int) *RollingAppender {
	return &RollingAppender{
		listLength:    listLength,
		messageList:   make([]Message, listLength),
		indexOfOldest: 0,
	}
}

// SortedCopyOfMessages returns the messages in the log starting with the
// oldest in a simple array, in order by timestamp.
func (rollingAppender *RollingAppender) SortedCopyOfMessages() []Message {
	logLength := rollingAppender.listLength
	sortedMessages := make([]Message, logLength)
	for messageIndex := 0; messageIndex < logLength; messageIndex++ {
		// We take the relevant message indexed with the oldest message at 0,
		// wrapping around if newer messages occupy earlier spots in the
		// actual array.
		adjustedIndex :=
			(messageIndex + rollingAppender.indexOfOldest) % logLength
		sortedMessages[messageIndex] = rollingAppender.messageList[adjustedIndex]
	}

	return sortedMessages
}

// AppendNewMessage adds the given message as the newest message, over-writing
// the oldest message and increasing the offset of the index to the oldest
// message.
func (rollingAppender *RollingAppender) AppendNewMessage(
	playerName string,
	textColor string,
	messageText string) {
	rollingAppender.mutualExclusion.Lock()

	// We over-write the oldest message.
	logMessage := &rollingAppender.messageList[rollingAppender.indexOfOldest]
	logMessage.CreationTime = time.Now()
	logMessage.PlayerName = playerName
	logMessage.TextColor = textColor
	logMessage.MessageText = messageText

	// Now we mark the next-oldest message as the oldest, thus implicitly
	// marking the updated message as the newest message.
	rollingAppender.indexOfOldest =
		(rollingAppender.indexOfOldest + 1) % rollingAppender.listLength

	rollingAppender.mutualExclusion.Unlock()
}
