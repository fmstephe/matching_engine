package trade

import (
	"errors"
	"fmt"
	"net"
)

type OrderKind int32
type ResponseKind int32

const (
	// OrderNodes
	BUY    = OrderKind(iota)
	SELL   = OrderKind(iota)
	CANCEL = OrderKind(iota)
)

const (
	// Responses
	//Received       = ResponseKind(iota)
	PARTIAL       = ResponseKind(iota)
	FULL          = ResponseKind(iota)
	CANCELLED     = ResponseKind(iota)
	NOT_CANCELLED = ResponseKind(iota)
	// ACK           = ResponseKind(iota) // Use this to allow clients to acknowledge message receipt
)

const (
	// Constant price indicating a market price sell
	MARKET_PRICE = 0
)

const (
	SizeofOrder    = 36 //int(unsafe.Sizeof(Order{}))
	SizeofResponse = 36 //int(unsafe.Sizeof(Response{}))
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

func MkGuid(traderId, tradeId uint32) int64 {
	return (int64(traderId) << 32) | int64(tradeId)
}

func GetTraderId(guid int64) uint32 {
	return uint32(guid >> 32)
}

func GetTradeId(guid int64) uint32 {
	return uint32(guid)
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
	Price   int64
	Amount  uint32
	Guid    int64
	StockId uint32
	Kind    OrderKind
	IP      [4]byte
	Port    int32
	// I think we need a checksum here
}

func (od *Order) WriteBuy(costData CostData, tradeData TradeData) {
	od.Write(costData, tradeData, BUY)
}

func (od *Order) WriteSell(costData CostData, tradeData TradeData) {
	od.Write(costData, tradeData, SELL)
}

func (od *Order) WriteCancel(tradeData TradeData) {
	od.Write(CostData{}, tradeData, CANCEL)
}

func (od *Order) WriteCancelFromOrderNode(o *Order) {
	od.Write(CostData{}, TradeData{TraderId: GetTraderId(o.Guid), TradeId: GetTradeId(o.Guid), StockId: o.StockId}, CANCEL)
}

func (od *Order) Write(costData CostData, tradeData TradeData, kind OrderKind) {
	od.Price = costData.Price
	od.Guid = MkGuid(tradeData.TraderId, tradeData.TradeId)
	od.Amount = costData.Amount
	od.StockId = tradeData.StockId
	od.Kind = kind
}

func (od *Order) UDPAddr() *net.UDPAddr {
	addr := &net.UDPAddr{}
	addr.IP = net.IPv4(od.IP[0], od.IP[1], od.IP[2], od.IP[3])
	addr.Port = int(od.Port)
	return addr
}

func (od *Order) SetUDPAddr(addr *net.UDPAddr) error {
	IP := addr.IP.To4()
	if IP == nil {
		return errors.New(fmt.Sprintf("IP address (%s) is not IPv4", addr.IP.String()))
	}
	od.IP[0], od.IP[1], od.IP[2], od.IP[3] = IP[0], IP[1], IP[2], IP[3]
	od.Port = int32(addr.Port)
	println(od.UDPAddr().String())
	return nil
}

type Response struct {
	Kind         ResponseKind
	Price        int64  // The actual trade price, will be negative if a purchase was made
	Amount       uint32 // The number of units actually bought or sold
	TraderId     uint32 // The trader-id of the trader to whom this response is directed
	TradeId      uint32 // Links this trade back to a previously submitted OrderNode
	CounterParty uint32 // The trader-id of the other half of this trade
	IP           [4]byte
	Port         int32
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

func (r *Response) UDPAddr() *net.UDPAddr {
	addr := &net.UDPAddr{}
	addr.IP = net.IPv4(r.IP[0], r.IP[1], r.IP[2], r.IP[3])
	addr.Port = int(r.Port)
	return addr
}
