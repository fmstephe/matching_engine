package rtree

import (
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
)

const (
	// NB: Ensure that 2^BLOCK_SHIFT == BLOCK_SIZE
	BLOCK_SHIFT = 14
	BLOCK_SIZE  = 16384
	EMPTY_IDX   = -1
)

type limit struct {
	head *trade.Order
	tail *trade.Order
}

func (l *limit) isEmpty() bool {
	return l.head == nil
}

func (l *limit) Peek() *trade.Order {
	return l.head
}

func (l *limit) Pop() *trade.Order {
	o := l.head
	l.head = o.Lower
	return o
}

func (l *limit) Push(o *trade.Order) {
	if l.head == nil {
		l.head = o
		l.tail = o
	} else {
		l.tail.Lower = o
		l.tail = o
	}
}

func (l *limit) remove(guid uint64) *trade.Order {
	return nil
}

type blocker interface {
	isEmpty() bool
	Peek() *trade.Order
	Pop() *trade.Order
	Push(*trade.Order)
	minPrice() int32
}

func newBlock(min int32, height uint, kind trade.OrderKind) blocker {
	if height == 0 {
		return newLeafBlock(min, kind)
	}
	return newNodeBlock(min, height, kind)
}

func pow(n int32, e uint) int32 {
	if e == 0 {
		return 1
	}
	p := n
	for ; e > 1; e-- {
		p *= n
	}
	return p
}

func minPrice(price int32, height uint) int32 {
	mask := int32(((BLOCK_SIZE) << (height * BLOCK_SHIFT))) - 1
	return price &^ mask
}

func maxPrice(minPrice int32, height uint) int32 {
	return minPrice + pow(BLOCK_SIZE, height+1) - 1
}

func minMaxPrice(price int32, height uint) (int32, int32) {
	min := minPrice(price, height)
	max := maxPrice(min, height)
	return min, max
}

func getIdx(price int32, height uint) int {
	return int(price>>(BLOCK_SHIFT*height)) & (BLOCK_SIZE - 1)
}

type R struct {
	min, max int32
	// If a heap contains only a single leafBlock then its height is 0
	// add one to height for every intermediate nodeBlock
	height uint
	size   int
	kind   trade.OrderKind
	block  blocker
}

func New(kind trade.OrderKind) *R {
	return &R{kind: kind}
}

func (r *R) isEmpty() bool {
	return r.size == 0
}

func (r *R) Size() int {
	return r.size
}

func (r *R) Peek() *trade.Order {
	if r.block == nil {
		return nil
	}
	return r.block.Peek()
}

func (r *R) Pop() *trade.Order {
	if r.block == nil {
		return nil
	}
	r.size--
	return r.block.Pop()
}

func (r *R) Push(o *trade.Order) {
	if r.block == nil {
		r.min, r.max = minMaxPrice(o.Price, 0)
		r.block = newBlock(r.min, 0, r.kind)
	} else if o.Price < r.min || o.Price > r.max {
		for o.Price < r.min || o.Price > r.max {
			newHeight := r.height + 1
			newMin, newMax := minMaxPrice(r.min, newHeight)
			block := newBlock(newMin, newHeight, r.kind)
			block.(*nodeBlock).PushBlock(r.block)
			r.block = block
			r.height = newHeight
			r.min = newMin
			r.max = newMax
		}
	}
	r.block.Push(o)
	r.size++
}

func (r *R) minPrice() int32 {
	return r.min
}

func (r *R) Kind() trade.OrderKind {
	return r.kind
}

func (r *R) Remove(o *trade.Order) {
	panic("Unsupported")
}

type nodeBlock struct {
	min     int32
	height  uint
	bestIdx int
	kind    trade.OrderKind
	blocks  [BLOCK_SIZE]blocker
}

func newNodeBlock(min int32, height uint, kind trade.OrderKind) *nodeBlock {
	return &nodeBlock{min: min, height: height, bestIdx: EMPTY_IDX, kind: kind}
}

func (node *nodeBlock) isEmpty() bool {
	return node.bestIdx == EMPTY_IDX
}

func (node *nodeBlock) Peek() *trade.Order {
	if node.bestIdx == EMPTY_IDX {
		return nil
	}
	return node.blocks[node.bestIdx].Peek()
}

func (node *nodeBlock) Pop() *trade.Order {
	if node.bestIdx == EMPTY_IDX {
		return nil
	}
	block := node.blocks[node.bestIdx]
	o := block.Pop()
	if block.isEmpty() {
		node.findIdx()
	}
	return o
}

