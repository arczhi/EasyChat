package controllers

import (
	"EasyChat/models"
	"EasyChat/utils/StringByte"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// 创建聊天室
func NewChatRoom(w http.ResponseWriter, r *http.Request) {
	//鉴权
	session, err := Store.Get(r, "user")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "auth error")
		return
	}

	// fmt.Println(session.Name())
	// fmt.Println(session.Values)

	if session.Values["id"] == nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ses error")
		return
	}

	//从session中提取用户id
	chatRoom := &models.ChatRoom{}
	chatRoom.RoomKey = models.NewRoomKey()
	chatRoom.AddMember(chatRoom.RoomKey, session.Values["id"].(int64))

	res, err := json.Marshal(&map[string]string{
		"room_key": chatRoom.RoomKey,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "inner error")
	}
	w.Write(res)

}

// 加入聊天室
func EntryChatRoom(w http.ResponseWriter, r *http.Request) {
	//鉴权
	session, err := Store.Get(r, "user")
	if err != nil {
		//secure cookie这个库在读取session的时候有问题（较频繁）
		//2023/01/21 20:28:38 securecookie: the value is not valid
		//是重启server,导致Store的 key发生了改变，默认使用了随机生成的key，现在需要固定这个key
		//https://github.com/gorilla/sessions/issues/16
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "auth error")
		return
	}

	//接收前端传递的聊天室key
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "binding error")
	}
	defer r.Body.Close()

	// room_key := StringByte.Bytes2String(buf)

	//从session中提取用户id
	chatRoom := &models.ChatRoom{}
	chatRoom.RoomKey = StringByte.Bytes2String(buf)
	//先查询该聊天室是否已存在该成员(MySQL)
	exists, err := chatRoom.CheckMember(chatRoom.RoomKey, session.Values["id"].(int64))
	if !exists {
		//添加成员
		chatRoom.AddMember(chatRoom.RoomKey, session.Values["id"].(int64))
	}

	fmt.Fprintf(w, "ok")
}

// 获取聊天室的成员数量
func RoomMemberNum(w http.ResponseWriter, r *http.Request) {
	//鉴权
	_, err := Store.Get(r, "user")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "auth error")
		return
	}
	// if session.Values["id"] == nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	fmt.Fprintf(w, "ses error")
	// 	return
	// }
	//获取前端传递的当前聊天室的room_key
	//接收前端传递的聊天室key
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "binding error")
	}
	defer r.Body.Close()

	//获取当前聊天室的成员数量
	chatRoom := &models.ChatRoom{}
	chatRoom.RoomKey = StringByte.Bytes2String(buf)
	member_list, err := chatRoom.GetMembers(chatRoom.RoomKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "db error")
		return
	}

	fmt.Fprintf(w, strconv.Itoa(len(member_list)))
}
