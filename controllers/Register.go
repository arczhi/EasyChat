package controllers

import (
	"EasyChat/models"
	"EasyChat/utils/AES"
	"EasyChat/utils/StringByte"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func Register(w http.ResponseWriter, r *http.Request) {

	//读取请求体
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "reading error")
		return
	}
	defer r.Body.Close()

	//反序列化
	user := &models.User{}
	err1 := json.Unmarshal(buf, &user)
	if err1 != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "binding error")
		return
	}

	//加密密码
	encPass, err := AES.AesEncrypt(StringByte.String2Bytes(user.Password), StringByte.String2Bytes(AES.Key))
	if err != nil {
		// fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "calculate error")
		return
	}
	user.Password = encPass

	//操作数据库
	err2 := user.Create()
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "db error")
		return
	}

	//创建session
	session, err := Store.Get(r, "user")
	if err != nil {
		log.Println(err, user.Id)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "auth error")
		return
	}
	session.Values["id"] = user.Id
	session.Values["username"] = user.Username
	err3 := session.Save(r, w)
	if err3 != nil {
		log.Println(err3, user.Id)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "auth create error")
		return
	}

	user.Password = "?"
	res, err := json.Marshal(&user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error")
		return
	}
	w.Write(res)

}

// 检查用户名是否重复
func CheckUsername(w http.ResponseWriter, r *http.Request) {
	//读取请求体
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "reading error")
		return
	}
	defer r.Body.Close()

	//反序列化
	user := &models.User{}
	err1 := json.Unmarshal(buf, &user)
	if err1 != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "binding error")
		return
	}

	//查询数据库
	exist, err := user.CheckUsername()
	if exist {
		fmt.Fprintf(w, "existed")
		return
	}

	fmt.Fprintf(w, "not existed")
	return
}
