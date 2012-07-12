package matcher

import (
	"testing"
	"github.com/fmstephe/matching_engine/trade"
)

const (
	stockId = "Stock1"
	trader1 = "trader1"
	trader2 = "trader2"
)

func TestSimpleMatch(t *testing.T) {
	m := New(stockId)
	trader1Chan := make(chan *trade.Response, 256)
	trader2Chan := make(chan *trade.Response, 256)
	b := trade.NewBuy(1, 1, 1, stockId, trader1, trader1Chan)
	s := trade.NewSell(2, 1, 1, stockId, trader2, trader2Chan)
	m.AddBuy(b)
	m.AddSell(s)
	rb := <- trader1Chan
	rs := <- trader2Chan
	if rb.TradeId != 1 || rb.Amount != 1 || rb.Price != 1 || rb.CounterParty != trader2 {
		t.Errorf("Buy response broken - %v", rb)
	}
	if rs.TradeId != 2 || rs.Amount != 1 || rs.Price != 1 || rs.CounterParty != trader1 {
		t.Errorf("Sell response broken - %v", rs)
	}
}
