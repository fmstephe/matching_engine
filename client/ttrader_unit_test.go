package client

import (
	//	"testing"
	"github.com/fmstephe/matching_engine/msg"
)

func setupTrader(traderId uint32) *trader {
	intoSvr := make(chan *msg.Message)
	outOfSvr := make(chan *msg.Message)
	tdr, _ := newTrader(traderId, intoSvr, outOfSvr)
	return tdr
}
