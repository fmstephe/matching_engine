package trade

import (
	"fmt"
	"github.com/fmstephe/fstrconv"
)

const (
	BUY              = OrderKind(1)
	SELL             = OrderKind(-1)
	DELETE           = OrderKind(2)
	EXECUTE          = ResponseKind(3)
	CANCEL           = ResponseKind(2)
	FULL             = ResponseKind(3)
	X                = ResponseKind(4)
	TRANSPARENT      = ResponseKind(5)
	MARKET_PRICE     = 0
	NO_COUNTER_PARTY = 0
)

type OrderKind int32

type ResponseKind int32

func (k OrderKind) String() string {
	switch k {
	case BUY:
		return "BUY"
	case SELL:
		return "SELL"
	case DELETE:
		return "DELETE"
	default:
		return "Unkown OrderKind"
	}
	panic("Unreachable")
}

// For readable constructors
type CostData struct {
	Price  int64  // The highest/lowest acceptable price for a buy/sell
	Amount uint32 // The number of units desired to buy/sell
}

// For readable constructors
type TradeData struct {
	TraderId uint32 // Identifies the submitting trader
	TradeId  uint32 // Identifies this trade to the submitting trader
	StockId  uint32 // Identifies the stock for trade
}

type Order struct {
	Guid      int64
	Price     int64
	Amount    uint32
	TraderId  uint32
	TradeId   uint32
	StockId   uint32
	Kind      OrderKind
	LimitNode Node
	GuidNode  Node
}

func (o *Order) setup() {
	o.Guid = int64((uint64(o.TraderId) << 32) | uint64(o.TradeId))
	initNode(o, o.Price, &o.LimitNode)
	initNode(o, o.Guid, &o.GuidNode)
}

func (o *Order) String() string {
	if o == nil {
		return "<nil>"
	}
	price := fstrconv.Itoa64Delim(int64(o.Price), ',')
	amount := fstrconv.Itoa64Delim(int64(o.Amount), ',')
	traderId := fstrconv.Itoa64Delim(int64(o.TraderId), '-')
	tradeId := fstrconv.Itoa64Delim(int64(o.TradeId), '-')
	stockId := fstrconv.Itoa64Delim(int64(o.StockId), '-')
	return fmt.Sprintf("%s, price %s, amount %s, trader %s, trade %s, stock %s", o.Kind.String(), price, amount, traderId, tradeId, stockId)
}

func NewBuy(costData CostData, tradeData TradeData) *Order {
	return NewOrder(costData, tradeData, BUY)
}

func NewSell(costData CostData, tradeData TradeData) *Order {
	return NewOrder(costData, tradeData, SELL)
}

func NewDelete(tradeData TradeData) *Order {
	return NewOrder(CostData{}, tradeData, DELETE)
}

func NewOrder(costData CostData, tradeData TradeData, orderKind OrderKind) *Order {
	o := &Order{Price: costData.Price, Amount: costData.Amount, TraderId: tradeData.TraderId, TradeId: tradeData.TradeId, StockId: tradeData.StockId, Kind: orderKind, LimitNode: Node{}}
	o.setup()
	return o
}

type Response struct {
	Kind         ResponseKind
	Price        int64  // The actual trade price, will be negative if a purchase was made
	Amount       uint32 // The number of units actually bought or sold
	TradeId      uint32 // Links this trade back to a previously submitted Order
	CounterParty uint32 // The trader-id of the other half of this trade
}
