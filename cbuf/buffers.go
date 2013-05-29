// Despite not being used, only half implemented and not not thread safe this package will be kept around because replacement queues for channels will probably have a very similar interface

package cbuf

import (
	"errors"
	"github.com/fmstephe/matching_engine/msg"
)

var WriteErr = errors.New("Cannot write to cbuf.Message")
var ReadErr = errors.New("Cannot read from cbuf.Message")

type Message struct {
	sizeMask    int
	read, write int
	responses   []msg.Message
}

func New(size int) *Message {
	realSize := 2
	for realSize < size {
		realSize *= 2
	}
	return &Message{sizeMask: realSize - 1, responses: make([]msg.Message, realSize, realSize)}
}

func (rb *Message) GetForWrite() (*msg.Message, error) {
	w := rb.write & rb.sizeMask
	r := rb.read & rb.sizeMask
	if rb.write != rb.read && w == r {
		return nil, WriteErr
	}
	resp := &rb.responses[w]
	rb.write++
	return resp, nil
}

func (rb *Message) GetForRead() (*msg.Message, error) {
	if rb.read == rb.write {
		return nil, ReadErr
	}
	r := rb.read & rb.sizeMask
	resp := &rb.responses[r]
	rb.read++
	return resp, nil
}

func (rb *Message) Clear() {
	for i := 0; i < len(rb.responses); i++ {
		rb.responses[i] = msg.Message{}
	}
	rb.read = 0
	rb.write = 0
}

func (rv *Message) Reads() int {
	return rv.read
}

func (rv *Message) Writes() int {
	return rv.write
}
