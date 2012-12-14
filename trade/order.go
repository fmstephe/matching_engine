package trade

import (
	"fmt"
	"github.com/fmstephe/fstrconv"
)

type OrderKind int32
type ResponseKind int32

const (
	BUY           = OrderKind(1)
	SELL          = OrderKind(2)
	CANCEL        = OrderKind(3)
	PARTIAL       = ResponseKind(0)
	FULL          = ResponseKind(1)
	CANCELLED     = ResponseKind(2)
	NOT_CANCELLED = ResponseKind(3)
	MARKET_PRICE  = 0
)

func (k OrderKind) String() string {
	switch k {
	case BUY:
		return "BUY"
	case SELL:
		return "SELL"
	case CANCEL:
		return "CANCEL"
	}
	panic("Unreachable")
}

func (k ResponseKind) String() string {
	switch k {
	case PARTIAL:
		return "PARTIAL"
	case FULL:
		return "FULL"
	case CANCELLED:
		return "CANCELLED"
	case NOT_CANCELLED:
		return "NOT_CANCELLED"
	}
	panic("Uncreachable")
}

func mkGuid(traderId, tradeId uint32) int64 {
	return int64((uint64(traderId) << 32) | uint64(tradeId))
}

func getTraderId(guid int64) uint32 {
	return uint32(uint64(guid >> 32)) // untested
}

func getTradeId(guid int64) uint32 {
	return uint32(uint64(guid ^ int64(1)<<32)) // untested
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

// Flat description of an incoming order
type OrderData struct {
	Price   int64
	Guid    int64
	Amount  uint32
	StockId uint32
	Kind    OrderKind
}

func NewBuyData(costData CostData, tradeData TradeData) *OrderData {
	od := &OrderData{}
	WriteBuyData(costData, tradeData, od)
	return od
}

func NewSellData(costData CostData, tradeData TradeData) *OrderData {
	od := &OrderData{}
	WriteSellData(costData, tradeData, od)
	return od
}

func CancelOrderData(o *Order) *OrderData {
	return NewCancelData(TradeData{TraderId: o.TraderId(), TradeId: o.TradeId(), StockId: o.StockId()})
}

func NewCancelData(tradeData TradeData) *OrderData {
	od := &OrderData{}
	WriteCancelData(tradeData, od)
	return od
}

func NewOrderData(costData CostData, tradeData TradeData, kind OrderKind) *OrderData {
	od := &OrderData{}
	WriteOrderData(costData, tradeData, kind, od)
	return od
}

func WriteBuyData(costData CostData, tradeData TradeData, od *OrderData) {
	WriteOrderData(costData, tradeData, BUY, od)
}

func WriteSellData(costData CostData, tradeData TradeData, od *OrderData) {
	WriteOrderData(costData, tradeData, SELL, od)
}

func WriteCancelOrderData(o *Order, od *OrderData) {
	WriteCancelData(TradeData{TraderId: o.TraderId(), TradeId: o.TradeId(), StockId: o.StockId()}, od)
}

func WriteCancelData(tradeData TradeData, od *OrderData) {
	WriteOrderData(CostData{}, tradeData, CANCEL, od)
}

func WriteOrderData(costData CostData, tradeData TradeData, kind OrderKind, od *OrderData) {
	od.Price = costData.Price
	od.Guid = mkGuid(tradeData.TraderId, tradeData.TradeId)
	od.Amount = costData.Amount
	od.StockId = tradeData.StockId
	od.Kind = kind
}

// Description of an order which can live inside a guid and price tree
type Order struct {
	priceNode node
	guidNode  node
	amount    uint32
	stockId   uint32
	kind      OrderKind
	nextFree  *Order
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

func CancelOrder(o *Order) *Order {
	return NewCancel(TradeData{TraderId: o.TraderId(), TradeId: o.TradeId(), StockId: o.StockId()})
}

func NewOrder(costData CostData, tradeData TradeData, orderKind OrderKind) *Order {
	o := &Order{amount: costData.Amount, stockId: tradeData.StockId, kind: orderKind, priceNode: node{}}
	guid := mkGuid(tradeData.TraderId, tradeData.TradeId)
	o.setup(costData.Price, guid)
	return o
}

func (o *Order) setup(price, guid int64) {
	initNode(o, price, &o.priceNode, &o.guidNode)
	initNode(o, guid, &o.guidNode, &o.priceNode)
}

func NewOrderFromData(od *OrderData) *Order {
	o := &Order{}
	o.CopyFrom(od)
	return o
}

func (o *Order) CopyFrom(from *OrderData) {
	o.amount = from.Amount
	o.stockId = from.StockId
	o.kind = from.Kind
	o.setup(from.Price, from.Guid)
}

func (o *Order) Price() int64 {
	return o.priceNode.val
}

func (o *Order) Guid() int64 {
	return o.guidNode.val
}

func (o *Order) TraderId() uint32 {
	return getTraderId(o.guidNode.val)
}

func (o *Order) TradeId() uint32 {
	return getTradeId(o.guidNode.val)
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
