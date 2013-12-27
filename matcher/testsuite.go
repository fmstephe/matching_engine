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
	testTradeSeparateStocksI(t, mkr)
	testTradeSeparateStocksII(t, mkr)
	testSellCancelBuyNoMatch(t, mkr)
	testBuyCancelSellNoMatch(t, mkr)
	testBadCancelNotCancelled(t, mkr)
	testThreeBuysMatchedToOneSell(t, mkr)
}

func testSellBuyMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Sell
	s := &Message{Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s)
	// Add Buy
	b := &Message{Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b)
	// Full match
	es := &Message{Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es)
	eb := &Message{Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb)
}

func testBuySellMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &Message{Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b)
	// Add Sell
	s := &Message{Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s)
	// Full match
	eb := &Message{Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb)
	es := &Message{Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es)
}

func testBuyDoubleSellMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &Message{Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 2}
	mt.Send(t, b)
	// Add Sell1
	s1 := &Message{Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s1)
	// Full match
	eb1 := &Message{Kind: PARTIAL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb1)
	es1 := &Message{Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es1)
	// Add Sell2
	s2 := &Message{Kind: SELL, TraderId: 2, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s2)
	// Full Match II
	eb2 := &Message{Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb2)
	es2 := &Message{Kind: FULL, TraderId: 2, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es2)
}

func testSellDoubleBuyMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Sell
	s := &Message{Kind: SELL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 2}
	mt.Send(t, s)
	// Add Buy1
	b1 := &Message{Kind: BUY, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b1)
	// Full match on the buy
	eb1 := &Message{Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb1)
	// Partial match on the sell
	es1 := &Message{Kind: PARTIAL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es1)
	// Add Buy2
	b2 := &Message{Kind: BUY, TraderId: 2, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b2)
	// Full Match II
	eb2 := &Message{Kind: FULL, TraderId: 2, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb2)
	es2 := &Message{Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es2)
}

func testMidPrice(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &Message{Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 9, Amount: 1}
	mt.Send(t, b)
	// Add Sell
	s := &Message{Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 6, Amount: 1}
	mt.Send(t, s)
	// Full match
	eb := &Message{Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb)
	es := &Message{Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es)
}

func testMidPriceBigSell(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &Message{Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 9, Amount: 1}
	mt.Send(t, b)
	// Add Sell
	s := &Message{Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 6, Amount: 10}
	mt.Send(t, s)
	// Full match
	eb := &Message{Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb)
	es := &Message{Kind: PARTIAL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es)
}

func testMidPriceBigBuy(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &Message{Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 9, Amount: 10}
	mt.Send(t, b)
	// Add Sell
	s := &Message{Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 6, Amount: 1}
	mt.Send(t, s)
	// Full match
	eb := &Message{Kind: PARTIAL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb)
	es := &Message{Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es)
}

func testTradeSeparateStocksI(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Sell stock 1
	s1 := &Message{Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s1)
	// Add Buy stock 2
	b2 := &Message{Kind: BUY, TraderId: 1, TradeId: 1, StockId: 2, Price: 7, Amount: 1}
	mt.Send(t, b2)
	// Add Sell stock 2
	s2 := &Message{Kind: SELL, TraderId: 2, TradeId: 2, StockId: 2, Price: 7, Amount: 1}
	mt.Send(t, s2)
	// Full match stock 2
	es2 := &Message{Kind: FULL, TraderId: 1, TradeId: 1, StockId: 2, Price: 7, Amount: 1}
	mt.Expect(t, es2)
	eb2 := &Message{Kind: FULL, TraderId: 2, TradeId: 2, StockId: 2, Price: 7, Amount: 1}
	mt.Expect(t, eb2)
	// Add Buy stock 1
	b1 := &Message{Kind: BUY, TraderId: 1, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b1)
	// Full match stock 1
	eb1 := &Message{Kind: FULL, TraderId: 1, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb1)
	es1 := &Message{Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es1)
}

