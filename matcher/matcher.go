package matcher

import (
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
	"github.com/fmstephe/heap"
)

type M struct {
	buys, sells heap.Interface
	stockId string
}

func New(stockId string) *M {
	buys := heap.New()
	sells := heap.New()
	return &M{buys: buys, sells: sells, stockId: stockId}
}

func (m *M) AddSell(s *trade.Order) {
	if s.BuySell != trade.SELL {
		panic("Added non-sell trade as a sell")
	}
	if s.StockId != m.stockId {
		panic(fmt.Sprintf("Added sell trade with stock-id %s expecting %s", s.StockId, m.stockId))
	}
	m.sells.Push(s)
	m.process()
}

func (m *M) AddBuy(b *trade.Order) {
	if b.BuySell != trade.BUY {
		panic("Added non-buy trade as a buy")
	}
	if b.StockId != m.stockId {
		panic(fmt.Sprintf("Added buy trade with stock-id %s expecting %s", b.StockId, m.stockId))
	}
	m.buys.Push(b)
	m.process()
}

// Currently when a match is made we favour the seller by always matching at the buyer's limit price
// Does not currently allow for 'market' trades - will do in the future 
func (m *M) process() {
	for {
		if m.buys.Len() == 0 || m.sells.Len() == 0 {
			return
		}
		b := m.buys.Peek().(*trade.Order)
		s := m.sells.Peek().(*trade.Order)
		if b.Price >= s.Price {
			if b.Amount > s.Amount {
				amount := s.Amount
				price := b.Price
				m.sells.Pop()
				b.Amount -= amount // Dangerous in place modification
				completeTrade(b,s,amount, price)
				continue
			}
			if s.Amount > b.Amount {
				amount := b.Amount
				price := b.Price
				m.buys.Pop()
				s.Amount -= amount // Dangerous in place modification
				completeTrade(b,s,amount, price)
				continue
			}
			if s.Amount == b.Amount {
				amount := s.Amount
				price := b.Price
				m.buys.Pop()
				m.sells.Pop()
				completeTrade(b,s,amount, price)
				continue
			}
		} else {
			return
		}
	}
}

func completeTrade(b, s *trade.Order, amount, price int64) {
	b.ResponseChan <- trade.NewResponse(b.TradeId, amount, price, s.Trader)
	s.ResponseChan <- trade.NewResponse(s.TradeId, amount, price, b.Trader)
}
