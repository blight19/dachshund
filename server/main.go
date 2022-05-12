package main

import (
	"dachshund-server/server/send"
	"dachshund-server/server/worker"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

type conf struct {
	RedisHost   string `json:"redisHost"`
	RedisPort   string `json:"redisPort"`
	RedisPass   string `json:"redisPass"`
	WorkerId    int    `json:"worker-id"`
	DingSecret  string `json:"ding-secret"`
	DingUrl     string `json:"ding-url"`
	DingProxy   string `json:"ding-proxy"`
	StreamKey   string `json:"stream-key"`
	HeartKey    string `json:"heart-key"`
	SheartKey   string `json:"sheart-key"`
	NetCardName string `json:"netCardName"`
	GroupName   string `json:"groupName"`
}

const (
	workerNum = 5
)

func init() {
	file, err := os.Create("monitor.log")
	if err != nil {
		log.Fatalln("fail to create monitor.log file!")
	}
	log.SetLevel(log.ErrorLevel)
	log.SetOutput(file)
}
func main() {
	log.Info("running...")
	var config conf
	confFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(confFile, &config)
	if err != nil {
		fmt.Println(err)
	}
	dingsender := send.DingDing{
		Secret:  config.DingSecret,
		Url:     config.DingUrl,
		Proxies: config.DingProxy,
	}
	w := worker.Worker{RedisConn: &worker.RedisConn{
		Host:     config.RedisHost,
		Port:     config.RedisPort,
		Password: config.RedisPass,
	},
		S:         dingsender,
		StreamKey: config.StreamKey,
		HeartKey:  config.HeartKey,
		WorkerID:  config.WorkerId,
		SheartKey: config.SheartKey}

	w.Run(workerNum, config.GroupName, config.NetCardName)

	select {}

}
