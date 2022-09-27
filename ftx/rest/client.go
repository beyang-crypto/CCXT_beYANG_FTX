package rest

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

const (
	RestEndpointURL = "https://ftx.com/api"
)

type Configuration struct {
	Addr      string `json:"addr"`
	ApiKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
	DebugMode bool   `json:"debug_mode"`
}

type FTXWS struct {
	cfg *Configuration
}

func New(config *Configuration) *FTXWS {

	// 	потом тут добавятся различные другие настройки
	b := &FTXWS{
		cfg: config,
	}
	return b
}

func (ex *FTXWS) GetBalance() interface{} {
	//	https://docs.ftx.com/#get-balances
	//	получение времяни
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	ts := time.Now().UTC().Unix() * 1000
	url := ex.cfg.Addr + "/wallet/balances"
	apiKey := ex.cfg.ApiKey
	secretKey := ex.cfg.SecretKey

	signature_payload := fmt.Sprintf("%dGET/api/wallet/balances", ts)
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(signature_payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	//	реализация метода GET
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("FTX-KEY", apiKey)
	req.Header.Set("FTX-SIGN", signature)
	req.Header.Set("FTX-TS", fmt.Sprintf("%d", ts))
	//	код для вывода полученных данных
	if err != nil {
		log.Fatalln(err)
	}
	response, err := client.Do(req)
	data, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
	if ex.cfg.DebugMode {
		log.Printf("BinanceWalletBalance %v", string(data))
	}

	var walletBalance WalletBalance
	err = json.Unmarshal(data, &walletBalance)
	if err != nil {
		log.Printf(`
			{
				"Status" : "Error",
				"Path to file" : "CCXT_beYANG_FTX/ftx",
				"File": "api.go",
				"Functions" : "(ex *FTXWS) GetBalance() (WalletBalance)",
				"Function where err" : "json.Unmarshal",
				"Exchange" : "Ftx",
				"Comment" : %s to WalletBalance struct,
				"Error" : %s
			}`, string(data), err)
		log.Fatal()
	}

	return walletBalance

}
