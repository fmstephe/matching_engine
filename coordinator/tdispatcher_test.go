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
	appMsgs  chan *Message
	dispatch chan *Message
}

func (m *tMatcher) Run() {
	panic("Not runnable")
}

func (m *tMatcher) SetDispatch(dispatch chan *Message) {
	m.dispatch = dispatch
}

func (m *tMatcher) SetAppMsgs(appMsgs chan *Message) {
	m.appMsgs = appMsgs
}

func setup() (*tListener, *tResponder, *tMatcher) {
	l := &tListener{}
	r := &tResponder{}
	m := &tMatcher{}
	d := connect(l, r, m, "Test System", false)
	go d.Run()
	return l, r, m
}

func TestInAppGoToMatcher(t *testing.T) {
	listener, _, matcher := setup()
	m := &Message{Direction: IN, Route: APP, Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	listener.dispatch <- m
	validate(t, <-matcher.appMsgs, m, 1)
}

func TestOutAppGoesToresponder(t *testing.T) {
	listener, responder, _ := setup()
	m := &Message{Direction: OUT, Route: APP, Kind: FULL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	listener.dispatch <- m
	validate(t, <-responder.responses, m, 1)
}

func TestErrorsGoToresponder(t *testing.T) {
	listener, responder, _ := setup()
	m := &Message{Status: READ_ERROR, Direction: OUT}
	listener.dispatch <- m
	validate(t, <-responder.responses, m, 1)
}

func TestOutAckGoesToresponder(t *testing.T) {
	listener, responder, _ := setup()
	m := &Message{Direction: OUT, Route: ACK, Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	listener.dispatch <- m
	validate(t, <-responder.responses, m, 1)
}

func TestInAckGoesToresponder(t *testing.T) {
	listener, responder, _ := setup()
	m := &Message{Direction: OUT, Route: ACK, Kind: FULL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	listener.dispatch <- m
	validate(t, <-responder.responses, m, 1)
}

func TestInCommandGoesToMatcher(t *testing.T) {
	listener, _, matcher := setup()
	m := &Message{Direction: IN, Route: SHUTDOWN}
	listener.dispatch <- m
	validate(t, <-matcher.appMsgs, m, 1)
}

func TestOutCommandGoesToresponder(t *testing.T) {
	listener, responder, _ := setup()
	m := &Message{Direction: OUT, Route: SHUTDOWN}
	listener.dispatch <- m
	validate(t, <-responder.responses, m, 1)
}

func TestInvalidMessageGoesToresponder(t *testing.T) {
	listener, responder, _ := setup()
	m := &Message{Route: APP, Kind: BUY}
	listener.dispatch <- m
	im := &Message{}
	*im = *m
	im.Status = INVALID_MSG_ERROR
	im.Direction = OUT
	validate(t, <-responder.responses, im, 1)
}