func (node *nodeBlock) findIdx() {
	if node.kind == trade.BUY {
		for i := node.bestIdx; i >= 0; i-- {
			block := node.blocks[i]
			if block != nil && !block.isEmpty() {
				node.bestIdx = i
				return
			}
		}
	} else {
		for i := node.bestIdx; i < BLOCK_SIZE; i++ {
			block := node.blocks[i]
			if block != nil && !block.isEmpty() {
				node.bestIdx = i
				return
			}
		}
	}
	node.bestIdx = EMPTY_IDX
	return
}

func (node *nodeBlock) Push(o *trade.Order) {
	if o.Price < node.min {
		panic(fmt.Sprintf("Trying to Push order with price (%v) into block with minPrice (%v)", o.Price, node.minPrice()))
	}
	idx := getIdx(o.Price, node.height)
	if idx < 0 || idx >= BLOCK_SIZE {
		panic(fmt.Sprintf("This crazy bit operation produced an index: %v from price: %d", idx, o.Price))
	}
	block := node.blocks[idx]
	if block == nil {
		newHeight := node.height - 1
		newMin := minPrice(o.Price, newHeight)
		block = newBlock(newMin, newHeight, node.kind)
		node.blocks[idx] = block
	}
	block.Push(o)
	if (idx < node.bestIdx || node.bestIdx == EMPTY_IDX) && node.kind == trade.SELL {
		node.bestIdx = idx
	}
	if idx > node.bestIdx && node.kind == trade.BUY {
		node.bestIdx = idx
	}
}

func (node *nodeBlock) minPrice() int32 {
	return node.min
}

func (node *nodeBlock) PushBlock(block blocker) {
	if block.minPrice() < node.min {
		panic(fmt.Sprintf("Trying to Push block with min-price (%v) into block with min-price (%v)", block.minPrice(), node.minPrice()))
	}
	idx := getIdx(block.minPrice(), node.height)
	if idx < 0 || idx >= BLOCK_SIZE {
		panic(fmt.Sprintf("This crazy bit operation produced an index: %v from min-price: %d", idx, block.minPrice()))
	}
	node.blocks[idx] = block
}

type leafBlock struct {
	min     int32
	bestIdx int
	kind    trade.OrderKind
	limits  [BLOCK_SIZE]limit
}

func newLeafBlock(min int32, kind trade.OrderKind) *leafBlock {
	return &leafBlock{min: min, bestIdx: EMPTY_IDX, kind: kind}
}

func (leaf *leafBlock) isEmpty() bool {
	return leaf == nil || leaf.bestIdx == EMPTY_IDX
}

func (leaf *leafBlock) Peek() *trade.Order {
	if leaf.bestIdx == EMPTY_IDX {
		return nil
	}
	return leaf.limits[leaf.bestIdx].Peek()
}

func (leaf *leafBlock) Pop() *trade.Order {
	if leaf.bestIdx == EMPTY_IDX {
		return nil
	}
	o := leaf.limits[leaf.bestIdx].Pop()
	if leaf.limits[leaf.bestIdx].isEmpty() {
		leaf.findIdx()
	}
	return o
}

func (leaf *leafBlock) findIdx() {
	if leaf.kind == trade.BUY {
		for i := leaf.bestIdx; i >= 0; i-- {
			lim := &leaf.limits[i]
			if !lim.isEmpty() {
				leaf.bestIdx = i
				return
			}
		}
	} else {
		for i := leaf.bestIdx; i < BLOCK_SIZE; i++ {
			lim := &leaf.limits[i]
			if !lim.isEmpty() {
				leaf.bestIdx = i
				return
			}
		}
	}
	leaf.bestIdx = EMPTY_IDX
	return
}

func (leaf *leafBlock) Push(o *trade.Order) {
	if o.Price < leaf.min || o.Price >= leaf.min+BLOCK_SIZE {
		panic(fmt.Sprintf("Attempting to Push an order with price %v into leaf-block with minPrice %v and BLOCK_SIZE %v", o.Price, leaf.min, BLOCK_SIZE))
	}
	idx := getIdx(o.Price, 0)
	if idx < 0 || idx >= BLOCK_SIZE {
		panic(fmt.Sprintf("This crazy bit operation produced an index: %v from price: %v", idx, o.Price))
	}
	leaf.limits[idx].Push(o)
	if (idx < leaf.bestIdx || leaf.bestIdx == EMPTY_IDX) && leaf.kind == trade.SELL {
		leaf.bestIdx = idx
	}
	if idx > leaf.bestIdx && leaf.kind == trade.BUY {
		leaf.bestIdx = idx
	}
}

func (leaf *leafBlock) minPrice() int32 {
	return leaf.min
}
