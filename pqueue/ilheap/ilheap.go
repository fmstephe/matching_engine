package ilheap

import (
	"github.com/fmstephe/matching_engine/trade"
	"math"
)

type elem struct {
	order *trade.Order
	val   int64
}

func (e *elem) zero() {
	e.order = nil
	e.val = 0
}

type H struct {
	kind   trade.OrderKind
	seq    int32
	seqInc int32
	idx    int
	elems  []elem
}

func New(kind trade.OrderKind, initCapacity int) *H {
	var seq int32
	var seqInc int32
	if kind == trade.BUY {
		seq = math.MaxInt32
		seqInc = -1
	} else {
		seq = 0
		seqInc = 1
	}
	return &H{kind: kind, seq: seq, seqInc: seqInc, idx: -1, elems: make([]elem, initCapacity)}
}

func (h *H) Size() int {
	return h.idx + 1
}

func (h *H) Push(o *trade.Order) {
	h.idx++
	h.seq += h.seqInc
	e := &h.elems[h.idx]
	e.order = o
	e.val = int64(uint64(o.Price)<<32|uint64(h.seq)) * int64(o.Kind)
	h.up(h.idx)
}

func (h *H) Pop() *trade.Order {
	if h.idx >= 0 {
		o := h.elems[0].order
		h.elems[0] = h.elems[h.idx]
		h.idx--
		h.down(0)
		return o
	}
	return nil
}

func (h *H) Kind() trade.OrderKind {
	return h.kind
}

func (h *H) Peek() *trade.Order {
	if h.idx >= 0 {
		return h.elems[0].order
	}
	return nil
}

func (h *H) Remove(o *trade.Order) {
	panic("Unsupported")
}

func (h *H) up(c int) {
	elems := h.elems
	for {
		p := (c - 1) / 2
		if p == c || elems[p].val > elems[c].val {
			break
		}
		elems[p], elems[c] = elems[c], elems[p]
		c = p
	}
}

func (h *H) down(p int) {
	n := h.idx
	elems := h.elems
	for {
		c := 2*p + 1
		if c >= n {
			break
		}
		lc := c
		if rc := lc + 1; rc < n && elems[lc].val <= elems[rc].val {
			c = rc
		}
		if elems[p].val > elems[c].val {
			break
		}
		elems[p], elems[c] = elems[c], elems[p]
		p = c
	}
}
