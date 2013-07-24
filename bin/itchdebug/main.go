package main

func main() {
	println("itchdebug not supported at this time.")
}

/*

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/fmstephe/fstrconv"
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/msg"
	"github.com/fmstephe/matching_engine/itch"
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
		ir := itch.NewItchReader(*filePath)
		defer func() {
			if r := recover(); r != nil {
				println(fmt.Sprintf("Panic at line %d", ir.LineCount()))
				print(fmt.Sprintf("%#v", r))
				panic("Repanic")
			}
		}()
		in := bufio.NewReader(os.Stdin)
		dispatch := make(chan interface{}, 20)
		orders := make(chan *msg.Message)
		m := matcher.NewMatcher(1000)
		m.SetDispatch(dispatch)
		m.SetAppMsgs(appMsgs)
		go m.Run()
		//
		var o *msg.OrderNode
		var err error
		for {
			o, _, err = ir.ReadOrderNode()
			if err != nil {
				panic(err)
			}
			if o != nil && (o.Kind() == msg.BUY || o.Kind() == msg.SELL || o.Kind() == msg.CANCEL) {
				orders<-o
				clear(dispatch)
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

func checkPause(in *bufio.Reader, ir *ItchReader, o *msg.OrderNode, bLine uint) byte {
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

func checkPrint(ir *ItchReader, o *msg.OrderNode, m *matcher.M, bLine uint) {
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

func printInfo(ir *ItchReader, o *msg.OrderNode, m *matcher.M) {
	buys, sells, orders, executions := m.Survey()
	println("OrderNode       ", o.String())
	println("Line        ", ir.LineCount())
	println("Max Buy     ", fstrconv.Itoa64Delim(int64(ir.MaxBuy()), ','))
	println("Min Sell    ", fstrconv.Itoa64Delim(int64(ir.MinSell()), ','))
	println("Executions  ", executions)
	println("...")
	println("Total       ", orders.Size())
	println("Buy Limits  ", formatLimits(buys))
	println("Sell Limits ", formatLimits(sells))
	println()
}

func formatLimits(limits []*msg.SurveyLimit) string {
	var b bytes.Buffer
	for _, l := range limits {
		b.WriteString(fmt.Sprintf("(%s, %s)", fstrconv.Itoa64Delim(int64(l.Price), ','), fstrconv.Itoa64Delim(int64(l.Size), ',')))
		b.WriteString(", ")
	}
	return b.String()
}

func drain(c chan interface{}) {
	for len(c) > 0 {
		<-c
	}
}
*/
