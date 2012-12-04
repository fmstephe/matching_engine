package itch

import (
	"bufio"
	"github.com/fmstephe/matching_engine/trade"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

type ItchReader struct {
	lineCount uint
	maxBuy    int64
	minSell   int64
	r         *bufio.Reader
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
	return &ItchReader{lineCount: 1, minSell: math.MaxInt32, r: r}
}

func (i *ItchReader) ReadOrder() (o *trade.Order, line string, err error) {
	i.lineCount++
	for {
		line, err = i.r.ReadString('\n')
		if err != nil {
			return
		}
		if line != "" {
			break
		}
	}
	o, err = mkOrder(line)
	if o != nil && o.Kind == trade.BUY && o.Price() > i.maxBuy {
		i.maxBuy = o.Price()
	}
	if o != nil && o.Kind == trade.SELL && o.Price() < i.minSell {
		i.minSell = o.Price()
	}
	return
}

func (i *ItchReader) ReadAll() (orders []*trade.Order, err error) {
	orders = make([]*trade.Order, 0)
	var o *trade.Order
	for err == nil {
		o, _, err = i.ReadOrder()
		if o != nil {
			orders = append(orders, o)
		}
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func (i *ItchReader) LineCount() uint {
	return i.lineCount
}

func (i *ItchReader) MaxBuy() int64 {
	return i.maxBuy
}

func (i *ItchReader) MinSell() int64 {
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
	cd, td, err := mkData(useful)
	if err != nil {
		return
	}
	switch useful[3] {
	case "B":
		o = trade.NewBuy(cd, td)
		return
	case "S":
		o = trade.NewSell(cd, td)
		return
	case "D":
		o = trade.NewCancel(td)
		return
	default:
		return
	}
	panic("Unreachable")
}

func mkData(useful []string) (cd trade.CostData, td trade.TradeData, err error) {
	//      print("ID: ", useful[2], " Type: ", useful[3], " Price: ",  useful[4], " Amount: ", useful[5])
	//      println()
	var price, amount, traderId, tradeId int
	amount, err = strconv.Atoi(useful[4])
	price, err = strconv.Atoi(useful[5])
	traderId, err = strconv.Atoi(useful[2])
	tradeId, err = strconv.Atoi(useful[2])
	if err != nil {
		return
	}
	cd = trade.CostData{Price: int64(price), Amount: uint32(amount)}
	td = trade.TradeData{TraderId: uint32(traderId), TradeId: uint32(tradeId), StockId: uint32(1)}
	return
}
