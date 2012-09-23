package binheap

import (
	"github.com/fmstephe/matching_engine/trade"
	"math"
)

func better(o1, o2 *trade.Order) bool {
	return o1.Compare > o2.Compare
}

type heap struct {
	buySell trade.TradeType
	seq     int32
	seqInc  int32
	orders  []*trade.Order
}

func NewHeap(buySell trade.TradeType, initCapacity int) *heap {
	var seq int32
	var seqInc int32
	if buySell == trade.BUY {
		seq = math.MaxInt32
		seqInc = -1
	} else {
		seq = 0
		seqInc = 1
	}
	return &heap{buySell: buySell, seq: seq, seqInc: seqInc, orders: make([]*trade.Order, 0, initCapacity)}
}

func (h *heap) heapLen() int {
	return len(h.orders)
}

func (h *heap) push(o *trade.Order) {
	o.Compare = int64(uint64(o.Price)<<32|uint64(h.seq)) * int64(o.BuySell)
	h.seq += h.seqInc
	h.orders = append(h.orders, o)
	h.up(len(h.orders) - 1)
}

func (h *heap) pop() *trade.Order {
	if len(h.orders) == 0 {
		return nil
	}
	o := h.orders[0]
	h.orders[0] = h.orders[len(h.orders)-1]
	h.orders[len(h.orders)-1] = nil
	h.orders = h.orders[:len(h.orders)-1]
	h.down(0)
	return o
}

func (h *heap) peek() *trade.Order {
	if len(h.orders) == 0 {
		return nil
	}
	return h.orders[0]
}

func (h *heap) up(c int) {
	orders := h.orders
	for {
		p := (c - 1) / 2
		if p == c || better(orders[p], orders[c]) {
			break
		}
		orders[p], orders[c] = orders[c], orders[p]
		c = p
	}
}

func (h *heap) down(p int) {
	n := len(h.orders)
	orders := h.orders
	for {
		c := 2*p + 1
		if c >= n {
			break
		}
		lc := c
		if rc := lc + 1; rc < n && !better(orders[lc], orders[rc]) {
			c = rc
		}
		if better(orders[p], orders[c]) {
			break
		}
		orders[p], orders[c] = orders[c], orders[p]
		p = c
	}
}
