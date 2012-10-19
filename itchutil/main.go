package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/prioq/limitheap"
	"github.com/fmstephe/matching_engine/trade"
	"os"
)

var (
	filePath = flag.String("f", "", "Relative path to an ITCH file to read")
	mode     = flag.String("m", "step", "Running mode. Currently supporting 'step', 'exec' and 'list'")
	line     = flag.Uint("l", 0, "First line to break on. Mode is ignored until line l is reached, then normal excution continues")
)

func main() {
	flag.Parse()
	loop()
}

func loop() {
	l := *line
	for {
		in := bufio.NewReader(os.Stdin)
		ir := NewItchReader(*filePath)
		buysQ := limitheap.New(trade.BUY, 2000, 10000)
		sellsQ := limitheap.New(trade.SELL, 2000, 10000)
		buffer := matcher.NewResponseBuffer(2)
		m := matcher.NewMatcher(buysQ, sellsQ, buffer)
		//
		var o *trade.Order
		var err error
		for {
			o, _, err = ir.ReadOrder()
			if err != nil {
				println(err.Error())
				return
			}
			if o != nil && (o.Kind == trade.BUY || o.Kind == trade.SELL || o.Kind == trade.DELETE) {
				m.Submit(o)
			}
			checkPrint(ir, o, m, l)
			c := checkPause(in, ir, o, l)
			if c == 'k' {
				l = ir.LineCount() - 1
				break
			}
		}
	}
}

func checkPause(in *bufio.Reader, ir *ItchReader, o *trade.Order, bLine uint) byte {
	if bLine > ir.LineCount() {
		return 'z'
	}
	if *mode == "step" {
		return pause(in)
	}
	if *mode == "exec" && o == nil {
		return pause(in)
	}
	return 'z'
}

func pause(in *bufio.Reader) byte {
	c, err := in.ReadByte()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	return c
}

func checkPrint(ir *ItchReader, o *trade.Order, m *matcher.M, bLine uint) {
	if bLine > ir.LineCount() {
		return
	}
	if *mode == "step" {
		printInfo(ir, o, m)
	}
	if *mode == "exec" && o == nil {
		printInfo(ir, o, m)
	}
}

func printInfo(ir *ItchReader, o *trade.Order, m *matcher.M) {
	buys, sells, orders, executions := m.Survey()
	println("Order       ", o.String())
	println("Line        ", ir.LineCount())
	println("Max Buy     ", ir.MaxBuy())
	println("Min Sell    ", ir.MinSell())
	println("Executions  ", executions)
	println("...")
	println("Total       ", orders.Size())
	println("Buy Limits  ", formatLimits(buys))
	println("Sell Limits ", formatLimits(sells))
	println()
}

func formatLimits(limits []*trade.SurveyLimit) string {
	var b bytes.Buffer
	for _, l := range limits {
		b.WriteString(fmt.Sprintf("(%d, %d)", l.Price, l.Size))
		b.WriteString(", ")
	}
	return b.String()
}
