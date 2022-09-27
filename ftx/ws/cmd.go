package ws

//	Необходим для удобного создания подписок
type Cmd struct {
	Op      string `json:"op"`
	Channel string `json:"channel"`
	Market  string `json:"market"`
}

type Auth struct {
	Op   string `json:"op"`
	Args Args   `json:"args"`
}
type Args struct {
	Key  string `json:"key"`
	Sign string `json:"sign"`
	Time int64  `json:"time"`
}
