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

const serverPort = "1200"
const clientPort = "1201"

type mockMatcher struct {
	submit chan interface{}
	orders chan *trade.OrderData
}

func (m *mockMatcher) SetSubmit(submit chan interface{}) {
	m.submit = submit
}

func (m *mockMatcher) SetOrders(orders chan *trade.OrderData) {
	m.orders = orders
}

func (m *mockMatcher) Run() {
	for {
		od := <-m.orders
		r := &trade.Response{}
		r.Price = od.Price
		r.Amount = od.Amount
		r.TraderId = trade.GetTraderId(od.Guid)
		r.TradeId = trade.GetTradeId(od.Guid)
		r.IP = od.IP
		r.Port = od.Port
		r.CounterParty = trade.GetTraderId(od.Guid)
		m.submit <- r
	}
}

func TestOrdersAndResponse(t *testing.T) {
	setRunning()
	read := readConn(clientPort)
	write := writeConn(serverPort)
	confirmOrder(t, read, write, 1, 2, 3, 4, 5)
	confirmOrder(t, read, write, 6, 7, 8, 9, 10)
	confirmOrder(t, read, write, 11, 12, 13, 14, 15)
}

func setRunning() {
	listener, err := NewListener(serverPort)
	if err != nil {
		panic(err)
	}
	responder := NewResponder()
	matcher := &mockMatcher{}
	coordinator.Coordinate(listener, responder, matcher, false)
}

func writeConn(port string) *net.UDPConn {
	addr, err := net.ResolveUDPAddr("udp", ":"+serverPort)
	if err != nil {
		panic(err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	return conn
}

func readConn(port string) *net.UDPConn {
	addr, err := net.ResolveUDPAddr("upd", ":"+clientPort)
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	return conn
}

func confirmOrder(t *testing.T, read, write *net.UDPConn, price int64, amount uint32, traderId, tradeId uint32, stockId uint32) {
	od := &trade.OrderData{}
	od.WriteBuy(trade.CostData{Price: price, Amount: amount}, trade.TradeData{TraderId: traderId, TradeId: tradeId, StockId: stockId})
	err := sendOrderData(t, write, od)
	if err != nil {
		t.Error(err.Error())
		return
	}
	r, err := receiveResponse(t, read, od)
	if err != nil {
		t.Error(err.Error())
		return
	}
	validate(t, od, r)
}

func sendOrderData(t *testing.T, write *net.UDPConn, od *trade.OrderData) error {
	od.IP = [4]byte{127, 0, 0, 1}
	port, err := strconv.Atoi(clientPort)
	if err != nil {
		return err
	}
	od.Port = int32(port)
	buf := bytes.NewBuffer(make([]byte, 0))
	binary.Write(buf, binary.BigEndian, od)
	write.Write(buf.Bytes())
	return nil
}

func receiveResponse(t *testing.T, read *net.UDPConn, od *trade.OrderData) (*trade.Response, error) {
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

func validate(t *testing.T, od *trade.OrderData, r *trade.Response) {
	if od.Price != r.Price {
		t.Errorf("Price mismatch, expecting %d, found %d", od.Price, r.Price)
	}
	if od.Amount != r.Amount {
		t.Errorf("Amount mismatch, expecting %d, found %d", od.Amount, r.Amount)
	}
	if trade.GetTraderId(od.Guid) != r.TraderId {
		t.Errorf("TraderId mismatch, expecting %d, found %d", trade.GetTraderId(od.Guid), r.TraderId)
	}
	if trade.GetTradeId(od.Guid) != r.TradeId {
		t.Errorf("TradeId mismatch, expecting %d, found %d", trade.GetTradeId(od.Guid), r.Price)
	}
	if trade.GetTraderId(od.Guid) != r.CounterParty {
		t.Errorf("Counterparty mismatch, expecting %d, found %d", trade.GetTraderId(od.Guid), r.CounterParty)
	}
	if od.IP != r.IP {
		t.Errorf("IP mismatch, expecting %d, found %d", od.IP, r.IP)
	}
	if od.Port != r.Port {
		t.Errorf("Port mismatch, expecting %d, found %d", od.Port, r.Port)
	}
	if trade.GetTraderId(od.Guid) != r.CounterParty {
		t.Errorf("Counterparty mismatch, expecting %d, found %d", trade.GetTraderId(od.Guid), r.CounterParty)
	}
}
