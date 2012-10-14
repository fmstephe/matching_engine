package trade

import ()

type Limit struct {
	Price int32
	dummy Order
}

func NewLimit(price int32) *Limit {
	l := &Limit{Price: price}
	l.dummy.Higher = &l.dummy
	l.dummy.Lower = &l.dummy
	return l
}

func (l *Limit) Push(newTail *Order) {
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
func (l *Limit) Pop() *Order {
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

func (l *Limit) Peek() *Order {
	if l.IsEmpty() {
		return nil
	}
	return l.dummy.Lower
}

func (l *Limit) IsEmpty() bool {
	return l.dummy.Higher == &l.dummy
}
