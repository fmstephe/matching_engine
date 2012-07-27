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

type verifyVals struct {
	price        int64
	amount       uint32
	tradeId      uint32
	counterParty uint32
}

func verify(t *testing.T, r *trade.Response, vals verifyVals) {
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

func responseChan() chan *trade.Response {
	return make(chan *trade.Response, 256)
}

func responseFunc(rc chan *trade.Response) func(*trade.Response) {
	return func(response *trade.Response) {
		rc <- response
	}
}

func TestMidPoint(t *testing.T) {
	midpoint(t, 1, 1, 1)
	midpoint(t, 2, 1, 1)
	midpoint(t, 3, 1, 2)
	midpoint(t, 4, 1, 2)
	midpoint(t, 5, 1, 3)
	midpoint(t, 6, 1, 3)
	midpoint(t, 20, 10, 15)
	midpoint(t, 21, 10, 15)
	midpoint(t, 22, 10, 16)
	midpoint(t, 23, 10, 16)
	midpoint(t, 24, 10, 17)
	midpoint(t, 25, 10, 17)
	midpoint(t, 26, 10, 18)
	midpoint(t, 27, 10, 18)
	midpoint(t, 28, 10, 19)
	midpoint(t, 29, 10, 19)
	midpoint(t, 30, 10, 20)
}

func midpoint(t *testing.T, bPrice, sPrice, expected int64) {
	result := price(bPrice, sPrice)
	if result != expected {
		t.Errorf("price(%d,%d) does not equal %d, got %d instead.", bPrice, sPrice, expected, result)
	}
}

// Basic test matches lonely buy/sell trade pair which match exactly
func TestSimpleMatch(t *testing.T) {
	m := New(stockId)
	addLowBuys(m, 5)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	// Add Buy
	costData := trade.CostData{Price: 7, Amount: 1}
	tradeData := trade.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	m.AddBuy(trade.NewBuy(costData, tradeData, responseFunc(trader1Chan)))
	// Add sell
	costData = trade.CostData{Price: 7, Amount: 1}
	tradeData = trade.TradeData{TraderId: trader2, TradeId: 2, StockId: stockId}
	m.AddSell(trade.NewSell(costData, tradeData, responseFunc(trader2Chan)))
	// Verify
	verify(t, <-trader1Chan, verifyVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verify(t, <-trader2Chan, verifyVals{price: 7, amount: 1, tradeId: 2, counterParty: trader1})
}

// Test matches one buy order to two separate sells
func TestDoubleSellMatch(t *testing.T) {
	m := New(stockId)
	addLowBuys(m, 5)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	trader3Chan := responseChan()
	// Add Buy
	costData := trade.CostData{Price: 7, Amount: 2}
	tradeData := trade.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	m.AddBuy(trade.NewBuy(costData, tradeData, responseFunc(trader1Chan)))
	// Add Sell
	costData = trade.CostData{Price: 7, Amount: 1}
	tradeData = trade.TradeData{TraderId: trader2, TradeId: 2, StockId: stockId}
	m.AddSell(trade.NewSell(costData, tradeData, responseFunc(trader2Chan)))
	// Verify
	verify(t, <-trader1Chan, verifyVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verify(t, <-trader2Chan, verifyVals{price: 7, amount: 1, tradeId: 2, counterParty: trader1})
	// Add Sell
	costData = trade.CostData{Price: 7, Amount: 1}
	tradeData = trade.TradeData{TraderId: trader3, TradeId: 3, StockId: stockId}
	m.AddSell(trade.NewSell(costData, tradeData, responseFunc(trader3Chan)))
	// Verify
	verify(t, <-trader1Chan, verifyVals{price: -7, amount: 1, tradeId: 1, counterParty: trader3})
	verify(t, <-trader3Chan, verifyVals{price: 7, amount: 1, tradeId: 3, counterParty: trader1})
}

// Test matches two buy orders to one sell
func TestDoubleBuyMatch(t *testing.T) {
	m := New(stockId)
	addLowBuys(m, 5)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	trader3Chan := responseChan()
	// Add Sell
	costData := trade.CostData{Price: 7, Amount: 2}
	tradeData := trade.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	m.AddSell(trade.NewSell(costData, tradeData, responseFunc(trader1Chan)))
	// Add Buy
	costData = trade.CostData{Price: 7, Amount: 1}
	tradeData = trade.TradeData{TraderId: trader2, TradeId: 2, StockId: stockId}
	m.AddBuy(trade.NewBuy(costData, tradeData, responseFunc(trader2Chan)))
	verify(t, <-trader1Chan, verifyVals{price: 7, amount: 1, tradeId: 1, counterParty: trader2})
	verify(t, <-trader2Chan, verifyVals{price: -7, amount: 1, tradeId: 2, counterParty: trader1})
	// Add Buy
	costData = trade.CostData{Price: 7, Amount: 1}
	tradeData = trade.TradeData{TraderId: trader3, TradeId: 3, StockId: stockId}
	m.AddBuy(trade.NewBuy(costData, tradeData, responseFunc(trader3Chan)))
	verify(t, <-trader1Chan, verifyVals{price: 7, amount: 1, tradeId: 1, counterParty: trader3})
	verify(t, <-trader3Chan, verifyVals{price: -7, amount: 1, tradeId: 3, counterParty: trader1})
}

// Test matches lonely buy/sell pair, with same quantity, uses the mid-price point for trade price
func TestMidPrice(t *testing.T) {
	m := New(stockId)
	addLowBuys(m, 5)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	// Add Buy
	costData := trade.CostData{Price: 9, Amount: 1}
	tradeData := trade.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	m.AddBuy(trade.NewBuy(costData, tradeData, responseFunc(trader1Chan)))
	// Add Sell
	costData = trade.CostData{Price: 6, Amount: 1}
	tradeData = trade.TradeData{TraderId: trader2, TradeId: 1, StockId: stockId}
	m.AddSell(trade.NewSell(costData, tradeData, responseFunc(trader2Chan)))
	verify(t, <-trader1Chan, verifyVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verify(t, <-trader2Chan, verifyVals{price: 7, amount: 1, tradeId: 1, counterParty: trader1})
}

// Test matches lonely buy/sell pair, sell > quantity, and uses the mid-price point for trade price
func TestMidPriceBigSell(t *testing.T) {
	m := New(stockId)
	addLowBuys(m, 5)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	// Add Buy
	costData := trade.CostData{Price: 9, Amount: 1}
	tradeData := trade.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	m.AddBuy(trade.NewBuy(costData, tradeData, responseFunc(trader1Chan)))
	// Add Sell
	costData = trade.CostData{Price: 6, Amount: 10}
	tradeData = trade.TradeData{TraderId: trader2, TradeId: 1, StockId: stockId}
	m.AddSell(trade.NewSell(costData, tradeData, responseFunc(trader2Chan)))
	// Verify
	verify(t, <-trader1Chan, verifyVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verify(t, <-trader2Chan, verifyVals{price: 7, amount: 1, tradeId: 1, counterParty: trader1})
}

// Test matches lonely buy/sell pair, buy > quantity, and uses the mid-price point for trade price
func TestMidPriceBigBuy(t *testing.T) {
	m := New(stockId)
	addLowBuys(m, 5)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	// Add Buy
	costData := trade.CostData{Price: 9, Amount: 10}
	tradeData := trade.TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	m.AddBuy(trade.NewBuy(costData, tradeData, responseFunc(trader1Chan)))
	// Add Sell
	costData = trade.CostData{Price: 6, Amount: 1}
	tradeData = trade.TradeData{TraderId: trader2, TradeId: 1, StockId: stockId}
	m.AddSell(trade.NewSell(costData, tradeData, responseFunc(trader2Chan)))
	verify(t, <-trader1Chan, verifyVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verify(t, <-trader2Chan, verifyVals{price: 7, amount: 1, tradeId: 1, counterParty: trader1})
}

func addLowBuys(m *M, highestPrice int64) {
	buys := mkBuys(10, 1, highestPrice)
	for _, buy := range buys {
		m.AddBuy(buy)
	}
}

func addHighSells(m *M, lowestPrice int64) {
	sells := mkSells(10, lowestPrice, lowestPrice+10000)
	for _, sell := range sells {
		m.AddSell(sell)
	}
}
