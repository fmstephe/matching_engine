package pqueue

import (
	"github.com/fmstephe/matching_engine/trade"
)

type Q interface {
	Push(*trade.Order)
	Pop() *trade.Order
	Peek() *trade.Order
	Size() int
	BuySell() trade.TradeType
	//Remove(guid int64) *trade.Order
}
