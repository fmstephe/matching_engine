package trade

import (
	"fmt"
	"github.com/fmstephe/fstrconv"
)

const (
	BUY           = OrderKind(0)
	SELL          = OrderKind(1)
	CANCEL        = OrderKind(2)
	PARTIAL       = ResponseKind(0)
	FULL          = ResponseKind(1)
	CANCELLED     = ResponseKind(2)
	NOT_CANCELLED = ResponseKind(3)
	MARKET_PRICE  = 0
)

type OrderKind int32

type ResponseKind int32

func (k OrderKind) String() string {
	switch k {
	case BUY:
		return "BUY"
	case SELL:
		return "SELL"
	case CANCEL:
		return "CANCEL"
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
	amount    uint32
	stockId   uint32
	kind      OrderKind
	priceNode node
	guidNode  node
	nextFree  *Order
}

func (o *Order) setup(price, guid int64) {
	initNode(o, price, &o.priceNode, &o.guidNode)
	initNode(o, guid, &o.guidNode, &o.priceNode)
}

func (o *Order) CopyInto(into *Order) {
	into.amount = o.Amount()
	into.stockId = o.StockId()
	into.kind = o.Kind()
	into.setup(o.Price(), o.Guid())
}

func (o *Order) Price() int64 {
	return o.priceNode.val
}

func (o *Order) Guid() int64 {
	return o.guidNode.val
}

func (o *Order) TraderId() uint32 {
	return uint32(uint64(o.guidNode.val) >> 32) // untested
}

func (o *Order) TradeId() uint32 {
	return uint32(uint64(o.guidNode.val ^ int64(1)<<32)) // untested
}

func (o *Order) Amount() uint32 {
	return o.amount
}

func (o *Order) ReduceAmount(s uint32) {
	o.amount -= s
}

func (o *Order) StockId() uint32 {
	return o.stockId
}

func (o *Order) Kind() OrderKind {
	return o.kind
}

func (o *Order) String() string {
	if o == nil {
		return "<nil>"
	}
	price := fstrconv.Itoa64Delim(int64(o.Price()), ',')
	amount := fstrconv.Itoa64Delim(int64(o.Amount()), ',')
	traderId := fstrconv.Itoa64Delim(int64(o.TraderId()), '-')
	tradeId := fstrconv.Itoa64Delim(int64(o.TradeId()), '-')
	stockId := fstrconv.Itoa64Delim(int64(o.StockId()), '-')
	return fmt.Sprintf("%s, price %s, amount %s, trader %s, trade %s, stock %s", o.Kind().String(), price, amount, traderId, tradeId, stockId)
}

func NewBuy(costData CostData, tradeData TradeData) *Order {
	return NewOrder(costData, tradeData, BUY)
}

func NewSell(costData CostData, tradeData TradeData) *Order {
	return NewOrder(costData, tradeData, SELL)
}

func NewCancel(tradeData TradeData) *Order {
	return NewOrder(CostData{}, tradeData, CANCEL)
}

func NewOrder(costData CostData, tradeData TradeData, orderKind OrderKind) *Order {
	o := &Order{amount: costData.Amount, stockId: tradeData.StockId, kind: orderKind, priceNode: node{}}
	guid := int64((uint64(tradeData.TraderId) << 32) | uint64(tradeData.TradeId))
	o.setup(costData.Price, guid)
	return o
}

type Response struct {
	Kind         ResponseKind
	Price        int64  // The actual trade price, will be negative if a purchase was made
	Amount       uint32 // The number of units actually bought or sold
	TraderId     uint32 // The trader-id of the trader to whom this response is directed
	TradeId      uint32 // Links this trade back to a previously submitted Order
	CounterParty uint32 // The trader-id of the other half of this trade
}

func (r *Response) WriteTrade(kind ResponseKind, price int64, amount, traderId, tradeId, counterParty uint32) {
	r.Kind = kind
	r.Price = price
	r.Amount = amount
	r.TraderId = traderId
	r.TradeId = tradeId
	r.CounterParty = counterParty
}

func (r *Response) WriteCancel(kind ResponseKind, traderId, tradeId uint32) {
	r.Kind = kind
	r.TraderId = traderId
	r.TradeId = tradeId
}
