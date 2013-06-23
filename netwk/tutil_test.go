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

// Because we are communicating via UDP, messages could arrive out of order, in practice they travel in-order via localhost

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
	port := strconv.Itoa(serverPort)
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	listener := NewListener(conn)
	responder := NewResponder(&udpWriter{})
	match := matcher.NewMatcher(100)
	coordinator.Coordinate(listener, responder, match, false)
	timeout := time.Duration(1) * time.Second
	clientMap := make(map[uint32]*client)
	return &netwkTester{ip: m.ip, serverPort: serverPort, maxPort: m.port - 1, freePort: minPort, clientMap: clientMap, timeout: timeout}
}

type netwkTester struct {
	ip         [4]byte
	serverPort int
	freePort   int
	maxPort    int
	clientMap  map[uint32]*client
	timeout    time.Duration
}

func (nt *netwkTester) Send(t *testing.T, m *msg.Message) {
	c := nt.getClient(m.TraderId)
	if err := c.writeMsg(m); err != nil {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("\n%s\n%s:%d", err.Error(), fname, lnum)
	}
	ref := &msg.Message{}
	ref.WriteServerAckFor(m)
	nt.simpleExpect(t, ref)
}

func (nt *netwkTester) SendNoAck(t *testing.T, m *msg.Message) {
	c := nt.getClient(m.TraderId)
	if err := c.writeMsg(m); err != nil {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("\n%s\n%s:%d", err.Error(), fname, lnum)
	}
}

func (nt *netwkTester) simpleExpect(t *testing.T, e *msg.Message) {
	c := nt.getClient(e.TraderId)
	c.writeNetwkInfo(e)
	r, err := c.receive()
	if err != nil {
		_, fname, lnum, _ := runtime.Caller(2)
		t.Errorf("Failure %s\n%s:%d", err.Error(), fname, lnum)
		return
	}
	validate(t, r, e, 3)
}

func (nt *netwkTester) Expect(t *testing.T, e *msg.Message) {
	c := nt.getClient(e.TraderId)
	nt.simpleExpect(t, e)
	ca := &msg.Message{}
	ca.WriteClientAckFor(e)
	if err := c.writeMsg(ca); err != nil {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("\n%s\n%s:%d", err.Error(), fname, lnum)
	}
}

func (nt *netwkTester) ExpectNoAck(t *testing.T, e *msg.Message) {
	nt.simpleExpect(t, e)
}

func (nt *netwkTester) ExpectEmpty(t *testing.T, traderId uint32) {
	c := nt.getClient(traderId)
	_, err := c.receive()
	if err != nil {
		expectTimeoutErr(t, err)
	}
	return
}

func (nt *netwkTester) ExpectTimeout(t *testing.T, traderId uint32) {
	c := nt.getClient(traderId)
	r, err := c.receive()
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
	for _, c := range nt.clientMap {
		m := &msg.Message{}
		m.WriteShutdown()
		c.writeMsg(m)
		return
	}
}

func (nt *netwkTester) getClient(traderId uint32) *client {
	c := nt.clientMap[traderId]
	if c == nil {
		nt.freePort++
		if nt.freePort > nt.maxPort {
			panic(fmt.Sprintf("Too many ports used. Only allowed %d", portAllocation))
		}
		read := mkReadConn(nt.freePort)
		write := mkWriteConn(nt.serverPort)
		c = &client{read: read, write: write, clientPort: nt.freePort, ip: nt.ip}
		nt.clientMap[traderId] = c
	}
	c.write.SetDeadline(time.Now().Add(nt.timeout))
	c.read.SetDeadline(time.Now().Add(nt.timeout))
	return c
}

type client struct {
	ip         [4]byte
	clientPort int
	read       *net.UDPConn
	write      *net.UDPConn
}

func (c *client) writeMsg(m *msg.Message) error {
	c.writeNetwkInfo(m)
	buf := bytes.NewBuffer(make([]byte, 0))
	binary.Write(buf, binary.BigEndian, m)
	_, err := c.write.Write(buf.Bytes())
	return err
}

func (c *client) writeNetwkInfo(m *msg.Message) {
	m.IP = c.ip
	m.Port = int32(c.clientPort)
}

func (c *client) receive() (*msg.Message, error) {
	s := make([]byte, msg.SizeofMessage)
	_, _, err := c.read.ReadFromUDP(s)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(s)
	r := &msg.Message{}
	binary.Read(buf, binary.BigEndian, r)
	return r, nil
}

func mkWriteConn(port int) *net.UDPConn {
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

func mkReadConn(port int) *net.UDPConn {
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

func validate(t *testing.T, m, e *msg.Message, stackOffset int) {
	if *m != *e {
		_, fname, lnum, _ := runtime.Caller(stackOffset)
		t.Errorf("\nExpecting: %v\nFound:     %v \n%s:%d", e, m, fname, lnum)
	}
}
