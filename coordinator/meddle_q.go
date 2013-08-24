package coordinator

import (
	"container/list"
	"runtime"
)

type netMeddler interface {
	meddle(*list.List)
}

type noMeddler struct{}

func (m *noMeddler) meddle(l *list.List) {}

type meddleQ struct {
	writeChan chan []byte
	readChan  chan []byte
	buf       *list.List
	meddler   netMeddler
}

func NewNetSim(meddler netMeddler) *meddleQ {
	return &meddleQ{writeChan: make(chan []byte), readChan: make(chan []byte), buf: list.New(), meddler: meddler}
}

func DefaultNetSim() *meddleQ {
	return &meddleQ{meddler: &noMeddler{}}
}

func (n *meddleQ) Read(p []byte) (int, error) {
	c := <-n.readChan
	copy(p, c)
	if len(p) < len(c) {
		return len(p), nil
	}
	return len(c), nil
}

func (n *meddleQ) Close() error {
	return nil
}

func (n *meddleQ) Write(p []byte) (int, error) {
	c := make([]byte, len(p))
	copy(c, p)
	n.writeChan <- c
	return len(c), nil
}

func (n *meddleQ) run() {
	for {
		n.read()
		n.meddler.meddle(n.buf)
		n.write()
		// TODO why is this necessary?
		runtime.Gosched()
	}
}

func (n *meddleQ) read() {
	select {
	case r := <-n.writeChan:
		n.buf.PushBack(r)
	default:
	}
}

func (n *meddleQ) write() {
	if n.buf.Len() > 0 {
		head := n.buf.Front()
		n.buf.Remove(head)
		n.readChan <- head.Value.([]byte)
	}
}
