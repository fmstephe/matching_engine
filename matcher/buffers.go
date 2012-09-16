package matcher

import ()

type ResponseBuffer struct {
	sizeMask    int
	read, write int
	responses   []Response
}

func NewResponseBuffer(size int) *ResponseBuffer {
	realSize := 2
	for realSize < size {
		realSize *= 2
	}
	return &ResponseBuffer{sizeMask: realSize - 1, responses: make([]Response, realSize, realSize)}
}

func (rb *ResponseBuffer) getForWrite() *Response {
	r := &rb.responses[rb.write&rb.sizeMask]
	rb.write++
	return r
}

func (rb *ResponseBuffer) getForRead() *Response {
	r := &rb.responses[rb.read&rb.sizeMask]
	rb.read++
	return r
}

func (rb *ResponseBuffer) clear() {
	for i := 0; i < len(rb.responses); i++ {
		rb.responses[i] = Response{}
	}
	rb.read = 0
	rb.write = 0
}
