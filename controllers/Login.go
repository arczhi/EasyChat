package controllers

import (
	"EasyChat/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func Login(w http.ResponseWriter, r *http.Request) {
	//读取请求体
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "reading error")
		return
	}
	defer r.Body.Close()

	//反序列化
	var pack struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err1 := json.Unmarshal(buf, &pack)
	if err1 != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "binding error")
		return
	}

	//校验密码
	user := &models.User{}
	err2 := user.Check(pack.Username)
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "db error")
		return
	}
	if pack.Password != user.Password {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "login error")
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

	// fmt.Println(session.Values["id"].(int64), session.Values["username"].(string))

	//登录成功
	user.Password = "?"
	res, err := json.Marshal(&user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error")
		return
	}
	w.Write(res)

}
