package coordinator

import (
	"container/list"
	. "github.com/fmstephe/matching_engine/msg"
	"github.com/fmstephe/matching_engine/q"
	"testing"
	"time"
)

type badNetMeddler struct {
	lastMeddle time.Time
}

func newBadNetMeddler() *badNetMeddler {
	return &badNetMeddler{time.Now()}
}

func (m *badNetMeddler) meddle(buf *list.List) {
}

const msgsToSend = 10 * 1000

type echoClient struct {
	AppMsgHelper
	received []*Message
	complete chan bool
}

func newEchoClient(complete chan bool) *echoClient {
	return &echoClient{received: make([]*Message, msgsToSend), complete: complete}
}

func (c *echoClient) Run() {
	go sendAll(c.Out)
	for {
		m, shutdown := c.Process(<-c.In)
		if shutdown {
			return
		}
		if m != nil {
			if c.received[m.TradeId-1] != nil {
				panic("Duplicate message received")
			}
			c.received[m.TradeId-1] = m
			if full(c.received) {
				c.complete <- true
				return
			}
		}
	}
}

func full(received []*Message) bool {
	for _, rm := range received {
		if rm == nil {
			return false
		}
	}
	return true
}

func sendAll(out chan<- *Message) {
	for i := uint32(1); i <= msgsToSend; i++ {
		m := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 1, TradeId: i, StockId: 1, Price: 7, Amount: 1}
		out <- m
	}
}

type echoServer struct {
	AppMsgHelper
}

func (s *echoServer) Run() {
	for {
		m, shutdown := s.Process(<-s.In)
		if shutdown {
			return
		}
		if m != nil {
			r := &Message{}
			*r = *m
			r.Direction = OUT
			s.Out <- r
		}
	}
}

func TestBadNetwork(t *testing.T) {
	complete := make(chan bool)
	c := newEchoClient(complete)
	s := &echoServer{}
	// TODO with the use of simpleQ this doesn't actually test a bad network
	clientToServer := q.NewSimpleQ("clientToServer", msgsToSend*2)
	serverToClient := q.NewSimpleQ("serverToClient", msgsToSend*2)
	Coordinate(serverToClient, clientToServer, c, "Echo Client", true)
	Coordinate(clientToServer, serverToClient, s, "Echo Server", true)
	<-complete
}
