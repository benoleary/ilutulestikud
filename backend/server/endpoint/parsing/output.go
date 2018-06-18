package parsing

// Types emitted by server.State:

// ErrorForBody allows a string to be expressed as JSON stating that it is an error.
type ErrorForBody struct {
	Error string
}

// Types emitted by server.playerEndpointHandler:

// PlayerList ensures that the PlayerState list is encapsulated within a single JSON object.
type PlayerList struct {
	Players []PlayerState
}

// ChatColorList ensures that the list of available chat colors is encapsulated within a single JSON object.
type ChatColorList struct {
	Colors []string
}

// Types emitted by server.gameEndpointHandler:

// SelectableRuleset contains the information required to enable a player to select a ruleset,
// plus the pertinent information from the ruleset to allow the frontend to form a valid request
// to create a new game with the ruleset.
type SelectableRuleset struct {
	Identifier  int
	Description string

	// This saves the dialog having to find out how many players can play when the ruleset is chosen.
	MinimumNumberOfPlayers int
	MaximumNumberOfPlayers int
}

// RulesetList lists the rulesets which are available for the creation of a game.
type RulesetList struct {
	Rulesets []SelectableRuleset
}

// TurnSummary contains the information to determine what games involve a player and whose turn it is.
// All the fields need to be public so that the JSON encoder can see them to serialize them.
type TurnSummary struct {
	GameIdentifier string
	GameName       string
	IsPlayerTurn   bool
}

// TurnSummaryList ensures that the TurnSummary list is encapsulated within a single JSON object.
type TurnSummaryList struct {
	TurnSummaries []TurnSummary
}

// LogMessage is a struct to hold the details of a single outgoing log message.
type LogMessage struct {
	TimestampInSeconds int64
	PlayerName         string
	TextColor          string
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
	PlayerName  string
	PlayerColor string
	HandCards   []VisibleCard
}

// CardFromBehind is a struct to hold the details of a single outgoing card as known
// to the player who is holding the card.
type CardFromBehind struct {
	PossibleColorSuits      []string
	PossibleSequenceIndices []int
}

// GameView contains the information of what a player can see about a game.
// The hands are arrranged into three groups:
// 1) those for the players whose next turn is before this player's next
//    turn, in order;
// 2) that of this player (as known to this player);
// 3) those for the players whose next turn is after this player's next
//    turn, in order.
// The lists for before and after may be empty, if this player is the first
// or last in order at the moment, respectively.
type GameView struct {
	ChatLog                            []LogMessage
	ActionLog                          []LogMessage
	ScoreSoFar                         int
	NumberOfReadyHints                 int
	MaximumNumberOfHints               int
	NumberOfMistakesMade               int
	NumberOfMistakesIndicatingGameOver int
	NumberOfCardsLeftInDeck            int
	PlayedCards                        [][]VisibleCard
	DiscardedCards                     []VisibleCard
	HandsBeforeThisPlayer              []VisibleHand
	HandOfThisPlayer                   []CardFromBehind
	HandsAfterThisPlayer               []VisibleHand
}
