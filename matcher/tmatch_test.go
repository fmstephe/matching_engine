package matcher

import (
	"github.com/fmstephe/matching_engine/trade"
	"testing"
)

const (
	stockId = 1
	trader1 = 1
	trader2 = 2
	trader3 = 3
)

var tmatchOrderMaker = trade.NewOrderMaker()

type responseVals struct {
	price        int64
	amount       uint32
	tradeId      uint32
	counterParty uint32
}

func verifyResponse(t *testing.T, rc chan interface{}, vals responseVals) {
	i := <-rc
	r := i.(*trade.Response)
	price := vals.price
	amount := vals.amount
	tradeId := vals.tradeId
	counterParty := vals.counterParty
	if r.TradeId != tradeId {
		t.Errorf("Expecting %d trade-id, got %d instead", tradeId, r.TradeId)
	}
	if r.Amount != amount {
		t.Errorf("Expecting %d amount, got %d instead", amount, r.Amount)
	}
	if r.Price != price {
		t.Errorf("Expecting %d price, got %d instead", price, r.Price)
	}
	if r.CounterParty != counterParty {
		t.Errorf("Expecting %d counter party, got %d instead", counterParty, r.CounterParty)
	}
}

func TestPrice(t *testing.T) {
	testPrice(t, 1, 1, 1)
	testPrice(t, 2, 1, 1)
	testPrice(t, 3, 1, 2)
	testPrice(t, 4, 1, 2)
	testPrice(t, 5, 1, 3)
	testPrice(t, 6, 1, 3)
	testPrice(t, 20, 10, 15)
	testPrice(t, 21, 10, 15)
	testPrice(t, 22, 10, 16)
	testPrice(t, 23, 10, 16)
	testPrice(t, 24, 10, 17)
	testPrice(t, 25, 10, 17)
	testPrice(t, 26, 10, 18)
	testPrice(t, 27, 10, 18)
	testPrice(t, 28, 10, 19)
	testPrice(t, 29, 10, 19)
	testPrice(t, 30, 10, 20)
}

func testPrice(t *testing.T, bPrice, sPrice, expected int64) {
	result := price(bPrice, sPrice)
	if result != expected {
		t.Errorf("price(%d,%d) does not equal %d, got %d instead.", bPrice, sPrice, expected, result)
	}
}

