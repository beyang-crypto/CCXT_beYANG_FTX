package ws

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/buger/jsonparser"      //  Для вытаскивания одного значения из файла json
	"github.com/chuckpreslar/emission" // Эмитер необходим для удобного выполнения функции в какой-то момент
	"github.com/goccy/go-json"         // для создания собственных json файлов и преобразования json в структуру
	"github.com/gorilla/websocket"
)

const (
	HostMainnetPublicTopics = "wss://ftx.com/ws/"
)

const (
	ChannelTicker = "ticker"
)

type Configuration struct {
	Addr      string `json:"addr"`
	ApiKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
	DebugMode bool   `json:"debug_mode"`
}

type FTXWS struct {
	cfg  *Configuration
	conn *websocket.Conn

	mu            sync.RWMutex
	subscribeCmds []Cmd //	сохраняем все подписки у данной биржи, чтоб при переподключении можно было к ним повторно подключиться

	emitter *emission.Emitter
}

func (b *FTXWS) GetPair(args ...string) string {
	pair := args[0] + "/" + args[1]
	return strings.ToUpper(pair)
}

func New(config *Configuration) *FTXWS {

	// 	потом тут добавятся различные другие настройки
	b := &FTXWS{
		cfg:     config,
		emitter: emission.NewEmitter(),
	}
	return b
}

func (b *FTXWS) Subscribe(channel string, coins []string) {
	switch len(coins) {
	case 1:
		cmd := Cmd{
			Op:      "subscribe",
			Channel: channel,
			Market:  coins[0],
		}
		b.subscribeCmds = append(b.subscribeCmds, cmd)
		if b.cfg.DebugMode {
			log.Printf("Создание json сообщения на подписку part 1")
		}
		b.SendCmd(cmd)
	default:
		log.Printf(`
			{
				"Status" : "Error",
				"Path to file" : "CCXT_beYANG_FTX/ftx/ws",
				"File": "client.go",
				"Functions" : "(b *FTXWS) Subscribe(channel string, coins []string)
				"Exchange" : "FTX",
				"Data" : [%v. %v],
				"Comment" : "Слишком много аргументов (за один раз можно подписаться только на один рынок) "
			}`, channel, coins)
		log.Fatal()
	}
}

func (b *FTXWS) Subscribe2(channel string, coin string) {
	cmd := Cmd{
		Op:      "subscribe",
		Channel: channel,
		Market:  coin,
	}
	b.subscribeCmds = append(b.subscribeCmds, cmd)
	if b.cfg.DebugMode {
		log.Printf("Создание json сообщения на подписку part 1")
	}
	b.SendCmd(cmd)
}

//	отправка команды на сервер в отдельной функции для того, чтобы при переподключении быстро подписаться на все предыдущие каналы
func (b *FTXWS) SendCmd(cmd Cmd) {
	data, err := json.Marshal(cmd)
	if err != nil {
		log.Printf(`
			{
				"Status" : "Error",
				"Path to file" : "CCXT_beYANG_FTX/ftx/ws",
				"File": "client.go",
				"Functions" : "(b *FTXWS) sendCmd(cmd Cmd)",
				"Function where err" : "json.Marshal",
				"Exchange" : "FTX",
				"Data" : [%s],
				"Error" : %s
			}`, cmd, err)
		log.Fatal()
	}
	if b.cfg.DebugMode {
		log.Printf("Создание json сообщения на подписку part 2")
	}
	b.Send(string(data))
}

func (b *FTXWS) Send(msg string) (err error) {
	defer func() {
		// recover необходим для корректной обработки паники
		if r := recover(); r != nil {
			if err != nil {
				log.Printf(`
					{
						"Status" : "Error",
						"Path to file" : "CCXT_beYANG_FTX/ftx/ws",
						"File": "client.go",
						"Functions" : "(b *FTXWS) Send(msg string) (err error)",
						"Function where err" : "json.Marshal",
						"Exchange" : "FTX",
						"Data" : [websocket.TextMessage, %s],
						"Error" : %s,
						"Recover" : %v
					}`, msg, err, r)
				log.Fatal()
			}
			err = errors.New(fmt.Sprintf("FTXWs send error: %v", r))
		}
	}()
	if b.cfg.DebugMode {
		log.Printf("Отправка сообщения на сервер. текст сообщения:%s", msg)
	}

	err = b.conn.WriteMessage(websocket.TextMessage, []byte(msg))
	return
}

// подключение к серверу и постоянное чтение приходящих ответов
func (b *FTXWS) Start() error {
	if b.cfg.DebugMode {
		log.Printf("Начало подключения к серверу")
	}
	b.connect()

	cancel := make(chan struct{})

	go func() {
		t := time.NewTicker(time.Second * 15)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				b.ping()
			case <-cancel:
				return
			}
		}
	}()

	go func() {
		defer close(cancel)

		for {
			_, data, err := b.conn.ReadMessage()
			if err != nil {

				if websocket.IsCloseError(err, 1006) {
					b.closeAndReconnect()
					//Необходим вызв SubscribeToTicker в отдельной горутине, рекурсия, думаю, тут неуместна
					log.Printf("Status: INFO	ошибка 1006 начинается переподключение к серверу")

				} else {
					log.Printf(`{
						"Status" : "Error",
						"Path to file" : "CCXT_beYANG_FTX/ftx/ws",
						"File": "client.go",
						"Functions" : "(b *FTXWS) Start() error",
						"Function where err" : "b.conn.ReadMessage",
						"Exchange" : "FTX",
						"Error" : %s
					}`, err)
					log.Fatal()
				}
			} else {
				b.messageHandler(data)
			}
		}
	}()

	return nil
}

