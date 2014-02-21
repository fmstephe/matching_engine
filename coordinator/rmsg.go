package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
)

// TODO this can probably be removed. If we log all reads but don't pass the malformed ones into the application then this is redundant.
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
	panic("Bad Value")
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
	panic("Bad Value")
}

type MsgRoute byte

const (
	NO_ROUTE     = MsgRoute(iota)
	APP          = MsgRoute(iota)
	ACK          = MsgRoute(iota)
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
	}
	panic("Bad Value")
}

// Flat description of an incoming message
type RMessage struct {
	// Body
	message msg.Message
	// Headers
	status    MsgStatus
	direction MsgDirection
	route     MsgRoute
	originId  uint32
	msgId     uint32
}

func (rm *RMessage) Valid() bool {
	// A message must always have a direction
	if rm.direction == NO_DIRECTION {
		return false
	}
	// Any message in an error status is valid
	if rm.status != NORMAL {
		return true
	}
	// Zero values for origin and Id are not valid
	if rm.originId == 0 || rm.msgId == 0 {
		return false
	}
	return rm.message.Valid()
}

func (rm *RMessage) WriteAckFor(orm *RMessage) {
	*rm = *orm
	rm.route = ACK
	rm.direction = OUT
}

func (rm *RMessage) String() string {
	if rm == nil {
		return "<nil>"
	}
	status := ""
	if rm.status != NORMAL {
		status = rm.status.String() + "! "
	}
	return fmt.Sprintf("(%s%v %v, %d, %d), %s", status, rm.direction, rm.route, rm.originId, rm.msgId, rm.message.String())
}
