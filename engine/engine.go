package engine

import (
	"github.com/fmstephe/matching_engine/trade"
	//"container/heap"
)

var tradeChan = make(chan *trade.Order, 256)

func StartEngine() {
	//stockMap := make(map[int])
	for {
	//	t := <- tradeChan
	}
}
