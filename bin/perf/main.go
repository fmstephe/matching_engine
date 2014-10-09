package main

import (
	"unsafe"
	"flag"
	"github.com/fmstephe/flib/fstrconv"
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/msg"
	"github.com/fmstephe/flib/queues/spscq"
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
	orderNum   = flag.Int64("o", 1, "The number of orders to generate (in millions). Ignored if -f is provided")
	delDelay   = flag.Int64("d", 10, "The number of orders generated before we begin deleting existing orders")
	perfRand   = rand.New(rand.NewSource(1))
	orderMaker = msg.NewMessageMaker(1)
)

func main() {
	doPerf(true)
}

func doPerf(log bool) {
	flag.Parse()
	orderData := getData()
	orderCount := fstrconv.ItoaComma(int64(len(orderData)))
	if log {
		println(orderCount, "OrderNodes Built")
	}
	in, err := spscq.NewPointerQ(1024, 1000 * 1000)
	if err != nil {
		panic(err.Error())
	}
	out, err := spscq.NewPointerQ(1024, 1000 * 1000)
	if err != nil {
		panic(err.Error())
	}
	m := matcher.NewMatcher(*delDelay * 2)
	m.Config("Perf Matcher", in, out)
	go m.Run()
	startProfile()
	defer endProfile()
	go write(in, orderData)
	read(out)
}

func write(q *spscq.PointerQ, orders []msg.Message) {
	for i := range orders {
		q.WriteSingleBlocking(unsafe.Pointer(&orders[i]))
	}
	s := &msg.Message{Kind: msg.SHUTDOWN}
	q.WriteSingleBlocking(unsafe.Pointer(s))
}

func read(q *spscq.PointerQ) {
	start := time.Now().UnixNano()
	for {
		m := (*msg.Message)(q.ReadSingleBlocking())
		if m.Kind == msg.SHUTDOWN {
			break
		}
	}
	total := time.Now().UnixNano() - start
	println("Nanos\t", fstrconv.ItoaComma(total))
	println("Micros\t", fstrconv.ItoaComma(total/1000))
	println("Millis\t", fstrconv.ItoaComma(total/(1000*1000)))
	println("Seconds\t", fstrconv.ItoaComma(total/(1000*1000*1000)))
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
	orders, err := orderMaker.RndTradeSet(*orderNum * 1000 * 1000, *delDelay, 1000, 1500)
	if err != nil {
		panic(err.Error())
	}
	return orders
}
