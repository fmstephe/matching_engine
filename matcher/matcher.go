package matcher

import (
	"fmt"
	"github.com/fmstephe/matching_engine/pqueue/limitheap"
	"github.com/fmstephe/matching_engine/trade"
)

type M struct {
	buys   *limitheap.H
	sells  *limitheap.H
	orders *trade.OrderSet
	output *ResponseBuffer
}

func NewMatcher(buys, sells *limitheap.H, output *ResponseBuffer) *M {
	if buys.Kind() != trade.BUY {
		panic("Provided a buy priority queue that was not accepting buys!")
	}
	if sells.Kind() != trade.SELL {
		panic("Provided a sell priority queue that was not accepting sells!")
	}
	orders := trade.NewOrderSet(1000) // Need a data structure that doesn't require an initialisation constant
	return &M{buys: buys, sells: sells, orders: orders, output: output}
}

func (m *M) Survey() (buys []*trade.SurveyLimit, sells []*trade.SurveyLimit) {
	buys = m.buys.Survey()
	sells = m.sells.Survey()
	return
}

func (m *M) Submit(o *trade.Order) {
	switch o.Kind {
	case trade.BUY:
		m.addBuy(o)
	case trade.SELL:
		m.addSell(o)
	case trade.DELETE:
		m.remove(o)
	default:
		panic(fmt.Sprintf("OrderKind %#v not currently supported", o.Kind))
	}
}

func (m *M) addBuy(b *trade.Order) {
	if b.Price == trade.MARKET_PRICE {
		panic("It is illegal to submit a buy at market price")
	}
	if !m.fillableBuy(b) {
		m.orders.Put(b)
		m.buys.Push(b)
	}
}

func (m *M) addSell(s *trade.Order) {
	if !m.fillableSell(s) {
		m.orders.Put(s)
		m.sells.Push(s)
	}
}

func (m *M) remove(o *trade.Order) {
	ro := m.orders.Remove(o.Guid)
	ro.RemoveFromLimit()
}

func (m *M) fillableBuy(b *trade.Order) bool {
	for {
		s := m.sells.Peek()
		if s == nil {
			return false
		}
		if b.Price >= s.Price {
			if b.Amount > s.Amount {
				amount := s.Amount
				price := price(b.Price, s.Price)
				m.clearHead(m.sells)
				b.Amount -= amount
				m.completeTrade(b, s, price, amount)
				continue
			}
			if s.Amount > b.Amount {
				amount := b.Amount
				price := price(b.Price, s.Price)
				s.Amount -= amount
				m.completeTrade(b, s, price, amount)
				return true // The buy has been used up
			}
			if s.Amount == b.Amount {
				amount := b.Amount
				price := price(b.Price, s.Price)
				m.completeTrade(b, s, price, amount)
				m.clearHead(m.sells)
				return true // The buy has been used up
			}
		} else {
			return false
		}
	}
	panic("Unreachable")
}

func (m *M) fillableSell(s *trade.Order) bool {
	for {
		b := m.buys.Peek()
		if b == nil {
			return false
		}
		if b.Price >= s.Price {
			if b.Amount > s.Amount {
				amount := s.Amount
				price := price(b.Price, s.Price)
				b.Amount -= amount
				m.completeTrade(b, s, price, amount)
				return true // The sell has been used up
			}
			if s.Amount > b.Amount {
				amount := b.Amount
				price := price(b.Price, s.Price)
				s.Amount -= amount
				m.completeTrade(b, s, price, amount)
				m.clearHead(m.buys)
				continue
			}
			if s.Amount == b.Amount {
				amount := b.Amount
				price := price(b.Price, s.Price)
				m.completeTrade(b, s, price, amount)
				m.clearHead(m.buys)
				return true // The sell has been used up
			}
		} else {
			return false
		}
	}
	panic("Unreachable")
}

func (m *M) clearHead(q *limitheap.H) {
	o := q.Pop()
	m.orders.Remove(o.Guid)
}

func price(bPrice, sPrice int32) int32 {
	if sPrice == trade.MARKET_PRICE {
		return bPrice
	}
	d := bPrice - sPrice
	return sPrice + (d >> 1)
}

func (m *M) completeTrade(b, s *trade.Order, price int32, amount uint32) {
	// TODO write the response type into these responses
	// Write the buy response
	rb := m.output.getForWrite()
	rb.Price = -price
	rb.Amount = amount
	rb.TradeId = b.TradeId
	rb.CounterParty = s.TraderId
	// Write the sell response
	rs := m.output.getForWrite()
	rs.Price = price
	rs.Amount = amount
	rs.TradeId = s.TradeId
	rs.CounterParty = b.TraderId
}
