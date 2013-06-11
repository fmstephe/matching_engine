package netwk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
	"net"
	"time"
)

const RESEND_MILLIS = time.Duration(100) * time.Millisecond

type Responder struct {
	responses chan *msg.Message
	unacked   []*msg.Message
}

func NewResponder() *Responder {
	return &Responder{unacked: make([]*msg.Message, 0, 100)}
}

func (r *Responder) SetResponses(responses chan *msg.Message) {
	r.responses = responses
}

func (r *Responder) Run() {
	defer shutdown()
	t := time.NewTimer(RESEND_MILLIS)
	for {
		select {
		case resp := <-r.responses:
			switch {
			case resp.Route == msg.COMMAND && resp.Kind == msg.SHUTDOWN:
				return
			case resp.Route == msg.CLIENT_ACK:
				r.handleClientAck(resp)
			case resp.Route == msg.RESPONSE, resp.Route == msg.SERVER_ACK:
				r.writeResponse(resp)
			}
		case <-t.C:
			r.resend()
			t = time.NewTimer(RESEND_MILLIS)
		}
	}
}

func (r *Responder) handleClientAck(ca *msg.Message) {
	unacked := r.unacked
	for i, uResp := range unacked {
		if ca.TraderId == uResp.TraderId && ca.TradeId == uResp.TradeId {
			unacked[i] = unacked[len(unacked)-1]
			unacked = unacked[:len(unacked)-1]
			// Corner cases?
		}
	}
	r.unacked = unacked
}

func (r *Responder) writeResponse(resp *msg.Message) {
	if resp.Route == msg.RESPONSE {
		r.unacked = append(r.unacked, resp)
	}
	err := r.write(resp)
	if err != nil {
		// TODO this should be a message sent back to the coordinator
		println("Responder - ", err.Error())
	}
}

func (r *Responder) resend() {
	for _, resp := range r.unacked {
		r.write(resp)
	}
}

func (r *Responder) write(resp *msg.Message) error {
	nbuf := &bytes.Buffer{}
	err := binary.Write(nbuf, binary.BigEndian, resp)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, resp.UDPAddr())
	if err != nil {
		return err
	}
	n, err := conn.Write(nbuf.Bytes())
	if err != nil {
		return err
	}
	if n != msg.SizeofMessage {
		return errors.New(fmt.Sprintf("Insufficient bytes written. Expecting %d, found %d", msg.SizeofMessage, n))
	}
	return nil
}

func shutdown() {
}
