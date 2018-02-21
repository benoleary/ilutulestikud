package backendjson

// Types accepted by game.Handler:

// GameDefinition encapsulates the necessary information to create a new game.
type GameDefinition struct {
	Name    string
	Players []string
}
