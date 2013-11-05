package msg

import (
	"fmt"
	"github.com/fmstephe/fstrconv"
	"unsafe"
)

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
	SHUTDOWN      = MsgKind(iota)
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
	case REJECTED:
		return "REJECTED"
	case SHUTDOWN:
		return "SHUTDOWN"
	}
	panic("Uncreachable")
}

const (
	// Constant price indicating a market price sell
	MARKET_PRICE = 0
)

// Flat description of an incoming message
type Message struct {
	Kind     MsgKind
	Price    uint64
	Amount   uint32
	TraderId uint32
	TradeId  uint32
	StockId  uint32
}

const (
	SizeofMessage = int(unsafe.Sizeof(Message{}))
)

func (m *Message) Valid() bool {
	if m.Kind == SHUTDOWN {
		return m.Price == 0 && m.Amount == 0 && m.TraderId == 0 && m.TradeId == 0 && m.StockId == 0
	}
	// Only sells (and messages cancelling sells) are allowed to have a price of 0
	isValid := (m.Price != 0 || m.Kind == SELL || m.Kind == CANCEL || m.Kind == CANCELLED || m.Kind == NOT_CANCELLED)
	// Remaining fields are never allowed to be 0
	isValid = isValid && m.Amount != 0 && m.TraderId != 0 && m.TradeId != 0 && m.StockId != 0
	// must have a kind
	isValid = isValid && m.Kind != NO_KIND
	return isValid
}

func (m *Message) WriteCancelFor(om *Message) {
	*m = *om
	m.Kind = CANCEL
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
	traderId := fstrconv.Itoa64Delim(int64(m.TraderId), ' ')
	tradeId := fstrconv.Itoa64Delim(int64(m.TradeId), ' ')
	stockId := fstrconv.Itoa64Delim(int64(m.StockId), ' ')
	return fmt.Sprintf("%v, price %s, amount %s, trader %s, trade %s, stock %s", m.Kind, price, amount, traderId, tradeId, stockId)
}
