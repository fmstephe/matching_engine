package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
	"io"
	"os"
	"strconv"
	"strings"
)

type ItchReader struct {
	lineCount int
	r *bufio.Reader
}

func NewItchReader(fName string) *ItchReader {
	f, err := os.Open(fName)
	if err != nil {
		panic(err.Error())
	}
	r := bufio.NewReader(f)
	// Clear column headers
	if _, err := r.ReadString('\n'); err != nil {
		panic(err.Error())
	}
	return &ItchReader{r: r}
}

func (i *ItchReader) ReadOrder() (o *trade.Order, line string, err error) {
	i.lineCount++
	line, err = i.r.ReadString('\n')
	if err != nil {
		return
	}
	o, err = mkOrder(line)
	return
}

func (i *ItchReader) LineCount() int {
	return i.lineCount
}

func mkOrder(line string) (o *trade.Order, err error) {
	ss := strings.Split(line, " ")
	var useful []string
	for _, w := range ss {
		if w != "" && w != "\n" {
			useful = append(useful, w)
		}
	}
	cd, td := mkData(useful)
	switch useful[3] {
	case "B":
		o = trade.NewBuy(cd, td)
		return
	case "S":
		o = trade.NewSell(cd, td)
		return
	case "D":
		o = trade.NewDelete(td)
		return
	default:
		err = errors.New(fmt.Sprintf("Unsupported Line Type %s", useful[3]))
		return
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

