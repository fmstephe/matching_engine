package netwk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
	"io"
	"net"
	"os"
	"time"
)

const RESEND_MILLIS = time.Duration(100) * time.Millisecond

type Responder struct {
	responses chan *msg.Message
	dispatch  chan *msg.Message
	unacked   []*msg.Message
	writer    io.WriteCloser
}

func NewResponder(writer io.WriteCloser) *Responder {
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
			case resp.Status == msg.ERROR, resp.Route == msg.RESPONSE, resp.Route == msg.SERVER_ACK:
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

func (r *Responder) handleClientAck(ca *msg.Message) {
	unacked := r.unacked
	for i, uResp := range unacked {
		if ca.TraderId == uResp.TraderId && ca.TradeId == uResp.TradeId {
			unacked[i] = unacked[len(unacked)-1]
			unacked = unacked[:len(unacked)-1]
			r.unacked = unacked
			return
		}
	}
}

func (r *Responder) writeResponse(resp *msg.Message) {
	r.addToUnacked(resp)
	r.write(resp)
}

func (r *Responder) addToUnacked(resp *msg.Message) {
	if resp.Route == msg.RESPONSE {
		r.unacked = append(r.unacked, resp)
	}
}

func (r *Responder) resend() {
	for _, resp := range r.unacked {
		r.write(resp)
	}
}

func (r *Responder) write(resp *msg.Message) {
	nbuf := &bytes.Buffer{}
	err := binary.Write(nbuf, binary.BigEndian, resp)
	if err != nil {
		r.handleError(resp, err)
	}
	n, err := r.writer.Write(nbuf.Bytes())
	if err != nil {
		r.handleError(resp, err)
	}
	if n != msg.SizeofMessage {
		cerr := errors.New(fmt.Sprintf("Insufficient data written. Expecting %d, found %d", msg.SizeofMessage, n))
		r.handleError(resp, cerr)
	}
}

func (r *Responder) handleError(resp *msg.Message, err error) {
	em := &msg.Message{}
	*em = *resp
	em.WriteStatus(msg.ERROR)
	if e, ok := err.(net.Error); ok && !e.Temporary() {
		os.Exit(1)
	}
	r.dispatch <- em
}

func (r *Responder) shutdown() {
	r.writer.Close()
}