func testTradeSeparateStocksII(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Sell stock 1
	s1 := &Message{Kind: SELL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s1)
	// Add Buy stock 2
	b1 := &Message{Kind: BUY, TraderId: 2, TradeId: 1, StockId: 2, Price: 7, Amount: 1}
	mt.Send(t, b1)
	// Add buy stock 1
	s2 := &Message{Kind: BUY, TraderId: 3, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s2)
	// Expect match on stock 1
	eb1 := &Message{Kind: FULL, TraderId: 3, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb1)
	es1 := &Message{Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es1)
	// Add sell stock 2
	b2 := &Message{Kind: SELL, TraderId: 4, TradeId: 1, StockId: 2, Price: 7, Amount: 1}
	mt.Send(t, b2)
	// Expect match on stock 2
	eb2 := &Message{Kind: FULL, TraderId: 2, TradeId: 1, StockId: 2, Price: 7, Amount: 1}
	mt.Expect(t, eb2)
	es2 := &Message{Kind: FULL, TraderId: 4, TradeId: 1, StockId: 2, Price: 7, Amount: 1}
	mt.Expect(t, es2)
}

func testSellCancelBuyNoMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Sell
	s := &Message{Kind: SELL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s)
	// Cancel Sell
	cs := &Message{Kind: CANCEL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, cs)
	// Expect Cancelled
	ec := &Message{Kind: CANCELLED, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, ec)
	// Add Buy
	b := &Message{Kind: BUY, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b)
	// Add Sell
	s2 := &Message{Kind: SELL, TraderId: 3, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s2)
	// Expect match for traders 1 and 3
	eb := &Message{Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb)
	es := &Message{Kind: FULL, TraderId: 3, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es)
}

func testBuyCancelSellNoMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &Message{Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b)
	// Cancel Buy
	cb := &Message{Kind: CANCEL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, cb)
	// Expect Cancelled
	ec := &Message{Kind: CANCELLED, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, ec)
	// Add Sell
	s := &Message{Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, s)
	// Add Buy
	b2 := &Message{Kind: BUY, TraderId: 3, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b2)
	// Expect match for traders 1 and 3
	eb := &Message{Kind: FULL, TraderId: 3, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb)
	es := &Message{Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es)
}

func testBadCancelNotCancelled(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Cancel Buy that doesn't exist
	cb := &Message{Kind: CANCEL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, cb)
	// Expect Not Cancelled
	ec := &Message{Kind: NOT_CANCELLED, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, ec)
}

// Defect found where the second PARTIAL sell match is getting filtered because it is identical to the first
func testThreeBuysMatchedToOneSell(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Three buys
	b1 := &Message{Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b1)
	b2 := &Message{Kind: BUY, TraderId: 1, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b2)
	b3 := &Message{Kind: BUY, TraderId: 1, TradeId: 3, StockId: 1, Price: 7, Amount: 1}
	mt.Send(t, b3)
	// One big sell
	s := &Message{Kind: SELL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 3}
	mt.Send(t, s)
	// Expect full matches on all three buys
	eb1 := &Message{Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb1)
	es1 := &Message{Kind: PARTIAL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es1)
	eb2 := &Message{Kind: FULL, TraderId: 1, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb2)
	es2 := &Message{Kind: PARTIAL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es2)
	eb3 := &Message{Kind: FULL, TraderId: 1, TradeId: 3, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, eb3)
	es3 := &Message{Kind: FULL, TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	mt.Expect(t, es3)
}

func addLowBuys(t *testing.T, mt MatchTester, highestPrice uint64, stockId uint64) {
	buys := suiteMaker.MkBuys(suiteMaker.ValRangeFlat(10, 1, highestPrice), stockId)
	for i := range buys {
		mt.Send(t, &buys[i])
	}
}

func addHighSells(t *testing.T, mt MatchTester, lowestPrice uint64, stockId uint64) {
	sells := suiteMaker.MkSells(suiteMaker.ValRangeFlat(10, lowestPrice, lowestPrice+10000), stockId)
	for i := range sells {
		mt.Send(t, &sells[i])
	}
}
