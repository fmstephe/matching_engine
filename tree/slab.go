package tree

import ()

type Slab struct {
	free   *Order
	orders []Order
}

func NewSlab(size int) *Slab {
	s := &Slab{orders: make([]Order, size)}
	s.free = &s.orders[0]
	prev := s.free
	for i := 1; i < len(s.orders); i++ {
		curr := &s.orders[i]
		prev.nextFree = curr
		prev = curr
	}
	return s
}

func (s *Slab) Malloc() *Order {
	o := s.free
	if o == nil {
		o = &Order{}
	}
	s.free = o.nextFree
	o.nextFree = o // Slab allocated order marker
	return o
}

func (s *Slab) Free(o *Order) {
	if o.nextFree == o {
		o.nextFree = s.free
		s.free = o
	}
	// Orders that were not slab allocated are left to the garbage collector
}
