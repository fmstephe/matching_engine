package msg

import (
	"errors"
	"fmt"
	"github.com/fmstephe/fstrconv"
	"net"
)

var routesToKinds = map[MsgRoute] map[MsgKind] bool {
	NO_ROUTE: map[MsgKind] bool {},
	ORDER: map[MsgKind] bool {BUY: true, SELL: true, CANCEL: true},
	CLIENT_ACK: map[MsgKind] bool {PARTIAL: true, FULL: true, CANCELLED: true, NOT_CANCELLED: true},
	COMMAND: map[MsgKind] bool {SHUTDOWN: true},
	RESPONSE: map[MsgKind] bool {PARTIAL: true, FULL: true, CANCELLED: true, NOT_CANCELLED: true},
	SERVER_ACK: map[MsgKind] bool {BUY: true, SELL: true, CANCEL: true, SHUTDOWN: true},
	ERROR: nil, // TODO
}

type MsgRoute int32

const (
	NO_ROUTE = MsgRoute(0)
	// Incoming
	ORDER      = MsgRoute(1)
	CLIENT_ACK = MsgRoute(2)
	COMMAND    = MsgRoute(3)
	// Outgoing
	RESPONSE   = MsgRoute(4)
	SERVER_ACK = MsgRoute(5)
	// Internal
	ERROR = MsgRoute(6)
)

func (r MsgRoute) String() string {
	switch r {
	case NO_ROUTE:
		return "NO_ROUTE"
	case ORDER:
		return "ORDER"
	case CLIENT_ACK:
		return "CLIENT_ACK"
	case COMMAND:
		return "COMMAND"
	case RESPONSE:
		return "RESPONSE"
	case SERVER_ACK:
		return "SERVER_ACK"
	case ERROR:
		return "ERROR"
	}
	panic("Uncreachable")
}

type MsgKind int32

// TODO replace PARTIAL and FULL with PARTIAL_SELL, PARTIAL_BUY and FULL_SELL, FULL_BUY
const (
	NO_KIND = MsgKind(0)
	// Incoming messages
	BUY      = MsgKind(1)
	SELL     = MsgKind(2)
	CANCEL   = MsgKind(3)
	SHUTDOWN = MsgKind(4)
	// Outgoing messages
	PARTIAL       = MsgKind(5)
	FULL          = MsgKind(6)
	CANCELLED     = MsgKind(7)
	NOT_CANCELLED = MsgKind(8)
)

func (k MsgKind) String() string {
	switch k {
	case NO_KIND:
		return "NO_KIND"
	case BUY:
		return "BUY"
	case SELL:
		return "SELL"
	case CANCEL:
		return "CANCEL"
	case SHUTDOWN:
		return "SHUTDOWN"
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

const (
	// Constant price indicating a market price sell
	MARKET_PRICE = 0
)

const (
	SizeofMessage = 40
)

// Flat description of an incoming message
type Message struct {
	Route    MsgRoute
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

func (m *Message) IsValid() bool {
	kinds := routesToKinds[m.Route]
	if !kinds[m.Kind] || m.IP == [4]byte{0,0,0,0} || m.Port == 0 {
		return false
	}
	if m.Kind == SHUTDOWN {
		return m.Price == 0 && m.Amount == 0 && m.TraderId == 0 && m.TradeId == 0 && m.StockId == 0
	} else {
		return (m.Price != 0 || m.Kind == SELL) && m.Amount != 0 && m.TraderId != 0 && m.TradeId != 0 && m.StockId != 0 && m.Port != 0
	}
	panic("Unreachable")
}

func (m *Message) WriteBuy() {
	m.Route = ORDER
	m.Kind = BUY
}

func (m *Message) WriteSell() {
	m.Route = ORDER
	m.Kind = SELL
}

func (m *Message) WriteCancelFor(om *Message) {
	*m = *om
	m.Route = ORDER
	m.Kind = CANCEL
}

func (m *Message) WriteResponse(kind MsgKind) {
	m.Route = RESPONSE
	m.Kind = kind
}

func (m *Message) WriteCancelled() {
	m.Route = RESPONSE
	m.Kind = CANCELLED
}

func (m *Message) WriteNotCancelled() {
	m.Route = RESPONSE
	m.Kind = NOT_CANCELLED
}

func (m *Message) WriteShutdown() {
	m.Route = COMMAND
	m.Kind = SHUTDOWN
}

func (m *Message) WriteServerAckFor(om *Message) {
	*m = *om
	m.Route = SERVER_ACK
}

func (m *Message) WriteClientAckFor(om *Message) {
	*m = *om
	m.Route = CLIENT_ACK
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
	return nil
}

func (m *Message) String() string {
	if m == nil {
		return "<nil>"
	}
	//	price := fstrconv.Itoa64Delim(int64(m.Price), ',')
	amount := fstrconv.Itoa64Delim(int64(m.Amount), ',')
	traderId := fstrconv.Itoa64Delim(int64(m.TraderId), '-')
	tradeId := fstrconv.Itoa64Delim(int64(m.TradeId), '-')
	stockId := fstrconv.Itoa64Delim(int64(m.StockId), '-')
	return fmt.Sprintf("(%v %v), price %v, amount %s, trader %s, trade %s, stock %s, ip %v, port %v", m.Route, m.Kind, m.Price, amount, traderId, tradeId, stockId, m.IP, m.Port)
}
