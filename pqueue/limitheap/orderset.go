package limitheap

import (
	"github.com/fmstephe/matching_engine/trade"
)

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
	key int64
	val *trade.Order
	next *orderEntry
	prev *orderEntry
	last *orderEntry
}

type orderset struct {
	entries  []orderEntry
	size   int32
	mask uint32
}

func newOrderSet(initCap int32) *orderset {
	capacity := toPowerOfTwo(initCap)
	entries := deadOrderElems(capacity)
	mask := uint32(capacity-1)
	return &orderset{entries: entries, mask: mask}
}

func (s *orderset) getIdx(key int64) uint32 {
	half := (uint32(key) ^ uint32(key >> 16)) ^ uint32(key >> 32)
	// This hash function selected at random from http://burtleburtle.net/bob/hash/integer.html
	half = (half+0x7ed55d16) + (half<<12)
	half = (half^0xc761c23c) ^ (half>>19)
	half = (half+0x165667b1) + (half<<5)
	half = (half+0xd3a2646c) ^ (half<<9)
	half = (half+0xfd7046c5) + (half<<3)
	half = (half^0xb55a4f09) ^ (half>>16)
	return half & s.mask
}

func (s *orderset) Size() int32 {
	return s.size
}

func (s *orderset) Put(key int64, val *trade.Order) {
	idx := s.getIdx(key)
	e := &s.entries[idx]
	if e.key == tombstone64 {
		e.key = key
		e.val = val
		e.last = e
		return
	} else {
	//	println("Hash Collision", idx, key)
		ne := &orderEntry{key: key, val: val, prev: e.last}
		e.last.next = ne
		ne.prev = e.last
		e.last = e
	}
	s.size++
}

func (s *orderset) Get(key int64) *trade.Order {
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

func (s *orderset) Remove(key int64) *trade.Order {
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
	e = e.next
	for e != nil {
		if e.key == key {
			e.prev.next = e.next
			e.next.prev = e.prev
			return e.val
		}
		e = e.next
	}
	return nil
}
