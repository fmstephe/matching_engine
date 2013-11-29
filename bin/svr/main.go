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
	var clientSvr *client.Server
	clientSvr, traderMaker = client.NewServer()
	coordinator.InMemory(serverToClient, clientToServer, clientSvr, clientOriginId, "Client.........", true)
	coordinator.InMemory(clientToServer, serverToClient, m, serverOriginId, "Matching Engine", true)
	http.Handle("/wsconn", websocket.Handler(handleTrader))
	http.Handle("/", http.FileServer(http.Dir(pwd+"/html/")))
	if err := http.ListenAndServe("127.0.0.1:8081", nil); err != nil {
		println(err.Error())
	}
}

func handleTrader(ws *websocket.Conn) {
	traderId := uint32(idMaker.Id())
	orders, responses := traderMaker.Make(traderId)
	go reader(ws, traderId, orders)
	writer(ws, responses)
}

func reader(ws *websocket.Conn, traderId uint32, orders chan<- *msg.Message) {
	defer close(orders)
	defer ws.Close()
	for {
		var data string
		if err := websocket.Message.Receive(ws, &data); err != nil {
			println("error", err.Error())
			return
		}
		m := &msg.Message{}
		if err := json.Unmarshal([]byte(data), m); err != nil {
			println("error", err.Error())
			return
		}
		m.TraderId = traderId
		println("WebSocket......: " + m.String())
		orders <- m
	}
}

func writer(ws *websocket.Conn, responses chan []byte) {
	defer ws.Close()
	for bytes := range responses {
		if _, err := ws.Write(bytes); err != nil {
			println("Writer Error", err.Error())
			return
		}
	}
}
