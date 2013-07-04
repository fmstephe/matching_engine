package coordinator

import (
	. "github.com/fmstephe/matching_engine/msg"
	"testing"
)

type tListener struct {
	dispatch chan *Message
}

func (l tListener) Run() {
	panic("Not runnable")
}

func (l *tListener) SetDispatch(dispatch chan *Message) {
	l.dispatch = dispatch
}

type tResponder struct {
	dispatch  chan *Message
	responses chan *Message
}

func (r *tResponder) Run() {
	panic("Not runnable")
}

func (r *tResponder) SetDispatch(dispatch chan *Message) {
	r.dispatch = dispatch
}

func (r *tResponder) SetResponses(responses chan *Message) {
	r.responses = responses
}

type tMatcher struct {
	orders   chan *Message
	dispatch chan *Message
}

func (m *tMatcher) Run() {
	panic("Not runnable")
}

func (m *tMatcher) SetDispatch(dispatch chan *Message) {
	m.dispatch = dispatch
}

func (m *tMatcher) SetOrders(orders chan *Message) {
	m.orders = orders
}

func setup() (*tListener, *tResponder, *tMatcher) {
	l := &tListener{}
	r := &tResponder{}
	m := &tMatcher{}
	d := connect(l, r, m, false)
	go d.Run()
	return l, r, m
}

func validate(t *testing.T, e, m *Message) {
	if *e != *m {
		t.Errorf("Expecting %v, found %v", e, m)
	}
}

func TestOrdersGoToMatcher(t *testing.T) {
	listener, _, matcher := setup()
	m := &Message{Status: NORMAL, Route: ORDER, Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	listener.dispatch <- m
	validate(t, <-matcher.orders, m)
}

func TestErrorsGoToResponder(t *testing.T) {
	listener, responder, _ := setup()
	m := &Message{Status: READ_ERROR}
	listener.dispatch <- m
	validate(t, <-responder.responses, m)
}

func TestServerAckGoesToResponder(t *testing.T) {
	listener, responder, _ := setup()
	m := &Message{Status: NORMAL, Route: SERVER_ACK, Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	listener.dispatch <- m
	validate(t, <-responder.responses, m)
}

func TestClientAckGoesToResponder(t *testing.T) {
	listener, responder, _ := setup()
	m := &Message{Status: NORMAL, Route: CLIENT_ACK, Kind: FULL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	listener.dispatch <- m
	validate(t, <-responder.responses, m)
}

func TestMatcherResponseGoesToResponder(t *testing.T) {
	listener, responder, _ := setup()
	m := &Message{Status: NORMAL, Route: MATCHER_RESPONSE, Kind: FULL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	listener.dispatch <- m
	validate(t, <-responder.responses, m)
}

func TestCommandGoesToResponderAndMatcher(t *testing.T) {
	listener, responder, matcher := setup()
	m := &Message{Status: NORMAL, Route: COMMAND, Kind: SHUTDOWN}
	listener.dispatch <- m
	validate(t, <-responder.responses, m)
	validate(t, <-matcher.orders, m)
}

func TestInvalidMessageGoesToResponder(t *testing.T) {
	listener, responder, _ := setup()
	m := &Message{Status: NORMAL, Route: ORDER, Kind: BUY}
	listener.dispatch <- m
	im := &Message{}
	*im = *m
	im.WriteStatus(INVALID_MSG_ERROR)
	validate(t, <-responder.responses, im)
}
