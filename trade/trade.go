package trade

import(
	"github.com/fmstephe/heap"
)

const (
	SELL = TradeType(-1)
	BUY = TradeType(1)
	MarketPrice = 0
)

type TradeType int32

type Order struct {
	TradeId int64 // Identifies this trade to the submitting trader
	Amount int64 // The number of units desired to buy/sell
	Price int64 // The highest/lowest acceptable price for a buy/sell
	StockId string
	Trader string
	BuySell TradeType // Indicates whether this trade is a buy or a sell
	ResponseChan chan *Response
	// Matching engine fields
	Seq int64 // Unique sequence number 
	Index int // The index of this trade in the heap
}

// TODO this does not account for Seq, which it should
func (t *Order) Less(e heap.Elem) bool {
	ot := e.(*Order)
	return (ot.Price - t.Price) * int32(t.BuySell) < 0
}

func (t *Order) SetIndex(i int) {
	t.Index = i
}

func NewBuy(tradeId, amount, price int64, stockId, trader string, rc chan *Response) *Order {
	return NewOrder(tradeId, amount, price, stockId, trader, rc, BUY)
}

func NewSell(tradeId, amount, price int64, stockId, trader string, rc chan *Response) *Order {
	return NewOrder(tradeId, amount, price, stockId, trader, rc, SELL)
}

func NewOrder(tradeId, amount, price int64, stockId, trader string, rc chan *Response, buySell TradeType) *Order {
	return &Order{TradeId: tradeId, Amount: amount, Price: price, StockId: stockId, Trader: trader, ResponseChan: rc, BuySell: buySell}
}

type Response struct {
	TradeId int64 // Links this trade back to a previously submitted Order
	Amount int64 // The number of units actually bought or sold
	Price int64 // The actual trade price, will be negative if a purchase was made
	CounterParty string // The trader-id of the other half of this trade
}

func NewResponse(tradeId, amount, price int64, counterParty string) *Response {
	return &Response{TradeId: tradeId, Amount: amount, Price: price, CounterParty: counterParty}
}
