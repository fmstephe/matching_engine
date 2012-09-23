package limitheap

import (
	"github.com/fmstephe/matching_engine/trade"
)

type limit struct {
	price int32
	head  *trade.Order
	tail  *trade.Order
}

func newLimit(price int32, o *trade.Order) *limit {
	limit := &limit{price: price, head: o, tail: o}
	return limit
}

func (l *limit) isEmpty() bool {
	return l.head == nil
}

func (l *limit) peek() *trade.Order {
	return l.head
}

func (l *limit) pop() *trade.Order {
	o := l.head
	l.head = o.Next
	return o
}

func (l *limit) push(o *trade.Order) {
	l.tail.Next = o
	l.tail = o
}

func (l *limit) removeOrder(guid uint64) *trade.Order {
	incoming := &l.head
	for o := l.head; o != nil; o = o.Next {
		if o == nil {
			return nil
		}
		if o.GUID() == guid {
			*incoming = o.Next
			return o
		}
		incoming = &o.Next
	}
	panic("Unreachable")
}

func better(l1, l2 *limit, buySell trade.TradeType) bool {
	if buySell == trade.BUY {
		return l2.price-l1.price < 0
	}
	return l1.price-l2.price < 0
}

type H struct {
	buySell  trade.TradeType
	priceMap map[int32]*limit // Maps existing limit prices to limits in the H
	limits   []*limit
}

func NewHeap(buySell trade.TradeType) *H {
	return &H{buySell: buySell, priceMap: make(map[int32]*limit), limits: make([]*limit, 0, 10)}
}

func (h *H) HLen() int {
	return len(h.limits)
}

func (h *H) Push(o *trade.Order) {
	lim := h.priceMap[o.Price]
	if lim == nil {
		lim = newLimit(o.Price, o)
		h.priceMap[o.Price] = lim
		h.limits = append(h.limits, lim)
		h.up(len(h.limits) - 1)
	} else {
		lim.push(o)
	}
}

func (h *H) Pop() *trade.Order {
	h.clearHead()
	if len(h.limits) == 0 {
		return nil
	}
	o := h.limits[0].pop()
	h.clearHead()
	return o
}

func (h *H) Peek() *trade.Order {
	h.clearHead()
	if len(h.limits) == 0 {
		return nil
	}
	return h.limits[0].peek()
}

func (h *H) clearHead() {
	for len(h.limits) > 0 {
		lim := h.limits[0]
		if !lim.isEmpty() {
			return
		}
		n := len(h.limits) - 1
		h.limits[0] = h.limits[n]
		h.limits[n] = nil
		h.limits = h.limits[0:n]
		h.down(0)
		delete(h.priceMap, lim.price)
	}
}

func (h *H) remove(guid uint64, price int32) *trade.Order {
	l := h.priceMap[price]
	o := l.removeOrder(guid)
	return o
}

func (h *H) up(c int) {
	limits := h.limits
	for {
		p := (c - 1) / 2
		if p == c || better(limits[p], limits[c], h.buySell) {
			break
		}
		limits[p], limits[c] = limits[c], limits[p]
		c = p
	}
}

func (h *H) down(p int) {
	n := len(h.limits)
	limits := h.limits
	for {
		c := 2*p + 1
		if c >= n {
			break
		}
		lc := c
		if rc := lc + 1; rc < n && !better(limits[lc], limits[rc], h.buySell) {
			c = rc
		}
		if better(limits[p], limits[c], h.buySell) {
			break
		}
		limits[p], limits[c] = limits[c], limits[p]
		p = c
	}
}
