package limitheap

import (
	"github.com/fmstephe/matching_engine/trade"
)

type limit struct {
	price int32
	head  *trade.Order
	tail  *trade.Order
	next  *limit
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

func better(l1, l2 *limit, buySell trade.TradeType) bool {
	if buySell == trade.BUY {
		return l2.price-l1.price < 0
	}
	return l1.price-l2.price < 0
}

type H struct {
	buySell  trade.TradeType
	limits *limitset
	heap   []*limit
	size     int
}

func New(buySell trade.TradeType, mapSize int32, heapSize int) *H {
	return &H{buySell: buySell, limits: newLimitSet(mapSize), heap: make([]*limit, 0, heapSize)}
}

func (h *H) Size() int {
	return h.size
}

func (h *H) Push(o *trade.Order) {
	//	h.orderSet.Put(o.Guid, o)
	lim := h.limits.Get(o.Price) // Shit this is a very slow operation
	if lim == nil {
		lim = newLimit(o.Price, o)
		h.limits.Put(o.Price, lim)
		h.heap = append(h.heap, lim)
		h.up(len(h.heap) - 1)
	} else {
		lim.push(o)
	}
	h.size++
}

func (h *H) Pop() *trade.Order {
	h.clearHead()
	if len(h.heap) == 0 {
		return nil
	}
	o := h.heap[0].pop()
	h.clearHead()
	//h.orderSet.Remove(o.Guid)
	h.size--
	return o
}

func (h *H) Peek() *trade.Order {
	h.clearHead()
	if len(h.heap) == 0 {
		return nil
	}
	return h.heap[0].peek()
}

func (h *H) clearHead() {
	for len(h.heap) > 0 {
		lim := h.heap[0]
		if !lim.isEmpty() {
			return
		}
		n := len(h.heap) - 1
		h.heap[0] = h.heap[n]
		h.heap[n] = nil
		h.heap = h.heap[0:n]
		h.down(0)
		h.limits.Remove(lim.price)
	}
}

func (h *H) Remove(guid int64) *trade.Order {
	//return h.orderSet.Remove(guid)
	return nil
}

func (h *H) BuySell() trade.TradeType {
	return h.buySell
}

func (h *H) up(c int) {
	heap := h.heap
	for {
		p := (c - 1) / 2
		if p == c || better(heap[p], heap[c], h.buySell) {
			break
		}
		heap[p], heap[c] = heap[c], heap[p]
		c = p
	}
}

func (h *H) down(p int) {
	n := len(h.heap)
	heap := h.heap
	for {
		c := 2*p + 1
		if c >= n {
			break
		}
		lc := c
		if rc := lc + 1; rc < n && !better(heap[lc], heap[rc], h.buySell) {
			c = rc
		}
		if better(heap[p], heap[c], h.buySell) {
			break
		}
		heap[p], heap[c] = heap[c], heap[p]
		p = c
	}
}
