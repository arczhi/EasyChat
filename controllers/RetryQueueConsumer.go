package controllers

import (
	"EasyChat/models"
	"encoding/json"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

// 数据持久化重试队列消费者
func MsgPersistenceRetryQueueConsumer() {
	for {
		rc := redisPool.Get()

		queue_len, err := redis.Int(rc.Do("LLEN", models.DataPersistenceRetryQueue))
		if err != nil {
			log.Println(err)
		}

		if queue_len != 0 {
			msgData, err := redis.Bytes(rc.Do("RPOP", models.DataPersistenceRetryQueue))
			if err != nil {
				log.Println(err)
			}
			msg := &models.Msg{}
			err1 := json.Unmarshal(msgData, msg)
			if err1 != nil {
				log.Println(err1)
			}
			//数据持久化重试
			msg.DataPersistenceRetry()
		}

		rc.Close()
		time.Sleep(10 * time.Second)
	}
}
