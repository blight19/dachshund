package main

import (
	"dachshund-server/client/matchLog"
	"dachshund-server/client/readLog"
	"dachshund-server/client/send"
	"dachshund-server/client/view"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

type conf struct {
	LogPath     string   `json:"log-path"`
	RedisHost   string   `json:"redisHost"`
	RedisPort   string   `json:"redisPort"`
	RedisPass   string   `json:"redisPass"`
	NetCardName string   `json:"netCardName"`
	MaxLen      int64    `json:"max-len"`
	StreamKey   string   `json:"stream-key"`
	HeartKey    string   `json:"heart-key"`
	Keywords    []string `json:"keywords"`
	SkipOnce    bool     `json:"skip-once"`
}

func init() {

	file, err := os.OpenFile("monitor.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.SetLevel(log.ErrorLevel)
}
func main() {
	log.Info("starting...")
	var config conf
	confFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(confFile, &config)
	if err != nil {
		fmt.Println(err)
	}
	var sender = send.NewRedisClient(config.RedisHost,
		config.RedisPort, config.RedisPass,
		config.NetCardName, config.MaxLen,
		config.StreamKey,
		config.HeartKey,
	)
	//start http server
	prvKeyL, err := os.ReadFile("prvKey.pem")

	if err != nil {
		panic(err)
	}
	var matcher = matchLog.Matcher{KeyWords: config.Keywords}
	reader := readLog.LogReader{
		Sender:   &sender,
		LogPath:  config.LogPath,
		Matcher:  matcher,
		SkipOnce: config.SkipOnce,
	}
	go reader.Run()
	s := view.NewServer(prvKeyL, config.LogPath, config.NetCardName)
	s.Serve()

}
