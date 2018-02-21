package backendjson

import (
	"github.com/benoleary/ilutulestikud/game/chat"
)

// Types emitted by player.Handler:

// PlayerStateList ensures that the PlayerState list is encapsulated within a single JSON object.
type PlayerStateList struct {
	Players []PlayerState
}

// Types emitted by game.Handler:

// TurnSummary contains the information to determine what games involve a player and whose turn it is.
// All the fields need to be public so that the JSON encoder can see them to serialize them.
// The creation timestamp is int64 because that is what time.Unix() returns.
type TurnSummary struct {
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

// PlayerKnowledge contains the information of what a player can see about a game.
type PlayerKnowledge struct {
	ChatLog []chat.Message
}
