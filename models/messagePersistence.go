package models

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/garyburd/redigo/redis"
)

/*
*数据持久化，服务端每收到一次客户端发来的信息，就执行一次
* 只负责将消息加入队列，并从队列中消费一条消息。
 */
func (m *Msg) DataPersistence(room_key string, user_id int64) error {

	//维护一个数据持久化的队列，使用redis的列表List,key为 room_key + user_id
	rc := redisPool.Get()
	defer rc.Close()
	ListKey := DataPersistenceQueue + room_key + ":" + strconv.Itoa(int(user_id))

	//判断列表里面的消息数量
	queue_len, err := redis.Int(rc.Do("LLEN", ListKey))
	if err != nil {
		return err
	}

	//队列不为空，出现消息堆积，需要批量消费
	if queue_len > 0 {
		if err := batchConsumption(ListKey); err != nil {
			log.Println(err)
			return err
		}
		return nil
	}

	//TODO：设计一段逻辑，根据队列的长度，判断队列是否已满
	//，若队列已满则要立即将当前的消息存入Mysql，并将堆积的消息批量持久化
	//TODO: 需要处理队列中消息堆积的问题
	//TODO: 关键在于，根据queue_len,动态调整消息是否加入队列、是否需要批量消费数据
	//solution: 专门启动消费者协程，监听queue_len,动态调整消息消费策略
	//log.Println("queue_len", queue_len)

	//将当前的消息加入队列
	msgData, err := json.Marshal(&m)
	if err != nil {
		return err
	}
	resp, err := redis.Bool(rc.Do("LPUSH", ListKey, msgData))
	if resp == false {
		return fmt.Errorf("LPUSH error")
	}

	//提取消息队列的队头（右侧）的单条消息
	msgBytes, err := redis.Bytes(rc.Do("RPOP", ListKey))
	if err != nil {
		return err
	}
	if len(msgBytes) == 0 {
		return fmt.Errorf("bytes len = 0")
	}
	//将提取的单条消息存入Mysql
	msg := &Msg{}
	err1 := json.Unmarshal(msgBytes, msg)
	if err != nil {
		return err1
	}
	err2 := msg.SaveOne()
	if err2 != nil {
		log.Println(err2)
		return err2
	}

	return nil

}

// 数据持久化队列-批量消费方案
func batchConsumption(list_key string) error {

	//加锁解锁，确保协程并发安全
	//此处的消费者，每次websocket连接只启动一个协程，设置锁是非必需的
	var operation sync.Mutex
	operation.Lock()
	defer operation.Unlock()

	//从连接池中获取redis连接
	rc := redisPool.Get()
	defer rc.Close()

	//判断队列里面的消息数量
	queue_len, err := redis.Int(rc.Do("LLEN", list_key))
	if err != nil {
		return err
	}
	//根据queue_len调整消费策略
	err1 := consumptionSolution(accumulationLevel(queue_len), list_key, queue_len, rc)
	if err1 != nil {
		return err1
	}

	return nil
}

// 消息堆积程度计算
func accumulationLevel(queue_len int) rune {
	if 0 < queue_len && queue_len <= 5 {
		return 'A'
	}
	if 5 < queue_len && queue_len <= 10 {
		return 'B'
	}
	if 10 < queue_len && queue_len <= 20 {
		return 'C'
	}
	if 20 < queue_len && queue_len <= 30 {
		return 'D'
	}
	log.Println("[NOW] queue_len: ", queue_len)
	return '1'
}

