package endpoint

// Types emitted and accepted by player.GetAndPostHandler:

// PlayerState encapsulates the information from player.State suitable for the front-end.
type PlayerState struct {
	Name  string
	Color string
}
