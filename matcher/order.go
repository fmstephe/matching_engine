package matcher

import ()

const (
	SELL        = TradeType(true)
	BUY         = TradeType(false)
	MarketPrice = 0
)

type TradeType bool

type CostData struct {
	Price  int64  // The highest/lowest acceptable price for a buy/sell
	Amount uint32 // The number of units desired to buy/sell
}

type TradeData struct {
	TraderId uint32 // Identifies the submitting trader
	TradeId  uint32 // Identifies this trade to the submitting trader
	StockId  uint32 // Identifies the stock for trade
}

type Order struct {
	CostData
	TradeData
	ResponseFunc func(*Response)
	BuySell      TradeType // Indicates whether this trade is a buy or a sell
	// Linked List fields
	next     *Order  // The next order in this limit
	incoming **Order // The pointer pointing to this order - used for deletion
}

func (o *Order) GUID() uint64 {
	return (uint64(o.TraderId) << 32) | uint64(o.TradeId)
}

func NewBuy(costData CostData, tradeData TradeData, responseFunc func(*Response)) *Order {
	return NewOrder(costData, tradeData, responseFunc, BUY)
}

func NewSell(costData CostData, tradeData TradeData, responseFunc func(*Response)) *Order {
	return NewOrder(costData, tradeData, responseFunc, SELL)
}

func NewOrder(costData CostData, tradeData TradeData, responseFunc func(*Response), buySell TradeType) *Order {
	return &Order{CostData: costData, TradeData: tradeData, ResponseFunc: responseFunc, BuySell: buySell}
}

type Response struct {
	Price        int64  // The actual trade price, will be negative if a purchase was made
	Amount       uint32 // The number of units actually bought or sold
	TradeId      uint32 // Links this trade back to a previously submitted Order
	CounterParty uint32 // The trader-id of the other half of this trade
}

func NewResponse(price int64, amount, tradeId, counterParty uint32) *Response {
	return &Response{Price: price, Amount: amount, TradeId: tradeId, CounterParty: counterParty}
}
