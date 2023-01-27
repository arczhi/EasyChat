package controllers

import (
	"EasyChat/config"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var (
	path               = "./sessions"
	KeyForSecureCookie = []byte("12345678abcdefgh87654321hgfedcba") //32位
)

var Store *sessions.FilesystemStore

func init() {
	//不存在该文件夹，则创建
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		panic(err)
	}
	// Store = sessions.NewFilesystemStore(path, securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32))
	// 使用随机的key会导致重启server之后，无法识别原有的cookie
	Store = sessions.NewFilesystemStore(path, KeyForSecureCookie, KeyForSecureCookie)
	if Store == nil {
		panic("Store is nil")
	}
	//设置session（存储了sessionID的cookie的）过期时间 单位为秒
	Store.MaxAge(604800) //7天
	//跨域存储session需要更改以下cookie设置
	//关闭http_only
	//Store.Options.HttpOnly = false
	//设置域
	Store.Options.Domain = config.Cfg.HOSTNAME //默认为127.0.0.1
	//设置same-site属性为None 允许跨站点使用
	Store.Options.SameSite = http.SameSiteNoneMode
	//同时开启secure选项
	Store.Options.Secure = true
}
