package models

import (
	"EasyChat/dbConn"
	sf "EasyChat/utils/snowFlake"
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	MsgTable                  = "EasyChat:Msg:"
	RoomTable                 = "EasyChat:Room:"
	DataPersistenceQueue      = "EasyChat:DataPersistenceQueue:"
	DataPersistenceRetryQueue = "EasyChat:DataPersistenceRetryQueue"
	InitialLastId             = 11111111    //初始化ws连接时，客户端发送的last_id
	RoomTableExpire           = 2 * 60 * 60 //聊天室消息Table缓存时间
	MsgTableExpire            = 2 * 60 * 60 //消息缓存时间
	ExpireWeight              = 1 * 60 * 60
)

type Msg struct {
	Id        int64     `gorm:"column:id" json:"id"`
	SenderId  int64     `gorm:"column:sender_id" json:"sender_id"`
	RoomKey   string    `gorm:"column:room_key" json:"room_key"`
	Content   string    `gorm:"column:content" json:"content"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (Msg) TableName() string {
	return "msg"
}

var redisPool = dbConn.NewRedisPool()

// 创建消息缓存（存入redis）
func (m *Msg) SetCache() error {
	//创建redis连接
	rc := redisPool.Get()
	defer rc.Close()
	//序列化消息
	msgData, err := json.Marshal(m)
	if err != nil {
		return err
	}
	//查询Mysql中近24小时的消息数量
	msg_num, err := m.CountLatestNum(m.RoomKey, time.Now().Add(-24*time.Hour))
	if err != nil {
		return err
	}
	//根据消息数量设置响应的缓存时间
	msg_expire := 1
	if 0 <= msg_num && msg_num <= 30 {
		msg_expire = 2 * ExpireWeight
	}
	if 30 < msg_num && msg_num <= 100 {
		msg_expire = 6 * ExpireWeight
	}
	if 100 < msg_num {
		msg_expire = 12 * ExpireWeight
	}
	//存储字符串 key:消息id, value:消息json数据
	_, err1 := rc.Do("SET", MsgTable+strconv.FormatInt(m.Id, 10), msgData, "EX", msg_expire)
	if err1 != nil {
		return err1
	}
	//存储有序集合（sorted set）key:聊天室id, score:生成消息时的时间戳,value:消息id
	// timestamp := time.Now().Unix()
	_, err2 := rc.Do("ZADD", RoomTable+m.RoomKey, m.CreatedAt.Unix(), m.Id)
	if err2 != nil {
		return err2
	}

	//设置有序集合成员的过期时间
	_, err3 := rc.Do("EXPIRE", RoomTable+m.RoomKey, msg_expire)
	if err3 != nil {
		return err3
	}
	//同时更新该聊天室中所有消息的缓存过期时间
	msgIdSlice, err := redis.Int64s(rc.Do("ZRANGE", RoomTable+m.RoomKey, 0, -1))
	if err != nil {
		return err
	}
	for _, msg_id := range msgIdSlice {
		rc.Do("EXPIRE", MsgTable+strconv.FormatInt(msg_id, 10), msg_expire)
	}

	return nil
}

// 批量缓存时间大于created_at的信息
func (m *Msg) SetBatchCache(msgBatch []*Msg, created_at time.Time) error {
	for _, msg := range msgBatch {
		if msg.CreatedAt.Unix() > created_at.Unix() {
			err := msg.SetCache()
			if err != nil {
				log.Println(err)
			}
		}
	}
	return nil
}

// 提取消息缓存
func (m Msg) GetCache(room_key string, last_id int64) ([]*Msg, error) {
	//创建redis连接
	rc := redisPool.Get()
	defer rc.Close()

	msgIdSlice := make([]int64, 0, 20)
	var err1 error

	//若last_id为11111111，即前端暂时未收到任何新消息,则提取所有消息
	if last_id == InitialLastId {
		// fmt.Println(RoomTable + room_key)
		//提取所有的消息id
		msgIdSlice, err1 = redis.Int64s(rc.Do("ZRANGE", RoomTable+room_key, 0, -1))
		if err1 != nil {
			return nil, err1
		}
	} else {
		//确定上一条所提取消息的下标
		lastMsgOrder, err := redis.Int64(rc.Do("ZRANK", RoomTable+room_key, strconv.FormatInt(last_id, 10)))
		if err != nil {
			return nil, err
		}
		//确定当前的总消息数
		MsgNum, err := redis.Int64(rc.Do("ZCARD", RoomTable+room_key))
		if err != nil {
			return nil, err
		}
		latestMsgOrder := MsgNum - 1
		// fmt.Println(lastMsgOrder, latestMsgOrder)
		//提取所需要的消息id(上一条提取的消息'下标+1'到最新的消息的下标)
		msgIdSlice, err1 = redis.Int64s(rc.Do("ZRANGE", RoomTable+room_key, lastMsgOrder+1, latestMsgOrder))
		if err1 != nil {
			return nil, err1
		}
	}

	//没有新消息
	if len(msgIdSlice) == 0 {
		return nil, nil
	}
	//反序列化消息
	ms := []*Msg{}
	for _, msgId := range msgIdSlice {
		msgData, err := redis.Bytes(rc.Do("GET", MsgTable+strconv.FormatInt(msgId, 10)))
		if err != nil {
			return nil, err
		}
		temp := &Msg{}
		if err = json.Unmarshal(msgData, &temp); err != nil {
			return nil, err
		}
		ms = append(ms, temp)
	}
	return ms, nil
}

// 保存单条数据到Mysql中
func (m *Msg) SaveOne() error {
	result := DB.Create(&m)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// 获取Mysql中单条的数据
func (m *Msg) GetOne(msg_id int64) error {
	result := DB.Where("id = ?", msg_id).Find(&m)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// 保存批量数据到Mysql中
func (m *Msg) SaveBatch(batchMsg []*Msg) {
	// result := DB.Create(&batchMsg)
	// if result.Error != nil {
	// 	return result.Error
	// }
	// return nil

	//获取redis连接
	rc := redisPool.Get()
	defer rc.Close()
	for _, msg := range batchMsg {
		//数据插入mysql
		result := DB.Create(&msg)
		if result.Error != nil {
			log.Println(result.Error)
			//数据写入Mysql发生错误，则将该消息放入（数据持久化）重试队列
			buf, err := json.Marshal(&msg)
			if err != nil {
				log.Println(err)
			}
			ok, _ := redis.Bool(rc.Do("LPUSH", DataPersistenceRetryQueue, buf))
			if ok == false {
				log.Println("DataPersistenceRetryQueue LPUSH Error")
			}
		}
	}
}

// 获取Mysql中的一批数据,小于created_at的msg_num条数据
func (m *Msg) GetBatch(room_key string, created_at time.Time, msg_num int) (batchMsg []*Msg, err error) {
	//升序排列
	query := DB.Where("room_key = ? and created_at <= ?", room_key, created_at).Order("created_at ASC")
	if msg_num != 0 {
		query = query.Limit(msg_num)
	}
	result := query.Find(&batchMsg)
	if result.Error != nil {
		return nil, result.Error
	}
	return batchMsg, nil
}

// 获取Mysql中的一批数据的数量
func (m *Msg) CountNum(room_key string, created_at time.Time) (count int64, err error) {
	//升序排列
	result := DB.Model(&Msg{}).Where("room_key = ? and created_at <= ?", room_key, created_at).Order("created_at ASC").Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

// 获取Mysql中的最近一段时间内的数据的数量
func (m *Msg) CountLatestNum(room_key string, created_at time.Time) (count int64, err error) {
	//升序排列
	result := DB.Model(&Msg{}).Where("room_key = ? and created_at >= ?", room_key, created_at).Order("created_at DESC").Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

// 获取Mysql中的最新一条消息的id
func (m *Msg) GetPersistenceLastId(room_key string) (msg_id int64, err error) {
	//升序排列
	result := DB.Where("room_key = ?", room_key).Order("created_at ASC").First(m)
	if result.Error != nil {
		return 0, result.Error
	}
	return m.Id, nil
}

// 数据持久化重试
func (m *Msg) DataPersistenceRetry() {

	//指数退避算法
	//重试次数
	var retry_times float64 = 1
	//随机数
	rand_num := 0
	//间隙时间
	slot_times := 5 * time.Second

	for retry_times <= 16 {
		//业务重试
		err := m.SaveOne()
		if err != nil {
			log.Println(err)
			continue
		}
		//等待
		//k=Min[重试次数，10]
		k := retry_times
		if k > 10 {
			k = 10
		}
		//使用雪花算法生成的id作为随机数种子
		rand.Seed(sf.GenID())
		//rand.Intn() 返回的随机数范围为 [0,n) ,即此处生成的随机数范围为(2^k)-1
		rand_num = rand.Intn(int(math.Pow(2, k)))
		for i := 0; i < rand_num; i++ {
			time.Sleep(slot_times)
		}
		//重试次数加一
		retry_times++
	}

}
