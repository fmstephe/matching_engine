package matcher

import (
	"fmt"
)

type M struct {
	buys, sells *heap
	stockId     uint32
}

func NewMatcher(stockId uint32) *M {
	buys := newHeap(BUY)
	sells := newHeap(SELL)
	return &M{buys: buys, sells: sells, stockId: stockId}
}

func (m *M) AddSell(s *Order) {
	if s.BuySell != SELL {
		panic("Added non-sell trade as a sell")
	}
	if s.StockId != m.stockId {
		panic(fmt.Sprintf("Added sell trade with stock-id %s expecting %s", s.StockId, m.stockId))
	}
	m.sells.push(s)
	m.process()
}

func (m *M) AddBuy(b *Order) {
	if b.BuySell != BUY {
		panic("Added non-buy trade as a buy")
	}
	if b.StockId != m.stockId {
		panic(fmt.Sprintf("Added buy trade with stock-id %s expecting %s", b.StockId, m.stockId))
	}
	if b.Price == MarketPrice {
		panic("It is illegal to submit a buy at market price")
	}
	m.buys.push(b)
	m.process()
}

func (m *M) process() {
	for {
		if m.buys.heapLen() == 0 || m.sells.heapLen() == 0 {
			return
		}
		b := m.buys.peek()
		s := m.sells.peek()
		if b.Price >= s.Price {
			if b.Amount > s.Amount {
				amount := s.Amount
				price := price(b.Price, s.Price)
				m.sells.pop()
				b.Amount -= amount // Dangerous in place modification
				completeTrade(b, s, price, amount)
				continue
			}
			if s.Amount > b.Amount {
				amount := b.Amount
				price := price(b.Price, s.Price)
				m.buys.pop()
				s.Amount -= amount // Dangerous in place modification
				completeTrade(b, s, price, amount)
				continue
			}
			if s.Amount == b.Amount {
				amount := s.Amount
				price := price(b.Price, s.Price)
				m.buys.pop()
				m.sells.pop()
				completeTrade(b, s, price, amount)
				continue
			}
		} else {
			return
		}
	}
}

func price(bPrice, sPrice int64) int64 {
	if sPrice == MarketPrice {
		return bPrice
	}
	d := bPrice - sPrice
	return sPrice + (d >> 1)
}

func completeTrade(b, s *Order, price int64, amount uint32) {
	b.ResponseFunc(NewResponse(-price, amount, b.TradeId, s.TraderId))
	s.ResponseFunc(NewResponse(price, amount, s.TradeId, b.TraderId))
}
