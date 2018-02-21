package chat

// LogSize gives the size of the list of the last chat messages.
const LogSize = 20

// Message is a struct to hold the details of a single chat message.
type Message struct {
	TimestampInSeconds int64
	PlayerName         string
	ChatColor          string
	MessageText        string
}
