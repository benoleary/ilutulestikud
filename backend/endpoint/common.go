package endpoint

// Types emitted and accepted by player.GetAndPostHandler:

// PlayerState encapsulates the information from player.State suitable for the frontend.
type PlayerState struct {
	Identifier string
	Name       string
	Color      string
}
