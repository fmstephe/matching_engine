package q

import (
	"container/list"
	"runtime"
)

type Meddler interface {
	Meddle(*list.List)
}

type meddleQ struct {
	writeChan chan []byte
	readChan  chan []byte
	shutdown  chan bool
	buf       *list.List
	meddler   Meddler
}

func NewNetSim(meddler Meddler) *meddleQ {
	return &meddleQ{writeChan: make(chan []byte), readChan: make(chan []byte), shutdown: make(chan bool), buf: list.New(), meddler: meddler}
}

func (q *meddleQ) Read(p []byte) (int, error) {
	c := <-q.readChan
	copy(p, c)
	if len(p) < len(c) {
		return len(p), nil
	}
	return len(c), nil
}

func (q *meddleQ) Close() error {
	q.shutdown <- true
	return nil
}

func (q *meddleQ) Write(p []byte) (int, error) {
	c := make([]byte, len(p))
	copy(c, p)
	q.writeChan <- c
	return len(c), nil
}

func (q *meddleQ) run() {
	for {
		q.read()
		q.meddler.Meddle(q.buf)
		q.write()
		select {
		case <-q.shutdown:
			return
		default:
		}
		runtime.Gosched()
	}
}

func (q *meddleQ) read() {
	select {
	case r := <-q.writeChan:
		q.buf.PushBack(r)
	default:
	}
}

func (q *meddleQ) write() {
	if q.buf.Len() > 0 {
		head := q.buf.Front()
		q.buf.Remove(head)
		q.readChan <- head.Value.([]byte)
	}
}
