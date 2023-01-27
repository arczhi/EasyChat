package easy

import (
	"log"
	"net/http"
)

/*
start http server
*/
func (r *router) Start(addr string) {
	log.Println("[easy] listening ", addr)
	signal := make(chan error)
	go func() {
		if err := http.ListenAndServe(addr, r); err != nil {
			signal <- err
		}
	}()
	log.Println(<-signal)
}
