package worker

import (
	"context"
	"dachshund-server/server/send"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

type RedisConn struct {
	Host        string
	Port        string
	Password    string
	redisClient *redis.Client
}
type Worker struct {
	*RedisConn
	S         send.Sender
	WorkerID  int
	HeartKey  string
	StreamKey string
	SheartKey string
}

func (r *RedisConn) conn() {
	rdb := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%s", r.Host, r.Port),
		Password:        r.Password, // no password set
		DB:              0,          // use default DB
		DialTimeout:     time.Second * 60,
		MinRetryBackoff: time.Second,
		MaxRetryBackoff: time.Second * 10,
		MaxRetries:      10,
		PoolSize:        10,
	})
	r.redisClient = rdb

}

func (w Worker) CreateGroup(streamkey, groupname string) {
	ctx := context.Background()
	w.RedisConn.conn()
	w.redisClient.XGroupCreate(ctx, streamkey, groupname, "0-0")
}

func (w Worker) Work(groupName string, cusId int, msgid string,
) chan redis.XMessage {
	out := make(chan redis.XMessage)
	ctx := context.Background()
	w.RedisConn.conn()
	cusName := fmt.Sprintf("Cus[%d-%d]", w.WorkerID, cusId)
	go func(msgId string) {
		for {
			if msgId == "0-0" {
				time.Sleep(time.Second * 5)
			}
			args := redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: cusName,
				Streams:  []string{w.StreamKey, msgId},
				Count:    1,
				Block:    time.Second,
				NoAck:    false,
			}
			xcmd := w.redisClient.XReadGroup(ctx, &args)

			res, err := xcmd.Result()
			if err != nil && err != redis.Nil {
				log.Error(err)
			}
			if err == redis.Nil {
				log.Debug("No Stream")
			}
			for _, x := range res {
				for _, y := range x.Messages {
					log.Infof("%s %s got message %s,ID%s", cusName, msgId, y.Values, y.ID)
					out <- y
				}
			}
		}
	}(msgid)
	return out
}

func (w Worker) Ack(groupName string) chan<- string {
	ctx := context.Background()
	idChan := make(chan string)
	go func() {
		for {
			x := <-idChan
			w.redisClient.XAck(ctx, w.StreamKey, groupName, x)
			log.Infof("Ack:%s", x)
		}
	}()

	return idChan
}

func (w Worker) SendMsg(ackChan chan<- string) (
	chan<- map[string]interface{}, chan<- map[string]interface{},
	chan<- map[string]interface{}) {
	msg := make(chan map[string]interface{}, 2000)
	ckHeart := make(chan map[string]interface{}, 100)
	sHeart := make(chan map[string]interface{}, 5)
	go func() {
		for true {
			select {
			case message := <-msg:
				log.Infof("Send message:%s", message)
				err := w.S.Send(message)
				if err != nil {
					log.Errorf("Send failed %s Error:%s", message, err)
					continue
				}
				if id, ok := message["ID"].(string); ok {
					ackChan <- id
				}
			case message := <-ckHeart:
				log.Infof("Send message:%s", message)
				err := w.S.Send(message)
				if err != nil {
					log.Errorf("Send failed %s Error:%s", message, err)
					continue
				}

			case message := <-sHeart:
				log.Infof("Send message:%s", message)
				err := w.S.Send(message)
				if err != nil {
					log.Errorf("Send failed %s Error:%s", message, err)
					continue
				}

			}
			<-time.Tick(time.Second * 3)
		}
	}()
	return msg, ckHeart, sHeart
}
