package worker

import (
	"dachshund-server/utils"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

// fact number of workers is numWorkers*2

func (w Worker) Run(numWorkers int, groupName string, NetCardName string) {
	w.CreateGroup(w.StreamKey, groupName)
	var workers []chan redis.XMessage
	var workersCk []chan redis.XMessage
	ip, err := utils.LocalIpByName(NetCardName)
	if err != nil {
		log.Fatal("Got server ip failed,check net card name")
	}
	for i := 0; i < numWorkers; i++ {
		workChanCk := w.Work(groupName, i, "0-0") //when cus failed this work
		workChan := w.Work(groupName, i, ">")
		workers = append(workers, workChan)
		workersCk = append(workersCk, workChanCk)
	}
	ackChan := w.Ack(groupName) //used to
	sendChan, ckHeartChan, sHeart := w.SendMsg(ackChan)

	w.ckHeart(ckHeartChan, w.HeartKey, 120) //client
	w.ckHeart(sHeart, w.SheartKey, 30)      //server
	w.sendHeart(w.SheartKey, ip)            //self send

	for i := 0; i < numWorkers; i++ {
		ww1 := workers[i]
		ww2 := workersCk[i]
		go func(w1, w2 chan redis.XMessage, i int) {
			for {
				select {
				case msg := <-w2:
					v := msg.Values
					if v != nil {
						v["ID"] = msg.ID
						sendChan <- v
					}
				case msg := <-w1:
					v := msg.Values
					v["ID"] = msg.ID
					sendChan <- v
				}
			}
		}(ww1, ww2, i)
	}
}
