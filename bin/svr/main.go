package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"github.com/fmstephe/matching_engine/bin/svr/client"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/msg"
	"github.com/fmstephe/matching_engine/q"
	"github.com/fmstephe/simpleid"
	"io"
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
	traderId := uint32(idMaker.Id()) // NB: A fussy man would check that the id generated fitted inside uint32
	orders, responses := traderMaker.Make(traderId)
	go reader(ws, traderId, orders)
	writer(ws, traderId, responses)
}

func reader(ws *websocket.Conn, traderId uint32, orders chan<- *msg.Message) {
	defer close(orders)
	defer ws.Close()
	for {
		var data string
		if err := websocket.Message.Receive(ws, &data); err != nil {
			logError(traderId, err)
			return
		}
		m := &msg.Message{}
		if err := json.Unmarshal([]byte(data), m); err != nil {
			logError(traderId, err)
			return
		}
		m.TraderId = traderId
		println("WebSocket......: " + m.String())
		orders <- m
	}
}

func writer(ws *websocket.Conn, traderId uint32, responses chan *client.Response) {
	defer ws.Close()
	for r := range responses {
		b, err := json.Marshal(r)
		if err != nil {
			logError(traderId, err)
			return
		}
		if _, err = ws.Write(b); err != nil {
			logError(traderId, err)
			return
		}
	}
}

func logError(traderId uint32, err error) {
	if err == io.EOF {
		println(fmt.Sprintf("Closing connection for trader %d", traderId))
	} else {
		println(fmt.Sprintf("Error for trader %d: %s", traderId, err.Error()))
	}
}
