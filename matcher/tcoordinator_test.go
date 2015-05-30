package matcher

import (
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/msg"
	"net"
	"runtime"
	"strconv"
	"testing"
)

// Because we are communicating via UDP, messages could arrive out of order, in practice they travel in-order via localhost

type netwkTesterMaker struct {
	ip       [4]byte
	freePort int
}

func newMatchTesterMaker() MatchTesterMaker {
	return &netwkTesterMaker{ip: [4]byte{127, 0, 0, 1}, freePort: 1201}
}

func (tm *netwkTesterMaker) Make() MatchTester {
	serverPort := tm.freePort
	tm.freePort++
	clientPort := tm.freePort
	tm.freePort++
	// Build matcher
	m := NewMatcher(100)
	coordinator.InMemory(mkReadConn(serverPort), mkWriteConn(clientPort), m, 0, "Matching Engine", false)
	// Build client
	fromListener, toResponder := coordinator.InMemoryListenerResponder(mkReadConn(clientPort), mkWriteConn(serverPort), "Test Client    ", false)
	return &netwkTester{receivedMsgs: fromListener, toSendMsgs: toResponder}
}

type netwkTester struct {
	receivedMsgs coordinator.MsgReader
	toSendMsgs   coordinator.MsgWriter
}

func (nt *netwkTester) Send(t *testing.T, m *msg.Message) {
	nt.toSendMsgs.Write(*m)
}

func (nt *netwkTester) Expect(t *testing.T, e *msg.Message) {
	r := &msg.Message{}
	*r = nt.receivedMsgs.Read()
	validate(t, r, e, 2)
}

func (nt *netwkTester) ExpectOneOf(t *testing.T, es ...*msg.Message) {
	r := &msg.Message{}
	*r = nt.receivedMsgs.Read()
	for _, e := range es {
		if *e == *r {
			return
		}
	}
	t.Errorf("Expecting one of %v, received %v instead", es, r)
}

func (nt *netwkTester) Cleanup(t *testing.T) {
	nt.toSendMsgs.Write(msg.Message{Kind: msg.SHUTDOWN})
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
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
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

func TestRunCoordinatedTestSuite(t *testing.T) {
	RunTestSuite(t, newMatchTesterMaker())
}