// 消费方案
func consumptionSolution(level rune, redis_list_key string, queue_len int, rc redis.Conn) error {
	defer rc.Close()

	//每批操作的列表下标
	var list_begin, list_end int
	switch level {
	case 'A':
		//一次性操作队列内的所有消息
		list_begin, list_end = 0, -1
		//获取列表内的所有消息
		batchMsgData, err := redis.ByteSlices(rc.Do("LRANGE", redis_list_key, list_begin, list_end))
		if err != nil {
			return err
		}
		//反序列化
		batchMsg, err := batchMsgUnmarshal(batchMsgData)
		if err != nil {
			return err
		}
		//数据持久化
		msg := &Msg{}
		msg.SaveBatch(batchMsg)

	case 'B':
		//分两批进行操作
		gap := 5
		last_order := 0
		for i := 0; i < 2; i++ {
			//第一批操作5条
			if i == 0 {
				list_begin, list_end = queue_len-1-(gap-1), queue_len-1
				last_order = queue_len - 1 - (gap - 1)
			}
			//第二批操作余下的
			if i == 1 {
				list_begin, list_end = 0, last_order-1
			}
			//测试
			log.Println(list_begin, list_end)
			//获取要持久化的信息
			batchMsgData, err := redis.ByteSlices(rc.Do("LRANGE", redis_list_key, list_begin, list_end))
			if err != nil {
				return err
			}
			//反序列化
			batchMsg, err := batchMsgUnmarshal(batchMsgData)
			if err != nil {
				return err
			}
			//数据持久化
			msg := &Msg{}
			msg.SaveBatch(batchMsg)
		}

	case 'C':
		//分两批进行操作
		gap := 10
		last_order := 0
		for i := 0; i < 2; i++ {
			//第一批操作10条
			if i == 0 {
				list_begin, list_end = queue_len-1-(gap-1), queue_len-1
				last_order = queue_len - 1 - (gap - 1)
			}
			//第二批操作余下的
			if i == 1 {
				list_begin, list_end = 0, last_order-1
			}
			//获取要持久化的信息
			batchMsgData, err := redis.ByteSlices(rc.Do("LRANGE", redis_list_key, list_begin, list_end))
			if err != nil {
				return err
			}
			//反序列化
			batchMsg, err := batchMsgUnmarshal(batchMsgData)
			if err != nil {
				return err
			}
			//数据持久化
			msg := &Msg{}
			msg.SaveBatch(batchMsg)

		}
	case 'D':
		//分两批进行操作
		gap := 10
		last_order := queue_len - 1
		for i := 0; i < 3; i++ {

			//第一批操作10条
			if i == 0 {
				list_begin, list_end = queue_len-1-(gap-1), queue_len-1
				last_order = queue_len - 1 - (gap - 1)
			}
			//第二批继续操作10条
			if i == 1 {
				list_begin, list_end = last_order-1-(gap-1), last_order-1
				last_order = last_order - 1 - (gap - 1)
			}
			//第三批继续操作余下
			if i == 2 {
				list_begin, list_end = 0, last_order-1
			}
			//获取要持久化的信息
			batchMsgData, err := redis.ByteSlices(rc.Do("LRANGE", redis_list_key, list_begin, list_end))
			if err != nil {
				return err
			}
			//反序列化
			batchMsg, err := batchMsgUnmarshal(batchMsgData)
			if err != nil {
				return err
			}
			//数据持久化
			msg := &Msg{}
			msg.SaveBatch(batchMsg)
			// if err1 != nil {
			// 	log.Println(err1)
			// 	if !strings.Contains(err1.Error(), "Duplicate entry") {
			// 		return err1
			// 	}
			// }
		}
	default:
		log.Println("TOO MANY MSG!!!")
		return fmt.Errorf("too many msg, queue_len: %d", queue_len)
	}

	//数据持久化完成，清空该数据持久化队列(清空列表)
	ok, err := redis.Bool(rc.Do("DEL", redis_list_key))
	if ok != true {
		log.Println(err)
		return fmt.Errorf("del list error")
	}

	return nil
}

// 一批信息Msg反序列化
func batchMsgUnmarshal(batchMsgData [][]byte) (batchMsg []*Msg, err error) {
	for _, msgData := range batchMsgData {
		msg := &Msg{}
		if err = json.Unmarshal(msgData, &msg); err != nil {
			return nil, err
		}
		batchMsg = append(batchMsg, msg)
	}
	return batchMsg, nil
}
