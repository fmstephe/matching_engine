package main

import (
	"github.com/fmstephe/matching_engine/client"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/q"
	"github.com/fmstephe/simpleid"
	"net/http"
	"os"
	"strconv"
)

var traderMaker *client.TraderMaker
var traderMap map[uint32]*client.Trader
var tradeId = uint32(0)
var idMaker = simpleid.NewIdMaker()

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
	traderMap = make(map[uint32]*client.Trader)
	traderMaker = tm
	coordinator.Coordinate(serverToClient, clientToServer, c, "Client.........", true)
	coordinator.Coordinate(clientToServer, serverToClient, m, "Matching Engine", true)
	http.HandleFunc("/buy", handleBuy)
	http.HandleFunc("/sell", handleSell)
	http.HandleFunc("/cancel", handleCancel)
	http.Handle("/", http.FileServer(http.Dir(pwd+"/html/")))
	if err := http.ListenAndServe("127.0.0.1:8081", nil); err != nil {
		println(err.Error())
	}
}

func handleBuy(w http.ResponseWriter, r *http.Request) {
	price, traderId, tradeId, amount, stockId, err := getFields(r)
	if err != nil {
		w.Write([]byte("Bad Parameters\n"))
	} else {
		trader := getTrader(traderId)
		trader.Buy(price, tradeId, amount, stockId)
		w.Write([]byte("BUY\n"))
	}
}

func handleSell(w http.ResponseWriter, r *http.Request) {
	price, traderId, tradeId, amount, stockId, err := getFields(r)
	if err != nil {
		w.Write([]byte("Bad Parameters\n"))
	} else {
		trader := getTrader(traderId)
		trader.Sell(price, tradeId, amount, stockId)
		w.Write([]byte("SELL\n"))
	}
}

func handleCancel(w http.ResponseWriter, r *http.Request) {
	price, traderId, tradeId, amount, stockId, err := getFields(r)
	if err != nil {
		w.Write([]byte("Bad Parameters\n"))
	} else {
		trader := getTrader(traderId)
		trader.Cancel(price, tradeId, amount, stockId)
		w.Write([]byte("CANCEL\n"))
	}
}

func getTrader(traderId uint32) *client.Trader {
	trader := traderMap[traderId]
	if trader == nil {
		trader = traderMaker.Make(traderId)
		traderMap[traderId] = trader
	}
	return trader
}

func getFields(r *http.Request) (price int64, traderId, tradeId, amount, stockId uint32, err error) {
	price, err = getFieldInt64(r, "price")
	if err != nil {
		return
	}
	traderId, err = getFieldUint32(r, "traderId")
	if err != nil {
		return
	}
	tradeId, err = getFieldUint32(r, "tradeId")
	if err != nil {
		return
	}
	amount, err = getFieldUint32(r, "amount")
	if err != nil {
		return
	}
	stockId, err = getFieldUint32(r, "stockId")
	return
}

func getFieldUint32(r *http.Request, name string) (uint32, error) {
	field64, err := getFieldInt64(r, name)
	// TODO do some bounds checking for max and min here
	fieldu32 := uint32(field64)
	return fieldu32, err
}

func getFieldInt64(r *http.Request, name string) (int64, error) {
	fieldStr := r.FormValue(name)
	field, err := strconv.Atoi(fieldStr)
	field64 := int64(field)
	if err != nil {
		println(err.Error(), name)
	}
	return field64, err
}
