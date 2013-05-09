package netwk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
	"net"
)

type Responder struct {
	responses chan *trade.Response
}

func NewResponder() *Responder {
	return &Responder{}
}

func (r *Responder) SetResponses(responses chan *trade.Response) {
	r.responses = responses
}

func (r *Responder) Run() {
	for {
		resp := <-r.responses
		nbuf := &bytes.Buffer{}
		err := binary.Write(nbuf, binary.BigEndian, resp)
		if err != nil {
			println("Responder: ", err.Error())
			continue
		}
		conn, err := net.DialUDP("udp", nil, resp.UDPAddr())
		if err != nil {
			println("Responder: ", err.Error())
			continue
		}
		n, err := conn.Write(nbuf.Bytes())
		if err != nil {
			println("Responder: ", err.Error())
			continue
		}
		if n != trade.SizeofResponse {
			println(fmt.Sprintf("Responder: Insufficient bytes written in response. Expecting %d, found %d", trade.SizeofResponse, n))
			continue
		}
	}
}
