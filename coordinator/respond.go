package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
	"github.com/fmstephe/matching_engine/msg/msgutil"
	"io"
	"net"
	"os"
	"time"
)

const RESEND_MILLIS = time.Duration(10) * time.Millisecond

type stdResponder struct {
	responses chan *msg.Message
	dispatch  chan *msg.Message
	unacked   *msgutil.Set
	writer    io.WriteCloser
}

func newResponder(writer io.WriteCloser) responder {
	return &stdResponder{unacked: msgutil.NewSet(), writer: writer}
}

func (r *stdResponder) SetResponses(responses chan *msg.Message) {
	r.responses = responses
}

func (r *stdResponder) SetDispatch(dispatch chan *msg.Message) {
	r.dispatch = dispatch
}

func (r *stdResponder) Run() {
	defer r.shutdown()
	t := time.NewTimer(RESEND_MILLIS)
	for {
		select {
		case resp := <-r.responses:
			switch {
			case resp.Direction == msg.IN && resp.Route == msg.ACK:
				r.handleInAck(resp)
			case resp.Direction == msg.OUT && (resp.Status != msg.NORMAL || resp.Route == msg.APP || resp.Route == msg.ACK):
				r.writeResponse(resp)
			case resp.Route == msg.SHUTDOWN:
				return
			default:
				panic(fmt.Sprintf("Unhandleable response %v", resp))
			}
		case <-t.C:
			r.resend()
			t = time.NewTimer(RESEND_MILLIS)
		}
	}
}

func (r *stdResponder) handleInAck(ca *msg.Message) {
	r.unacked.Remove(ca)
}

func (r *stdResponder) writeResponse(resp *msg.Message) {
	resp.Direction = msg.IN
	r.addToUnacked(resp)
	r.write(resp)
}

func (r *stdResponder) addToUnacked(resp *msg.Message) {
	if resp.Route == msg.APP {
		r.unacked.Add(resp)
	}
}

func (r *stdResponder) resend() {
	// There is a way to turn r.Write into a closure directly - but needs go 1.1
	r.unacked.Do(func(m *msg.Message) {
		r.write(m)
	})
}

func (r *stdResponder) write(resp *msg.Message) {
	b := make([]byte, msg.SizeofMessage)
	resp.WriteTo(b)
	n, err := r.writer.Write(b)
	if err != nil {
		r.handleError(resp, err, msg.WRITE_ERROR)
	}
	if n != msg.SizeofMessage {
		r.handleError(resp, err, msg.SMALL_WRITE_ERROR)
	}
}

func (r *stdResponder) handleError(resp *msg.Message, err error, s msg.MsgStatus) {
	em := &msg.Message{}
	*em = *resp
	em.Status = s
	r.dispatch <- em
	println(err.Error())
	if e, ok := err.(net.Error); ok && !e.Temporary() {
		os.Exit(1)
	}
}

func (r *stdResponder) shutdown() {
	r.writer.Close()
}
