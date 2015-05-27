package coordinator

import (
	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/queues/spscq"
	"github.com/fmstephe/matching_engine/msg"
	"unsafe"
)

// Is Thread-Safe in a single-reader/single-writer context
type MsgReader interface {
	Read() *msg.Message
}

// Is Thread-Safe in a single-reader/single-writer context
type MsgWriter interface {
	Write(*msg.Message)
}

// A MsgReader/MsgWriter implementation using channels
type ChanReaderWriter struct {
	inout chan *msg.Message
}

func NewChanReaderWriter(size int) *ChanReaderWriter {
	inout := make(chan *msg.Message, size)
	return &ChanReaderWriter{
		inout: inout,
	}
}

func (rw *ChanReaderWriter) Read() *msg.Message {
	return <-rw.inout
}

func (rw *ChanReaderWriter) Write(m *msg.Message) {
	rw.inout <- m
}

// A MsgReader/MsgWriter implementation using spscq.PointerQ
// Should be fast
type SPSCQReaderWriter struct {
	q *spscq.PointerQ
}

func NewSPSCQReaderWriter(size int64) *SPSCQReaderWriter {
	p2Size := fmath.NxtPowerOfTwo(size)
	q, err := spscq.NewPointerQ(p2Size, 1024*1024)
	if err != nil {
		panic(err)
	}
	return &SPSCQReaderWriter{
		q: q,
	}
}

func (rw *SPSCQReaderWriter) Read() *msg.Message {
	return (*msg.Message)(rw.q.ReadSingleBlocking())
}

func (rw *SPSCQReaderWriter) Write(m *msg.Message) {
	rw.q.WriteSingleBlocking(unsafe.Pointer(m))
}

func (rw *SPSCQReaderWriter) Fails() (int64, int64) {
	return rw.q.FailedReads(), rw.q.FailedWrites()
}

// A Reader loaded with a set of test *msg.Message
// When the messages slice is exhausted Read() returns
// SHUTDOWN messages
type PreloadedReaderWriter struct {
	idx int
	ms  []msg.Message
}

func NewPreloadedReaderWriter(ms []msg.Message) *PreloadedReaderWriter {
	return &PreloadedReaderWriter{
		ms: ms,
	}
}

func (r *PreloadedReaderWriter) Read() *msg.Message {
	if r.idx >= len(r.ms) {
		return &msg.Message{Kind: msg.SHUTDOWN}
	}
	m := r.ms[r.idx]
	r.idx++
	return &m
}

func (r *PreloadedReaderWriter) Write(m *msg.Message) {
}

// A Writer which does nothing. Good for some performance testing.
type ShutdownReaderWriter struct {
	out chan *msg.Message
}

func NewShutdownReaderWriter() *ShutdownReaderWriter {
	return &ShutdownReaderWriter{
		out: make(chan *msg.Message),
	}
}

func (rw *ShutdownReaderWriter) Read() *msg.Message {
	return <-rw.out
}

func (rw *ShutdownReaderWriter) Write(m *msg.Message) {
	if m.Kind == msg.SHUTDOWN {
		rw.out <- m
	}
}
