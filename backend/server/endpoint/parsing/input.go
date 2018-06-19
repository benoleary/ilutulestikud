package parsing

// Types accepted by server.gameEndpointHandler:

// GameDefinition encapsulates the necessary information to create a new game.
type GameDefinition struct {
	GameName          string
	RulesetIdentifier int
	PlayerNames       []string
}

// PlayerChatMessage is a struct to hold a single chat message from a player to a game.
type PlayerChatMessage struct {
	GameName    string
	PlayerName  string
	ChatMessage string
}

// PlayerCardIndication is a struct to hold a single indication of a card in the hand of
// a player, from that player to a game.
type PlayerCardIndication struct {
	GameName   string
	PlayerName string
	CardIndex  int
}
