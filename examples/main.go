package main

import (
	"log"
	"time"

	ftxRest "github.com/TestingAccMar/CCXT_beYANG_FTX/ftx/rest"
	ftxWS "github.com/TestingAccMar/CCXT_beYANG_FTX/ftx/ws"
)

func main() {
	cfg := &ftxRest.Configuration{
		Addr:      ftxRest.RestEndpointURL,
		ApiKey:    "SFcbrCf_Xwgg9E8nmWJ5t8wOPU4y8zqGIl-LY6cs",
		SecretKey: "lgXdP8mjQpoiHPWkjAXxI2Pz3fPMtU1vxn892MPk",
		DebugMode: true,
	}
	b := ftxRest.New(cfg)
	// cfg := &ftxWS.Configuration{
	// 	Addr:      ftxWS.HostMainnetPublicTopics,
	// 	ApiKey:    "",
	// 	SecretKey: "",
	// 	DebugMode: true,
	// }
	// b := ftxWS.New(cfg)
	// b.Start()

	// b.Subscribe(ftxWS.ChannelTicker, b.GetPair("BTC", "USDT"))

	// b.On(ftxWS.ChannelTicker, handleBookTicker)

	go func() {
		time.Sleep(5 * time.Second)
		balance := ftxRest.FTXToWalletBalance(b.GetBalance())

		for _, coins := range balance.Result {
			log.Printf("coin = %s, total = %f", coins.Coin, coins.Total)
		}
	}()
	//	не дает прекратить работу программы
	forever := make(chan struct{})
	<-forever
}

func handleBookTicker(symbol string, data ftxWS.Ticker) {
	log.Printf("Ftx Ticker  %s: %v", symbol, data)
}

func handleBestBidPrice(symbol string, data ftxWS.Ticker) {
	log.Printf("Ftx BookTicker  %s: BestBidPrice : %f", symbol, data.Data.Bid)
}
