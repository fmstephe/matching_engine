package listen

import (
	"bytes"
	"encoding/binary"
	"github.com/fmstephe/matching_engine/trade"
	"net"
	"testing"
)

const port = "1200"

func TestSingleOrder(t *testing.T) {
	orderChan := make(chan *trade.OrderData, 10)
	go runListener(orderChan)
	udpAddr, _ := net.ResolveUDPAddr("udp", ":"+port)
	conn, _ := net.DialUDP("udp", nil, udpAddr)
	od := &trade.OrderData{}
	od.WriteBuy(trade.CostData{Price: 1, Amount: 2}, trade.TradeData{TraderId: 3, TradeId: 4, StockId: 5})
	buf := bytes.NewBuffer(make([]byte, 0, orderSize))
	binary.Write(buf, binary.LittleEndian, od)
	conn.Write(buf.Bytes())
	newOd := <-orderChan
	if od == nil {
		t.Error("Test failed. Most likely unable to construct listener")
	} else {
		if *newOd != *od {
			t.Errorf("Expecting %s, found %s", trade.NewOrderFromData(od).String(), trade.NewOrderFromData(newOd).String())
		}
	}
}

func runListener(orderChan chan *trade.OrderData) {
	l, err := NewOrderListener(port, orderChan)
	if err != nil {
		orderChan <- nil
	} else {
		l.Listen()
	}
}
