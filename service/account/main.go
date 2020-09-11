package main

import (
	"time"
	"log"
	"github.com/micro/go-micro"
	proto "filestore-hsz/service/account/proto"

	"filestore-hsz/service/account/handler"
	"filestore-hsz/common"
	dbproxy "filestore-hsz/service/dbproxy/client"
)

func main()  {
	// 创建一个service
	service := micro.NewService(
		micro.Name("go.micro.service.user"),
		micro.RegisterTTL(time.Second * 10),
		micro.RegisterInterval(time.Second * 5),
		//micro.Registry(config.RegistryConsul()),
		micro.Flags(common.CustomFlags...),
	)
	// 初始化命令行参数, 解析命令行参数
	service.Init()

	// 初始化dbproxy client
	dbproxy.Init(service)

	proto.RegisterUserServiceHandler(service.Server(), new(handler.User))
	if err := service.Run(); err != nil {
		log.Println(err)
	}
}
