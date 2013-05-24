package trade

import (
	"errors"
	"fmt"
	"net"
)

type MsgKind int32

const (
	ILLEGAL = MsgKind(0) // 0 is never a valid kind
	// Incoming messages
	CLIENT_ACK = MsgKind(1) // Indicates the client has received a message from the matcher
	BUY        = MsgKind(2)
	SELL       = MsgKind(3)
	CANCEL     = MsgKind(4)
	SHUTDOWN   = MsgKind(5)
	// Outgoing messages
	MATCHER_ACK   = MsgKind(6) // Used to acknowledge a message back to the client
	PARTIAL       = MsgKind(7)
	FULL          = MsgKind(8)
	CANCELLED     = MsgKind(9)
	NOT_CANCELLED = MsgKind(10)
	// Error message
	ERROR = MsgKind(11) // TODO error messages are currently unused
)

const (
	// Constant price indicating a market price sell
	MARKET_PRICE = 0
)

const (
	SizeofOrder    = 36
	SizeofResponse = 36
)

func (k MsgKind) String() string {
	switch k {
	case ILLEGAL:
		return "ILLEGAL"
	case CLIENT_ACK:
		return "CLIENT_ACK"
	case BUY:
		return "BUY"
	case SELL:
		return "SELL"
	case CANCEL:
		return "CANCEL"
	case SHUTDOWN:
		return "SHUTDOWN"
	case MATCHER_ACK:
		return "MATCHER_ACK"
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
type Order struct {
	Kind     MsgKind
	Price    int64
	Amount   uint32
	TraderId uint32
	TradeId  uint32
	StockId  uint32
	IP       [4]byte
	Port     int32
	// I think we need a checksum here
}

func (o *Order) WriteBuy(costData CostData, tradeData TradeData) {
	o.Write(costData, tradeData, BUY)
}

func (o *Order) WriteSell(costData CostData, tradeData TradeData) {
	o.Write(costData, tradeData, SELL)
}

func (o *Order) WriteCancel(tradeData TradeData) {
	o.Write(CostData{}, tradeData, CANCEL)
}

func (o *Order) WriteCancelFromOrder(oo *Order) {
	o.Write(CostData{}, TradeData{TraderId: oo.TraderId, TradeId: oo.TradeId, StockId: oo.StockId}, CANCEL)
}

func (o *Order) WriteShutdown() {
	o.Write(CostData{}, TradeData{}, SHUTDOWN)
}

func (o *Order) Write(costData CostData, tradeData TradeData, kind MsgKind) {
	o.Price = costData.Price
	o.TraderId = tradeData.TraderId
	o.TradeId = tradeData.TradeId
	o.Amount = costData.Amount
	o.StockId = tradeData.StockId
	o.Kind = kind
}

func (o *Order) UDPAddr() *net.UDPAddr {
	addr := &net.UDPAddr{}
	addr.IP = net.IPv4(o.IP[0], o.IP[1], o.IP[2], o.IP[3])
	addr.Port = int(o.Port)
	return addr
}

func (o *Order) SetUDPAddr(addr *net.UDPAddr) error {
	IP := addr.IP.To4()
	if IP == nil {
		return errors.New(fmt.Sprintf("IP address (%s) is not IPv4", addr.IP.String()))
	}
	o.IP[0], o.IP[1], o.IP[2], o.IP[3] = IP[0], IP[1], IP[2], IP[3]
	o.Port = int32(addr.Port)
	println(o.UDPAddr().String())
	return nil
}

type Response struct {
	Kind         MsgKind
	Price        int64  // The actual trade price, will be negative if a purchase was made
	Amount       uint32 // The number of units actually bought or sold
	TraderId     uint32 // The trader-id of the trader to whom this response is directed
	TradeId      uint32 // Links this trade back to a previously submitted OrderNode
	CounterParty uint32 // The trader-id of the other half of this trade
	IP           [4]byte
	Port         int32
}

func (r *Response) WriteTrade(kind MsgKind, price int64, amount, traderId, tradeId, counterParty uint32) {
	r.Kind = kind
	r.Price = price
	r.Amount = amount
	r.TraderId = traderId
	r.TradeId = tradeId
	r.CounterParty = counterParty
}

func (r *Response) WriteCancel(kind MsgKind, traderId, tradeId uint32) {
	r.Kind = kind
	r.TraderId = traderId
	r.TradeId = tradeId
}

func (r *Response) WriteAck(o *Order) {
	r.Kind = MATCHER_ACK
	r.Price = o.Price
	r.Amount = o.Amount
	r.TraderId = o.TraderId
	r.TradeId = o.TradeId
	r.CounterParty = 0
	r.IP = o.IP
	r.Port = o.Port
}

func (r *Response) WriteShutdown() {
	r.Kind = SHUTDOWN
}

func (r *Response) UDPAddr() *net.UDPAddr {
	addr := &net.UDPAddr{}
	addr.IP = net.IPv4(r.IP[0], r.IP[1], r.IP[2], r.IP[3])
	addr.Port = int(r.Port)
	return addr
}
