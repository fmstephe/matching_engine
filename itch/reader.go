package itch

import (
	"bufio"
	"github.com/fmstephe/matching_engine/msg"
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

func (i *ItchReader) ReadMessage() (o *msg.Message, line string, err error) {
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
	o, err = mkMessage(line)
	if o != nil && o.Kind == msg.BUY && o.Price > i.maxBuy {
		i.maxBuy = o.Price
	}
	if o != nil && o.Kind == msg.SELL && o.Price < i.minSell {
		i.minSell = o.Price
	}
	return
}

func (i *ItchReader) ReadAll() (orders []*msg.Message, err error) {
	orders = make([]*msg.Message, 0)
	var o *msg.Message
	for err == nil {
		o, _, err = i.ReadMessage()
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

func mkMessage(line string) (o *msg.Message, err error) {
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
		o.WriteBuy(cd, td, msg.NetData{})
		return
	case "S":
		o.WriteSell(cd, td, msg.NetData{})
		return
	case "D":
		o.WriteCancel(td, msg.NetData{})
		return
	default:
		return
	}
	panic("Unreachable")
}

func mkData(useful []string) (cd msg.CostData, td msg.TradeData, err error) {
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
	cd = msg.CostData{Price: int64(price), Amount: uint32(amount)}
	td = msg.TradeData{TraderId: uint32(traderId), TradeId: uint32(tradeId), StockId: uint32(1)}
	return
}
