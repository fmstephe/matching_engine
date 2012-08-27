package matcher

import(
)

type OrderBuffer struct {
	read, write int
	orders []Order
}

func NewOrderBuffer(size int) *OrderBuffer {
	return &OrderBuffer{orders: make([]Order, size, size)}
}

func (ob *OrderBuffer) getForWrite() *Order {
	o := &ob.orders[ob.write]
	ob.write++
	return o
}

func (ob *OrderBuffer) getForRead() *Order {
	o := &ob.orders[ob.read]
	ob.read++
	return o
}

func (ob *OrderBuffer) clear() {
	for i := 0; i < len(ob.orders); i++ {
		ob.orders[i] = Order{}
	}
	ob.read = 0
	ob.write = 0
}

type ResponseBuffer struct {
	read, write int
	responses []Response
}

func NewResponseBuffer(size int) *ResponseBuffer {
	return &ResponseBuffer{responses: make([]Response, size, size)}
}

func (rb *ResponseBuffer) getForWrite() *Response {
	r := &rb.responses[rb.write]
	rb.write++
	return r
}

func (rb *ResponseBuffer) getForRead() *Response {
	r := &rb.responses[rb.read]
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
