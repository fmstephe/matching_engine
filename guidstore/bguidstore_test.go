package guidstore

import (
	"math"
	"math/rand"
	"testing"
)

const pushes = 1000

func BenchmarkOneBlock(b *testing.B) {
	benchmarker(b, blockSize)
}

func Benchmark2Blocks(b *testing.B) {
	benchmarker(b, blockSize*2)
}

func Benchmark4Blocks(b *testing.B) {
	benchmarker(b, blockSize*4)
}

func Benchmark8Blocks(b *testing.B) {
	benchmarker(b, blockSize*8)
}

func Benchmark16Blocks(b *testing.B) {
	benchmarker(b, blockSize*16)
}

func Benchmark32Blocks(b *testing.B) {
	benchmarker(b, blockSize*32)
}

func Benchmark64Blocks(b *testing.B) {
	benchmarker(b, blockSize*64)
}

func BenchmarkManyBlocks(b *testing.B) {
	benchmarker(b, math.MaxInt64)
}

func benchmarker(b *testing.B, randMax int64) {
	b.StopTimer()
	r := rand.New(rand.NewSource(1))
	a := make([]int64, b.N*pushes)
	s := NewStore()
	for i := 0; i < b.N*pushes; i++ {
		a[i] = r.Int63n(randMax)
	}
	b.StartTimer()
	for i := 0; i < b.N*pushes; i++ {
		s.Push(a[i])
	}
}
