package game

import (
	"time"

	"github.com/benoleary/ilutulestikud/backendjson"
	"github.com/benoleary/ilutulestikud/chat"
	"github.com/benoleary/ilutulestikud/player"
)

// State is a struct meant to encapsulate all the state required for a single game to function.
type State struct {
	gameName             string
	creationTime         time.Time
	participatingPlayers []*player.State
	turnNumber           int
	chatLog              *chat.Log
}

// NewState constructs a State object with a non-nil, non-empty slice of player.State objects,
// returning a pointer to the newly-created object.
func NewState(gameName string, playerHandler *player.GetAndPostHandler, playerNames []string) *State {
	numberOfPlayers := len(playerNames)
	playerStates := make([]*player.State, numberOfPlayers)
	for playerIndex := 0; playerIndex < numberOfPlayers; playerIndex++ {
		playerStates[playerIndex], _ = playerHandler.GetPlayerByName(playerNames[playerIndex])
	}

	return &State{
		gameName:             gameName,
		creationTime:         time.Now(),
		participatingPlayers: playerStates,
		turnNumber:           1,
		chatLog:              chat.NewLog(),
	}
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
	state.chatLog.AppendNewMessage(chattingPlayer.Name, chattingPlayer.Color, chatMessage)
}

// ForPlayer creates a PlayerKnowledge object encapsulating the knowledge of the
// given player for the receiver game.
func (state *State) ForPlayer(playerName string) backendjson.PlayerKnowledge {
	return backendjson.PlayerKnowledge{ChatLog: state.chatLog.ForFrontend()}
}
