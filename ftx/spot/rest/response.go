package rest

import "log"

//https://docs.ftx.com/#get-balances
type WalletBalance struct {
	Success bool                  `json:"success"`
	Result  []resultWalletBalance `json:"result"`
}
type resultWalletBalance struct {
	Coin                   string  `json:"coin"`
	Free                   float64 `json:"free"`
	SpotBorrow             float64 `json:"spotBorrow"`
	Total                  float64 `json:"total"`
	UsdValue               float64 `json:"usdValue"`
	AvailableWithoutBorrow float64 `json:"availableWithoutBorrow"`
}

func FTXToWalletBalance(data interface{}) WalletBalance {
	bt, ok := data.(WalletBalance)
	if !ok {
		log.Printf(`
			{
				"Status" : "Error",
				"Path to file" : "CCXT_beYANG_FTX/ftx/rest",
				"File": "response.go",
				"Functions" : "FTXToWalletBalance(data interface{}) WalletBalance
				"Exchange" : "FTX",
				"Comment" : "Ошибка преобразования %v в WalletBalance"
			}`, data)
		log.Fatal()
	}
	return bt
}
