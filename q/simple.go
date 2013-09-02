package q

type simpleQ struct {
	name string
	c    chan []byte
}

func NewSimpleQ(name string, size int) *simpleQ {
	return &simpleQ{name: name, c: make(chan []byte, size)}
}

func (q *simpleQ) Read(p []byte) (int, error) {
	b := <-q.c
	copy(p, b)
	if len(p) < len(b) {
		return len(p), nil
	}
	return len(b), nil
}

func (q *simpleQ) Write(p []byte) (int, error) {
	b := make([]byte, len(p))
	copy(b, p)
	q.c <- b
	return len(b), nil
}

func (r *simpleQ) Close() error {
	r.c = nil
	return nil
}
