package backendjson

// Types emitted and accepted by player.Handler:

// PlayerState encapsulates the information from player.State suitable for the front-end.
type PlayerState struct {
	Name  string
	Color string
}
