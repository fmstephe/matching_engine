package netwk

import (
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

type netwkTesterMaker struct {
	ip       [4]byte
	freePort int
}

func newMatchTesterMaker() matcher.MatchTesterMaker {
	return &netwkTesterMaker{ip: [4]byte{127, 0, 0, 1}, freePort: 1201}
}

func (m *netwkTesterMaker) Make() matcher.MatchTester {
	serverPort := m.freePort
	m.freePort++
	clientPort := m.freePort
	m.freePort++
	matcherRead := mkReadConn(serverPort)
	matcherWrite := mkWriteConn(clientPort)
	clientRead := mkReadConn(clientPort)
	clientWrite := mkWriteConn(serverPort)
	listener := NewListener(matcherRead)
	responder := NewResponder(matcherWrite)
	match := matcher.NewMatcher(100)
	coordinator.Coordinate(listener, responder, match, false)
	timeout := time.Duration(1) * time.Second
	return &netwkTester{ip: m.ip, write: clientWrite, read: clientRead, timeout: timeout}
}

type netwkTester struct {
	ip      [4]byte
	write   *net.UDPConn
	read    *net.UDPConn
	timeout time.Duration
}

func (nt *netwkTester) Send(t *testing.T, m *msg.Message) {
	if err := nt.writeMsg(m); err != nil {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("\n%s\n%s:%d", err.Error(), fname, lnum)
	}
	ref := &msg.Message{}
	ref.WriteServerAckFor(m)
	nt.simpleExpect(t, ref)
}

func (nt *netwkTester) SendNoAck(t *testing.T, m *msg.Message) {
	if err := nt.writeMsg(m); err != nil {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("\n%s\n%s:%d", err.Error(), fname, lnum)
	}
}

func (nt *netwkTester) simpleExpect(t *testing.T, e *msg.Message) {
	r, err := nt.receive()
	if err != nil {
		_, fname, lnum, _ := runtime.Caller(2)
		t.Errorf("Failure %s\n%s:%d", err.Error(), fname, lnum)
		return
	}
	validate(t, r, e, 3)
}

func (nt *netwkTester) Expect(t *testing.T, e *msg.Message) {
	nt.simpleExpect(t, e)
	ca := &msg.Message{}
	ca.WriteClientAckFor(e)
	if err := nt.writeMsg(ca); err != nil {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("\n%s\n%s:%d", err.Error(), fname, lnum)
	}
}

func (nt *netwkTester) ExpectOneOf_NoAck(t *testing.T, es ...*msg.Message) {
	r, err := nt.receive()
	if err != nil {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("Failure %s\n%s:%d", err.Error(), fname, lnum)
		return
	}
	for _, e := range es {
		if *e == *r {
			return
		}
	}
	t.Errorf("Expecting one of %v, received %v instead", es, r)
}

func (nt *netwkTester) ExpectEmpty(t *testing.T, traderId uint32) {
	_, err := nt.receive()
	if err != nil {
		expectTimeoutErr(t, err)
	}
	return
}

func (nt *netwkTester) ExpectTimeout(t *testing.T, traderId uint32) {
	r, err := nt.receive()
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
	nt.writeMsg(m)
	return
}

func (nt *netwkTester) writeMsg(m *msg.Message) error {
	b := make([]byte, msg.SizeofMessage)
	m.WriteTo(b)
	_, err := nt.write.Write(b)
	return err
}

func (nt *netwkTester) receive() (*msg.Message, error) {
	nt.read.SetDeadline(time.Now().Add(nt.timeout))
	b := make([]byte, msg.SizeofMessage)
	_, _, err := nt.read.ReadFromUDP(b)
	if err != nil {
		return nil, err
	}
	r := &msg.Message{}
	r.WriteFrom(b)
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
