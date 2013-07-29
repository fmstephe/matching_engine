package matcher

import (
	. "github.com/fmstephe/matching_engine/msg"
	"testing"
)

var suiteMaker = NewMessageMaker(100)

type MatchTester interface {
	Send(*testing.T, *Message)
	Expect(*testing.T, *Message)
	Cleanup(*testing.T)
}

type MatchTesterMaker interface {
	Make() MatchTester
}

func RunTestSuite(t *testing.T, mkr MatchTesterMaker) {
	testSellBuyMatch(t, mkr)
	testBuySellMatch(t, mkr)
	testBuyDoubleSellMatch(t, mkr)
	testSellDoubleBuyMatch(t, mkr)
	testMidPrice(t, mkr)
	testMidPriceBigSell(t, mkr)
	testMidPriceBigBuy(t, mkr)
	testTradeSeparateStocks(t, mkr)
	testSeparateStocksNotMatched(t, mkr)
	testSellCancelBuyNoMatch(t, mkr)
	testBuyCancelSellNoMatch(t, mkr)
	testBadCancelNotCancelled(t, mkr)
}

func testSellBuyMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Sell
	s := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s)
	// Add Buy
	b := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b)
	// Full match
	es := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	mt.Expect(t, es)
	eb := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	eb.Route = APP
	eb.Kind = FULL
	mt.Expect(t, eb)
}

func testBuySellMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b)
	// Add Sell
	s := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s)
	// Full match
	eb := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	mt.Expect(t, eb)
	es := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es)
}

func testBuyDoubleSellMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 2}
	mt.Send(t, b)
	// Add Sell1
	s1 := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s1)
	// Full match
	eb1 := &Message{Route: APP, Direction: IN, Kind: PARTIAL, TraderId: 1, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	mt.Expect(t, eb1)
	es1 := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es1)
	// Add Sell2
	s2 := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s2)
	// Full Match II
	eb2 := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	mt.Expect(t, eb2)
	es2 := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 2, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es2)
}

func testSellDoubleBuyMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Sell
	s := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 2}
	mt.Send(t, s)
	// Add Buy1
	b1 := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b1)
	// Full match on the buy
	eb1 := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	mt.Expect(t, eb1)
	// Partial match on the sell
	es1 := &Message{Route: APP, Direction: IN, Kind: PARTIAL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es1)
	// Add Buy2
	b2 := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 2, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b2)
	// Full Match II
	eb2 := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 2, TradeId: 2, StockId: 1, Price: -7, Amount: 1}
	mt.Expect(t, eb2)
	es2 := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es2)
}

func testMidPrice(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 9, Amount: 1}
	mt.Send(t, b)
	// Add Sell
	s := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 6, Amount: 1}
	mt.Send(t, s)
	// Full match
	eb := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	mt.Expect(t, eb)
	es := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es)
}

func testMidPriceBigSell(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 9, Amount: 1}
	mt.Send(t, b)
	// Add Sell
	s := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 6, Amount: 10}
	mt.Send(t, s)
	// Full match
	eb := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	mt.Expect(t, eb)
	es := &Message{Route: APP, Direction: IN, Kind: PARTIAL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es)
}

func testMidPriceBigBuy(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 9, Amount: 10}
	mt.Send(t, b)
	// Add Sell
	s := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 6, Amount: 1}
	mt.Send(t, s)
	// Full match
	eb := &Message{Route: APP, Direction: IN, Kind: PARTIAL, TraderId: 1, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	mt.Expect(t, eb)
	es := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es)
}

func testTradeSeparateStocks(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Sell stock 1
	s1 := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s1)
	// Add Buy stock 2
	b2 := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 1, TradeId: 1, StockId: 2, Price: 7, Amount: 1}
	mt.Send(t, b2)
	// Add Sell stock 2
	s2 := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 2, StockId: 2, Price: 7, Amount: 1}
	mt.Send(t, s2)
	// Full match stock 2
	es2 := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 1, TradeId: 1, StockId: 2, Price: -7, Amount: 1}
	mt.Expect(t, es2)
	eb2 := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 2, TradeId: 2, StockId: 2, Price: 7, Amount: 1}
	mt.Expect(t, eb2)
	// Add Buy stock 1
	b1 := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 1, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b1)
	// Full match stock 1
	eb1 := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 1, TradeId: 2, StockId: 1, Price: -7, Amount: 1}
	mt.Expect(t, eb1)
	es1 := &Message{Route: APP, Direction: IN, Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es1)
}

func testSeparateStocksNotMatched(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Sell
	s1 := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s1)
	// Add Sell
	// Add Buy
	b1 := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 1, TradeId: 1, StockId: 2, Price: 7, Amount: 1}
	mt.Send(t, b1)
	// Expect empty message flushing
	// Add Trader 2 order to flush trader 2 messages
	s2 := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 2, StockId: 1, Price: 70, Amount: 1}
	mt.Send(t, s2)
	// Add Trader 1 order to flush trader 1 messages
	b2 := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 1, TradeId: 2, StockId: 1, Price: 1, Amount: 1}
	mt.Send(t, b2)
}

func testSellCancelBuyNoMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Sell
	s := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s)
	// Cancel Sell
	cs := &Message{Route: APP, Direction: OUT, Kind: CANCEL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, cs)
	// Expect Cancelled
	ec := &Message{Route: APP, Direction: IN, Kind: CANCELLED, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, ec)
	// Add Buy
	b := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b)
	// Add Trader 1 order to flush trader 1 messages
	b2 := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 1, TradeId: 2, StockId: 1, Price: 1, Amount: 1}
	mt.Send(t, b2)
	// Add Trader 2 order to flush trader 2 messages
	s2 := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 2, StockId: 1, Price: 70, Amount: 1}
	mt.Send(t, s2)
}

func testBuyCancelSellNoMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b)
	// Cancel Buy
	cb := &Message{Route: APP, Direction: OUT, Kind: CANCEL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, cb)
	// Expect Cancelled
	ec := &Message{Route: APP, Direction: IN, Kind: CANCELLED, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, ec)
	// Add Sell
	s := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s)
	// Expect empty message flushing
	// Add Trader 1 order to flush trader 1 messages
	b2 := &Message{Route: APP, Direction: OUT, Kind: BUY, TraderId: 1, TradeId: 2, StockId: 1, Price: 1, Amount: 1}
	mt.Send(t, b2)
	// Add Trader 2 order to flush trader 2 messages
	s2 := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 2, TradeId: 2, StockId: 1, Price: 70, Amount: 1}
	mt.Send(t, s2)
}

func testBadCancelNotCancelled(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Cancel Buy that doesn't exist
	cb := &Message{Route: APP, Direction: OUT, Kind: CANCEL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, cb)
	// Expect Not Cancelled
	ec := &Message{Route: APP, Direction: IN, Kind: NOT_CANCELLED, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, ec)
}

func addLowBuys(t *testing.T, mt MatchTester, highestPrice int64, stockId uint32) {
	buys := suiteMaker.MkBuys(suiteMaker.ValRangeFlat(10, 1, highestPrice), stockId)
	for i := range buys {
		mt.Send(t, &buys[i])
	}
}

func addHighSells(t *testing.T, mt MatchTester, lowestPrice int64, stockId uint32) {
	sells := suiteMaker.MkSells(suiteMaker.ValRangeFlat(10, lowestPrice, lowestPrice+10000), stockId)
	for i := range sells {
		mt.Send(t, &sells[i])
	}
}
