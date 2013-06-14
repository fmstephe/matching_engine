package netwk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/msg"
	"net"
	"runtime"
	"strconv"
	"testing"
	"time"
)

// NB: There are a number of problems with these tests which are currently being ignored
// 1: Because we are communicating via UDP messages could arrive out of order, in practice they travel in-order via localhost
// 2: In the main test suite messages are currently not being acked, which means that responses may be re-sent - which would
//    confuse response checking. This doesn't impact the tests right now because the resend rate is so slow that the test is
//    complete and the system shut down before unacked messages are resent, this is pretty delicate.

const (
	portAllocation = 100
)

type netwkTesterMaker struct {
	ip   [4]byte
	port int
}

func newMatchTesterMaker() matcher.MatchTesterMaker {
	return &netwkTesterMaker{ip: [4]byte{127, 0, 0, 1}, port: 1201}
}

func (m *netwkTesterMaker) Make() matcher.MatchTester {
	m.port++
	serverPort := m.port
	m.port++
	minPort := m.port
	m.port = m.port + portAllocation
	listener, err := NewListener(strconv.Itoa(serverPort))
	if err != nil {
		panic(err)
	}
	responder := NewResponder()
	match := matcher.NewMatcher(100)
	coordinator.Coordinate(listener, responder, match, false)
	timeout := time.Duration(1) * time.Second
	return &netwkTester{ip: m.ip, serverPort: serverPort, maxPort: m.port - 1, freePort: minPort, connsMap: make(map[uint32]*conns), timeout: timeout}
}

type netwkTester struct {
	ip         [4]byte
	serverPort int
	freePort   int
	maxPort    int
	connsMap   map[uint32]*conns
	timeout    time.Duration
}

type conns struct {
	clientPort int
	read       *net.UDPConn
	write      *net.UDPConn
}

func (nt *netwkTester) Send(t *testing.T, m *msg.Message) {
	nt.writeNetwk(m)
	buf := bytes.NewBuffer(make([]byte, 0))
	binary.Write(buf, binary.BigEndian, m)
	nt.getConns(m.TraderId).write.Write(buf.Bytes())
	// We always expect a server ack when sending a message
	ref := &msg.Message{}
	ref.WriteServerAckFor(m)
	nt.Expect(t, ref)
}

func (nt *netwkTester) SendNoAck(t *testing.T, m *msg.Message) {
	nt.writeNetwk(m)
	buf := bytes.NewBuffer(make([]byte, 0))
	binary.Write(buf, binary.BigEndian, m)
	nt.getConns(m.TraderId).write.Write(buf.Bytes())
}

func (nt *netwkTester) Expect(t *testing.T, e *msg.Message) {
	nt.writeNetwk(e)
	r, err := receive(nt.getConns(e.TraderId).read)
	if err != nil {
		t.Error(err.Error())
		return
	}
	validate(t, r, e)
}

func (nt *netwkTester) ExpectFiniteResends(t *testing.T, e *msg.Message) {
	nt.writeNetwk(e)
	limit := 20
	for i := 0; i < limit; i++ {
		r, err := receive(nt.getConns(e.TraderId).read)
		if err == nil && !validate(t, r, e) {
			return
		}
		if err != nil {
			expectTimeoutErr(t, err)
			return
		}
	}
	_, fname, lnum, _ := runtime.Caller(1)
	t.Errorf("\nExpecting timeout\nReceived %d responses\n%s:%d", limit, fname, lnum)
}

func (nt *netwkTester) ExpectTimeout(t *testing.T, traderId uint32) {
	r, err := receive(nt.getConns(traderId).read)
	if err == nil {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Error("\nExpecting timeout\n Recieved %v\n%s%d", r, fname, lnum)
	}
	expectTimeoutErr(t, err)
}

func expectTimeoutErr(t *testing.T, err error) {
	e, ok := err.(net.Error)
	if !ok || !e.Timeout() {
		_, fname, lnum, _ := runtime.Caller(2)
		t.Errorf("\nUnexpected error %v\n%s:%d", err, fname, lnum)
	}
}

func (nt *netwkTester) Cleanup(t *testing.T) {
	m := &msg.Message{}
	m.WriteShutdown()
	nt.Send(t, m)
}

func (nt *netwkTester) writeNetwk(m *msg.Message) {
	m.IP = nt.ip
	m.Port = int32(nt.getConns(m.TraderId).clientPort)
}

func (nt *netwkTester) getConns(traderId uint32) *conns {
	c := nt.connsMap[traderId]
	if c == nil {
		nt.freePort++
		if nt.freePort > nt.maxPort {
			panic(fmt.Sprintf("Too many ports used. Only allowed %d", portAllocation))
		}
		read := readConn(nt.freePort)
		write := writeConn(nt.serverPort)
		c = &conns{read: read, write: write, clientPort: nt.freePort}
		nt.connsMap[traderId] = c
	}
	c.write.SetDeadline(time.Now().Add(nt.timeout))
	c.read.SetDeadline(time.Now().Add(nt.timeout))
	return c
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

func validate(t *testing.T, m, e *msg.Message) bool {
	if *m != *e {
		_, fname, lnum, _ := runtime.Caller(2)
		t.Errorf("\nExpecting: %v\nFound:     %v \n%s:%d", e, m, fname, lnum)
		return false
	}
	return true
}
