package endpoint

// Types emitted by player.GetAndPostHandler:

// PlayerList ensures that the PlayerState list is encapsulated within a single JSON object.
type PlayerList struct {
	Players []PlayerState
}

// ChatColorList ensures that the list of available chat colors is encapsulated within a single JSON object.
type ChatColorList struct {
	Colors []string
}

// Types emitted by game.GetAndPostHandler:

// TurnSummary contains the information to determine what games involve a player and whose turn it is.
// All the fields need to be public so that the JSON encoder can see them to serialize them.
// The creation timestamp is int64 because that is what time.Unix() returns.
type TurnSummary struct {
	GameIdentifier             string
	GameName                   string
	CreationTimestampInSeconds int64
	TurnNumber                 int
	PlayersInNextTurnOrder     []string
	IsPlayerTurn               bool
}

// TurnSummaryList ensures that the TurnSummary list is encapsulated within a single JSON object.
type TurnSummaryList struct {
	TurnSummaries []TurnSummary
}

// ChatLogMessage is a struct to hold the details of a single outgoing chat message.
type ChatLogMessage struct {
	TimestampInSeconds int64
	PlayerName         string
	ChatColor          string
	MessageText        string
}

// GameView contains the information of what a player can see about a game.
type GameView struct {
	ChatLog []ChatLogMessage
}
