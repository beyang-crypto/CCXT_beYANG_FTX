package main

import (
	"log"

	//ftxRest "github.com/TestingAccMar/CCXT_beYANG_FTX/ftx/rest"
	ftxWS "github.com/TestingAccMar/CCXT_beYANG_FTX/ftx/spot/ws"
)

func main() {
	// cfg := &ftxRest.Configuration{
	// 	Addr:      ftxRest.RestEndpointURL,
	// 	ApiKey:    "SFcbrCf_Xwgg9E8nmWJ5t8wOPU4y8zqGIl-LY6cs",
	// 	SecretKey: "lgXdP8mjQpoiHPWkjAXxI2Pz3fPMtU1vxn892MPk",
	// 	DebugMode: true,
	// }
	// b := ftxRest.New(cfg)
	cfg := &ftxWS.Configuration{
		Addr:      ftxWS.HostMainnetPublicTopics,
		ApiKey:    "",
		SecretKey: "",
		DebugMode: true,
	}
	b := ftxWS.New(cfg)
	b.Start()

	pair1 := b.GetPair("BTC", "USDT")
	pair2 := b.GetPair("eth", "USDT")
	b.Subscribe(ftxWS.ChannelTicker, []string{pair1})
	b.Subscribe(ftxWS.ChannelTicker, []string{pair2})

	b.On(ftxWS.ChannelTicker, handleBestBidPrice)

	// go func() {
	// 	time.Sleep(5 * time.Second)
	// 	balance := ftxRest.FTXToWalletBalance(b.GetBalance())

	// 	for _, coins := range balance.Result {
	// 		log.Printf("coin = %s, total = %f", coins.Coin, coins.Total)
	// 	}
	// }()
	//	не дает прекратить работу программы
	forever := make(chan struct{})
	<-forever
}

func handleBookTicker(name string, symbol string, data ftxWS.Ticker) {
	log.Printf("%s Ticker  %s: %v", name, symbol, data)
}

func handleBestBidPrice(name string, symbol string, data ftxWS.Ticker) {
	log.Printf("%s BookTicker  %s: BestBidPrice : %f", name, symbol, data.Data.Bid)
}
