package client

import (
	"github.com/fmstephe/matching_engine/msg"
	"testing"
)

func runTrader(traderId uint32, t *testing.T) (intoSvr, outOfSvr, orders chan *msg.Message, responses chan *Response, tc traderComm) {
	intoSvr = make(chan *msg.Message)
	outOfSvr = make(chan *msg.Message)
	tdr, tc := newTrader(traderId, intoSvr, outOfSvr)
	go tdr.run()
	orders = make(chan *msg.Message)
	responses = make(chan *Response)
	con := connect{traderId: traderId, orders: orders, responses: responses}
	tc.connecter <- con
	if resp := <-responses; resp.Comment != connectedComment {
		t.Errorf("Expecting '" + connectedComment + "' found '" + resp.Comment + "'")
	}
	return intoSvr, outOfSvr, orders, responses, tc
}

func TestTraderDisconnect(t *testing.T) {
	traderId := uint32(1)
	_, _, orders, responses, _ := runTrader(traderId, t)
	close(orders)
	if resp := <-responses; resp.Comment != ordersClosedComment {
		t.Errorf("Expecting '" + ordersClosedComment + "' found '" + resp.Comment + "'")
	}
	if <-responses != nil {
		t.Errorf("Expecting nil response indicating responses had closed")
	}
}

func TestTraderNewConnection(t *testing.T) {
	traderId := uint32(1)
	_, _, _, responses, tc := runTrader(traderId, t)
	newOrders := make(chan *msg.Message)
	newResponses := make(chan *Response)
	con := connect{traderId: traderId, orders: newOrders, responses: newResponses}
	tc.connecter <- con
	// Old connection disconnects
	if resp := <-responses; resp.Comment != replacedComment {
		t.Errorf("Expecting '" + replacedComment + "' found '" + resp.Comment + "'")
	}
	if <-responses != nil {
		t.Errorf("Expecting nil response indicating responses had closed")
	}
	// New connection connected
	if <-newResponses == nil {
		t.Errorf("Expecting initial state response, recieved nil")
	}
}
