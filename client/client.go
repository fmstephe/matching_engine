package client

import (
	"errors"
	"fmt"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/msg"
)

type traderRegMsg struct {
	traderId    uint32
	outOfClient chan *msg.Message
}

type C struct {
	coordinator.AppMsgHelper
	clientMap  map[uint32]chan *msg.Message
	intoClient chan interface{}
}

func NewClient() (*C, *CommMaker) {
	intoClient := make(chan interface{})
	return &C{intoClient: intoClient, clientMap: make(map[uint32]chan *msg.Message)}, &CommMaker{intoClient: intoClient}
}

func (c *C) Run() {
	for {
		select {
		case m := <-c.In:
			if m.Kind == msg.SHUTDOWN {
				c.Out <- m
				return
			}
			if m != nil {
				clientChan := c.clientMap[m.TraderId]
				if clientChan == nil {
					println("Missing traderId", m.TraderId)
					continue
				}
				clientChan <- m
			}
		case i := <-c.intoClient:
			switch i := i.(type) {
			case *traderRegMsg:
				if c.clientMap[i.traderId] != nil {
					panic(fmt.Sprintf("Attempted to register a trader (%i) twice", i.traderId))
				}
				c.clientMap[i.traderId] = i.outOfClient
			case *msg.Message:
				c.Out <- i
			default:
				panic(fmt.Sprintf("%T is not a legal type", i))
			}
		}
	}
}

type CommMaker struct {
	intoClient chan interface{}
}

func (cm *CommMaker) NewComm(traderId uint32) *Comm {
	outOfClient := make(chan *msg.Message)
	cr := &traderRegMsg{traderId: traderId, outOfClient: outOfClient}
	cm.intoClient <- cr // Register this Comm
	return &Comm{traderId: traderId, intoClient: cm.intoClient, outOfClient: outOfClient}
}

type Comm struct {
	traderId    uint32
	intoClient  chan interface{}
	outOfClient chan *msg.Message
}

func (c *Comm) Buy(price int64, tradeId, amount, stockId uint32) error {
	return c.submit(msg.BUY, price, tradeId, amount, stockId)
}

func (c *Comm) Sell(price int64, tradeId, amount, stockId uint32) error {
	return c.submit(msg.SELL, price, tradeId, amount, stockId)
}

func (c *Comm) Cancel(price int64, tradeId, amount, stockId uint32) error {
	return c.submit(msg.CANCEL, price, tradeId, amount, stockId)
}

func (c *Comm) Out() chan *msg.Message {
	return c.outOfClient
}

func (c *Comm) submit(kind msg.MsgKind, price int64, tradeId, amount, stockId uint32) error {
	m := &msg.Message{Kind: kind, Price: price, Amount: amount, TraderId: c.traderId, TradeId: tradeId, StockId: stockId}
	if !m.Valid() {
		return errors.New(fmt.Sprintf("Invalid Message %v", m))
	}
	c.intoClient <- m
	return nil
}
