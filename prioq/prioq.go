package prioq

import (
	"github.com/fmstephe/matching_engine/trade"
)

type Q interface {
	Push(*trade.Order)
	Pop() *trade.Order
	Peek() *trade.Order
	Size() int
	Kind() trade.OrderKind
	Remove(*trade.Order)
}
