package itchutil

import (
	"bufio"
	"fmt"
	"os"
	"io"
	"github.com/fmstephe/matching_engine/trade"
	"strings"
	"strconv"
)

func PrintLineCount(fName string) {
	f, _ := os.Open(fName)
	r := bufio.NewReader(f)
	i := 0
	for {
		if _, err := r.ReadString('\n'); err != nil {
			if err == io.EOF {
				break
			}
			panic(err.Error())
		}
		i++
	}
	println(i)
}

func ReadOrders(fName string) []*trade.Order{
	f, _ := os.Open(fName)
	r := bufio.NewReader(f)
	orders := make([]*trade.Order, 10)
	i := 0
	// Read column headers
	if _, err := r.ReadString('\n'); err != nil {
		panic(err.Error())
	}
	for {
		var line string
		var err error
		if line, err = r.ReadString('\n'); err != nil {
			panic(err.Error())
		}
		orders = append(orders, mkOrder(line))
		i++
		if i > 1000 {
			break
		}
	}
	return orders
}

func mkOrder(line string) *trade.Order {
	ss := strings.Split(line, " ")
	var useful []string
	for _, w := range ss {
		if w != "" && w != "\n" {
			useful = append(useful, w)
		}
	}
	cd, td := mkData(useful)
	switch useful[3] {
		case "B" : return trade.NewBuy(cd, td)
		case "S" : return trade.NewSell(cd, td)
		case "D" : return trade.NewDelete(td)
		default : panic(fmt.Sprintf("Unrecognised Trade Type %s", useful[3]))
	}
	panic("Unreachable")
}

func mkData(useful []string) (cd trade.CostData, td trade.TradeData) {
	//      print("ID: ", useful[2], " Type: ", useful[3], " Price: ",  useful[4], " Amount: ", useful[5])
	//      println()
	var price, amount, traderId, tradeId, stockId int
	var err error
	if price, err = strconv.Atoi(useful[4]); err != nil {
		panic(err.Error())
	}
	if amount, err = strconv.Atoi(useful[5]); err != nil {
		panic(err.Error())
	}
	if traderId, err = strconv.Atoi(useful[2]); err != nil {
		panic(err.Error())
	}
	if tradeId, err = strconv.Atoi(useful[2]); err != nil {
		panic(err.Error())
	}
	stockId = 1
	cd = trade.CostData{Price: int32(price), Amount: uint32(amount)}
	td = trade.TradeData{TraderId: uint32(traderId), TradeId: uint32(tradeId), StockId: uint32(stockId)}
	return
}

