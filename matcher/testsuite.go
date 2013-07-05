package matcher

import (
	"github.com/fmstephe/matching_engine/msg"
	"testing"
)

var suiteMaker = msg.NewMessageMaker(100)

type MatchTester interface {
	Send(*testing.T, *msg.Message)
	Expect(*testing.T, *msg.Message)
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
	s := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	s.Route = msg.ORDER
	s.Kind = msg.SELL
	mt.Send(t, s)
	// Add Buy
	b := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	b.Route = msg.ORDER
	b.Kind = msg.BUY
	mt.Send(t, b)
	// Full match
	es := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	es.Route = msg.MATCHER_RESPONSE
	es.Kind = msg.FULL
	mt.Expect(t, es)
	eb := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	eb.Route = msg.MATCHER_RESPONSE
	eb.Kind = msg.FULL
	mt.Expect(t, eb)
}

func testBuySellMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	b.Route = msg.ORDER
	b.Kind = msg.BUY
	mt.Send(t, b)
	// Add Sell
	s := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	s.Route = msg.ORDER
	s.Kind = msg.SELL
	mt.Send(t, s)
	// Full match
	eb := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	eb.Route = msg.MATCHER_RESPONSE
	eb.Kind = msg.FULL
	mt.Expect(t, eb)
	es := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	es.Route = msg.MATCHER_RESPONSE
	es.Kind = msg.FULL
	mt.Expect(t, es)
}

func testBuyDoubleSellMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 2}
	b.Route = msg.ORDER
	b.Kind = msg.BUY
	mt.Send(t, b)
	// Add Sell1
	s1 := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	s1.Route = msg.ORDER
	s1.Kind = msg.SELL
	mt.Send(t, s1)
	// Full match
	eb1 := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	eb1.Route = msg.MATCHER_RESPONSE
	eb1.Kind = msg.PARTIAL
	mt.Expect(t, eb1)
	es1 := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	es1.Route = msg.MATCHER_RESPONSE
	es1.Kind = msg.FULL
	mt.Expect(t, es1)
	// Add Sell2
	s2 := &msg.Message{TraderId: 2, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	s2.Route = msg.ORDER
	s2.Kind = msg.SELL
	mt.Send(t, s2)
	// Full Match II
	eb2 := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	eb2.Route = msg.MATCHER_RESPONSE
	eb2.Kind = msg.FULL
	mt.Expect(t, eb2)
	es2 := &msg.Message{TraderId: 2, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	es2.Route = msg.MATCHER_RESPONSE
	es2.Kind = msg.FULL
	mt.Expect(t, es2)
}

func testSellDoubleBuyMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Sell
	s := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 2}
	s.Route = msg.ORDER
	s.Kind = msg.SELL
	mt.Send(t, s)
	// Add Buy1
	b1 := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	b1.Route = msg.ORDER
	b1.Kind = msg.BUY
	mt.Send(t, b1)
	// Full match
	eb1 := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	eb1.Route = msg.MATCHER_RESPONSE
	eb1.Kind = msg.FULL
	mt.Expect(t, eb1)
	es1 := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	es1.Route = msg.MATCHER_RESPONSE
	es1.Kind = msg.PARTIAL
	mt.Expect(t, es1)
	// Add Buy2
	b2 := &msg.Message{TraderId: 2, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	b2.Route = msg.ORDER
	b2.Kind = msg.BUY
	mt.Send(t, b2)
	// Full Match II
	eb2 := &msg.Message{TraderId: 2, TradeId: 2, StockId: 1, Price: -7, Amount: 1}
	eb2.Route = msg.MATCHER_RESPONSE
	eb2.Kind = msg.FULL
	mt.Expect(t, eb2)
	es2 := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	es2.Route = msg.MATCHER_RESPONSE
	es2.Kind = msg.FULL
	mt.Expect(t, es2)
}

