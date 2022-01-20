package send

import (
	"context"
	"dachshund-server/utils"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

type RedisSender struct {
	host          string
	port          string
	password      string
	localIp       string
	localHostName string
	c             *redis.Client
	maxlen        int64 //the max length of stream
	StreamKey     string
	HeartKey      string
}

var ctx = context.Background()

type RedisMsg struct {
}

// the factory method

func NewRedisClient(host string, port string, password string, netCardName string, maxlen int64, streamKey, heartKey string) RedisSender {
	ip, err := utils.LocalIpByName(netCardName)
	if err != nil {
		fmt.Println(netCardName)
		panic(err)
	}
	localHost, err := utils.GetHost()
	if err != nil {
		return RedisSender{}
	}

	r := RedisSender{
		host:          host,
		port:          port,
		password:      password,
		localIp:       ip,
		localHostName: localHost,
		maxlen:        maxlen,
		StreamKey:     streamKey,
		HeartKey:      heartKey,
	}
	r.conn()

	return r
}

// heart check ,every 3 second send timestamp to redis

func (r *RedisSender) ping() {
	go func() {
		for true {
			<-time.After(time.Second * 3)
			now := time.Now().Unix()
			r.c.HSet(ctx, r.HeartKey, map[string]interface{}{r.localIp: now})
			log.Debugf("Ping:Send TIMESTAMP[%d]", now)
		}
	}()
}

// send message to redis stream

func (r *RedisSender) Send(msg map[string]interface{}) error {
	xargs := redis.XAddArgs{}
	sendInfo := map[string]interface{}{"IP": r.localIp, "HostName": r.localHostName,
		"Error": msg["Error"], "Offset": msg["Offset"]}
	xargs.Values = sendInfo
	xargs.Stream = r.StreamKey
	xargs.MaxLen = r.maxlen

	r.c.XAdd(ctx, &xargs)
	log.Infof("Send to redis: Error: 【%s】 HostName: 【%s】 IP: 【%s】 Offset: 【%d】",
		sendInfo["Error"], sendInfo["HostName"], sendInfo["IP"], sendInfo["Offset"])
	return nil
}

// connection to the redis server and start heart check

func (r *RedisSender) conn() {
	rdb := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%s", r.host, r.port),
		Password:        r.password, // no password set
		DB:              0,          // use default DB
		IdleTimeout:     20,
		MinRetryBackoff: time.Second,
		MaxRetryBackoff: time.Second * 10,
		MaxRetries:      10,
	})
	r.c = rdb
	//开启心跳检测
	r.ping()
}