//	Необходим для приватных каналов
func (b *FTXWS) Auth() {
	expires := time.Now().Unix()*1000 + 10000
	req := fmt.Sprintf("GET/realtime%d", expires)
	sig := hmac.New(sha256.New, []byte(b.cfg.SecretKey))
	sig.Write([]byte(req))
	signature := hex.EncodeToString(sig.Sum(nil))

	auth := Auth{
		Op: "auth",
		Args: Args{
			Key:  b.cfg.ApiKey,
			Sign: signature,
			Time: expires,
		},
	}
	data, err := json.Marshal(auth)
	if err != nil {
		log.Printf(`
			{
				"Status" : "Error",
				"Path to file" : "CCXT_beYANG_FTX/ftx/ws",
				"File": "client.go",
				"Functions" : "(b *FTXWS) Auth()",
				"Function where err" : "json.Marshal",
				"Exchange" : "FTX",
				"Data" : [%v],
				"Error" : %s
			}`, auth, err)
		log.Fatal()
	}
	if b.cfg.DebugMode {
		log.Printf("Создание json сообщения на подписку")
	}
	b.Send(string(data))
}

func (b *FTXWS) connect() {

	c, _, err := websocket.DefaultDialer.Dial(b.cfg.Addr, nil)
	if err != nil {
		log.Printf(`{
						"Status" : "Error",
						"Path to file" : "CCXT_beYANG_FTX/ftx/ws",
						"File": "client.go",
						"Functions" : "(b *FTXWS) connect()",
						"Function where err" : "websocket.DefaultDialer.Dial",
						"Exchange" : "FTX",
						"Data" : [%s, nil],
						"Error" : %s
					}`, b.cfg.Addr, err)
		log.Fatal()
	}
	b.conn = c
	for _, cmd := range b.subscribeCmds {
		b.SendCmd(cmd)
	}
}

func (b *FTXWS) closeAndReconnect() {
	b.conn.Close()
	b.connect()
}

func (b *FTXWS) ping() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("FTXWs ping error: %v", r)
		}
	}()

	//	https://docs.ftx.com/?python#websocket-api
	err := b.conn.WriteMessage(websocket.TextMessage, []byte(`{"op":"ping"}`))
	if err != nil {
		log.Printf("FTXWs ping error: %v", err)
	}
}

func (b *FTXWS) messageHandler(data []byte) {

	if b.cfg.DebugMode {
		log.Printf("FTXWs %v", string(data))
	}

	//	в ошибке нет необходимости, т.к. она выходит каждый раз, когда не найдет элемент
	typeJSON, _ := jsonparser.GetString(data, "type")

	switch typeJSON {
	case "update":
		channel, _ := jsonparser.GetString(data, "channel")
		switch channel {
		case "ticker":
			market, _ := jsonparser.GetString(data, "market")
			var ticker Ticker
			err := json.Unmarshal(data, &ticker)
			if err != nil {
				log.Printf(`
					{
						"Status" : "Error",
						"Path to file" : "CCXT_beYANG_FTX/ftx/ws",
						"File": "client.go",
						"Functions" : "(b *FTXWS) messageHandler(data []byte)",
						"Function where err" : "json.Unmarshal",
						"Exchange" : "FTX",
						"Comment" : %s to BookTicker struct,
						"Error" : %s
					}`, string(data), err)
				log.Fatal()
			}
			b.processTicker(market, ticker)
		default:
			log.Printf(`
				{
					"Status" : "INFO",
					"Path to file" : "CCXT_beYANG_FTX/ftx/ws",
					"File": "client.go",
					"Functions" : "(b *FTXWS) messageHandler(data []byte)",
					"Exchange" : "FTX",
					"Comment" : "Ответ от неизвестного канала"
					"Message" : %s
				}`, string(data))
			log.Fatal()
		}
	case "error":
		log.Printf(`
			{
				"Status" : "Error",
				"Path to file" : "CCXT_beYANG_FTX/ftx/ws",
				"File": "client.go",
				"Functions" : "(b *FTXWS) messageHandler(data []byte)",
				"Exchange" : "FTX",
				"Message" : %s
			}`, string(data))
		log.Fatal()
	case "subscribed":
		//	Ну что сказать, хорошо, что подписались успешно
	case "pong":
		//	По этому поводу у меня тоже нет слов
	default:
		log.Printf(`
			{
				"Status" : "INFO",
				"Path to file" : "CCXT_beYANG_FTX/ftx/ws",
				"File": "client.go",
				"Functions" : "(b *FTXWS) messageHandler(data []byte)",
				"Exchange" : "FTX",
				"Comment" : "не известный ответ от сервера"
				"Message" : %s
			}`, string(data))
		log.Fatal()
	}
}
