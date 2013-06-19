package msg

import (
	"errors"
	"fmt"
	"github.com/fmstephe/fstrconv"
	"net"
)

type MsgStatus uint32

const (
	NORMAL             = MsgStatus(0)
	SENDABLE_ERROR     = MsgStatus(1)
	NOT_SENDABLE_ERROR = MsgStatus(2)
)

func (s MsgStatus) String() string {
	switch s {
	case NORMAL:
		return "NORMAL"
	case SENDABLE_ERROR:
		return "SENDABLE_ERROR"
	case NOT_SENDABLE_ERROR:
		return "NOT_SENDABLE_ERROR"
	}
	panic("Unreachable")
}

type MsgRoute uint32

const (
	NO_ROUTE = MsgRoute(0)
	// Incoming
	ORDER      = MsgRoute(1)
	CLIENT_ACK = MsgRoute(2)
	COMMAND    = MsgRoute(3)
	// Outgoing
	RESPONSE   = MsgRoute(4)
	SERVER_ACK = MsgRoute(5)
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
	}
	panic("Uncreachable")
}

type MsgKind uint32

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
	SizeofMessage = 44
)

var routesToKinds = map[MsgRoute]map[MsgKind]bool{
	NO_ROUTE:   map[MsgKind]bool{},
	ORDER:      map[MsgKind]bool{BUY: true, SELL: true, CANCEL: true},
	CLIENT_ACK: map[MsgKind]bool{PARTIAL: true, FULL: true, CANCELLED: true, NOT_CANCELLED: true},
	COMMAND:    map[MsgKind]bool{SHUTDOWN: true},
	RESPONSE:   map[MsgKind]bool{PARTIAL: true, FULL: true, CANCELLED: true, NOT_CANCELLED: true},
	SERVER_ACK: map[MsgKind]bool{BUY: true, SELL: true, CANCEL: true, SHUTDOWN: true},
}

// Flat description of an incoming message
type Message struct {
	Status   MsgStatus
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

func (m *Message) Valid() bool {
	if m.Status == NOT_SENDABLE_ERROR {
		return true
	}
	if m.Status == SENDABLE_ERROR {
		return m.Networked()
	}
	kinds := routesToKinds[m.Route]
	if !kinds[m.Kind] || !m.Networked() {
		return false
	}
	if m.Kind == SHUTDOWN {
		return m.Price == 0 && m.Amount == 0 && m.TraderId == 0 && m.TradeId == 0 && m.StockId == 0
	} else {
		isValid := (m.Price != 0 || m.Kind == SELL || m.Kind == CANCEL || m.Kind == CANCELLED || m.Kind == NOT_CANCELLED)
		isValid = isValid && m.Amount != 0 && m.TraderId != 0 && m.TradeId != 0 && m.StockId != 0
		return isValid
	}
	panic("Unreachable")
}

func (m *Message) Networked() bool {
	return m.IP != [4]byte{} && m.Port != 0
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

func (m *Message) WriteStatus(status MsgStatus) {
	m.Status = status
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
	price := fstrconv.Itoa64Delim(int64(m.Price), ',')
	amount := fstrconv.Itoa64Delim(int64(m.Amount), ',')
	traderId := fstrconv.Itoa64Delim(int64(m.TraderId), '-')
	tradeId := fstrconv.Itoa64Delim(int64(m.TradeId), '-')
	stockId := fstrconv.Itoa64Delim(int64(m.StockId), '-')
	status := ""
	if m.Status != NORMAL {
		status = m.Status.String() + "! "
	}
	return fmt.Sprintf("%s(%v %v), price %v, amount %s, trader %s, trade %s, stock %s, ip %v, port %v", status, m.Route, m.Kind, price, amount, traderId, tradeId, stockId, m.IP, m.Port)
}
