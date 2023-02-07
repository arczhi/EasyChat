package main

import (
	"EasyChat/controllers"
	"EasyChat/easy"
	"log"
	"net/http"
)

func main() {

	// _, err := os.OpenFile("./statics/index.html", os.O_RDWR|os.O_CREATE, 0755)
	// if err != nil {
	// 	panic(err)
	// }
	//

	//前端
	go func() {
		log.Println("frontend listening  :9999")
		err := http.ListenAndServe(":9999", http.FileServer(http.Dir("./statics")))
		if err != nil {
			log.Println(err)
		}
	}()

	//后端
	e := easy.New()

	e.All("/Msg", controllers.MsgExchange)
	e.POST("/Login", controllers.Login)
	e.POST("/CheckUsername", controllers.CheckUsername)
	e.POST("/Register", controllers.Register)
	e.POST("/ChatRoom/New", controllers.NewChatRoom)
	e.POST("/ChatRoom/Entry", controllers.EntryChatRoom)
	// e.POST("/OppoUsername", controllers.GetOppoUsername)
	e.POST("/GetUsername", controllers.GetUsername)
	e.POST("/RoomMemberNum", controllers.RoomMemberNum)

	e.Start(":8283")
}
