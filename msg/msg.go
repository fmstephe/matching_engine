package msg

import (
	"fmt"
	"github.com/fmstephe/fstrconv"
	"unsafe"
)

type MsgStatus byte

const (
	NORMAL            = MsgStatus(iota)
	INVALID_MSG_ERROR = MsgStatus(iota)
	READ_ERROR        = MsgStatus(iota)
	SMALL_READ_ERROR  = MsgStatus(iota)
	WRITE_ERROR       = MsgStatus(iota)
	SMALL_WRITE_ERROR = MsgStatus(iota)
)

func (s MsgStatus) String() string {
	switch s {
	case NORMAL:
		return "NORMAL"
	case INVALID_MSG_ERROR:
		return "INVALID_MSG_ERROR"
	case READ_ERROR:
		return "READ_ERROR"
	case SMALL_READ_ERROR:
		return "SMALL_READ_ERROR"
	case WRITE_ERROR:
		return "WRITE_ERROR"
	case SMALL_WRITE_ERROR:
		return "SMALL_WRITE_ERROR"
	}
	panic("Unreachable")
}

type MsgRoute byte

const (
	NO_ROUTE = MsgRoute(0)
	// Incoming
	ORDER      = MsgRoute(1)
	CLIENT_ACK = MsgRoute(2)
	COMMAND    = MsgRoute(3)
	// Outgoing
	MATCHER_RESPONSE = MsgRoute(4)
	SERVER_ACK       = MsgRoute(5)
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
	case MATCHER_RESPONSE:
		return "MATCHER_RESPONSE"
	case SERVER_ACK:
		return "SERVER_ACK"
	}
	panic("Uncreachable")
}

type MsgKind byte

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

var routesToKinds = map[MsgRoute]map[MsgKind]bool{
	NO_ROUTE:         map[MsgKind]bool{},
	ORDER:            map[MsgKind]bool{BUY: true, SELL: true, CANCEL: true},
	CLIENT_ACK:       map[MsgKind]bool{PARTIAL: true, FULL: true, CANCELLED: true, NOT_CANCELLED: true},
	COMMAND:          map[MsgKind]bool{SHUTDOWN: true},
	MATCHER_RESPONSE: map[MsgKind]bool{PARTIAL: true, FULL: true, CANCELLED: true, NOT_CANCELLED: true},
	SERVER_ACK:       map[MsgKind]bool{BUY: true, SELL: true, CANCEL: true, SHUTDOWN: true},
}

// Flat description of an incoming message
type Message struct {
	pad8     byte
	Status   MsgStatus
	Route    MsgRoute
	Kind     MsgKind
	pad32    uint32
	Price    int64
	Amount   uint32
	TraderId uint32
	TradeId  uint32
	StockId  uint32
	// I think we need a checksum here
}

const (
	SizeofMessage = int(unsafe.Sizeof(Message{}))
)

func (m *Message) Valid() bool {
	if m.Status != NORMAL {
		return true
	}
	kinds := routesToKinds[m.Route]
	if !kinds[m.Kind] {
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
	m.Route = MATCHER_RESPONSE
	m.Kind = kind
}

func (m *Message) WriteCancelled() {
	m.Route = MATCHER_RESPONSE
	m.Kind = CANCELLED
}

func (m *Message) WriteNotCancelled() {
	m.Route = MATCHER_RESPONSE
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
	return fmt.Sprintf("%s(%v %v), price %v, amount %s, trader %s, trade %s, stock %s", status, m.Route, m.Kind, price, amount, traderId, tradeId, stockId)
}

func (m *Message) WriteTo(b []byte) {
	p := unsafe.Pointer(m)
	mb := (*([SizeofMessage]byte))(p)[:]
	copy(b, mb)
}

func (m *Message) WriteFrom(b []byte) {
	p := unsafe.Pointer(m)
	mb := (*([SizeofMessage]byte))(p)[:]
	copy(mb, b)
}
