package netwk

import (
	"bytes"
	"encoding/binary"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/msg"
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
	submit chan *msg.Message
	orders chan *msg.Message
}

func (m *mockMatcher) SetSubmit(submit chan *msg.Message) {
	m.submit = submit
}

func (m *mockMatcher) SetOrders(orders chan *msg.Message) {
	m.orders = orders
}

func (m *mockMatcher) Run() {
	for {
		o := <-m.orders
		r := &msg.Message{}
		r.Kind = msg.FULL
		r.Price = o.Price
		r.Amount = o.Amount
		r.TraderId = o.TraderId
		r.TradeId = o.TradeId
		r.IP = o.IP
		r.Port = o.Port
		r.StockId = o.StockId
		m.submit <- r
	}
}

func TestOrdersAndResponses(t *testing.T) {
	serverPort := 1201
	clientPort := 1202
	setRunning(serverPort)
	read := readConn(clientPort)
	write := writeConn(serverPort)
	confirmNewMessage(t, read, write, &msg.Message{msg.BUY, 1, 2, 3, 4, 5, localhost, int32(clientPort)})
	confirmNewMessage(t, read, write, &msg.Message{msg.BUY, 6, 7, 8, 9, 10, localhost, int32(clientPort)})
	confirmNewMessage(t, read, write, &msg.Message{msg.BUY, 11, 12, 13, 14, 15, localhost, int32(clientPort)})
	shutdownSystem(t, read, write, localhost, int32(clientPort))
}

func TestDuplicateOrders(t *testing.T) {
	serverPort := 1203
	clientPort := 1204
	setRunning(serverPort)
	read := readConn(clientPort)
	write := writeConn(serverPort)
	confirmNewMessage(t, read, write, &msg.Message{msg.BUY, 1, 2, 3, 4, 5, localhost, int32(clientPort)})
	confirmDupMessage(t, read, write, &msg.Message{msg.BUY, 1, 2, 3, 4, 5, localhost, int32(clientPort)})
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

func confirmNewMessage(t *testing.T, read, write *net.UDPConn, o *msg.Message) {
	sendMessage(t, write, o)
	ack, err := receiveMessage(t, read)
	if err != nil {
		t.Error(err.Error())
		return
	}
	validate(t, o, ack, true)
	r, err := receiveMessage(t, read)
	if err != nil {
		t.Error(err.Error())
		return
	}
	validate(t, o, r, false)
}

func confirmDupMessage(t *testing.T, read, write *net.UDPConn, o *msg.Message) {
	sendMessage(t, write, o)
	ack, err := receiveMessage(t, read)
	if err != nil {
		t.Error(err.Error())
		return
	}
	validate(t, o, ack, true)
}

func shutdownSystem(t *testing.T, read, write *net.UDPConn, ip [4]byte, port int32) {
	o := &msg.Message{}
	o.WriteShutdown()
	o.IP = ip
	o.Port = port
	sendMessage(t, write, o)
	ack, err := receiveMessage(t, read)
	if err != nil {
		t.Error(err.Error())
		return
	}
	validate(t, o, ack, true)
}

func sendMessage(t *testing.T, write *net.UDPConn, o *msg.Message) {
	buf := bytes.NewBuffer(make([]byte, 0))
	binary.Write(buf, binary.BigEndian, o)
	write.Write(buf.Bytes())
}

func receiveMessage(t *testing.T, read *net.UDPConn) (*msg.Message, error) {
	s := make([]byte, msg.SizeofMessage)
	_, _, err := read.ReadFromUDP(s)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(s)
	r := &msg.Message{}
	binary.Read(buf, binary.BigEndian, r)
	return r, nil
}

func validate(t *testing.T, order *msg.Message, resp *msg.Message, isAck bool) {
	if isAck && resp.Kind != msg.MATCHER_ACK {
		t.Errorf("Expecting %v kind response, found %v", msg.MATCHER_ACK, resp.Kind)
	}
	if !isAck && resp.Kind != msg.FULL {
		t.Errorf("Expecting %v kind response, found %v", msg.FULL, resp.Kind)
	}
	if order.Price != resp.Price {
		t.Errorf("Price mismatch, expecting %d, found %d", order.Price, resp.Price)
	}
	if order.Amount != resp.Amount {
		t.Errorf("Amount mismatch, expecting %d, found %d", order.Amount, resp.Amount)
	}
	if order.TraderId != resp.TraderId {
		t.Errorf("TraderId mismatch, expecting %d, found %d", order.TraderId, resp.TraderId)
	}
	if order.TradeId != resp.TradeId {
		t.Errorf("TradeId mismatch, expecting %d, found %d", order.TradeId, resp.Price)
	}
	if order.StockId != resp.StockId {
		t.Errorf("Counterparty mismatch, expecting %d, found %d", order.StockId, resp.StockId)
	}
	if order.IP != resp.IP {
		t.Errorf("IP mismatch, expecting %d, found %d", order.IP, resp.IP)
	}
	if order.Port != resp.Port {
		t.Errorf("Port mismatch, expecting %d, found %d", order.Port, resp.Port)
	}
}
