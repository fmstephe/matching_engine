package matcher

import (
)

type limit struct {
	price int64
	index int
	head *Order
	tail *Order
}

func (l *limit) appendOrder(o *Order) {
	tail := l.tail
	tail.Next = o
	l.tail = o
}

func better(l1, l2 *limit, buySell TradeType) bool {
	if buySell == BUY {
		return l2.price - l1.price < 0
	}
	return l1.price - l2.price < 0
}

type heap struct {
	buySell TradeType
	priceMap map[int64] *limit // Maps existing limit prices to limits in the heap
	limits []*limit
}

func newHeap(buySell TradeType) *heap {
	return &heap{buySell: buySell, priceMap: make(map[int64] *limit), limits: make([]*limit, 0, 10)}
}

func (h *heap) Len() int {
	return len(h.limits)
}

func (h *heap) Push(o *Order) {
	lim := h.priceMap[o.Price]
	if lim == nil {
		lim = &limit{price: o.Price, head: o, tail: o}
		h.priceMap[o.Price] = lim
		h.limits = append(h.limits, lim)
		h.up(len(h.limits)-1)
	} else {
		lim.appendOrder(o)
	}
}

func (h *heap) Pop() *Order {
	if len(h.limits) == 0 {
		return nil
	}
	lim := h.limits[0]
	o := lim.head
	lim.head = o.Next
	if (lim.head == nil) {
		n := len(h.limits)-1
		h.limits[0] = h.limits[n]
		h.limits[n] = nil
		h.limits = h.limits[0:n]
		h.down(0)
		delete(h.priceMap, o.Price)
	}
	return o
}

func (h *heap) Peek() *Order {
	if len(h.limits) == 0 {
		return nil
	}
	return h.limits[0].head
}

func (h *heap) Remove(guid uint64) *Order {
	panic("Remove not supported")
}

func (h *heap) up(j int) {
	limits := h.limits
	for {
		i := (j - 1) / 2 // parent
		if i == j || better(limits[i], limits[j], h.buySell) {
			break
		}
		limits[i], limits[j] = limits[j], limits[i]
		j = i
	}
}

func (h *heap) down(i int) {
	limits := h.limits
	n := len(limits)
	for {
		j1 := 2*i + 1
		if j1 >= n {
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && !better(limits[j1], limits[j2], h.buySell) {
			j = j2 // = 2*i + 2  // right child
		}
		if better(limits[i], limits[j], h.buySell) {
			break
		}
		limits[i], limits[j] = limits[j], limits[i]
		i = j
	}
}
