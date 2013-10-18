package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"github.com/fmstephe/matching_engine/client"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/msg"
	"github.com/fmstephe/matching_engine/q"
	"github.com/fmstephe/simpleid"
	"net/http"
	"os"
)

var traderMaker *client.TraderMaker
var idMaker = simpleid.NewIdMaker()

const (
	clientOriginId = iota
	serverOriginId = iota
)

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		println(err.Error())
		return
	}
	// Create matching engine + client
	clientToServer := q.NewSimpleQ("Client To Server")
	serverToClient := q.NewSimpleQ("Server To Client")
	// Matching Engine
	m := matcher.NewMatcher(100)
	c, tm := client.NewClient()
	traderMaker = tm
	coordinator.InMemory(serverToClient, clientToServer, c, clientOriginId, "Client.........", true)
	coordinator.InMemory(clientToServer, serverToClient, m, serverOriginId, "Matching Engine", true)
	http.Handle("/wsconn", websocket.Handler(handleTrader))
	http.Handle("/", http.FileServer(http.Dir(pwd+"/html/")))
	if err := http.ListenAndServe("127.0.0.1:8081", nil); err != nil {
		println(err.Error())
	}
}

type tradeInfo struct {
	Kind    string `json:"kind"`
	Price   int64  `json:"price"`
	Amount  uint32 `json:"amount"`
	StockId uint32 `json:"stockId"`
}

type order struct {
	tradeInfo
}

type response struct {
	tradeInfo
	TradeId  uint32 `json:"tradeId"`
	AvailBal int64  `json:"availBal"`
	CurBal   int64  `json:"curBal"`
}

type traderState struct {
	traderId uint32
	tradeId  uint32
	availBal int64
	curBal   int64
	trader   *client.Trader
}

func newTraderState() *traderState {
	traderId := uint32(idMaker.Id())
	trader := traderMaker.Make(traderId)
	bal := int64(100)
	return &traderState{traderId: traderId, tradeId: uint32(1), availBal: bal, curBal: bal, trader: trader}
}

func (ts *traderState) processOrder(o *order) *response {
	r := &response{tradeInfo: tradeInfo{Kind: o.Kind, Price: o.Price, Amount: o.Amount, StockId: o.StockId}, TradeId: ts.tradeId}
	if o.Kind == "BUY" {
		total := o.Price * int64(o.Amount)
		if total > ts.availBal {
			r.Kind = msg.REJECTED.String()
		} else {
			ts.trader.Buy(o.Price, ts.tradeId, o.Amount, o.StockId)
			ts.availBal -= total
		}
	}
	if o.Kind == "SELL" {
		ts.trader.Sell(o.Price, ts.tradeId, o.Amount, o.StockId)
	}
	ts.tradeId++
	r.AvailBal = ts.availBal
	r.CurBal = ts.curBal
	return r
}

func (ts *traderState) processMsg(m *msg.Message) *response {
	total := m.Price * int64(m.Amount)
	ts.curBal += total
	// If sell (>0) then available balance is updated
	// If buy (<0) then available balance has already been updated
	if total > 0 {
		ts.availBal += total
	}
	r := &response{tradeInfo: tradeInfo{Kind: m.Kind.String(), Price: m.Price, Amount: m.Amount, StockId: m.StockId}, TradeId: m.TradeId}
	r.AvailBal = ts.availBal
	r.CurBal = ts.curBal
	return r
}

func handleTrader(ws *websocket.Conn) {
	ts := newTraderState()
	orders := make(chan *order)
	responses := make(chan *response)
	defer close(responses)
	go reader(ws, orders)
	go writer(ws, responses)
	for {
		select {
		case o := <-orders:
			r := ts.processOrder(o)
			responses <- r
		case m := <-ts.trader.OutOfClient:
			r := ts.processMsg(m)
			responses <- r
		}
	}
}

func reader(ws *websocket.Conn, orders chan<- *order) {
	defer close(orders)
	defer ws.Close()
	for {
		var data string
		if err := websocket.Message.Receive(ws, &data); err != nil {
			println("error", err.Error())
			return
		}
		println(data)
		o := &order{}
		if err := json.Unmarshal([]byte(data), o); err != nil {
			println("error", err.Error())
			return
		}
		orders <- o
	}
}

func writer(ws *websocket.Conn, responses chan *response) {
	defer ws.Close()
	for r := range responses {
		bytes, err := json.Marshal(r)
		if err != nil {
			println("Writer Error", err.Error())
			return
		}
		if _, err := ws.Write(bytes); err != nil {
			println("Writer Error", err.Error())
			return
		}
	}
}
