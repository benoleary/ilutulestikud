package endpoint

// Types accepted by server.gameEndpointHandler:

// GameDefinition encapsulates the necessary information to create a new game.
type GameDefinition struct {
	GameName          string
	RulesetIdentifier int
	PlayerIdentifiers []string
}

// PlayerChatMessage is a struct to hold a single chat message from a player to a game.
type PlayerChatMessage struct {
	GameName    string
	PlayerName  string
	ChatMessage string
}

// PlayerAction is a struct to hold the details of an action performed by a player in a game.
type PlayerAction struct {
	PlayerIdentifier string
	GameIdentifier   string
	ActionType       string
	ChatMessage      string
	CardIndex        int
	HintRecipient    string
	HintNumber       int
	HintColor        string
}
