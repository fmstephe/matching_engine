package matcher

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
	"github.com/fmstephe/matching_engine/prioq"
)

type M struct {
	matchQueues prioq.MatchQueues // No constructor required
	slab        *prioq.Slab
	submit      chan *msg.Message
	orders      chan *msg.Message
}

func NewMatcher(slabSize int) *M {
	slab := prioq.NewSlab(slabSize)
	return &M{slab: slab}
}

func (m *M) SetSubmit(submit chan *msg.Message) {
	m.submit = submit
}

func (m *M) SetOrders(orders chan *msg.Message) {
	m.orders = orders
}

func (m *M) Run() {
	for {
		o := <-m.orders
		if o.Kind == msg.SHUTDOWN {
			r := &msg.Message{}
			r.WriteShutdown()
			m.submit <- r
			return
		}
		on := m.slab.Malloc()
		on.CopyFrom(o)
		switch on.Kind() {
		case msg.BUY:
			m.addBuy(on)
		case msg.SELL:
			m.addSell(on)
		case msg.CANCEL:
			m.cancel(on)
		default:
			// This should probably just be an message to m.submit
			panic(fmt.Sprintf("MsgKind %v not supported", on.Kind()))
		}
	}
}

func (m *M) addBuy(b *prioq.OrderNode) {
	if b.Price() == msg.MARKET_PRICE {
		// This should probably just be a message to m.submit
		panic("It is illegal to submit a buy at market price")
	}
	if !m.fillableBuy(b) {
		m.matchQueues.PushBuy(b)
	}
}

func (m *M) addSell(s *prioq.OrderNode) {
	if !m.fillableSell(s) {
		m.matchQueues.PushSell(s)
	}
}

func (m *M) cancel(o *prioq.OrderNode) {
	ro := m.matchQueues.Cancel(o)
	if ro != nil {
		completeCancelled(m.submit, ro)
		m.slab.Free(ro)
	} else {
		completeNotCancelled(m.submit, o)
	}
	m.slab.Free(o)
}

func (m *M) fillableBuy(b *prioq.OrderNode) bool {
	for {
		s := m.matchQueues.PeekSell()
		if s == nil {
			return false
		}
		if b.Price() >= s.Price() {
			if b.Amount() > s.Amount() {
				amount := s.Amount()
				price := price(b.Price(), s.Price())
				m.slab.Free(m.matchQueues.PopSell())
				b.ReduceAmount(amount)
				completeTrade(m.submit, msg.PARTIAL, msg.FULL, b, s, price, amount)
				continue
			}
			if s.Amount() > b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				s.ReduceAmount(amount)
				completeTrade(m.submit, msg.FULL, msg.PARTIAL, b, s, price, amount)
				m.slab.Free(b)
				return true // The buy has been used up
			}
			if s.Amount() == b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				completeTrade(m.submit, msg.FULL, msg.FULL, b, s, price, amount)
				m.slab.Free(m.matchQueues.PopSell())
				m.slab.Free(b)
				return true // The buy has been used up
			}
		} else {
			return false
		}
	}
	panic("Unreachable")
}

func (m *M) fillableSell(s *prioq.OrderNode) bool {
	for {
		b := m.matchQueues.PeekBuy()
		if b == nil {
			return false
		}
		if b.Price() >= s.Price() {
			if b.Amount() > s.Amount() {
				amount := s.Amount()
				price := price(b.Price(), s.Price())
				b.ReduceAmount(amount)
				completeTrade(m.submit, msg.PARTIAL, msg.FULL, b, s, price, amount)
				m.slab.Free(s)
				return true // The sell has been used up
			}
			if s.Amount() > b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				s.ReduceAmount(amount)
				completeTrade(m.submit, msg.FULL, msg.PARTIAL, b, s, price, amount)
				m.slab.Free(m.matchQueues.PopBuy())
				continue
			}
			if s.Amount() == b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				completeTrade(m.submit, msg.FULL, msg.FULL, b, s, price, amount)
				m.slab.Free(m.matchQueues.PopBuy())
				m.slab.Free(s)
				return true // The sell has been used up
			}
		} else {
			return false
		}
	}
	panic("Unreachable")
}

func price(bPrice, sPrice int64) int64 {
	if sPrice == msg.MARKET_PRICE {
		return bPrice
	}
	d := bPrice - sPrice
	return sPrice + (d / 2)
}

func completeTrade(submit chan *msg.Message, brk, srk msg.MsgKind, b, s *prioq.OrderNode, price int64, amount uint32) {
	br := &msg.Message{Price: -price, Amount: amount, TraderId: b.TraderId(), TradeId: b.TradeId(), StockId: b.StockId()}
	br.IP = b.IP()
	br.Port = b.Port()
	br.WriteResponse(brk)
	sr := &msg.Message{Price: price, Amount: amount, TraderId: s.TraderId(), TradeId: s.TradeId(), StockId: s.StockId()}
	sr.IP = b.IP()
	sr.Port = b.Port()
	sr.WriteResponse(srk)
	submit <- br
	submit <- sr
}

func completeCancelled(submit chan *msg.Message, c *prioq.OrderNode) {
	cr := writeMessage(c)
	cr.WriteCancelled()
	submit <- cr
}

func completeNotCancelled(submit chan *msg.Message, nc *prioq.OrderNode) {
	ncr := writeMessage(nc)
	ncr.WriteNotCancelled()
	submit <- ncr
}

func writeMessage(on *prioq.OrderNode) *msg.Message {
	m := &msg.Message{Price: on.Price(), Amount: on.Amount(), TraderId: on.TraderId(), TradeId: on.TradeId(), StockId: on.StockId()}
	m.IP = on.IP()
	m.Port = on.Port()
	return m
}
