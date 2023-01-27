package middlewares

import (
	"EasyChat/config"
	"net/http"
)

var hostname string

func init() {
	hostname = config.Cfg.HOSTNAME
	// fmt.Println("http://" + hostname + ":9999")
}

/*
跨域检查中间件
*/
func Cors(w http.ResponseWriter, r *http.Request) {
	// if !strings.Contains(r.RemoteAddr, "127.0.0.1") {
	// 	w.WriteHeader(http.StatusForbidden)
	// 	fmt.Fprintf(w, "forbidden")
	// 	return
	// }
	// log.Println(r.Header.Get("Origin"))
	w.Header().Add("Access-Control-Allow-Origin", "http://"+hostname+":9999")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type,X-CSRF-Token, Authorization, Token")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE ,OPTIONS")
	w.Header().Add("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
	w.Header().Add("Access-Control-Allow-Credentials", "true") //允许客户端发送cookie
}
