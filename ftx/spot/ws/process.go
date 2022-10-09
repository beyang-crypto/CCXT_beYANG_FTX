package ws

func (b *FTXWS) processTicker(name string, symbol string, data Ticker) {
	b.Emit(ChannelTicker, name, symbol, data)
}
