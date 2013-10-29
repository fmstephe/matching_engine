package ints

func Combine(high, low uint32) int64 {
	return (int64(high) << 32) | int64(low)
}

func High32(i int64) uint32 {
	return uint32(i >> 32)
}

func Low32(i int64) uint32 {
	return uint32(i)
}
