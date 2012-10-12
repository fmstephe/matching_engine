package trade

import ()

const (
	BUY              = OrderKind(1)
	SELL             = OrderKind(-1)
	DELETE_BUY       = OrderKind(2)
	DELETE_SELL      = OrderKind(-2)
	EXECUTE          = ResponseKind(1)
	CANCEL           = ResponseKind(2)
	FULL             = ResponseKind(3)
	X                = ResponseKind(4)
	TRANSPARENT      = ResponseKind(5)
	MARKET_PRICE     = 0
	NO_COUNTER_PARTY = 0
)

type OrderKind int32

type ResponseKind int32

// For readable constructors
type CostData struct {
	Price  int32  // The highest/lowest acceptable price for a buy/sell
	Amount uint32 // The number of units desired to buy/sell
}

// For readable constructors
type TradeData struct {
	TraderId uint32 // Identifies the submitting trader
	TradeId  uint32 // Identifies this trade to the submitting trader
	StockId  uint32 // Identifies the stock for trade
}

type Order struct {
	Guid     int64
	Price    int32
	Amount   uint32
	TraderId uint32
	TradeId  uint32
	StockId  uint32
	Kind     OrderKind
	Compare  int64  // Binary heap comparison value
	Higher   *Order // Next higher priority order in this limit
	Lower    *Order // Next lower priority order in this limit
}

func (o *Order) setup() {
	o.Guid = int64((uint64(o.TraderId) << 32) | uint64(o.TradeId))
}

func NewBuy(costData CostData, tradeData TradeData) *Order {
	return NewOrder(costData, tradeData, BUY)
}

func NewSell(costData CostData, tradeData TradeData) *Order {
	return NewOrder(costData, tradeData, SELL)
}

func NewDeleteBuy(tradeData TradeData) *Order {
	return NewOrder(CostData{}, tradeData, DELETE_BUY)
}

func NewDeleteSell(tradeData TradeData) *Order {
	return NewOrder(CostData{}, tradeData, DELETE_SELL)
}

func NewOrder(costData CostData, tradeData TradeData, orderKind OrderKind) *Order {
	o := &Order{Price: costData.Price, Amount: costData.Amount, TraderId: tradeData.TraderId, TradeId: tradeData.TradeId, StockId: tradeData.StockId, Kind: orderKind}
	o.setup()
	return o
}

type Response struct {
	Kind         ResponseKind
	Price        int32  // The actual trade price, will be negative if a purchase was made
	Amount       uint32 // The number of units actually bought or sold
	TradeId      uint32 // Links this trade back to a previously submitted Order
	CounterParty uint32 // The trader-id of the other half of this trade
}
