package limitheap

import (
	"github.com/fmstephe/matching_engine/trade"
)

func better(l1, l2 *trade.Limit, kind trade.OrderKind) bool {
	if kind == trade.BUY {
		return l2.Price-l1.Price < 0
	}
	return l1.Price-l2.Price < 0
}

type H struct {
	kind   trade.OrderKind
	limits *trade.LimitSet
	heap   []*trade.Limit
	size   int
}

func New(kind trade.OrderKind, limitSetSize int32, heapSize int) *H {
	return &H{kind: kind, limits: trade.NewLimitSet(limitSetSize), heap: make([]*trade.Limit, 0, heapSize)}
}

func (h *H) Survey() []*trade.SurveyLimit {
	survey := make([]*trade.SurveyLimit, 0, len(h.heap))
	heap := h.heap
	for i := range heap {
		if !heap[i].IsEmpty() {
			survey = append(survey, heap[i].Survey())
		}
	}
	return survey
}

func (h *H) Size() int {
	return h.size
}

func (h *H) Push(o *trade.Order) {
	lim := h.limits.Get(o.Price)
	if lim == nil {
		lim = trade.NewLimit(o.Price)
		lim.Push(o)
		h.limits.Put(o.Price, lim)
		h.heap = append(h.heap, lim)
		h.up(len(h.heap) - 1)
	} else {
		lim.Push(o)
	}
	h.size++
}

func (h *H) Pop() *trade.Order {
	h.clearHead()
	if len(h.heap) == 0 {
		return nil
	}
	o := h.heap[0].Pop()
	h.clearHead()
	h.size--
	return o
}

func (h *H) Peek() *trade.Order {
	h.clearHead()
	if len(h.heap) == 0 {
		return nil
	}
	return h.heap[0].Peek()
}

func (h *H) clearHead() {
	for len(h.heap) > 0 {
		lim := h.heap[0]
		if !lim.IsEmpty() {
			return
		}
		n := len(h.heap) - 1
		h.heap[0] = h.heap[n]
		h.heap[n] = nil
		h.heap = h.heap[0:n]
		h.down(0)
		h.limits.Remove(lim.Price)
	}
}

func (h *H) Remove(o *trade.Order) {
	// No op
}

func (h *H) Kind() trade.OrderKind {
	return h.kind
}

func (h *H) up(c int) {
	heap := h.heap
	for {
		p := (c - 1) / 2
		if p == c || better(heap[p], heap[c], h.kind) {
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
		if rc := lc + 1; rc < n && !better(heap[lc], heap[rc], h.kind) {
			c = rc
		}
		if better(heap[p], heap[c], h.kind) {
			break
		}
		heap[p], heap[c] = heap[c], heap[p]
		p = c
	}
}
