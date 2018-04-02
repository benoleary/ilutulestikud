package game

// PlayerView encapsulates the functions on a game's read-only state
// which provide the information available to a particular player for
// that state.
type PlayerView struct {
	GameState  ReadonlyState
	PlayerName string
}
