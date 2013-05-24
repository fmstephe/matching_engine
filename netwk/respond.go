package netwk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
	"net"
)

type Responder struct {
	responses chan *trade.Response
	resend    []*trade.Response
}

func NewResponder() *Responder {
	return &Responder{}
}

func (r *Responder) SetResponses(responses chan *trade.Response) {
	r.responses = responses
}

func (r *Responder) Run() {
	for {
		select {
		case resp := <-r.responses:
			err := r.write(resp)
			if err != nil {
				println("Responder - ", err.Error())
			}
			if resp.Kind == trade.SHUTDOWN {
				return
			}
		}
	}
}

func (r *Responder) write(resp *trade.Response) error {
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
	if n != trade.SizeofResponse {
		return errors.New(fmt.Sprintf("Insufficient bytes written. Expecting %d, found %d", trade.SizeofResponse, n))
	}
	return nil
}
