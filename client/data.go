package client

import ()

type ClientMessage struct {
	Kind    string `json:"kind"`
	Price   uint64 `json:"price"`
	Amount  uint32 `json:"amount"`
	StockId uint32 `json:"stockId"`
	TradeId uint32 `json:"tradeId"`
}

type receivedMessage struct {
	FromClient bool          `json:"fromClient"`
	Accepted   bool          `json:"accepted"`
	Message    ClientMessage `json:"message"`
}

type response struct {
	Balance     balanceManager  `json:"balance"`
	Stocks      stockManager    `json:"stocks"`
	Received    receivedMessage `json:"received"`
	Outstanding []ClientMessage `json:"outstanding"`
}
