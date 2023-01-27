package controllers

import (
	"EasyChat/config"
	"EasyChat/dbConn"
	"EasyChat/models"
	"EasyChat/utils/StringByte"
	sf "EasyChat/utils/snowFlake"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
)

const (
	snowFlakeMachineID = 12345678
	MsgStampTable      = "EasyChat:MsgStamp:"
)

// websocket相关配置
var connectionTimeout = config.Cfg.WEBSOCKET.ConnectionTimeout

// redis连接池
var redisPool = dbConn.NewRedisPool()

// 定义一个upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func MsgExchange(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI, " 调用: ", r.RemoteAddr)
	//跨域检查
	upgrader.CheckOrigin = func(r *http.Request) bool {
		// if strings.Contains(r.RemoteAddr, "127.0.0.1") {
		// 	return true
		// }
		// return false
		return true
	}
	//升级http连接为webSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer ws.Close()

	//监听并处理业务
	handle(ws)

}

// 消息接收与发送
func handle(conn *websocket.Conn) {

	//启动数据持久化重试队列消费者
	go MsgPersistenceRetryQueueConsumer()

	//客户端当前接收到的消息id
	lastId := make(chan int64)
	//当前用户的聊天室room_key
	var room_key string
	//当前用户的id
	var user_id int64

	//(1)持续向客户端发送消息
	go func() {
		for {
			err := sendMsg(conn, room_key, user_id, <-lastId)
			if err != nil {
				log.Println(err)
				continue
			}
			// time.Sleep(100 * time.Millisecond)
		}
	}()

	//(2)挂起并接收客户端的消息
	for {
		msg := &models.Msg{}
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
		}
		//正常接收消息的情况下，msgType可能是1或其他数字
		//若msgType为-1，则证明浏览器已经主动将连接关闭
		if msgType == -1 {
			//当客户端主动关闭连接时，删除last_id
			removeLastId(user_id, room_key)
			return
		}

		//测试
		// fmt.Println(string(data))
		// fmt.Println(string(data[:8]))    //11111111 初始化的last_id
		// fmt.Println(string(data[8:18]))  //10位 room_key
		// fmt.Println(string(data[18:23])) //5位 用户id

		//根据接收数据的前面一部分的内容，来区分执行相关操作

		//(i) 初始化连接，接收last_id(初始化为11111111)、room_key、用户id
		if string(data[0]) == "i" {
			num, _ := strconv.Atoi(StringByte.Bytes2String(data[1:9]))
			lastId <- int64(num)
			//记录当前连接用户的room_key和用户id(预计在客户端初始化ws连接、发送第一条数据的时候执行)
			if room_key == "" && user_id == 0 {
				room_key = StringByte.Bytes2String(data[9:19])
				id, _ := strconv.Atoi(StringByte.Bytes2String(data[19:24]))
				user_id = int64(id)
			}

		}
		//(ii) 接收Msg
		if string(data[0]) == "{" {

			if err := json.Unmarshal(data, &msg); err != nil {
				log.Println(err)
			}
			//雪花算法生成消息id
			msg.Id = sf.GenID()
			//缓存消息
			if msg.SenderId != 0 && msg.RoomKey != "" && msg.Content != "" {
				msg.SetCache()
			} else {
				continue
			}
			//数据持久化
			err1 := msg.DataPersistence(room_key, user_id)
			if err1 != nil {
				log.Println(err1)
			}
		}
		//(iii) 接收last_id
		if string(data[0]) != "[" && string(data[0]) != "{" {
			num, _ := strconv.Atoi(StringByte.Bytes2String(data))
			lastId <- int64(num)
		}

	}

}

