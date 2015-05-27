package q

import (
	"container/list"
	"runtime"
)

type Meddler interface {
	Meddle(*list.List)
}

type notMeddler struct{}

func (m *notMeddler) Meddle(l *list.List) {}

type meddleQ struct {
	name      string
	writeChan chan []byte
	readChan  chan []byte
	shutdown  chan bool
	buf       *list.List
	meddler   Meddler
}

func NewMeddleQ(name string, meddler Meddler) *meddleQ {
	q := &meddleQ{
		name:      name,
		writeChan: make(chan []byte, 100),
		readChan:  make(chan []byte, 100),
		shutdown:  make(chan bool),
		buf:       list.New(),
		meddler:   meddler}
	go q.run()
	return q
}

func NewSimpleQ(name string) *meddleQ {
	q := &meddleQ{name: name,
		writeChan: make(chan []byte, 100),
		readChan:  make(chan []byte, 100),
		shutdown:  make(chan bool),
		buf:       list.New(),
		meddler:   &notMeddler{}}
	go q.run()
	return q
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
	}
}

func (q *meddleQ) read() {
	if q.buf.Len() == 0 {
		r := <-q.writeChan
		q.buf.PushBack(r)
	} else {
		select {
		case r := <-q.writeChan:
			q.buf.PushBack(r)
		default:
			runtime.Gosched()
		}
	}
}

func (q *meddleQ) write() {
	if q.buf.Len() > 0 {
		head := q.buf.Front()
		val := head.Value.([]byte)
		select {
		case q.readChan <- val:
			q.buf.Remove(head)
		default:
			runtime.Gosched()
		}

	}
}
