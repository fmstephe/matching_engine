package netwk

import (
	"bytes"
	"encoding/binary"
	"github.com/fmstephe/matching_engine/msg"
	"net"
	"time"
)

const RESEND_MILLIS = time.Duration(100) * time.Millisecond

type ipWriter interface {
	Write(data []byte, ip [4]byte, port int) error
}

type Responder struct {
	responses chan *msg.Message
	dispatch  chan *msg.Message
	unacked   []*msg.Message
	writer    ipWriter
}

func NewResponder(writer ipWriter) *Responder {
	return &Responder{unacked: make([]*msg.Message, 0, 100), writer: writer}
}

func (r *Responder) SetResponses(responses chan *msg.Message) {
	r.responses = responses
}

func (r *Responder) SetDispatch(dispatch chan *msg.Message) {
	r.dispatch = dispatch
}

func (r *Responder) Run() {
	defer r.shutdown()
	t := time.NewTimer(RESEND_MILLIS)
	for {
		select {
		case resp := <-r.responses:
			switch {
			case resp.Status == msg.NOT_SENDABLE_ERROR:
				panic("Not sendable error sent to responder. Probable infinite loop.")
			case resp.Status == msg.SENDABLE_ERROR, resp.Route == msg.RESPONSE, resp.Route == msg.SERVER_ACK:
				r.writeResponse(resp)
			case resp.Route == msg.CLIENT_ACK:
				r.handleClientAck(resp)
			case resp.Route == msg.COMMAND && resp.Kind == msg.SHUTDOWN:
				return
			}
		case <-t.C:
			r.resend()
			t = time.NewTimer(RESEND_MILLIS)
		}
	}
}

// TODO write some pure unit tests around the unacked feature
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
		r.handleError(resp, err)
	}
}

func (r *Responder) handleError(resp *msg.Message, err error) {
	em := &msg.Message{}
	*em = *resp
	em.WriteStatus(msg.NOT_SENDABLE_ERROR)
	if e, ok := err.(net.Error); ok && e.Temporary() {
		em.WriteStatus(msg.SENDABLE_ERROR)
	}
	r.dispatch <- em
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
	return r.writer.Write(nbuf.Bytes(), resp.IP, int(resp.Port))
}

func (r *Responder) shutdown() {
}
