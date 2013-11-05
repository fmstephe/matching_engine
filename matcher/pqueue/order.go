package pqueue

import (
	"fmt"
	"github.com/fmstephe/fstrconv"
	"github.com/fmstephe/matching_engine/ints"
	"github.com/fmstephe/matching_engine/msg"
)

type OrderNode struct {
	priceNode node
	guidNode  node
	amount    uint32
	stockId   uint32
	kind      msg.MsgKind
	nextFree  *OrderNode
}

func (o *OrderNode) CopyFrom(from *msg.Message) {
	o.amount = from.Amount
	o.stockId = from.StockId
	o.kind = from.Kind
	o.setup(from.Price, ints.Combine(from.TraderId, from.TradeId))
}

func (o *OrderNode) CopyTo(to *msg.Message) {
	to.Kind = o.Kind()
	to.Price = o.Price()
	to.Amount = o.Amount()
	to.TraderId = o.TraderId()
	to.TradeId = o.TradeId()
	to.StockId = o.StockId()
}

func (o *OrderNode) setup(price, guid uint64) {
	initNode(o, price, &o.priceNode, &o.guidNode)
	initNode(o, guid, &o.guidNode, &o.priceNode)
}

func (o *OrderNode) Price() uint64 {
	return o.priceNode.val
}

func (o *OrderNode) Guid() uint64 {
	return o.guidNode.val
}

func (o *OrderNode) TraderId() uint32 {
	return ints.High32(o.guidNode.val)
}

func (o *OrderNode) TradeId() uint32 {
	return ints.Low32(o.guidNode.val)
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

func (o *OrderNode) Kind() msg.MsgKind {
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
