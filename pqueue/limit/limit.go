package limit

import (
	"github.com/fmstephe/matching_engine/trade"
)

type L struct {
	Price int32
	dummy trade.Order
}

func New(price int32) *L {
	l := &L{Price: price}
	l.dummy.Higher = &l.dummy
	l.dummy.Lower = &l.dummy
	return l
}

func (l *L) Push(newTail *trade.Order) {
	dummy := &l.dummy
	tail := dummy.Higher
	// newTail is lower than tail
	tail.Lower = newTail
	newTail.Higher = tail
	// reconnect newTail with dummy
	newTail.Lower = dummy
	dummy.Higher = newTail
}

// This function cannot be called safely on an empty limit
func (l *L) Pop() *trade.Order {
	dummy := &l.dummy
	head := dummy.Lower
	newHead := head.Lower
	// reconnect newHead with dummy
	newHead.Higher = dummy
	dummy.Lower = newHead
	// Clean out head
	head.Higher = nil
	head.Lower = nil
	return head
}

func (l *L) Peek() *trade.Order {
	if l.IsEmpty() {
		return nil
	}
	return l.dummy.Lower
}

func (l *L) IsEmpty() bool {
	return l.dummy.Higher == &l.dummy
}

func Remove(o *trade.Order) {
	o.Higher.Lower = o.Lower
	o.Lower.Higher = o.Higher
}
