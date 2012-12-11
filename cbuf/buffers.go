package cbuf

import (
	"errors"
	"github.com/fmstephe/matching_engine/trade"
)

var WriteErr = errors.New("Cannot write to cbuf.Response")
var ReadErr = errors.New("Cannot read from cbuf.Response")

type Response struct {
	sizeMask    int
	read, write int
	responses   []trade.Response
}

func New(size int) *Response {
	realSize := 2
	for realSize < size {
		realSize *= 2
	}
	return &Response{sizeMask: realSize - 1, responses: make([]trade.Response, realSize, realSize)}
}

func (rb *Response) GetForWrite() (*trade.Response, error) {
	w := rb.write & rb.sizeMask
	r := rb.read & rb.sizeMask
	if rb.write != rb.read && w == r {
		return nil, WriteErr
	}
	resp := &rb.responses[w]
	rb.write++
	return resp, nil
}

func (rb *Response) GetForRead() (*trade.Response, error) {
	if rb.read == rb.write {
		return nil, ReadErr
	}
	r := rb.read & rb.sizeMask
	resp := &rb.responses[r]
	rb.read++
	return resp, nil
}

func (rb *Response) Clear() {
	for i := 0; i < len(rb.responses); i++ {
		rb.responses[i] = trade.Response{}
	}
	rb.read = 0
	rb.write = 0
}

func (rv *Response) Reads() int {
	return rv.read
}

func (rv *Response) Writes() int {
	return rv.write
}
