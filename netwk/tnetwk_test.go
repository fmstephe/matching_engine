package netwk

import (
	"bytes"
	"encoding/binary"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/trade"
	"net"
	"strconv"
	"testing"
)

// NB: There are a number of problems with these tests which are currently being ignored
// 1: Because we are communicating via UDP messages could arrive out of order, in practice they don't travelling via localhost
// 2: The messages are currently not being acked, which means that responses may be resent - which would confuse response checking
//    The reason this doesn't impact the tests right now is that the resend rate is so slow that the test is complete and the system
//    shut down before unacked messages are resent, this is pretty delicate

var localhost = [4]byte{127, 0, 0, 1}

type mockMatcher struct {
	submit chan interface{}
	orders chan *trade.Order
}

func (m *mockMatcher) SetSubmit(submit chan interface{}) {
	m.submit = submit
}

func (m *mockMatcher) SetOrderNodes(orders chan *trade.Order) {
	m.orders = orders
}

func (m *mockMatcher) Run() {
	for {
		o := <-m.orders
		r := &trade.Response{}
		r.Kind = o.Kind
		r.Price = o.Price
		r.Amount = o.Amount
		r.TraderId = o.TraderId
		r.TradeId = o.TradeId
		r.IP = o.IP
		r.Port = o.Port
		r.CounterParty = o.TraderId
		m.submit <- r
	}
}

func TestOrdersAndResponse(t *testing.T) {
	serverPort := 1201
	clientPort := 1202
	setRunning(serverPort)
	read := readConn(clientPort)
	write := writeConn(serverPort)
	confirmNewOrder(t, read, write, &trade.Order{trade.BUY, 1, 2, 3, 4, 5, localhost, int32(clientPort)})
	confirmNewOrder(t, read, write, &trade.Order{trade.BUY, 6, 7, 8, 9, 10, localhost, int32(clientPort)})
	confirmNewOrder(t, read, write, &trade.Order{trade.BUY, 11, 12, 13, 14, 15, localhost, int32(clientPort)})
	shutdownSystem(t, read, write, localhost, int32(clientPort))
}

func TestDuplicateOrders(t *testing.T) {
	serverPort := 1203
	clientPort := 1204
	setRunning(serverPort)
	read := readConn(clientPort)
	write := writeConn(serverPort)
	confirmNewOrder(t, read, write, &trade.Order{trade.BUY, 1, 2, 3, 4, 5, localhost, int32(clientPort)})
	confirmDupOrder(t, read, write, &trade.Order{trade.BUY, 1, 2, 3, 4, 5, localhost, int32(clientPort)})
	shutdownSystem(t, read, write, localhost, int32(clientPort))
}

func setRunning(serverPort int) {
	listener, err := NewListener(strconv.Itoa(serverPort))
	if err != nil {
		panic(err)
	}
	responder := NewResponder()
	matcher := &mockMatcher{}
	coordinator.Coordinate(listener, responder, matcher, false)
}

func writeConn(port int) *net.UDPConn {
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	return conn
}

func readConn(port int) *net.UDPConn {
	addr, err := net.ResolveUDPAddr("upd", ":"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	return conn
}

func confirmNewOrder(t *testing.T, read, write *net.UDPConn, o *trade.Order) {
	sendOrder(t, write, o)
	ack, err := receiveResponse(t, read)
	if err != nil {
		t.Error(err.Error())
		return
	}
	validate(t, o, ack, true)
	r, err := receiveResponse(t, read)
	if err != nil {
		t.Error(err.Error())
		return
	}
	validate(t, o, r, false)
}

func confirmDupOrder(t *testing.T, read, write *net.UDPConn, o *trade.Order) {
	sendOrder(t, write, o)
	ack, err := receiveResponse(t, read)
	if err != nil {
		t.Error(err.Error())
		return
	}
	validate(t, o, ack, true)
}

func shutdownSystem(t *testing.T, read, write *net.UDPConn, ip [4]byte, port int32) {
	o := &trade.Order{}
	o.WriteShutdown()
	o.IP = ip
	o.Port = port
	sendOrder(t, write, o)
	ack, err := receiveResponse(t, read)
	if err != nil {
		t.Error(err.Error())
		return
	}
	validate(t, o, ack, true)
}

func sendOrder(t *testing.T, write *net.UDPConn, o *trade.Order) {
	buf := bytes.NewBuffer(make([]byte, 0))
	binary.Write(buf, binary.BigEndian, o)
	write.Write(buf.Bytes())
}

func receiveResponse(t *testing.T, read *net.UDPConn) (*trade.Response, error) {
	s := make([]byte, trade.SizeofResponse)
	_, _, err := read.ReadFromUDP(s)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(s)
	r := &trade.Response{}
	binary.Read(buf, binary.BigEndian, r)
	return r, nil
}

func validate(t *testing.T, o *trade.Order, r *trade.Response, isAck bool) {
	if isAck && r.Kind != trade.MATCHER_ACK {
		t.Errorf("Expecting %v kind response, found %v", trade.MATCHER_ACK, r.Kind)
	}
	if !isAck && r.Kind != o.Kind {
		t.Errorf("Expecting %v kind response, found %v", trade.FULL, r.Kind)
	}
	if o.Price != r.Price {
		t.Errorf("Price mismatch, expecting %d, found %d", o.Price, r.Price)
	}
	if o.Amount != r.Amount {
		t.Errorf("Amount mismatch, expecting %d, found %d", o.Amount, r.Amount)
	}
	if o.TraderId != r.TraderId {
		t.Errorf("TraderId mismatch, expecting %d, found %d", o.TraderId, r.TraderId)
	}
	if o.TradeId != r.TradeId {
		t.Errorf("TradeId mismatch, expecting %d, found %d", o.TradeId, r.Price)
	}
	if !isAck && o.TraderId != r.CounterParty {
		t.Errorf("Counterparty mismatch, expecting %d, found %d", o.TraderId, r.CounterParty)
	}
	if isAck && r.CounterParty != 0 {
		t.Errorf("Counterparty should be zero because this is an MATCHER_ACK message actual value %d", r.CounterParty)
	}
	if o.IP != r.IP {
		t.Errorf("IP mismatch, expecting %d, found %d", o.IP, r.IP)
	}
	if o.Port != r.Port {
		t.Errorf("Port mismatch, expecting %d, found %d", o.Port, r.Port)
	}
}
