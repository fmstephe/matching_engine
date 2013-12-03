package client

import (
	"github.com/fmstephe/matching_engine/msg"
)

type receivedMessage struct {
	Accepted bool        `json:"accepted"`
	Message  msg.Message `json:"message"`
}

type response struct {
	State    traderState     `json:"state"`
	Received receivedMessage `json:"received"`
	Comment  string          `json:"comment"`
}

type traderState struct {
	Balance     balanceManager `json:"balance"`
	Stocks      stockManager   `json:"stocks"`
	Outstanding []msg.Message  `json:"outstanding"`
}
