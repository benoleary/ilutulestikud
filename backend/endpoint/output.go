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

// SelectableRuleset contains the information required to enable a player to select a ruleset,
// plus the pertinent information from the ruleset to allow the frontend to form a valid request
// to create a new game with the ruleset.
type SelectableRuleset struct {
	Identifier             int
	Description            string
	MinimumNumberOfPlayers int
	MaximumNumberOfPlayers int
}

// RulesetList lists the rulesets which are available for the creation of a game.
type RulesetList struct {
	Rulesets []SelectableRuleset
}

// TurnSummary contains the information to determine what games involve a player and whose turn it is.
// All the fields need to be public so that the JSON encoder can see them to serialize them.
// The creation timestamp is int64 because that is what time.Unix() returns.
type TurnSummary struct {
	GameIdentifier             string
	GameName                   string
	CreationTimestampInSeconds int64
	TurnNumber                 int
	PlayerNamesInNextTurnOrder []string
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

// VisibleCard is a struct to hold the details of a single outgoing card when visible
// to a player.
type VisibleCard struct {
	ColorSuit     string
	SequenceIndex int
}

// VisibleHand is a struct to hold the details of the hand of cards held by a player
// other than the player who is viewing the game state.
type VisibleHand struct {
	PlayerIdentifier string
	PlayerName       string
	HandCards        []VisibleCard
}

// CardFromBehind is a struct to hold the details of a single outgoing card as known
// to the player who is holding the card.
type CardFromBehind struct {
	AllowedColorSuits       []string
	ExcludedColorSuits      []string
	AllowedSequenceIndices  []int
	ExcludedSequenceIndices []int
}

// GameView contains the information of what a player can see about a game.
type GameView struct {
	ChatLog                      []ChatLogMessage
	ScoreSoFar                   int
	NumberOfReadyHints           int
	NumberOfSpentHints           int
	NumberOfMistakesStillAllowed int
	NumberOfMistakesMade         int
	NumberOfCardsLeftInDeck      int
	PlayedCards                  [][]VisibleCard
	DiscardedCards               [][]VisibleCard
	ThisPlayerHand               []CardFromBehind
	OtherPlayerHands             []VisibleHand
}
