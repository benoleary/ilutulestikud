package endpoint

// Types accepted by game.GetAndPostHandler:

// GameDefinition encapsulates the necessary information to create a new game.
type GameDefinition struct {
	GameName          string
	RulesetIdentifier int
	PlayerIdentifiers []string

	// These are the same as GameName and PlayerIdentifiers respectively, and
	// are here for backwards compatibility with the frontend until it gets updated.
	Name    string
	Players []string
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

	// These are the same as PlayerIdentifier, GameIdentifier, and ActionType respectively, and
	// are here for backwards compatibility with the frontend until it gets updated.
	Player string
	Game   string
	Action string
}
