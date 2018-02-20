package game

import (
	"sync"
	"time"

	"github.com/benoleary/ilutulestikud/player"
)

// ChatLogSize gives the size of the list of the last chat messages.
const ChatLogSize = 20

// ChatMessage is a struct to hold the details of a single chat message.
type ChatMessage struct {
	TimestampInSeconds int64
	PlayerName         string
	ChatColor          string
	MessageText        string
}

// State is a struct meant to encapsulate all the state required for a single game to function.
type State struct {
	gameName             string
	creationTime         time.Time
	participatingPlayers []*player.State
	turnNumber           int
	chatLog              []ChatMessage
	mutualExclusion      sync.Mutex
}

// NewState constructs a State object with a non-nil, non-empty slice of player.State objects,
// returning a pointer to the newly-created object.
func NewState(gameName string, playerHandler *player.Handler, playerNames []string) *State {
	numberOfPlayers := len(playerNames)
	playerStates := make([]*player.State, numberOfPlayers)
	for playerIndex := 0; playerIndex < numberOfPlayers; playerIndex++ {
		playerStates[playerIndex], _ = playerHandler.GetPlayerByName(playerNames[playerIndex])
	}

	return &State{gameName, time.Now(), playerStates, 1, make([]ChatMessage, ChatLogSize), sync.Mutex{}}
}

// HasPlayerAsParticipant returns true if the given player name matches
// the name of any of the game's participating players.
func (state *State) HasPlayerAsParticipant(playerName string) bool {
	for _, participatingPlayer := range state.participatingPlayers {
		if participatingPlayer.Name == playerName {
			return true
		}
	}

	return false
}

// byCreationTime implements sort.Interface for []*State based on the creationTime field.
type byCreationTime []*State

// Len implements part of the sort.Interface for byCreationTime.
func (statePointerArray byCreationTime) Len() int {
	return len(statePointerArray)
}

// Swap implements part of the sort.Interface for byCreationTime.
func (statePointerArray byCreationTime) Swap(firstIndex int, secondIndex int) {
	statePointerArray[firstIndex], statePointerArray[secondIndex] =
		statePointerArray[secondIndex], statePointerArray[firstIndex]
}

// Less implements part of the sort.Interface for byCreationTime.
func (statePointerArray byCreationTime) Less(firstIndex int, secondIndex int) bool {
	return statePointerArray[firstIndex].creationTime.Before(
		statePointerArray[secondIndex].creationTime)
}

// RecordPlayerChatMessage adds the given new message to the end of the chat log
// and removes the oldest message from the top.
func (state *State) RecordPlayerChatMessage(chattingPlayer *player.State, chatMessage string) {
	state.mutualExclusion.Lock()

	// This could probably be more efficient, but is unlikely to be a performance
	// bottleneck...
	for messageIndex := 1; messageIndex < ChatLogSize; messageIndex++ {
		state.chatLog[messageIndex-1] = state.chatLog[messageIndex]
	}

	state.chatLog[ChatLogSize-1] =
		ChatMessage{time.Now().Unix(), chattingPlayer.Name, chattingPlayer.Color, chatMessage}
	state.mutualExclusion.Unlock()
}

// PlayerKnowledge contains the information of what a player can see about a game.
type PlayerKnowledge struct {
	ChatLog []ChatMessage
}

// ForPlayer creates a PlayerKnowledge object encapsulating the knowledge of the
// given player for the receiver game.
func (state *State) ForPlayer(playerName string) PlayerKnowledge {
	return PlayerKnowledge{state.chatLog}
}
