package heap

import (
	"testing"
)

type elem int

func (i elem) Less(j Elem) bool {
	ji := j.(elem)
	return int(i) < int(ji)
}

func (i elem) SetIndex(index int) {
}

func verify(h Interface, t *testing.T, i int) {
	n := h.Len()
	internal := *h.(*elems)
	j1 := 2*i + 1
	j2 := 2*i + 2
	if j1 < n {
		if internal[j1].Less(internal[i]) {
			t.Errorf("heap invariant invalidated [%d] = %d > [%d] = %d", i, internal[i], j1, internal[j1])
			return
		}
		verify(h, t, j1)
	}
	if j2 < n {
		if internal[j2].Less(internal[i]) {
			t.Errorf("heap invariant invalidated [%d] = %d > [%d] = %d", i, internal[i], j1, internal[j2])
			return
		}
		verify(h, t, j2)
	}
}

func TestInit0(t *testing.T) {
	h := New()
	for i := 20; i > 0; i-- {
		h.Push(elem(0)) // all elements are the same
	}
	verify(h, t, 0)

	for i := 1; h.Len() > 0; i++ {
		x := int(h.Pop().(elem))
		verify(h, t, 0)
		if x != 0 {
			t.Errorf("%d.th pop got %d; want %d", i, x, 0)
		}
	}
}

func TestInit1(t *testing.T) {
	h := New()
	for i := 20; i > 0; i-- {
		h.Push(elem(i)) // all elements are different
	}
	verify(h, t, 0)

	for i := 1; h.Len() > 0; i++ {
		x := int(h.Pop().(elem))
		verify(h, t, 0)
		if x != i {
			t.Errorf("%d.th pop got %d; want %d", i, x, i)
		}
	}
}

func Test(t *testing.T) {
	h := New()
	verify(h, t, 0)

	for i := 20; i > 10; i-- {
		h.Push(elem(i))
	}
	verify(h, t, 0)

	for i := 10; i > 0; i-- {
		h.Push(elem(i))
		verify(h, t, 0)
	}

	for i := 1; h.Len() > 0; i++ {
		x := int(h.Pop().(elem))
		if i < 20 {
			h.Push(elem(20 + i))
		}
		verify(h, t, 0)
		if x != i {
			t.Errorf("%d.th pop got %d; want %d", i, x, i)
		}
	}
}

func _TestRemove0(t *testing.T) {
	h := New()
	for i := 0; i < 10; i++ {
		h.Push(elem(i))
	}
	verify(h, t, 0)

	for h.Len() > 0 {
		i := h.Len() - 1
		x := int(h.Remove(i).(elem))
		if x != i {
			t.Errorf("Remove(%d) got %d; want %d", i, x, i)
		}
		verify(h, t, 0)
	}
}

func TestRemove1(t *testing.T) {
	h := New()
	for i := 0; i < 10; i++ {
		h.Push(elem(i))
	}
	verify(h, t, 0)

	for i := 0; h.Len() > 0; i++ {
		x := int(h.Remove(0).(elem))
		if x != i {
			t.Errorf("Remove(0) got %d; want %d", x, i)
		}
		verify(h, t, 0)
	}
}

func TestRemove2(t *testing.T) {
	N := 10

	h := New()
	for i := 0; i < N; i++ {
		h.Push(elem(i))
	}
	verify(h, t, 0)

	m := make(map[int]bool)
	for h.Len() > 0 {
		m[int(h.Remove((h.Len()-1)/2).(elem))] = true
		verify(h, t, 0)
	}

	if len(m) != N {
		t.Errorf("len(m) = %d; want %d", len(m), N)
	}
	for i := 0; i < len(m); i++ {
		if !m[i] {
			t.Errorf("m[%d] doesn't exist", i)
		}
	}
}
