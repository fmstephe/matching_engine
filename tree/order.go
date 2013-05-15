package tree

import (
	"fmt"
	"github.com/fmstephe/fstrconv"
	"github.com/fmstephe/matching_engine/trade"
)

// Description of an Order which can live inside a guid and price tree
type Order struct {
	priceNode node
	guidNode  node
	amount    uint32
	stockId   uint32
	kind      trade.OrderKind
	ip        [4]byte
	port      int32
	nextFree  *Order
}

func newBuy(costData trade.CostData, tradeData trade.TradeData) *Order {
	return newOrder(costData, tradeData, trade.BUY)
}

func newSell(costData trade.CostData, tradeData trade.TradeData) *Order {
	return newOrder(costData, tradeData, trade.SELL)
}

func newCancel(tradeData trade.TradeData) *Order {
	return newOrder(trade.CostData{}, tradeData, trade.CANCEL)
}

func cancelOrder(o *Order) *Order {
	return newCancel(trade.TradeData{TraderId: o.TraderId(), TradeId: o.TradeId(), StockId: o.StockId()})
}

func newOrder(costData trade.CostData, tradeData trade.TradeData, kind trade.OrderKind) *Order {
	o := &Order{amount: costData.Amount, stockId: tradeData.StockId, kind: kind, priceNode: node{}}
	guid := trade.MkGuid(tradeData.TraderId, tradeData.TradeId)
	o.setup(costData.Price, guid)
	return o
}

func (o *Order) setup(price, guid int64) {
	initNode(o, price, &o.priceNode, &o.guidNode)
	initNode(o, guid, &o.guidNode, &o.priceNode)
}

func (o *Order) CopyFrom(from *trade.OrderData) {
	o.amount = from.Amount
	o.stockId = from.StockId
	o.kind = from.Kind
	o.setup(from.Price, from.Guid)
	o.ip = from.IP
	o.port = from.Port
}

func (o *Order) Price() int64 {
	return o.priceNode.val
}

func (o *Order) Guid() int64 {
	return o.guidNode.val
}

func (o *Order) TraderId() uint32 {
	return trade.GetTraderId(o.guidNode.val)
}

func (o *Order) TradeId() uint32 {
	return trade.GetTradeId(o.guidNode.val)
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

func (o *Order) IP() [4]byte {
	return o.ip
}

func (o *Order) Port() int32 {
	return o.port
}

func (o *Order) Kind() trade.OrderKind {
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
	kind := o.kind
	return fmt.Sprintf("%v, price %s, amount %s, trader %s, trade %s, stock %s", kind, price, amount, traderId, tradeId, stockId)
}
