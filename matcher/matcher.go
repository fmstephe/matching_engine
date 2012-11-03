package matcher

import (
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
)

type M struct {
	buys   *trade.Tree
	sells  *trade.Tree
	orders *trade.Tree
	output *ResponseBuffer
}

func NewMatcher(output *ResponseBuffer) *M {
	buys := trade.NewTree()
	sells := trade.NewTree()
	orders := trade.NewTree()
	return &M{buys: buys, sells: sells, orders: orders, output: output}
}

/*
func (m *M) Survey() (buys []*trade.SurveyLimit, sells []*trade.SurveyLimit, orders *trade.OrderSet, executions int) {
	buys = m.buys.Survey()
	sells = m.sells.Survey()
	orders = m.orders
	executions = m.output.Writes()
	return
}
*/

func (m *M) Submit(o *trade.Order) {
	switch o.Kind {
	case trade.BUY:
		m.addBuy(o)
	case trade.SELL:
		m.addSell(o)
	case trade.DELETE:
		m.remove(o)
	default:
		panic(fmt.Sprintf("OrderKind %#v not supported", o.Kind))
	}
}

func (m *M) addBuy(b *trade.Order) {
	if b.Price == trade.MARKET_PRICE {
		panic("It is illegal to submit a buy at market price")
	}
	if !m.fillableBuy(b) {
		//	m.orders.Push(b)
		m.buys.Push(&b.LimitNode)
	}
}

func (m *M) addSell(s *trade.Order) {
	if !m.fillableSell(s) {
		//	m.orders.Put(s)
		m.sells.Push(&s.LimitNode)
	}
}

func (m *M) remove(o *trade.Order) {
	//ro := m.orders.Remove(o.Guid)
	/*
		if ro != nil {
			ro.RemoveFromLimit()
		}
	*/
}

func (m *M) fillableBuy(b *trade.Order) bool {
	for {
		s := m.sells.PopMin().O
		if s == nil {
			return false
		}
		if b.Price >= s.Price {
			if b.Amount > s.Amount {
				amount := s.Amount
				price := price(b.Price, s.Price)
				m.sells.PopMin()
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
				m.sells.PopMin()
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
		b := m.buys.PeekMax().O
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
				m.buys.PopMax()
				continue
			}
			if s.Amount == b.Amount {
				amount := b.Amount
				price := price(b.Price, s.Price)
				m.completeTrade(b, s, price, amount)
				m.buys.PopMax()
				return true // The sell has been used up
			}
		} else {
			return false
		}
	}
	panic("Unreachable")
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
