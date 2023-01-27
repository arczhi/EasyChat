package controllers

import (
	"EasyChat/models"
	"EasyChat/utils/StringByte"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// 根据前端传递的用户id,返回对应的用户名
func GetUsername(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "binding error")
		return
	}
	defer r.Body.Close()

	// id := int64(binary.BigEndian.Uint64(buf))

	id1, _ := strconv.Atoi(StringByte.Bytes2String(buf))
	id := int64(id1)
	//测试
	// fmt.Println(id)

	//根据对方的id查询用户名
	user := &models.User{}
	username, err := user.GetUsernameById(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "db error")
		return
	}

	fmt.Fprintf(w, username)

}

// // 获取通信对方的用户名(仅用于双人对话)
// func GetOppoUsername(w http.ResponseWriter, r *http.Request) {
// 	buf, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		fmt.Fprintf(w, "binding error")
// 		return
// 	}
// 	defer r.Body.Close()

// 	var pack struct {
// 		EntryRoomKey string `json:"entry_room_key"`
// 		OwnId        int64  `json:"own_id"`
// 	}

// 	err1 := json.Unmarshal(buf, &pack)
// 	if err1 != nil || pack.EntryRoomKey == "" || pack.OwnId == 0 {
// 		log.Println(err1)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		fmt.Fprintf(w, "binding error")
// 		return
// 	}

// 	//测试
// 	fmt.Println(pack.EntryRoomKey, pack.OwnId)

// 	//获取该room_key下所有成员的id
// 	chatRoom := models.ChatRoom{}
// 	ids, err := chatRoom.GetMembers(pack.EntryRoomKey)
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		fmt.Fprintf(w, "db error")
// 		return
// 	}

// 	//打印测试
// 	fmt.Println(ids)

// 	//筛选出沟通对方的id
// 	var oppo_id int64
// 	for _, v := range ids {
// 		if v != pack.OwnId {
// 			oppo_id = v
// 			break
// 		}
// 	}

// 	if oppo_id == 0 {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		fmt.Fprintf(w, "db error")
// 		return
// 	}

// 	//根据对方的id查询用户名
// 	user := &models.User{}
// 	username, err := user.GetUsernameById(oppo_id)
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		fmt.Fprintf(w, "db error")
// 		return
// 	}

// 	fmt.Fprintf(w, username)
// }
