package coordinator

import (
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/msg"
	"net"
	"runtime"
	"strconv"
	"testing"
)

// Because we are communicating via UDP, messages could arrive out of order, in practice they travel in-order via localhost

var testMkrUtil = newMatchTesterMaker()

type netwkTesterMaker struct {
	ip       [4]byte
	freePort int
}

func newMatchTesterMaker() matcher.MatchTesterMaker {
	return &netwkTesterMaker{ip: [4]byte{127, 0, 0, 1}, freePort: 1201}
}

func (tm *netwkTesterMaker) Make() matcher.MatchTester {
	serverPort := tm.freePort
	tm.freePort++
	clientPort := tm.freePort
	tm.freePort++
	// Build matcher
	m := matcher.NewMatcher(100)
	Coordinate(mkReadConn(serverPort), mkWriteConn(clientPort), m, "Matching Engine", false)
	// Build client
	receivedMsgs := make(chan *msg.Message, 1000)
	toSendMsgs := make(chan *msg.Message, 1000)
	c := newClient(receivedMsgs, toSendMsgs)
	Coordinate(mkReadConn(clientPort), mkWriteConn(serverPort), c, "Test Client    ", false)
	return &netwkTester{receivedMsgs: receivedMsgs, toSendMsgs: toSendMsgs}
}

type netwkTester struct {
	receivedMsgs chan *msg.Message
	toSendMsgs   chan *msg.Message
}

func (nt *netwkTester) Send(t *testing.T, m *msg.Message) {
	nt.toSendMsgs <- m
}

func (nt *netwkTester) Expect(t *testing.T, e *msg.Message) {
	r := <-nt.receivedMsgs
	validate(t, r, e, 2)
}

func (nt *netwkTester) ExpectOneOf(t *testing.T, es ...*msg.Message) {
	r := <-nt.receivedMsgs
	for _, e := range es {
		if *e == *r {
			return
		}
	}
	t.Errorf("Expecting one of %v, received %v instead", es, r)
}

func (nt *netwkTester) Cleanup(t *testing.T) {
	m := &msg.Message{}
	m.WriteShutdown()
	nt.toSendMsgs <- m
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
