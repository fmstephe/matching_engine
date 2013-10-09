package msg

// TODO rename these functions, msg doesn't seem like a good package for these functions anymore
func MkGuid(traderId, tradeId uint32) int64 {
	return (int64(traderId) << 32) | int64(tradeId)
}

func GetTraderId(guid int64) uint32 {
	return uint32(guid >> 32)
}

func GetTradeId(guid int64) uint32 {
	return uint32(guid)
}
