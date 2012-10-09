package matcher

import (
	"github.com/fmstephe/matching_engine/pqueue"
	"github.com/fmstephe/matching_engine/trade"
)

type M struct {
	buys, sells pqueue.Q
	output      *ResponseBuffer
}

func NewMatcher(buys, sells pqueue.Q, output *ResponseBuffer) *M {
	if buys.Kind() != trade.BUY {
		panic("Provided a buy priority queue that was not accepting buys!")
	}
	if sells.Kind() != trade.SELL {
		panic("Provided a sell priority queue that was not accepting sells!")
	}
	return &M{buys: buys, sells: sells, output: output}
}

func (m *M) AddSell(s *trade.Order) {
	if s.Kind != trade.SELL {
		panic("Added non-sell trade as a sell")
	}
	if !m.fillableSell(s) {
		m.sells.Push(s)
	}
}

func (m *M) AddBuy(b *trade.Order) {
	if b.Kind != trade.BUY {
		panic("Added non-buy trade as a buy")
	}
	if b.Price == trade.MARKET {
		panic("It is illegal to submit a buy at market price")
	}
	if !m.fillableBuy(b) {
		m.buys.Push(b)
	}
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
				m.sells.Pop()
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
				m.sells.Pop()
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
				m.sells.Pop()
				b.Amount -= amount
				m.completeTrade(b, s, price, amount)
				return true // The sell has been used up
			}
			if s.Amount > b.Amount {
				amount := b.Amount
				price := price(b.Price, s.Price)
				s.Amount -= amount
				m.completeTrade(b, s, price, amount)
				m.sells.Pop()
				continue
			}
			if s.Amount == b.Amount {
				amount := b.Amount
				price := price(b.Price, s.Price)
				m.completeTrade(b, s, price, amount)
				m.sells.Pop()
				return true // The sell has been used up
			}
		} else {
			return false
		}
	}
	panic("Unreachable")
}

func price(bPrice, sPrice int32) int32 {
	if sPrice == trade.MARKET {
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
