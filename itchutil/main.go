package main

import (
	"bufio"
	"flag"
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/prioq/limitheap"
	"github.com/fmstephe/matching_engine/trade"
	"os"
)

var (
	filePath = flag.String("f", "", "Relative path to an ITCH file to read")
)

func main() {
	flag.Parse()
	in := bufio.NewReader(os.Stdin)
	ir := NewItchReader(*filePath)
	buysQ := limitheap.New(trade.BUY, 2000, 10000)
	sellsQ := limitheap.New(trade.SELL, 2000, 10000)
	buffer := matcher.NewResponseBuffer(2)
	m := matcher.NewMatcher(buysQ, sellsQ, buffer)
	loop(in, ir, m, buffer)
}

func loop(in *bufio.Reader, ir *ItchReader, m *matcher.M, b *matcher.ResponseBuffer) {
	var o *trade.Order
	var err error
	for {
		o, _, err = ir.ReadOrder()
		if err != nil {
			println(err.Error())
			return
		}
		if o.Kind == trade.BUY || o.Kind == trade.SELL || o.Kind == trade.DELETE {
			m.Submit(o)
		}
		println("Line: ", ir.lineCount)
		println(b.Reads(), b.Writes())
		println(o.String())
		buys, sells, orders := m.Survey()
		print("Buys: [", len(buys), ",", cap(buys), "]\n")
		print("Sells: [", len(sells), ",", cap(sells), "]\n")
		println("Total: ", orders.Size())
		processInput(in, m)
	}
}

func processInput(in *bufio.Reader, m *matcher.M) {
	for {
		println("...")
		_, err := in.ReadString('\n')
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		break
	}
}
