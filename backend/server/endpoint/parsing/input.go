package parsing

// Types accepted by server.gameEndpointHandler:

// GameDefinition encapsulates the necessary information to create a new game.
type GameDefinition struct {
	GameName          string
	RulesetIdentifier int
	PlayerNames       []string
}

// PlayerInGameIndication is a struct to identify a player and a game together.
type PlayerInGameIndication struct {
	GameName   string
	PlayerName string
}

// PlayerChatMessage is a struct to hold a single chat message from a player to a game.
type PlayerChatMessage struct {
	PlayerInGameIndication
	ChatMessage string
}

// PlayerCardIndication is a struct to hold a single indication of a card in the hand of
// a player, from that player to a game.
type PlayerCardIndication struct {
	PlayerInGameIndication
	CardIndex int
}
