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
	NUM_OF_STATUS     = int32(iota)
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

type MsgDirection byte

const (
	NO_DIRECTION = MsgDirection(iota)
	OUT          = MsgDirection(iota)
	IN           = MsgDirection(iota)
)

func (d MsgDirection) String() string {
	switch d {
	case NO_DIRECTION:
		return "NO_DIRECTION"
	case IN:
		return "IN"
	case OUT:
		return "OUT"
	}
	panic("unreachable")
}

type MsgRoute byte

const (
	NO_ROUTE     = MsgRoute(iota)
	APP          = MsgRoute(iota)
	ACK          = MsgRoute(iota)
	SHUTDOWN     = MsgRoute(iota)
	NUM_OF_ROUTE = int32(iota)
)

func (r MsgRoute) String() string {
	switch r {
	case NO_ROUTE:
		return "NO_ROUTE"
	case APP:
		return "APP"
	case ACK:
		return "ACK"
	case SHUTDOWN:
		return "SHUTDOWN"
	}
	panic("Uncreachable")
}

type MsgKind byte

const (
	NO_KIND       = MsgKind(iota)
	BUY           = MsgKind(iota)
	SELL          = MsgKind(iota)
	CANCEL        = MsgKind(iota)
	PARTIAL       = MsgKind(iota)
	FULL          = MsgKind(iota)
	CANCELLED     = MsgKind(iota)
	NOT_CANCELLED = MsgKind(iota)
	REJECTED      = MsgKind(iota)
	NUM_OF_KIND   = int32(iota)
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

// Flat description of an incoming message
type Message struct {
	// Headers
	Status    MsgStatus
	Direction MsgDirection
	Route     MsgRoute
	// Body
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
	// A message must always have a direction
	if m.Direction == NO_DIRECTION {
		return false
	}
	// Any message in an error status is valid
	if m.Status != NORMAL {
		return true
	}
	// Shutdown messages must be blank
	if m.Route == SHUTDOWN {
		return m.Kind == NO_KIND && m.Price == 0 && m.Amount == 0 && m.TraderId == 0 && m.TradeId == 0 && m.StockId == 0
	}
	// (Ack, APP) must have a Kind
	if (m.Route != ACK && m.Route != APP) || m.Kind == NO_KIND {
		return false
	}
	// Only sells (and messages cancelling sells) are allowed to have a price of 0
	isValid := (m.Price != 0 || m.Kind == SELL || m.Kind == CANCEL || m.Kind == CANCELLED || m.Kind == NOT_CANCELLED)
	// Remaining fields are never allowed to be 0
	isValid = isValid && m.Amount != 0 && m.TraderId != 0 && m.TradeId != 0 && m.StockId != 0
	return isValid
}

func (m *Message) WriteApp(kind MsgKind) {
	m.Route = APP
	m.Kind = kind
	m.Direction = OUT
}

func (m *Message) WriteCancelFor(om *Message) {
	*m = *om
	m.Route = APP
	m.Kind = CANCEL
	m.Direction = OUT
}

func (m *Message) WriteAckFor(om *Message) {
	*m = *om
	m.Route = ACK
	m.Direction = OUT
}

func (m *Message) WriteShutdown() {
	m.Route = SHUTDOWN
	m.Kind = NO_KIND
	m.Direction = OUT
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
	return fmt.Sprintf("%s(%v %v %v), price %v, amount %s, trader %s, trade %s, stock %s", status, m.Direction, m.Route, m.Kind, price, amount, traderId, tradeId, stockId)
}