// Basic test matches lonely buy/sell trade pair which match exactly
func TestSimpleMatch(t *testing.T) {
	submit := make(chan interface{}, 20)
	orders := make(chan *trade.OrderData, 20)
	m := NewMatcher(100)
	m.SetSubmit(submit)
	m.SetOrders(orders)
	go m.Run()
	addLowBuys(m, 5)
	addHighSells(m, 10)
	// Add Buy
	costData := trade.CostData{Price: 7, Amount: 1}
	tradeData := trade.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	b := &trade.OrderData{}
	b.WriteBuy(costData, tradeData)
	orders <- b
	// Add sell
	costData = trade.CostData{Price: 7, Amount: 1}
	tradeData = trade.TradeData{TraderId: trader2, TradeId: 2, StockId: stockId}
	s := &trade.OrderData{}
	s.WriteSell(costData, tradeData)
	orders <- s
	// Verify
	verifyResponse(t, submit, responseVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verifyResponse(t, submit, responseVals{price: 7, amount: 1, tradeId: 2, counterParty: trader1})
}

// Test matches one buy order to two separate sells
func TestDoubleSellMatch(t *testing.T) {
	submit := make(chan interface{}, 20)
	orders := make(chan *trade.OrderData, 20)
	m := NewMatcher(100)
	m.SetSubmit(submit)
	m.SetOrders(orders)
	go m.Run()
	addLowBuys(m, 5)
	addHighSells(m, 10)
	// Add Buy
	costData := trade.CostData{Price: 7, Amount: 2}
	tradeData := trade.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	b := &trade.OrderData{}
	b.WriteBuy(costData, tradeData)
	orders <- b
	// Add Sell
	costData = trade.CostData{Price: 7, Amount: 1}
	tradeData = trade.TradeData{TraderId: trader2, TradeId: 2, StockId: stockId}
	s1 := &trade.OrderData{}
	s1.WriteSell(costData, tradeData)
	orders <- s1
	// Verify
	verifyResponse(t, submit, responseVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verifyResponse(t, submit, responseVals{price: 7, amount: 1, tradeId: 2, counterParty: trader1})
	// Add Sell
	costData = trade.CostData{Price: 7, Amount: 1}
	tradeData = trade.TradeData{TraderId: trader3, TradeId: 3, StockId: stockId}
	s2 := &trade.OrderData{}
	s2.WriteSell(costData, tradeData)
	orders <- s2
	// Verify
	verifyResponse(t, submit, responseVals{price: -7, amount: 1, tradeId: 1, counterParty: trader3})
	verifyResponse(t, submit, responseVals{price: 7, amount: 1, tradeId: 3, counterParty: trader1})
}

// Test matches two buy orders to one sell
func TestDoubleBuyMatch(t *testing.T) {
	submit := make(chan interface{}, 20)
	orders := make(chan *trade.OrderData, 20)
	m := NewMatcher(100)
	m.SetSubmit(submit)
	m.SetOrders(orders)
	go m.Run()
	addLowBuys(m, 5)
	addHighSells(m, 10)
	// Add Sell
	costData := trade.CostData{Price: 7, Amount: 2}
	tradeData := trade.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	s := &trade.OrderData{}
	s.WriteSell(costData, tradeData)
	orders <- s
	// Add Buy
	costData = trade.CostData{Price: 7, Amount: 1}
	tradeData = trade.TradeData{TraderId: trader2, TradeId: 2, StockId: stockId}
	b1 := &trade.OrderData{}
	b1.WriteBuy(costData, tradeData)
	orders <- b1
	verifyResponse(t, submit, responseVals{price: -7, amount: 1, tradeId: 2, counterParty: trader1})
	verifyResponse(t, submit, responseVals{price: 7, amount: 1, tradeId: 1, counterParty: trader2})
	// Add Buy
	costData = trade.CostData{Price: 7, Amount: 1}
	tradeData = trade.TradeData{TraderId: trader3, TradeId: 3, StockId: stockId}
	b2 := &trade.OrderData{}
	b2.WriteBuy(costData, tradeData)
	orders <- b2
	verifyResponse(t, submit, responseVals{price: -7, amount: 1, tradeId: 3, counterParty: trader1})
	verifyResponse(t, submit, responseVals{price: 7, amount: 1, tradeId: 1, counterParty: trader3})
}

// Test matches lonely buy/sell pair, with same quantity, uses the mid-price point for trade price
func TestMidPrice(t *testing.T) {
	submit := make(chan interface{}, 20)
	orders := make(chan *trade.OrderData, 20)
	m := NewMatcher(100)
	m.SetSubmit(submit)
	m.SetOrders(orders)
	go m.Run()
	addLowBuys(m, 5)
	addHighSells(m, 10)
	// Add Buy
	costData := trade.CostData{Price: 9, Amount: 1}
	tradeData := trade.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	b := &trade.OrderData{}
	b.WriteBuy(costData, tradeData)
	orders <- b
	// Add Sell
	costData = trade.CostData{Price: 6, Amount: 1}
	tradeData = trade.TradeData{TraderId: trader2, TradeId: 1, StockId: stockId}
	s := &trade.OrderData{}
	s.WriteSell(costData, tradeData)
	orders <- s
	verifyResponse(t, submit, responseVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verifyResponse(t, submit, responseVals{price: 7, amount: 1, tradeId: 1, counterParty: trader1})
}

// Test matches lonely buy/sell pair, sell > quantity, and uses the mid-price point for trade price
func TestMidPriceBigSell(t *testing.T) {
	submit := make(chan interface{}, 20)
	orders := make(chan *trade.OrderData, 20)
	m := NewMatcher(100)
	m.SetSubmit(submit)
	m.SetOrders(orders)
	go m.Run()
	addLowBuys(m, 5)
	addHighSells(m, 10)
	// Add Buy
	costData := trade.CostData{Price: 9, Amount: 1}
	tradeData := trade.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	b := &trade.OrderData{}
	b.WriteBuy(costData, tradeData)
	orders <- b
	// Add Sell
	costData = trade.CostData{Price: 6, Amount: 10}
	tradeData = trade.TradeData{TraderId: trader2, TradeId: 1, StockId: stockId}
	s := &trade.OrderData{}
	s.WriteSell(costData, tradeData)
	orders <- s
	// Verify
	verifyResponse(t, submit, responseVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verifyResponse(t, submit, responseVals{price: 7, amount: 1, tradeId: 1, counterParty: trader1})
}

// Test matches lonely buy/sell pair, buy > quantity, and uses the mid-price point for trade price
func TestMidPriceBigBuy(t *testing.T) {
	submit := make(chan interface{}, 20)
	orders := make(chan *trade.OrderData, 20)
	m := NewMatcher(100)
	m.SetSubmit(submit)
	m.SetOrders(orders)
	go m.Run()
	addLowBuys(m, 5)
	addHighSells(m, 10)
	// Add Buy
	costData := trade.CostData{Price: 9, Amount: 10}
	tradeData := trade.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	b := &trade.OrderData{}
	b.WriteBuy(costData, tradeData)
	orders <- b
	// Add Sell
	costData = trade.CostData{Price: 6, Amount: 1}
	tradeData = trade.TradeData{TraderId: trader2, TradeId: 1, StockId: stockId}
	s := &trade.OrderData{}
	s.WriteSell(costData, tradeData)
	orders <- s
	verifyResponse(t, submit, responseVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verifyResponse(t, submit, responseVals{price: 7, amount: 1, tradeId: 1, counterParty: trader1})
}

func addLowBuys(m *M, highestPrice int64) {
	buys := tmatchOrderMaker.MkBuys(tmatchOrderMaker.ValRangeFlat(10, 1, highestPrice))
	for _, buy := range buys {
		m.orders <- &buy
	}
}

func addHighSells(m *M, lowestPrice int64) {
	sells := tmatchOrderMaker.MkSells(tmatchOrderMaker.ValRangeFlat(10, lowestPrice, lowestPrice+10000))
	for _, sell := range sells {
		m.orders <- &sell
	}
}
