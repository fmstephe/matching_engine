package client

import (
	"fmt"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/msg"
)

const OUT_BUF_SIZE = 100
const IN_BUF_SIZE = 100

type traderRegMsg struct {
	traderId    uint32
	outOfClient chan *msg.Message
}

type Client struct {
	coordinator.AppMsgHelper
	clientMap  map[uint32]chan *msg.Message
	intoClient chan interface{}
}

func NewClient() (*Client, *TraderMaker) {
	intoClient := make(chan interface{}, IN_BUF_SIZE)
	return &Client{intoClient: intoClient, clientMap: make(map[uint32]chan *msg.Message)}, &TraderMaker{intoClient: intoClient}
}

func (c *Client) Run() {
	for {
		select {
		case m := <-c.In:
			m, shutdown := c.MsgProcessor(m, c.Out)
			if shutdown {
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

type TraderMaker struct {
	intoClient chan interface{}
}

func (tm *TraderMaker) Make(traderId uint32) *Trader {
	outOfClient := make(chan *msg.Message, OUT_BUF_SIZE)
	cr := &traderRegMsg{traderId: traderId, outOfClient: outOfClient}
	tm.intoClient <- cr // Performs the registration of this trader
	return &Trader{traderId: traderId, intoClient: tm.intoClient, OutOfClient: outOfClient}
}

type Trader struct {
	traderId    uint32
	intoClient  chan interface{}
	OutOfClient chan *msg.Message
}

func (t *Trader) Buy(price int64, tradeId, amount, stockId uint32) {
	t.submit(msg.BUY, price, tradeId, amount, stockId)
}

func (t *Trader) Sell(price int64, tradeId, amount, stockId uint32) {
	t.submit(msg.SELL, price, tradeId, amount, stockId)
}

func (t *Trader) Cancel(price int64, tradeId, amount, stockId uint32) {
	t.submit(msg.CANCEL, price, tradeId, amount, stockId)
}

func (t *Trader) submit(kind msg.MsgKind, price int64, tradeId, amount, stockId uint32) {
	m := &msg.Message{Direction: msg.OUT, Route: msg.APP, Kind: kind, Price: price, Amount: amount, TraderId: t.traderId, TradeId: tradeId, StockId: stockId}
	if !m.Valid() {
		panic(fmt.Sprintf("Invalid Message %v", m))
	}
	t.intoClient <- m
}
