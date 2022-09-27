package rest

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
	bt, _ := data.(WalletBalance)
	return bt
}
