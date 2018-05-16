package parsing

// Types emitted and accepted by server.playerEndpointHandler:

// PlayerState encapsulates the information from player.ReadonlyState suitable
// for the frontend, including the identifier to ensure that it can be passed
// as a URI segment when making a GET request.
type PlayerState struct {
	Identifier string
	Name       string
	Color      string
}
