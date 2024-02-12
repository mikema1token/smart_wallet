package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var pushLogChan = make(chan types.Log, 1024)
var config = Config{}

var funcSignature = map[string]string{
	"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef": "Transfer",
	"0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67": "Swap",
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			SimpleLog(fmt.Sprintf("%v", err))
			PushMessageToTg(fmt.Sprintf("%v", err))
		}
	}()
	SimpleLog("begin Listening")
	ListeningEoAddress()
}

func ListeningEoAddress() {
	dialer := GetWebSocketDial()
	websocketDialer := rpc.WithWebsocketDialer(dialer)
	url := config.EtherTest
	if !config.Test {
		url = config.EtherMain
	}
	client, err := rpc.DialOptions(context.Background(), url, websocketDialer)
	if err != nil {
		panic(err)
	}
	newClient := ethclient.NewClient(client)

	go LoopListen(newClient, true)
	go LoopListen(newClient, false)
	for log := range pushLogChan {
		PushLogToTg(log)
	}
}

func ListenEvent(client *ethclient.Client, from bool) error {
	filterQuery := ethereum.FilterQuery{
		Topics: make([][]common.Hash, 0),
	}
	filterQuery.Topics = append(filterQuery.Topics, nil)
	topicFilter := make([]common.Hash, 0)
	for _, address := range config.Address {
		topicFilter = append(topicFilter, common.HexToHash(address))
	}
	if from {
		filterQuery.Topics = append(filterQuery.Topics, topicFilter)
		filterQuery.Topics = append(filterQuery.Topics, nil)
	} else {
		filterQuery.Topics = append(filterQuery.Topics, nil)
		filterQuery.Topics = append(filterQuery.Topics, topicFilter)
	}
	logSubscribe, err := client.SubscribeFilterLogs(context.Background(), filterQuery, pushLogChan)
	if err != nil {
		return err
	}
	err = <-logSubscribe.Err()
	return err
}

func PushLogToTg(log types.Log) {
	event := ""
	from := ""
	to := ""
	for i, topic := range log.Topics {
		if i == 0 {
			event = funcSignature[topic.Hex()]
		}
		if i == 1 {
			from = topic.Hex()
		}
		if i == 2 {
			to = topic.Hex()
		}
	}
	message := "from:" + from + "\n" + "to:" + to + "\n" + "event:" + event + "\n" + "hash:" + log.TxHash.Hex()
	PushMessageToTg(message)
}

func GetProxyHttpClient() *http.Client {
	if !config.HttpProxy {
		return http.DefaultClient
	}
	proxyUrl, _ := url.Parse(config.HttpProxyUrl)
	return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
}
func GetWebSocketDial() websocket.Dialer {
	if !config.HttpProxy {
		return websocket.Dialer{}
	}
	dialer := websocket.Dialer{}
	proxyUrl, _ := url.Parse(config.HttpProxyUrl)
	dialer.Proxy = http.ProxyURL(proxyUrl)
	return dialer
}

type Config struct {
	Test         bool     `json:"test"`
	TgBotToken   string   `json:"tg_bot_token"`
	EtherMain    string   `json:"ether_main"`
	EtherTest    string   `json:"ether_test"`
	Address      []string `json:"address"`
	ChatId       string   `json:"chat_id"`
	HttpProxy    bool     `json:"http_proxy"`
	HttpProxyUrl string   `json:"http_proxy_url"`
}

func init() {
	configFile, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&config); err != nil {
		panic(err)
	}
}

func SimpleLog(message string) {
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}
	defer logFile.Close()
	fmt.Println(message)
	log.SetOutput(logFile)
	log.Println(fmt.Sprintf("%s:%s", time.Now().Format(time.RFC3339), message))
}

func PushMessageToTg(message string) {
	client := GetProxyHttpClient()
	urlPath := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?", config.TgBotToken)
	values := url.Values{}
	values.Set("chat_id", config.ChatId)
	values.Set("text", message)
	urlPath = urlPath + values.Encode()
	request, err := http.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		SimpleLog(err.Error())
		return
	}
	_, err = client.Do(request)
	if err != nil {
		SimpleLog(err.Error())
	}
}

func LoopListen(client *ethclient.Client, from bool) {
	go func() {
		for {
			err := ListenEvent(client, from)
			if err != nil {
				PushMessageToTg(fmt.Sprintf("%v", err))
			}
			time.Sleep(time.Minute)
		}
	}()
}
