package client

import (
	"github.com/fmstephe/matching_engine/msg"
)

type receivedMessage struct {
	Accepted bool        `json:"accepted"`
	Message  msg.Message `json:"message"`
}

type Response struct {
	State    traderState     `json:"state"`
	Received receivedMessage `json:"received"`
	Comment  string          `json:"comment"`
}

type traderState struct {
	CurrentBalance   uint64            `json:"currentBalance"`
	AvailableBalance uint64            `json:"availableBalance"`
	StocksHeld       map[string]uint64 `json:"stocksHeld"`
	StocksToSell     map[string]uint64 `json:"stocksToSell"`
	Outstanding      []msg.Message     `json:"outstanding"`
}
