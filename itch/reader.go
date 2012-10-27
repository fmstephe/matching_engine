package main

import (
	"bufio"
	"github.com/fmstephe/matching_engine/trade"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

type Reader struct {
	lineCount uint
	maxBuy    int32
	minSell   int32
	r         *bufio.Reader
}

func NewItchReader(fName string) *Reader {
	println(fName)
	f, err := os.Open(fName)
	if err != nil {
		panic(err.Error())
	}
	r := bufio.NewReader(f)
	// Clear column headers
	if _, err := r.ReadString('\n'); err != nil {
		panic(err.Error())
	}
	return &Reader{lineCount: 1, minSell: math.MaxInt32, r: r}
}

func (i *Reader) ReadOrder() (o *trade.Order, line string, err error) {
	i.lineCount++
	line, err = i.r.ReadString('\n')
	if err != nil {
		return
	}
	o, err = mkOrder(line)
	if o != nil && o.Kind == trade.BUY && o.Price > i.maxBuy {
		i.maxBuy = o.Price
	}
	if o != nil && o.Kind == trade.SELL && o.Price < i.minSell {
		i.minSell = o.Price
	}
	return
}

func (i *Reader) LineCount() uint {
	return i.lineCount
}

func (i *Reader) MaxBuy() int32 {
	return i.maxBuy
}

func (i *Reader) MinSell() int32 {
	return i.minSell
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
		return
	}
	panic("Unreachable")
}

func mkData(useful []string) (cd trade.CostData, td trade.TradeData) {
	//      print("ID: ", useful[2], " Type: ", useful[3], " Price: ",  useful[4], " Amount: ", useful[5])
	//      println()
	var price, amount, traderId, tradeId, stockId int
	var err error
	if amount, err = strconv.Atoi(useful[4]); err != nil {
		panic(err.Error())
	}
	if price, err = strconv.Atoi(useful[5]); err != nil {
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
