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

// PlayerHintToReceiver is a struct to hold a single hint from a (hinting) player to a
// receiving player.
type PlayerHintToReceiver struct {
	PlayerInGameIndication
	ReceiverName string
}

// PlayerColorHint is a struct to hold a single hint, from a player in a game to another
// player, about a color suit with respect to the receiver's hand.
type PlayerColorHint struct {
	PlayerHintToReceiver
	HintedColor string
}

// PlayerIndexHint is a struct to hold a single hint, from a player in a game to another
// player, about a sequence index with respect to the receiver's hand.
type PlayerIndexHint struct {
	PlayerHintToReceiver
	HintedNumber int
}
