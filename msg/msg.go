package msg

import (
	"errors"
	"fmt"
	"net"
)

type MsgKind int32

const (
	ILLEGAL = MsgKind(0) // 0 is never a valid kind
	// Incoming messages
	CLIENT_ACK = MsgKind(1) // Indicates the client has received a message from the matcher, NOT CURRENTLY USED
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
	SizeofMessage = 36
)

func (k MsgKind) IsOrder() bool {
	return k == BUY || k == SELL || k == CANCEL
}

func (k MsgKind) IsResponse() bool {
	return k == PARTIAL || k == FULL || k == CANCELLED || k == NOT_CANCELLED || k == MATCHER_ACK
}

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

// For readable constructors
type NetData struct {
	IP   [4]byte // IPv4 address of client
	Port int32   // Port of client
}

// Flat description of an incoming message
type Message struct {
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

func (m *Message) WriteBuy(costData CostData, tradeData TradeData, netData NetData) {
	m.Write(costData, tradeData, netData, BUY)
}

func (m *Message) WriteSell(costData CostData, tradeData TradeData, netData NetData) {
	m.Write(costData, tradeData, netData, SELL)
}

func (m *Message) WriteCancel(tradeData TradeData, netData NetData) {
	m.Write(CostData{}, tradeData, netData, CANCEL)
}

func (m *Message) WriteCancelFor(om *Message) {
	*m = *om
	m.Kind = CANCEL
}

func (m *Message) WriteShutdown() {
	m.Write(CostData{}, TradeData{}, NetData{}, SHUTDOWN)
}

func (m *Message) WriteMatcherAck(om *Message) {
	*m = *om
	m.Kind = MATCHER_ACK
}

func (m *Message) WriteClientAck(om *Message) {
	*m = *om
	m.Kind = CLIENT_ACK
}

func (m *Message) WriteCancelled(costData CostData, tradeData TradeData, netData NetData) {
	m.Write(costData, tradeData, netData, CANCELLED)
}

func (m *Message) WriteNotCancelled(costData CostData, tradeData TradeData, netData NetData) {
	m.Write(costData, tradeData, netData, NOT_CANCELLED)
}

// This function is used to cover PARTIAL and FULL messages
func (m *Message) WriteMatch(price int64, amount, traderId, tradeId, stockId uint32, ip [4]byte, port int32, kind MsgKind) {
	m.Kind = kind
	m.Price = price
	m.Amount = amount
	m.TraderId = traderId
	m.TradeId = tradeId
	m.StockId = stockId
	m.IP = ip
	m.Port = port
}

func (m *Message) Write(costData CostData, tradeData TradeData, netData NetData, kind MsgKind) {
	m.Kind = kind
	m.Price = costData.Price
	m.TraderId = tradeData.TraderId
	m.TradeId = tradeData.TradeId
	m.Amount = costData.Amount
	m.StockId = tradeData.StockId
	m.IP = netData.IP
	m.Port = netData.Port
}

func (m *Message) UDPAddr() *net.UDPAddr {
	addr := &net.UDPAddr{}
	addr.IP = net.IPv4(m.IP[0], m.IP[1], m.IP[2], m.IP[3])
	addr.Port = int(m.Port)
	return addr
}

func (m *Message) SetUDPAddr(addr *net.UDPAddr) error {
	IP := addr.IP.To4()
	if IP == nil {
		return errors.New(fmt.Sprintf("IP address (%s) is not IPv4", addr.IP.String()))
	}
	m.IP[0], m.IP[1], m.IP[2], m.IP[3] = IP[0], IP[1], IP[2], IP[3]
	m.Port = int32(addr.Port)
	println(m.UDPAddr().String())
	return nil
}
