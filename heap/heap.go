package heap

import (
	"github.com/fmstephe/matching_engine/trade"
)

type Interface interface {
	Len() int
	Push(*trade.Order)
	Pop() *trade.Order
	Peek() *trade.Order
	Remove(uint64) *trade.Order
}

type limit struct {
	price int64 
	index int
	head *trade.Order
	tail *trade.Order
}

func (l *limit) appendOrder(o *trade.Order) {
	tail := l.tail
	tail.Next = o
	l.tail = o
}

func better(l1, l2 *limit, buySell trade.TradeType) bool {
	if buySell == trade.BUY {
		return l2.price - l1.price < 0
	}
	return l1.price - l2.price < 0
}

type limitHeap struct {
	buySell trade.TradeType
	priceMap map[int64] *limit // Maps existing limit prices to limits in the heap
	limits []*limit
}

func New(buySell trade.TradeType) *limitHeap {
	return &limitHeap{buySell: buySell, priceMap: make(map[int64] *limit), limits: make([]*limit, 0, 10)}
}

func (lh *limitHeap) Len() int {
	return len(lh.limits)
}

func (lh *limitHeap) Push(o *trade.Order) {
	lim := lh.priceMap[o.Price]
	if lim == nil {
		lim = &limit{price: o.Price, head: o, tail: o}
		lh.priceMap[o.Price] = lim
		lh.limits = append(lh.limits, lim)
		lh.up(len(lh.limits)-1)
	} else {
		lim.appendOrder(o)
	}
}

func (lh *limitHeap) Pop() *trade.Order {
	if len(lh.limits) == 0 {
		return nil
	}
	lim := lh.limits[0]
	o := lim.head
	lim.head = o.Next
	if (lim.head == nil) {
		n := len(lh.limits)-1
		lh.limits[0] = lh.limits[n]
		lh.limits[n] = nil
		lh.limits = lh.limits[0:n]
		lh.down(0)
		delete(lh.priceMap, o.Price)
	}
	return o
}

func (lh *limitHeap) Peek() *trade.Order {
	if len(lh.limits) == 0 {
		return nil
	}
	return lh.limits[0].head
}

func (lh *limitHeap) Remove(guid uint64) *trade.Order {
	panic("Remove not supported")
}

func (lh *limitHeap) up(j int) {
	limits := lh.limits
	for {
		i := (j - 1) / 2 // parent
		if i == j || better(limits[i], limits[j], lh.buySell) {
			break
		}
		limits[i], limits[j] = limits[j], limits[i]
		j = i
	}
}

func (lh *limitHeap) down(i int) {
	limits := lh.limits
	n := len(limits)
	for {
		j1 := 2*i + 1
		if j1 >= n {
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && !better(limits[j1], limits[j2], lh.buySell) {
			j = j2 // = 2*i + 2  // right child
		}
		if better(limits[i], limits[j], lh.buySell) {
			break
		}
		limits[i], limits[j] = limits[j], limits[i]
		i = j
	}
}
