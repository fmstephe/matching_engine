package coordinator

import (
	"container/list"
	. "github.com/fmstephe/matching_engine/msg"
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

const msgsToSend = 150

type echoClient struct {
	appMsgs  chan *Message
	dispatch chan *Message
	complete chan bool
}

func (c *echoClient) SetAppMsgs(appMsgs chan *Message) {
	c.appMsgs = appMsgs
}

func (c *echoClient) SetDispatch(dispatch chan *Message) {
	c.dispatch = dispatch
}

func (c *echoClient) Run() {
	sent := make([]*Message, msgsToSend)
	received := make([]*Message, msgsToSend)
	count := uint32(0)
	for {
		count++
		if count <= msgsToSend {
			m := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 1, TradeId: count, StockId: 1, Price: 7, Amount: 1}
			c.dispatch <- m
			sent[count-1] = m
		}
		select {
		case r := <-c.appMsgs:
			if received[r.TradeId-1] != nil {
				panic("Duplicate message received")
			}
			received[r.TradeId-1] = r
			if full(received) {
				c.complete <- true
				return
			}
		default:
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

type echoServer struct {
	appMsgs  chan *Message
	dispatch chan *Message
}

func (s *echoServer) SetAppMsgs(appMsgs chan *Message) {
	s.appMsgs = appMsgs
}

func (s *echoServer) SetDispatch(dispatch chan *Message) {
	s.dispatch = dispatch
}

func (s *echoServer) Run() {
	for {
		m := <-s.appMsgs
		r := &Message{}
		*r = *m
		r.Direction = OUT
		s.dispatch <- r
	}
}

// TODO this test has revealed a deadlock scenario where the listener is able to fill up
// the dispatch queue and no-one else is able to communicate. This requires that each component
// in coordinator package needs its own queue to communicate with the dispatcher.
// Currently the tests will fail with a rather dramatic deadlock.
func TestBadNetwork(t *testing.T) {
	complete := make(chan bool)
	c := &echoClient{complete: complete}
	s := &echoServer{}
	clientToServer := NewNetSim(newBadNetMeddler())
	serverToClient := NewNetSim(newBadNetMeddler())
	go clientToServer.run()
	go serverToClient.run()
	Coordinate(serverToClient, clientToServer, c, "Echo Client", false)
	Coordinate(clientToServer, serverToClient, s, "Echo Server", false)
	<-complete
}