func testMidPrice(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 9, Amount: 1}
	b.Route = msg.ORDER
	b.Kind = msg.BUY
	mt.Send(t, b)
	// Add Sell
	s := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 6, Amount: 1}
	s.Route = msg.ORDER
	s.Kind = msg.SELL
	mt.Send(t, s)
	// Full match
	eb := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	eb.Route = msg.MATCHER_RESPONSE
	eb.Kind = msg.FULL
	mt.Expect(t, eb)
	es := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	es.Route = msg.MATCHER_RESPONSE
	es.Kind = msg.FULL
	mt.Expect(t, es)
}

func testMidPriceBigSell(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 9, Amount: 1}
	b.Route = msg.ORDER
	b.Kind = msg.BUY
	mt.Send(t, b)
	// Add Sell
	s := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 6, Amount: 10}
	s.Route = msg.ORDER
	s.Kind = msg.SELL
	mt.Send(t, s)
	// Full match
	eb := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	eb.Route = msg.MATCHER_RESPONSE
	eb.Kind = msg.FULL
	mt.Expect(t, eb)
	es := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	es.Route = msg.MATCHER_RESPONSE
	es.Kind = msg.PARTIAL
	mt.Expect(t, es)
}

func testMidPriceBigBuy(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 9, Amount: 10}
	b.Route = msg.ORDER
	b.Kind = msg.BUY
	mt.Send(t, b)
	// Add Sell
	s := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 6, Amount: 1}
	s.Route = msg.ORDER
	s.Kind = msg.SELL
	mt.Send(t, s)
	// Full match
	eb := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	eb.Route = msg.MATCHER_RESPONSE
	eb.Kind = msg.PARTIAL
	mt.Expect(t, eb)
	es := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	es.Route = msg.MATCHER_RESPONSE
	es.Kind = msg.FULL
	mt.Expect(t, es)
}

func testTradeSeparateStocks(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Sell stock 1
	s1 := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	s1.Route = msg.ORDER
	s1.Kind = msg.SELL
	mt.Send(t, s1)
	// Add Buy stock 2
	b2 := &msg.Message{TraderId: 1, TradeId: 1, StockId: 2, Price: 7, Amount: 1}
	b2.Route = msg.ORDER
	b2.Kind = msg.BUY
	mt.Send(t, b2)
	// Add Sell stock 2
	s2 := &msg.Message{TraderId: 2, TradeId: 2, StockId: 2, Price: 7, Amount: 1}
	s2.Route = msg.ORDER
	s2.Kind = msg.SELL
	mt.Send(t, s2)
	// Full match stock 2
	es2 := &msg.Message{TraderId: 1, TradeId: 1, StockId: 2, Price: -7, Amount: 1}
	es2.Route = msg.MATCHER_RESPONSE
	es2.Kind = msg.FULL
	mt.Expect(t, es2)
	eb2 := &msg.Message{TraderId: 2, TradeId: 2, StockId: 2, Price: 7, Amount: 1}
	eb2.Route = msg.MATCHER_RESPONSE
	eb2.Kind = msg.FULL
	mt.Expect(t, eb2)
	// Add Buy stock 1
	b1 := &msg.Message{TraderId: 1, TradeId: 2, StockId: 1, Price: 7, Amount: 1}
	b1.Route = msg.ORDER
	b1.Kind = msg.BUY
	mt.Send(t, b1)
	// Full match stock 1
	eb1 := &msg.Message{TraderId: 1, TradeId: 2, StockId: 1, Price: -7, Amount: 1}
	eb1.Route = msg.MATCHER_RESPONSE
	eb1.Kind = msg.FULL
	mt.Expect(t, eb1)
	es1 := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	es1.Route = msg.MATCHER_RESPONSE
	es1.Kind = msg.FULL
	mt.Expect(t, es1)
}

func testSeparateStocksNotMatched(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Sell
	s1 := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	s1.Route = msg.ORDER
	s1.Kind = msg.SELL
	mt.Send(t, s1)
	// Add Sell
	// Add Buy
	b1 := &msg.Message{TraderId: 1, TradeId: 1, StockId: 2, Price: 7, Amount: 1}
	b1.Route = msg.ORDER
	b1.Kind = msg.BUY
	mt.Send(t, b1)
	// Expect empty message flushing
	// Add Trader 2 order to flush trader 2 messages
	s2 := &msg.Message{TraderId: 2, TradeId: 2, StockId: 1, Price: 70, Amount: 1}
	s2.Route = msg.ORDER
	s2.Kind = msg.SELL
	mt.Send(t, s2)
	// Add Trader 1 order to flush trader 1 messages
	b2 := &msg.Message{TraderId: 1, TradeId: 2, StockId: 1, Price: 1, Amount: 1}
	b2.Route = msg.ORDER
	b2.Kind = msg.BUY
	mt.Send(t, b2)
}

func testSellCancelBuyNoMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Sell
	s := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	s.Route = msg.ORDER
	s.Kind = msg.SELL
	mt.Send(t, s)
	// Cancel Sell
	cs := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	cs.Route = msg.ORDER
	cs.Kind = msg.CANCEL
	mt.Send(t, cs)
	// Expect Cancelled
	ec := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	ec.Route = msg.MATCHER_RESPONSE
	ec.Kind = msg.CANCELLED
	mt.Expect(t, ec)
	// Add Buy
	b := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	b.Route = msg.ORDER
	b.Kind = msg.BUY
	mt.Send(t, b)
	// Add Trader 1 order to flush trader 1 messages
	b2 := &msg.Message{TraderId: 1, TradeId: 2, StockId: 1, Price: 1, Amount: 1}
	b2.Route = msg.ORDER
	b2.Kind = msg.BUY
	mt.Send(t, b2)
	// Add Trader 2 order to flush trader 2 messages
	s2 := &msg.Message{TraderId: 2, TradeId: 2, StockId: 1, Price: 70, Amount: 1}
	s2.Route = msg.ORDER
	s2.Kind = msg.SELL
	mt.Send(t, s2)
}

func testBuyCancelSellNoMatch(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Add Buy
	b := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	b.Route = msg.ORDER
	b.Kind = msg.BUY
	mt.Send(t, b)
	// Cancel Buy
	cb := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	cb.Route = msg.ORDER
	cb.Kind = msg.CANCEL
	mt.Send(t, cb)
	// Expect Cancelled
	ec := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	ec.Route = msg.MATCHER_RESPONSE
	ec.Kind = msg.CANCELLED
	mt.Expect(t, ec)
	// Add Sell
	s := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	s.Route = msg.ORDER
	s.Kind = msg.SELL
	mt.Send(t, s)
	// Expect empty message flushing
	// Add Trader 1 order to flush trader 1 messages
	b2 := &msg.Message{TraderId: 1, TradeId: 2, StockId: 1, Price: 1, Amount: 1}
	b2.Route = msg.ORDER
	b2.Kind = msg.BUY
	mt.Send(t, b2)
	// Add Trader 2 order to flush trader 2 messages
	s2 := &msg.Message{TraderId: 2, TradeId: 2, StockId: 1, Price: 70, Amount: 1}
	s2.Route = msg.ORDER
	s2.Kind = msg.SELL
	mt.Send(t, s2)
}

func testBadCancelNotCancelled(t *testing.T, mkr MatchTesterMaker) {
	mt := mkr.Make()
	defer mt.Cleanup(t)
	addLowBuys(t, mt, 5, 1)
	addHighSells(t, mt, 10, 1)
	// Cancel Buy that doesn't exist
	cb := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	cb.Route = msg.ORDER
	cb.Kind = msg.CANCEL
	mt.Send(t, cb)
	// Expect Not Cancelled
	ec := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	ec.Route = msg.MATCHER_RESPONSE
	ec.Kind = msg.NOT_CANCELLED
	mt.Expect(t, ec)
}

func addLowBuys(t *testing.T, mt MatchTester, highestPrice int64, stockId uint32) {
	buys := suiteMaker.MkBuys(suiteMaker.ValRangeFlat(10, 1, highestPrice), stockId)
	for _, buy := range buys {
		mt.Send(t, &buy)
	}
}

func addHighSells(t *testing.T, mt MatchTester, lowestPrice int64, stockId uint32) {
	sells := suiteMaker.MkSells(suiteMaker.ValRangeFlat(10, lowestPrice, lowestPrice+10000), stockId)
	for _, sell := range sells {
		mt.Send(t, &sell)
	}
}
