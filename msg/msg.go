package msg

import (
	"fmt"
	"github.com/fmstephe/fstrconv"
)

type MsgStatus uint32

const (
	NORMAL = MsgStatus(0)
	ERROR  = MsgStatus(1)
)

func (s MsgStatus) String() string {
	switch s {
	case NORMAL:
		return "NORMAL"
	case ERROR:
		return "ERROR"
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
	SizeofMessage = 36
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
	// I think we need a checksum here
}

func (m *Message) Valid() bool {
	if m.Status == ERROR {
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
