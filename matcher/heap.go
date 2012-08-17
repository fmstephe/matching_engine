package matcher

import ()

type limit struct {
	price int64
	head  *Order
	tail  *Order
}

func newLimit(price int64, o *Order) *limit {
	limit := &limit{price: price, head: o, tail: o}
	return limit
}

func (l *limit) isEmpty() bool {
	return l.head == nil
}

func (l *limit) peek() *Order {
	return l.head
}

func (l *limit) pop() *Order {
	o := l.head
	l.head = o.next
	return o
}

func (l *limit) push(o *Order) {
	l.tail.next = o
	l.tail = o
}

func (l *limit) removeOrder(guid uint64) *Order {
	incoming := &l.head
	for o := l.head; o != nil; o = o.next {
		if o == nil {
			return nil
		}
		if o.GUID() == guid {
			*incoming = o.next
			return o
		}
		incoming = &o.next
	}
	panic("Unreachable")
}

func better(l1, l2 *limit, buySell TradeType) bool {
	if buySell == BUY {
		return l2.price-l1.price < 0
	}
	return l1.price-l2.price < 0
}

type heap struct {
	buySell  TradeType
	priceMap map[int64]*limit // Maps existing limit prices to limits in the heap
	limits   []*limit
}

func newHeap(buySell TradeType) *heap {
	return &heap{buySell: buySell, priceMap: make(map[int64]*limit), limits: make([]*limit, 0, 10)}
}

func (h *heap) heapLen() int {
	return len(h.limits)
}

func (h *heap) push(o *Order) {
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

func (h *heap) pop() *Order {
	h.clearHead()
	if len(h.limits) == 0 {
		return nil
	}
	o := h.limits[0].pop()
	h.clearHead()
	return o
}

func (h *heap) peek() *Order {
	h.clearHead()
	if len(h.limits) == 0 {
		return nil
	}
	return h.limits[0].peek()
}

func (h *heap) clearHead() {
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

func (h *heap) remove(guid uint64, price int64) *Order {
	l := h.priceMap[price]
	o := l.removeOrder(guid)
	return o
}

func (h *heap) up(c int) {
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

func (h *heap) down(p int) {
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
