package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"github.com/fmstephe/matching_engine/trade"
)

var (
	filePath = flag.String("f", "", "Relative path to an ITCH file to read")
)

func main() {
	flag.Parse()
	loop()
}

func loop() {
	in := bufio.NewReader(os.Stdin)
	r := NewItchReader(*filePath)
	var o *trade.Order
	var l string
	var err error
	for {
		o, l, err = r.ReadOrder()
		if err != nil {
			println(err.Error())
			return
		}
		println(fmt.Sprintf("%#v", o))
		print(l)
		awaitNextLine(in)
	}
}

func awaitNextLine(in *bufio.Reader) {
	println("...")
	if _, err := in.ReadString('\n'); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
