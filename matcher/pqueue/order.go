package pqueue

import (
	"fmt"
	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/fstrconv"
	"github.com/fmstephe/matching_engine/msg"
)

type OrderNode struct {
	priceNode node
	guidNode  node
	amount    uint64
	stockId   uint64
	kind      msg.MsgKind
	nextFree  *OrderNode
}

func (o *OrderNode) CopyFrom(from *msg.Message) {
	o.amount = from.Amount
	o.stockId = from.StockId
	o.kind = from.Kind
	o.setup(from.Price, uint64(fmath.CombineInt32(int32(from.TraderId), int32(from.TradeId))))
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
	return uint32(fmath.HighInt32(int64(o.guidNode.val)))
}

func (o *OrderNode) TradeId() uint32 {
	return uint32(fmath.LowInt32(int64(o.guidNode.val)))
}

func (o *OrderNode) Amount() uint64 {
	return o.amount
}

func (o *OrderNode) ReduceAmount(s uint64) {
	o.amount -= s
}

func (o *OrderNode) StockId() uint64 {
	return o.stockId
}

func (o *OrderNode) Kind() msg.MsgKind {
	return o.kind
}

func (o *OrderNode) Remove() {
	o.priceNode.pop()
	o.guidNode.pop()
}

func (o *OrderNode) String() string {
	if o == nil {
		return "<nil>"
	}
	price := fstrconv.ItoaDelim(int64(o.Price()), ',')
	amount := fstrconv.ItoaDelim(int64(o.Amount()), ',')
	traderId := fstrconv.ItoaDelim(int64(o.TraderId()), '-')
	tradeId := fstrconv.ItoaDelim(int64(o.TradeId()), '-')
	stockId := fstrconv.ItoaDelim(int64(o.StockId()), '-')
	kind := o.kind
	return fmt.Sprintf("%v, price %s, amount %s, trader %s, trade %s, stock %s", kind, price, amount, traderId, tradeId, stockId)
}
