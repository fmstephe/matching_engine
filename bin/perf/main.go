package main

import (
	"flag"
	"github.com/fmstephe/flib/fstrconv"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/msg"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

const (
	StockId = uint32(1)
)

var (
	filePath   = flag.String("f", "", "Relative path to an ITCH file providing test data")
	profile    = flag.String("p", "", "Write out a profile of this application, 'cpu' and 'mem' supported")
	orderNum   = flag.Int("o", 1, "The number of orders to generate (in millions). Ignored if -f is provided")
	delDelay   = flag.Int("d", 10, "The number of orders generated before we begin deleting existing orders")
	perfRand   = rand.New(rand.NewSource(1))
	orderMaker = msg.NewMessageMaker(1)
)

func main() {
	doPerf(true)
}

func doPerf(log bool) {
	flag.Parse()
	data := getData()
	if log {
		orderCount := fstrconv.ItoaComma(int64(len(data)))
		println(orderCount, "OrderNodes Built")
	}
	start := time.Now()
	defer func() {
		if log {
			println("Running Time: ", time.Now().Sub(start).String())
		}
	}()
	startProfile()
	defer endProfile()
	singleThreaded(log, data)
}

func singleThreaded(log bool, data []msg.Message) {
	inout := coordinator.NewNoopReaderWriter()
	mchr := matcher.NewMatcher(*delDelay * 2)
	mchr.Config("Perf Matcher", inout, inout)
	for i := range data {
		mchr.Submit(&data[i])
	}
}

func multiThreadedChan(log bool, data []msg.Message) {
	in := coordinator.NewChanReaderWriter(1024)
	out := coordinator.NewChanReaderWriter(1024)
	multiThreaded(log, data, in, out)
}

func multiThreadedPreloaded(log bool, data []msg.Message) {
	in := coordinator.NewPreloadedReaderWriter(data)
	out := coordinator.NewShutdownReaderWriter()
	multiThreaded(log, data, in, out)
}

func multiThreadedSPSCQ(log bool, data []msg.Message) {
	in := coordinator.NewSPSCQReaderWriter(1024 * 1024)
	out := coordinator.NewSPSCQReaderWriter(1024 * 1024)
	multiThreaded(log, data, in, out)
}

func multiThreaded(log bool, data []msg.Message, in, out coordinator.MsgReaderWriter) {
	mchr := matcher.NewMatcher(*delDelay * 2)
	mchr.Config("Perf Matcher", in, out)
	go run(mchr)
	go write(in, data)
	// Read all messages coming out of the matching engine
	read(out)
}

func read(reader coordinator.MsgReader) {
	for {
		m := reader.Read()
		if m.Kind == msg.SHUTDOWN {
			break
		}
	}
}

type runner interface {
	Run()
}

func run(r runner) {
	r.Run()
}

func write(in coordinator.MsgWriter, msgs []msg.Message) {
	for i := range msgs {
		in.Write(msgs[i])
	}
	in.Write(msg.Message{Kind: msg.SHUTDOWN})
}

func startProfile() {
	if *profile == "cpu" {
		f, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}
}

func endProfile() {
	if *profile == "cpu" {
		pprof.StopCPUProfile()
	}
	if *profile == "mem" {
		f, err := os.Create("mem.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
	}
}

func getData() []msg.Message {
	orders, err := orderMaker.RndTradeSet(*orderNum*1000*1000, *delDelay, 1000, 1500)
	if err != nil {
		panic(err.Error())
	}
	return orders
}
