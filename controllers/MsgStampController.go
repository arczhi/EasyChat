package controllers

// import (
// 	"EasyChat/dbConn"
// 	"encoding/binary"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"strconv"

// 	"github.com/gomodule/redigo/redis"
// )

// const (
// 	MsgStampTable  = "EasyChat:MsgStamp:"
// 	UserIdExample  = 666
// 	RoomKeyExample = "SGA62adgS6"
// )

// var redisPool = dbConn.NewRedisPool()

// // 接收并记录客户端传回的last_id（上一条接收到的消息的id）
// func MsgStamp(w http.ResponseWriter, r *http.Request) {
// 	data, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	defer r.Body.Close()

// 	last_id := int64(binary.BigEndian.Uint64(data))
// 	fmt.Println(last_id)

// 	//将last_id存入redis
// 	rc := redisPool.Get()
// 	defer rc.Close()

// 	//字符串存储 key: userId + RoomKey, value: last_id
// 	status, err := redis.String(rc.Do("SET", MsgStampTable+":"+strconv.Itoa(UserIdExample)+":"+RoomKeyExample, last_id))
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	if status != "OK" {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		fmt.Fprintf(w, "db error")
// 		return
// 	}

// 	fmt.Fprintf(w, "OK")

// }
