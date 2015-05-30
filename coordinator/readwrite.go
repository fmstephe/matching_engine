package coordinator

import (
	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/queues/spscq"
	"github.com/fmstephe/matching_engine/msg"
	"unsafe"
)

//TODO need a testsuite for this

// Is Thread-Safe in a single-reader/single-writer context
type MsgReader interface {
	Read(*msg.Message)
}

// Is Thread-Safe in a single-reader/single-writer context
type MsgWriter interface {
	GetForWrite() *msg.Message
	Write()
}

// Is Thread-Safe in a single-reader/single-writer context
type MsgReaderWriter interface {
	MsgReader
	MsgWriter
}

type msgWriterCache struct {
	cachedM msg.Message
}

func (c *msgWriterCache) GetForWrite() *msg.Message {
	return &c.cachedM
}

// A MsgReader/MsgWriter implementation using channels
type ChanReaderWriter struct {
	msgWriterCache
	inout chan msg.Message
}

func NewChanReaderWriter(size int) *ChanReaderWriter {
	inout := make(chan msg.Message, size)
	return &ChanReaderWriter{
		inout: inout,
	}
}

func (rw *ChanReaderWriter) Read(m *msg.Message) {
	*m = <-rw.inout
}

func (rw *ChanReaderWriter) Write() {
	rw.inout <- rw.cachedM
}

// A MsgReader/MsgWriter implementation using spscq.PointerQ
// Should be fast
type SPSCQReaderWriter struct {
	msgWriterCache
	q *spscq.PointerQ
}

func NewSPSCQReaderWriter(size int64) *SPSCQReaderWriter {
	p2Size := fmath.NxtPowerOfTwo(size)
	q, err := spscq.NewPointerQ(p2Size, 1024)
	if err != nil {
		panic(err)
	}
	return &SPSCQReaderWriter{
		q: q,
	}
}

func (rw *SPSCQReaderWriter) Read(m *msg.Message) {
	*m = *((*msg.Message)(rw.q.ReadSingleBlocking()))
}

func (rw *SPSCQReaderWriter) Write() {
	m := &msg.Message{}
	*m = rw.cachedM
	rw.q.WriteSingleBlocking(unsafe.Pointer(m))
}

func (rw *SPSCQReaderWriter) Fails() (int64, int64) {
	return rw.q.FailedReads(), rw.q.FailedWrites()
}

// A Reader loaded with a set of test *msg.Message
// When the messages slice is exhausted Read() returns
// SHUTDOWN messages
type PreloadedReaderWriter struct {
	msgWriterCache
	idx int
	ms  []msg.Message
}

func NewPreloadedReaderWriter(ms []msg.Message) *PreloadedReaderWriter {
	return &PreloadedReaderWriter{
		ms: ms,
	}
}

func (r *PreloadedReaderWriter) Read(m *msg.Message) {
	if r.idx >= len(r.ms) {
		*m = msg.Message{Kind: msg.SHUTDOWN}
	} else {
		*m = r.ms[r.idx]
		r.idx++
	}
}

func (r *PreloadedReaderWriter) Write() {
}

// A Writer which does almost nothing. Good for some performance testing.
type ShutdownReaderWriter struct {
	msgWriterCache
	out chan *msg.Message
}

func NewShutdownReaderWriter() *ShutdownReaderWriter {
	return &ShutdownReaderWriter{
		out: make(chan *msg.Message, 1),
	}
}

func (rw *ShutdownReaderWriter) Read(m *msg.Message) {
	*m = *(<-rw.out)
}

func (rw *ShutdownReaderWriter) Write() {
	if rw.cachedM.Kind == msg.SHUTDOWN {
		m := &msg.Message{}
		*m = rw.cachedM
		rw.out <- m
	}
}

// A Writer which does absolutely nothing. Good for single threaded performance testing.
type NoopReaderWriter struct {
	msgWriterCache
}

func NewNoopReaderWriter() *NoopReaderWriter {
	return &NoopReaderWriter{}
}

func (rw *NoopReaderWriter) Read(m *msg.Message) {
}

func (rw *NoopReaderWriter) Write() {
}
