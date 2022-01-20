package worker

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// every 10 second,choice 10 machine to check heart

func (w Worker) ckHeart(x chan<- map[string]interface{}, key string, th int64) {
	ctx := context.Background()
	var cursor uint64 = 0
	go func(key string, th int64) {
		for true {
			r := w.redisClient.HScan(ctx, key, cursor, "*", 10)
			result, c, err := r.Result()
			cursor = c
			for i := 1; i < len(result)+1; i = i + 2 {
				u, _ := strconv.Atoi(result[i])
				if time.Now().Unix()-int64(u) > th {
					x <- map[string]interface{}{"IP": result[i-1], "Last": u}
					w.redisClient.HDel(ctx, key, result[i-1])
				}
			}
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(time.Second * 10)
		}
	}(key, th)

}

// every 3 second send timestamp to redis
func (w Worker) sendHeart(key, ip string) {
	go func(key, ip string) {
		ctx := context.Background()
		for true {
			<-time.After(time.Second * 3)
			now := time.Now().Unix()
			w.redisClient.HSet(ctx, key, map[string]interface{}{ip: now})
		}
	}(key, ip)
}
