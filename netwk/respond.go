package netwk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
	"net"
)

type Responder struct {
	respChan chan *trade.Response
}

func NewResponder(respChan chan *trade.Response) *Responder {
	return &Responder{respChan: respChan}
}

func (r *Responder) Respond() {
	for {
		resp := <-r.respChan
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
