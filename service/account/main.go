package main

import (
	"time"
	"log"
	"github.com/micro/go-micro"
	proto "filestore-hsz/service/account/proto"

	"filestore-hsz/service/account/handler"
)

func main()  {
	// 创建一个service
	service := micro.NewService(
		micro.Name("go.micro.service.user"),
		micro.RegisterTTL(time.Second * 10),
		micro.RegisterInterval(time.Second * 5),
	)

	service.Init()

	proto.RegisterUserServiceHandler(service.Server(), new(handler.User))
	if err := service.Run(); err != nil {
		log.Println(err)
	}
}
