package ws

func (b *FTXWS) processTicker(symbol string, data Ticker) {
	b.Emit(ChannelTicker, symbol, data)
}
