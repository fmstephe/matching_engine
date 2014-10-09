package matcher

import (
	"unsafe"
	"github.com/fmstephe/flib/queues/spscq"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/msg"
	"net"
	"runtime"
	"strconv"
	"testing"
)

// Because we are communicating via UDP, messages could arrive out of order, in practice they travel in-order via localhost

const (
	matcherOrigin = iota
	clientOrigin  = iota
)

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
	coordinator.InMemory(mkReadConn(serverPort), mkWriteConn(clientPort), m, matcherOrigin, "Matching Engine", false)
	// Build client
	receivedMsgs, _:= spscq.NewPointerQ(1024, 1000 * 1000)
	toSendMsgs, _ := spscq.NewPointerQ(1024, 1000 * 1000)
	c := newClient(receivedMsgs, toSendMsgs)
	coordinator.InMemory(mkReadConn(clientPort), mkWriteConn(serverPort), c, clientOrigin, "Test Client    ", false)
	return &netwkTester{receivedMsgs: receivedMsgs, toSendMsgs: toSendMsgs}
}

type netwkTester struct {
	receivedMsgs *spscq.PointerQ
	toSendMsgs   *spscq.PointerQ
}

func (nt *netwkTester) Send(t *testing.T, m *msg.Message) {
	nt.toSendMsgs.WriteSingleBlocking(unsafe.Pointer(m))
}

func (nt *netwkTester) Expect(t *testing.T, e *msg.Message) {
	r := (*msg.Message)(nt.receivedMsgs.ReadSingleBlocking())
	validate(t, r, e, 2)
}

func (nt *netwkTester) ExpectOneOf(t *testing.T, es ...*msg.Message) {
	r := (*msg.Message)(nt.receivedMsgs.ReadSingleBlocking())
	for _, e := range es {
		if *e == *r {
			return
		}
	}
	t.Errorf("Expecting one of %v, received %v instead", es, r)
}

func (nt *netwkTester) Cleanup(t *testing.T) {
	m := &msg.Message{}
	m.Kind = msg.SHUTDOWN
	nt.toSendMsgs.WriteSingleBlocking(unsafe.Pointer(m))
}

type client struct {
	coordinator.AppMsgHelper
	receivedMsgs *spscq.PointerQ
	toSendMsgs   *spscq.PointerQ
}

func newClient(receivedMsgs, toSendMsgs *spscq.PointerQ) *client {
	return &client{receivedMsgs: receivedMsgs, toSendMsgs: toSendMsgs}
}

func (c *client) Run() {
	for {
		m := (*msg.Message)(c.In.ReadSingle())
		if m != nil {
			c.receivedMsgs.WriteSingleBlocking(unsafe.Pointer(m))
			if m.Kind == msg.SHUTDOWN {
				c.Out.WriteSingleBlocking(unsafe.Pointer(m))
				return
			}
		}
		m = (*msg.Message)(c.toSendMsgs.ReadSingle())
		if m != nil {
			c.Out.WriteSingleBlocking(unsafe.Pointer(m))
		}
	}
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
