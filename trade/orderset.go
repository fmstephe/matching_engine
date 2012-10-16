package trade

import ()

const (
	tombstone64 = int64(-1)
)

func deadOrderElems(size int32) []orderEntry {
	entries := make([]orderEntry, size)
	for i := int32(0); i < size; i++ {
		entries[i].key = tombstone64
	}
	return entries
}

type orderEntry struct {
	key  int64
	val  *Order
	next *orderEntry
	prev *orderEntry
	last *orderEntry
}

type OrderSet struct {
	entries []orderEntry
	size    int32
	mask    uint32
}

func NewOrderSet(initCap int32) *OrderSet {
	capacity := toPowerOfTwo(initCap)
	entries := deadOrderElems(capacity)
	mask := uint32(capacity - 1)
	return &OrderSet{entries: entries, mask: mask}
}

func (s *OrderSet) getIdx(key int64) uint32 {
	half := (uint32(key) ^ uint32(key>>16)) ^ uint32(key>>32)
	// This hash function selected at random from http://burtleburtle.net/bob/hash/integer.html
	half = (half + 0x7ed55d16) + (half << 12)
	half = (half ^ 0xc761c23c) ^ (half >> 19)
	half = (half + 0x165667b1) + (half << 5)
	half = (half + 0xd3a2646c) ^ (half << 9)
	half = (half + 0xfd7046c5) + (half << 3)
	half = (half ^ 0xb55a4f09) ^ (half >> 16)
	return half & s.mask
}

func (s *OrderSet) Size() int32 {
	return s.size
}

func (s *OrderSet) Put(val *Order) {
	key := val.Guid
	idx := s.getIdx(key)
	e := &s.entries[idx]
	if e.key == tombstone64 {
		e.key = key
		e.val = val
		e.last = e
		return
	} else {
		//println("Hash Collision", idx, key)
		ne := &orderEntry{key: key, val: val, prev: e.last}
		e.last.next = ne
		e.last = ne
	}
	s.size++
}

func (s *OrderSet) Get(key int64) *Order {
	idx := s.getIdx(key)
	e := &s.entries[idx]
	for e != nil {
		if e.key == key {
			return e.val
		}
		e = e.next
	}
	return nil
}

func (s *OrderSet) Remove(key int64) *Order {
	idx := s.getIdx(key)
	e := &s.entries[idx]
	if e.next == nil && e.key == key {
		val := e.val
		e.val = nil
		e.key = tombstone64
		return val
	}
	if e.key == key {
		lkey := e.last.key
		lval := e.last.val
		e.last.prev.next = nil
		val := e.val
		e.key = lkey
		e.val = lval
		return val
	}
	for e = e.next; e != nil; e = e.next {
		if e.key == key {
			e.prev.next = e.next
			if e.next != nil {
				e.next.prev = e.prev
			}
			return e.val
		}
	}
	return nil
}
