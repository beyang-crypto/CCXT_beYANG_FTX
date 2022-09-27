package ws

type Ticker struct {
	Channel string     `json:"channel"`
	Market  string     `json:"market"`
	Type    string     `json:"type"`
	Data    dataTicker `json:"data"`
}
type dataTicker struct {
	Bid     float64 `json:"bid"`
	Ask     float64 `json:"ask"`
	BidSize float64 `json:"bidSize"`
	AskSize float64 `json:"askSize"`
	Last    float64 `json:"last"`
	Time    float64 `json:"time"`
}
