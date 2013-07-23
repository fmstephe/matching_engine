package matcher

import (
	"fmt"
	"github.com/fmstephe/matching_engine/matcher/pqueue"
	"github.com/fmstephe/matching_engine/msg"
)

type M struct {
	matchQueues map[uint32]*pqueue.MatchQueues
	slab        *pqueue.Slab
	dispatch    chan *msg.Message
	orders      chan *msg.Message
}

func NewMatcher(slabSize int) *M {
	matchQueues := make(map[uint32]*pqueue.MatchQueues)
	slab := pqueue.NewSlab(slabSize)
	return &M{matchQueues: matchQueues, slab: slab}
}

func (m *M) SetDispatch(dispatch chan *msg.Message) {
	m.dispatch = dispatch
}

func (m *M) SetOrders(orders chan *msg.Message) {
	m.orders = orders
}

func (m *M) getMatchQueues(stockId uint32) *pqueue.MatchQueues {
	q := m.matchQueues[stockId]
	if q == nil {
		q = &pqueue.MatchQueues{}
		m.matchQueues[stockId] = q
	}
	return q
}

func (m *M) Run() {
	for {
		o := <-m.orders
		if o.Route == msg.SHUTDOWN {
			r := &msg.Message{}
			r.WriteShutdown()
			m.dispatch <- r
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
			panic(fmt.Sprintf("MsgKind %v not supported", o))
		}
	}
}

func (m *M) addBuy(b *pqueue.OrderNode) {
	if b.Price() == msg.MARKET_PRICE {
		panic("It is illegal to send a buy at market price")
	}
	q := m.getMatchQueues(b.StockId())
	if !m.fillableBuy(b, q) {
		q.PushBuy(b)
	}
}

func (m *M) addSell(s *pqueue.OrderNode) {
	q := m.getMatchQueues(s.StockId())
	if !m.fillableSell(s, q) {
		q.PushSell(s)
	}
}

func (m *M) cancel(o *pqueue.OrderNode) {
	q := m.getMatchQueues(o.StockId())
	ro := q.Cancel(o)
	if ro != nil {
		completeCancelled(m.dispatch, ro)
		m.slab.Free(ro)
	} else {
		completeNotCancelled(m.dispatch, o)
	}
	m.slab.Free(o)
}

func (m *M) fillableBuy(b *pqueue.OrderNode, q *pqueue.MatchQueues) bool {
	for {
		s := q.PeekSell()
		if s == nil {
			return false
		}
		if b.Price() >= s.Price() {
			if b.Amount() > s.Amount() {
				amount := s.Amount()
				price := price(b.Price(), s.Price())
				m.slab.Free(q.PopSell())
				b.ReduceAmount(amount)
				completeTrade(m.dispatch, msg.PARTIAL, msg.FULL, b, s, price, amount)
				continue
			}
			if s.Amount() > b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				s.ReduceAmount(amount)
				completeTrade(m.dispatch, msg.FULL, msg.PARTIAL, b, s, price, amount)
				m.slab.Free(b)
				return true // The buy has been used up
			}
			if s.Amount() == b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				completeTrade(m.dispatch, msg.FULL, msg.FULL, b, s, price, amount)
				m.slab.Free(q.PopSell())
				m.slab.Free(b)
				return true // The buy has been used up
			}
		} else {
			return false
		}
	}
	panic("Unreachable")
}

func (m *M) fillableSell(s *pqueue.OrderNode, q *pqueue.MatchQueues) bool {
	for {
		b := q.PeekBuy()
		if b == nil {
			return false
		}
		if b.Price() >= s.Price() {
			if b.Amount() > s.Amount() {
				amount := s.Amount()
				price := price(b.Price(), s.Price())
				b.ReduceAmount(amount)
				completeTrade(m.dispatch, msg.PARTIAL, msg.FULL, b, s, price, amount)
				m.slab.Free(s)
				return true // The sell has been used up
			}
			if s.Amount() > b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				s.ReduceAmount(amount)
				completeTrade(m.dispatch, msg.FULL, msg.PARTIAL, b, s, price, amount)
				m.slab.Free(q.PopBuy())
				continue
			}
			if s.Amount() == b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				completeTrade(m.dispatch, msg.FULL, msg.FULL, b, s, price, amount)
				m.slab.Free(q.PopBuy())
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

func completeTrade(dispatch chan *msg.Message, brk, srk msg.MsgKind, b, s *pqueue.OrderNode, price int64, amount uint32) {
	br := &msg.Message{Price: -price, Amount: amount, TraderId: b.TraderId(), TradeId: b.TradeId(), StockId: b.StockId()}
	br.WriteApp(brk)
	sr := &msg.Message{Price: price, Amount: amount, TraderId: s.TraderId(), TradeId: s.TradeId(), StockId: s.StockId()}
	sr.WriteApp(srk)
	dispatch <- br
	dispatch <- sr
}

func completeCancelled(dispatch chan *msg.Message, c *pqueue.OrderNode) {
	cr := writeMessage(c)
	cr.WriteApp(msg.CANCELLED)
	dispatch <- cr
}

func completeNotCancelled(dispatch chan *msg.Message, nc *pqueue.OrderNode) {
	ncr := writeMessage(nc)
	ncr.WriteApp(msg.NOT_CANCELLED)
	dispatch <- ncr
}

func writeMessage(on *pqueue.OrderNode) *msg.Message {
	m := &msg.Message{Price: on.Price(), Amount: on.Amount(), TraderId: on.TraderId(), TradeId: on.TradeId(), StockId: on.StockId()}
	return m
}
