package tree

import (
	"fmt"
	"github.com/fmstephe/fstrconv"
	"github.com/fmstephe/matching_engine/trade"
)

// Description of an OrderNode which can live inside a guid and price tree
type OrderNode struct {
	priceNode node
	guidNode  node
	amount    uint32
	stockId   uint32
	kind      trade.OrderNodeKind
	ip        [4]byte
	port      int32
	nextFree  *OrderNode
}

func newBuy(costData trade.CostData, tradeData trade.TradeData) *OrderNode {
	return newOrderNode(costData, tradeData, trade.BUY)
}

func newSell(costData trade.CostData, tradeData trade.TradeData) *OrderNode {
	return newOrderNode(costData, tradeData, trade.SELL)
}

func newCancel(tradeData trade.TradeData) *OrderNode {
	return newOrderNode(trade.CostData{}, tradeData, trade.CANCEL)
}

func cancelOrderNode(o *OrderNode) *OrderNode {
	return newCancel(trade.TradeData{TraderId: o.TraderId(), TradeId: o.TradeId(), StockId: o.StockId()})
}

func newOrderNode(costData trade.CostData, tradeData trade.TradeData, kind trade.OrderNodeKind) *OrderNode {
	o := &OrderNode{amount: costData.Amount, stockId: tradeData.StockId, kind: kind, priceNode: node{}}
	guid := trade.MkGuid(tradeData.TraderId, tradeData.TradeId)
	o.setup(costData.Price, guid)
	return o
}

func (o *OrderNode) setup(price, guid int64) {
	initNode(o, price, &o.priceNode, &o.guidNode)
	initNode(o, guid, &o.guidNode, &o.priceNode)
}

func (o *OrderNode) CopyFrom(from *trade.Order) {
	o.amount = from.Amount
	o.stockId = from.StockId
	o.kind = from.Kind
	o.setup(from.Price, from.Guid)
	o.ip = from.IP
	o.port = from.Port
}

func (o *OrderNode) Price() int64 {
	return o.priceNode.val
}

func (o *OrderNode) Guid() int64 {
	return o.guidNode.val
}

func (o *OrderNode) TraderId() uint32 {
	return trade.GetTraderId(o.guidNode.val)
}

func (o *OrderNode) TradeId() uint32 {
	return trade.GetTradeId(o.guidNode.val)
}

func (o *OrderNode) Amount() uint32 {
	return o.amount
}

func (o *OrderNode) ReduceAmount(s uint32) {
	o.amount -= s
}

func (o *OrderNode) StockId() uint32 {
	return o.stockId
}

func (o *OrderNode) IP() [4]byte {
	return o.ip
}

func (o *OrderNode) Port() int32 {
	return o.port
}

func (o *OrderNode) Kind() trade.OrderNodeKind {
	return o.kind
}

func (o *OrderNode) String() string {
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
