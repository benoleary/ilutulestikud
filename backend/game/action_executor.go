package game

// ActionExecutor encapsulates the write functions on a game's state
// which update the state based on player actions.
type ActionExecutor struct {
	GameState  readAndWriteState
	PlayerName string
}