// 向客户端发送信息
func sendMsg(conn *websocket.Conn, room_key string, user_id int64, last_id int64) error {

	rc := dbConn.NewRedisPool().Get()
	defer rc.Close()

	//在redis中存储last_id的作用是：让后端服务获知客户端接收到的最后一条消息的id，以确定是否向客户端发送最新的消息
	//先判断last_id是否已设置
	old_id, _ := redis.Int64(rc.Do("GET", MsgStampTable+strconv.Itoa(int(user_id))+":"+room_key))
	// fmt.Println("old_id: ", old_id, "last_id: ", last_id)
	//未设置
	if old_id != last_id { //初始化连接的时候，redis中没有last_id,即old_id为0
		//字符串存储 key: userId + RoomKey, value: last_id
		rc.Do("SET", MsgStampTable+strconv.Itoa(int(user_id))+":"+room_key, last_id)
	}

	//（1）先尝试从redis中提取数据
	msg := &models.Msg{}
	// msgList, err := msg.GetCache(room_key, last_id)
	// if err != nil {
	// 	return err
	// }
	// isCached, _ := redis.Bool(rc.Do("EXISTS", models.RoomTable+room_key))
	// if isCached == true {
	// 	if msgList != nil {
	// 		//消息列表不为空，则返回消息列表
	// 		if err := conn.WriteJSON(&msgList); err != nil {
	// 			return err
	// 		}
	// 		// fmt.Println("msgList cache hit")
	// 	}
	// 	return nil
	// }
	msgList := []*models.Msg{}
	err := errors.New("error")
	isCached, _ := redis.Bool(rc.Do("EXISTS", models.RoomTable+room_key))
	//缓存的key存在
	if isCached == true {
		//获取Mysql中的有效数据条数
		persistence_msg_num, err := msg.CountNum(room_key, time.Now())
		if err != nil {
			log.Println(err)
		}
		//获取redis中的有效数据条数
		cache_msg_num, _ := redis.Int(rc.Do("ZCARD", models.RoomTable+room_key))
		//测试
		// fmt.Println("persistence_msg_num", persistence_msg_num, "cache_msg_num", cache_msg_num)
		//两处的有效数据条数误差不大
		//（考虑到客户端轮询服务端数据的间隔较短，客户端发送消息后，服务端还没完成数据写入，就已接收了下一次轮询请求）
		if (int64(cache_msg_num) - persistence_msg_num) <= 1 {
			msgList, err = msg.GetCache(room_key, last_id)
			if err != nil {
				return err
			}
			if msgList != nil && len(msgList) != 0 {
				//发送给websocket客户端
				if err := conn.WriteJSON(&msgList); err != nil {
					return err
				}
			}
			// fmt.Println("msgList cache hit")
			return nil
		} else {
			//两处的有效数据条数误差较大，则清空缓存
			rc.Do("DEL", models.RoomTable+room_key)
		}
	}

	fmt.Println("msgList cache missing")

	//（2）再尝试读取Mysql的数据
	//全量提取数据
	msgList, err = msg.GetBatch(room_key, time.Now(), 0)
	if err != nil {
		return err
	}
	//全量缓存
	msg.SetBatchCache(msgList, time.Now().Add(-7*24*time.Hour))
	//发送给websocket客户端
	if err := conn.WriteJSON(&msgList); err != nil {
		return err
	}
	// //缓存近24个小时的消息
	// msg.SetBatchCache(msgList, time.Now().Add(-24*time.Hour))
	// if err := conn.WriteJSON(&msgList); err != nil {
	// 	return err
	// }

	//若Mysql中也没有消息，则直接返回
	if msgList == nil {
		return fmt.Errorf("no message")
	}

	//其余情况
	return nil
}

// 关闭连接前删除redis中的last_id
func removeLastId(user_id int64, room_key string) {
	rc := dbConn.NewRedisPool().Get()
	defer rc.Close()
	status := false
	var err error
	for !status {
		status, err = redis.Bool(rc.Do("DEL", MsgStampTable+strconv.Itoa(int(user_id))+":"+room_key))
		if err != nil {
			log.Println(err)
		}
	}
	fmt.Println("last_id deleted ok")
}
