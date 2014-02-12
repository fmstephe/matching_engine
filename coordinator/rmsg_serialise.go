package coordinator

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
)

const (
	msgOffset       = 0                 // msg.ByteSize bytes (40)
	statusOffset    = msg.ByteSize + 0  // 1 byte
	directionOffset = msg.ByteSize + 1  // 1 byte
	routeOffset     = msg.ByteSize + 2  // 1 byte
	originIdOffset  = msg.ByteSize + 3  // 4 bytes
	msgIdOffset     = msg.ByteSize + 7  // 4 bytes
	rmsgByteSize    = msg.ByteSize + 11 // (51)
)

var binCoder = binary.LittleEndian

// Populate NMessage with *Message values
func (rm *RMessage) Marshal(b []byte) error {
	if len(b) != rmsgByteSize {
		return errors.New(fmt.Sprintf("Wrong sized byte buffer. Expecting %d, found %d", rmsgByteSize, len(b)))
	}
	(&rm.message).Marshal(b[:msg.ByteSize])
	b[statusOffset] = byte(rm.status)
	b[directionOffset] = byte(rm.direction)
	b[routeOffset] = byte(rm.route)
	binCoder.PutUint32(b[originIdOffset:msgIdOffset], rm.originId)
	binCoder.PutUint32(b[msgIdOffset:], rm.msgId)
	return nil
}

// Populate *Message with NMessage values
func (rm *RMessage) Unmarshal(b []byte) error {
	if len(b) != rmsgByteSize {
		return errors.New(fmt.Sprintf("Wrong sized byte buffer. Expecting %d, found %d", rmsgByteSize, len(b)))
	}
	(&rm.message).Unmarshal(b[:msg.ByteSize])
	rm.status = MsgStatus(b[statusOffset])
	rm.direction = MsgDirection(b[directionOffset])
	rm.route = MsgRoute(b[routeOffset])
	rm.originId = binCoder.Uint32(b[originIdOffset:msgIdOffset])
	rm.msgId = binCoder.Uint32(b[msgIdOffset:])
	return nil
}
