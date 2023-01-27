package easy

import (
	"EasyChat/middlewares"
	"fmt"
	"net/http"
	"strings"
)

/*
http methods which are allowed
*
*/
func (r *router) POST(addr string, handlers ...func(http.ResponseWriter, *http.Request)) {
	r.HandleFunc(addr, func(w http.ResponseWriter, r *http.Request) {
		middlewares.Cors(w, r) //使用Nginx可以直接add_header,不需要在后端服务里设置规则
		if strings.ToUpper(r.Method) == "OPTIONS" {
			return
		}
		if strings.ToUpper(r.Method) != "POST" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "found not found")
			return
		}
		for _, h := range handlers {
			h(w, r)
		}
	})
}

func (r *router) GET(addr string, handlers ...func(http.ResponseWriter, *http.Request)) {
	r.HandleFunc(addr, func(w http.ResponseWriter, r *http.Request) {
		middlewares.Cors(w, r)
		if strings.ToUpper(r.Method) == "OPTIONS" {
			return
		}
		if strings.ToUpper(r.Method) != "GET" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "found not found")
			return
		}
		for _, h := range handlers {
			h(w, r)
		}
	})
}

func (r *router) PUT(addr string, handlers ...func(http.ResponseWriter, *http.Request)) {
	r.HandleFunc(addr, func(w http.ResponseWriter, r *http.Request) {
		middlewares.Cors(w, r) //使用Nginx可以直接add_header,不需要在后端服务里设置规则
		if strings.ToUpper(r.Method) == "OPTIONS" {
			return
		}
		if strings.ToUpper(r.Method) != "PUT" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "found not found")
			return
		}
		for _, h := range handlers {
			h(w, r)
		}
	})
}

func (r *router) DELETE(addr string, handlers ...func(http.ResponseWriter, *http.Request)) {
	r.HandleFunc(addr, func(w http.ResponseWriter, r *http.Request) {
		middlewares.Cors(w, r) //使用Nginx可以直接add_header,不需要在后端服务里设置规则
		if strings.ToUpper(r.Method) == "OPTIONS" {
			return
		}
		if strings.ToUpper(r.Method) != "DELETE" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "found not found")
			return
		}
		for _, h := range handlers {
			h(w, r)
		}
	})
}

/*
accept all requests
*
*/
func (r *router) All(addr string, handlers ...func(http.ResponseWriter, *http.Request)) {
	r.HandleFunc(addr, func(w http.ResponseWriter, r *http.Request) {
		for _, h := range handlers {
			h(w, r)
		}
	})
}
