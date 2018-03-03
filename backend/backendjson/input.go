package backendjson

// Types accepted by game.Handler:

// GameDefinition encapsulates the necessary information to create a new game.
type GameDefinition struct {
	Name    string
	Players []string
}

// PlayerChatMessage is a struct to hold the details of a single incoming chat message.
type PlayerChatMessage struct {
	Player  string
	Game    string
	Message string
}
