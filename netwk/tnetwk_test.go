package netwk

import (
	"bytes"
	"encoding/binary"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/matcher"
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

func TestOrdersAndAck(t *testing.T) {
	serverPort := 1201
	clientPort := 1202
	setRunning(serverPort)
	read := readConn(clientPort)
	write := writeConn(serverPort)
	send(t, write, &msg.Message{msg.ORDER, msg.BUY, 1, 2, 3, 4, 5, localhost, int32(clientPort)})
	expect(t, read, msg.SERVER_ACK, msg.BUY)
	send(t, write, &msg.Message{msg.ORDER, msg.BUY, 6, 7, 8, 9, 10, localhost, int32(clientPort)})
	expect(t, read, msg.SERVER_ACK, msg.BUY)
	send(t, write, &msg.Message{msg.ORDER, msg.BUY, 11, 12, 13, 14, 15, localhost, int32(clientPort)})
	expect(t, read, msg.SERVER_ACK, msg.BUY)
	shutdownSystem(t, read, write, localhost, int32(clientPort))
}

func TestDuplicateOrders(t *testing.T) {
	serverPort := 1203
	clientPort := 1204
	setRunning(serverPort)
	read := readConn(clientPort)
	write := writeConn(serverPort)
	send(t, write, &msg.Message{msg.ORDER, msg.BUY, 1, 2, 3, 4, 5, localhost, int32(clientPort)})
	expect(t, read, msg.SERVER_ACK, msg.BUY)
	send(t, write, &msg.Message{msg.ORDER, msg.BUY, 1, 2, 3, 4, 5, localhost, int32(clientPort)})
	expect(t, read, msg.SERVER_ACK, msg.BUY)
	shutdownSystem(t, read, write, localhost, int32(clientPort))
}

func setRunning(serverPort int) {
	listener, err := NewListener(strconv.Itoa(serverPort))
	if err != nil {
		panic(err)
	}
	responder := NewResponder()
	match := matcher.NewMatcher(100)
	coordinator.Coordinate(listener, responder, match, false)
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

func send(t *testing.T, write *net.UDPConn, o *msg.Message) {
	buf := bytes.NewBuffer(make([]byte, 0))
	binary.Write(buf, binary.BigEndian, o)
	write.Write(buf.Bytes())
}

func expect(t *testing.T, read *net.UDPConn, route msg.MsgRoute, kind msg.MsgKind) {
	r, err := receive(read)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if r.Route != route {
		t.Errorf("Expecting %v route in response, found %v", route, r.Route)
	}
	if r.Kind != kind {
		t.Errorf("Expecting %v kind in response, found %v", kind, r.Kind)
	}
}

func shutdownSystem(t *testing.T, read, write *net.UDPConn, ip [4]byte, port int32) {
	o := &msg.Message{}
	o.WriteShutdown()
	o.IP = ip
	o.Port = port
	send(t, write, o)
	expect(t, read, msg.SERVER_ACK, msg.SHUTDOWN)
}

func receive(read *net.UDPConn) (*msg.Message, error) {
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

// TODO review this method to ensure it still makes sense
func validate(t *testing.T, order *msg.Message, resp *msg.Message, route msg.MsgRoute, kind msg.MsgKind) {
	if resp.Route != route {
		t.Errorf("Expecting %v route in response, found %v", route, resp.Route)
	}
	if resp.Kind != kind {
		t.Errorf("Expecting %v kind response, found %v", kind, resp.Kind)
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
