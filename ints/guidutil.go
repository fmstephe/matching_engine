package ints

func Combine(high, low uint32) uint64 {
	return (uint64(high) << 32) | uint64(low)
}

func High32(i uint64) uint32 {
	return uint32(i >> 32)
}

func Low32(i uint64) uint32 {
	return uint32(i)
}
