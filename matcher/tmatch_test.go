package matcher

import (
	"github.com/fmstephe/matching_engine/msg"
	"testing"
)

const (
	stockId = 1
	trader1 = 1
	trader2 = 2
	trader3 = 3
)

var matchMaker = msg.NewMessageMaker()

type responseVals struct {
	price   int64
	amount  uint32
	tradeId uint32
	stockId uint32
}

func verifyMessage(t *testing.T, rc chan *msg.Message, vals responseVals) {
	r := <-rc
	price := vals.price
	amount := vals.amount
	tradeId := vals.tradeId
	stockId := vals.stockId
	if r.TradeId != tradeId {
		t.Errorf("Expecting %d trade-id, got %d instead", tradeId, r.TradeId)
	}
	if r.Amount != amount {
		t.Errorf("Expecting %d amount, got %d instead", amount, r.Amount)
	}
	if r.Price != price {
		t.Errorf("Expecting %d price, got %d instead", price, r.Price)
	}
	if r.StockId != stockId {
		t.Errorf("Expecting %d stock id, got %d instead", stockId, r.StockId)
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
	submit := make(chan *msg.Message, 20)
	orders := make(chan *msg.Message, 20)
	m := NewMatcher(100)
	m.SetSubmit(submit)
	m.SetOrders(orders)
	go m.Run()
	addLowBuys(m, 5)
	addHighSells(m, 10)
	// Add Buy
	costData := msg.CostData{Price: 7, Amount: 1}
	tradeData := msg.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	b := &msg.Message{}
	b.WriteBuy(costData, tradeData, msg.NetData{})
	orders <- b
	// Add sell
	costData = msg.CostData{Price: 7, Amount: 1}
	tradeData = msg.TradeData{TraderId: trader2, TradeId: 2, StockId: stockId}
	s := &msg.Message{}
	s.WriteSell(costData, tradeData, msg.NetData{})
	orders <- s
	// Verify
	verifyMessage(t, submit, responseVals{price: -7, amount: 1, tradeId: 1, stockId: stockId})
	verifyMessage(t, submit, responseVals{price: 7, amount: 1, tradeId: 2, stockId: stockId})
}

// Test matches one buy order to two separate sells
func TestDoubleSellMatch(t *testing.T) {
	submit := make(chan *msg.Message, 20)
	orders := make(chan *msg.Message, 20)
	m := NewMatcher(100)
	m.SetSubmit(submit)
	m.SetOrders(orders)
	go m.Run()
	addLowBuys(m, 5)
	addHighSells(m, 10)
	// Add Buy
	costData := msg.CostData{Price: 7, Amount: 2}
	tradeData := msg.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	b := &msg.Message{}
	b.WriteBuy(costData, tradeData, msg.NetData{})
	orders <- b
	// Add Sell
	costData = msg.CostData{Price: 7, Amount: 1}
	tradeData = msg.TradeData{TraderId: trader2, TradeId: 2, StockId: stockId}
	s1 := &msg.Message{}
	s1.WriteSell(costData, tradeData, msg.NetData{})
	orders <- s1
	// Verify
	verifyMessage(t, submit, responseVals{price: -7, amount: 1, tradeId: 1, stockId: stockId})
	verifyMessage(t, submit, responseVals{price: 7, amount: 1, tradeId: 2, stockId: stockId})
	// Add Sell
	costData = msg.CostData{Price: 7, Amount: 1}
	tradeData = msg.TradeData{TraderId: trader3, TradeId: 3, StockId: stockId}
	s2 := &msg.Message{}
	s2.WriteSell(costData, tradeData, msg.NetData{})
	orders <- s2
	// Verify
	verifyMessage(t, submit, responseVals{price: -7, amount: 1, tradeId: 1, stockId: stockId})
	verifyMessage(t, submit, responseVals{price: 7, amount: 1, tradeId: 3, stockId: stockId})
}

// Test matches two buy orders to one sell
func TestDoubleBuyMatch(t *testing.T) {
	submit := make(chan *msg.Message, 20)
	orders := make(chan *msg.Message, 20)
	m := NewMatcher(100)
	m.SetSubmit(submit)
	m.SetOrders(orders)
	go m.Run()
	addLowBuys(m, 5)
	addHighSells(m, 10)
	// Add Sell
	costData := msg.CostData{Price: 7, Amount: 2}
	tradeData := msg.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	s := &msg.Message{}
	s.WriteSell(costData, tradeData, msg.NetData{})
	orders <- s
	// Add Buy
	costData = msg.CostData{Price: 7, Amount: 1}
	tradeData = msg.TradeData{TraderId: trader2, TradeId: 2, StockId: stockId}
	b1 := &msg.Message{}
	b1.WriteBuy(costData, tradeData, msg.NetData{})
	orders <- b1
	verifyMessage(t, submit, responseVals{price: -7, amount: 1, tradeId: 2, stockId: stockId})
	verifyMessage(t, submit, responseVals{price: 7, amount: 1, tradeId: 1, stockId: stockId})
	// Add Buy
	costData = msg.CostData{Price: 7, Amount: 1}
	tradeData = msg.TradeData{TraderId: trader3, TradeId: 3, StockId: stockId}
	b2 := &msg.Message{}
	b2.WriteBuy(costData, tradeData, msg.NetData{})
	orders <- b2
	verifyMessage(t, submit, responseVals{price: -7, amount: 1, tradeId: 3, stockId: stockId})
	verifyMessage(t, submit, responseVals{price: 7, amount: 1, tradeId: 1, stockId: stockId})
}

// Test matches lonely buy/sell pair, with same quantity, uses the mid-price point for trade price
func TestMidPrice(t *testing.T) {
	submit := make(chan *msg.Message, 20)
	orders := make(chan *msg.Message, 20)
	m := NewMatcher(100)
	m.SetSubmit(submit)
	m.SetOrders(orders)
	go m.Run()
	addLowBuys(m, 5)
	addHighSells(m, 10)
	// Add Buy
	costData := msg.CostData{Price: 9, Amount: 1}
	tradeData := msg.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	b := &msg.Message{}
	b.WriteBuy(costData, tradeData, msg.NetData{})
	orders <- b
	// Add Sell
	costData = msg.CostData{Price: 6, Amount: 1}
	tradeData = msg.TradeData{TraderId: trader2, TradeId: 1, StockId: stockId}
	s := &msg.Message{}
	s.WriteSell(costData, tradeData, msg.NetData{})
	orders <- s
	verifyMessage(t, submit, responseVals{price: -7, amount: 1, tradeId: 1, stockId: stockId})
	verifyMessage(t, submit, responseVals{price: 7, amount: 1, tradeId: 1, stockId: stockId})
}

// Test matches lonely buy/sell pair, sell > quantity, and uses the mid-price point for trade price
func TestMidPriceBigSell(t *testing.T) {
	submit := make(chan *msg.Message, 20)
	orders := make(chan *msg.Message, 20)
	m := NewMatcher(100)
	m.SetSubmit(submit)
	m.SetOrders(orders)
	go m.Run()
	addLowBuys(m, 5)
	addHighSells(m, 10)
	// Add Buy
	costData := msg.CostData{Price: 9, Amount: 1}
	tradeData := msg.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	b := &msg.Message{}
	b.WriteBuy(costData, tradeData, msg.NetData{})
	orders <- b
	// Add Sell
	costData = msg.CostData{Price: 6, Amount: 10}
	tradeData = msg.TradeData{TraderId: trader2, TradeId: 1, StockId: stockId}
	s := &msg.Message{}
	s.WriteSell(costData, tradeData, msg.NetData{})
	orders <- s
	// Verify
	verifyMessage(t, submit, responseVals{price: -7, amount: 1, tradeId: 1, stockId: stockId})
	verifyMessage(t, submit, responseVals{price: 7, amount: 1, tradeId: 1, stockId: stockId})
}

// Test matches lonely buy/sell pair, buy > quantity, and uses the mid-price point for trade price
func TestMidPriceBigBuy(t *testing.T) {
	submit := make(chan *msg.Message, 20)
	orders := make(chan *msg.Message, 20)
	m := NewMatcher(100)
	m.SetSubmit(submit)
	m.SetOrders(orders)
	go m.Run()
	addLowBuys(m, 5)
	addHighSells(m, 10)
	// Add Buy
	costData := msg.CostData{Price: 9, Amount: 10}
	tradeData := msg.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	b := &msg.Message{}
	b.WriteBuy(costData, tradeData, msg.NetData{})
	orders <- b
	// Add Sell
	costData = msg.CostData{Price: 6, Amount: 1}
	tradeData = msg.TradeData{TraderId: trader2, TradeId: 1, StockId: stockId}
	s := &msg.Message{}
	s.WriteSell(costData, tradeData, msg.NetData{})
	orders <- s
	verifyMessage(t, submit, responseVals{price: -7, amount: 1, tradeId: 1, stockId: stockId})
	verifyMessage(t, submit, responseVals{price: 7, amount: 1, tradeId: 1, stockId: stockId})
}

func addLowBuys(m *M, highestPrice int64) {
	buys := matchMaker.MkBuys(matchMaker.ValRangeFlat(10, 1, highestPrice))
	for _, buy := range buys {
		m.orders <- &buy
	}
}

func addHighSells(m *M, lowestPrice int64) {
	sells := matchMaker.MkSells(matchMaker.ValRangeFlat(10, lowestPrice, lowestPrice+10000))
	for _, sell := range sells {
		m.orders <- &sell
	}
}
