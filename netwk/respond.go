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
		var conn *net.UDPConn
		var n int
		resp := <-r.responses
		nbuf := &bytes.Buffer{}
		err := binary.Write(nbuf, binary.BigEndian, resp)
		if err != nil {
			println("Responder: ", err.Error())
			goto SHUTDOWN_CHECK
		}
		conn, err = net.DialUDP("udp", nil, resp.UDPAddr())
		if err != nil {
			println("Responder: ", err.Error())
			goto SHUTDOWN_CHECK
		}
		n, err = conn.Write(nbuf.Bytes())
		if err != nil {
			println("Responder: ", err.Error())
			goto SHUTDOWN_CHECK
		}
		if n != trade.SizeofResponse {
			println(fmt.Sprintf("Responder: Insufficient bytes written in response. Expecting %d, found %d", trade.SizeofResponse, n))
			goto SHUTDOWN_CHECK
		}
	SHUTDOWN_CHECK:
		if resp.Kind == trade.SHUTDOWN {
			return
		}
	}
}
