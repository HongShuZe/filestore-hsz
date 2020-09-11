package main

import (
	"filestore-hsz/service/upload/route"
	cfg "filestore-hsz/service/upload/config"
	"github.com/micro/go-micro"
	"time"
	upProto "filestore-hsz/service/upload/proto"
	upRpc "filestore-hsz/service/upload/rpc"
	"fmt"
	"filestore-hsz/common"
	"github.com/micro/cli"
	"log"
	"filestore-hsz/mq"
	dbproxy "filestore-hsz/service/dbproxy/client"
	"filestore-hsz/config"
)

func startRPCService()  {
	service := micro.NewService(
		micro.Name("go.micro.service.upload"), 	// 在注册中心中的服务名称
		micro.RegisterTTL(time.Second * 10), 		// 声明超时时间, 避免consul不主动删除掉已失去心跳的服务节点
		micro.RegisterInterval(time.Second * 5),	// 让服务在指定时间内重新注册, 保持TTL获取的注册时间有效
		micro.Registry(config.RegistryConsul()), 	// micro(v1.18) 需要显示指定consul
		micro.Flags(common.CustomFlags...),
	)
	service.Init(
		micro.Action(func(c *cli.Context) {
			// 检查是否指定mqhost
			mqhost := c.String("mqhost")
			if len(mqhost) > 0 {
				log.Println("custom mq address: " + mqhost)
				mq.UpdateRabbitHost(mqhost)
			}
		}),
	)

	dbproxy.Init(service)
	mq.Init()

	upProto.RegisterUploadServiceHandler(service.Server(), new(upRpc.Upload))
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

func startAPIService()  {
	router := route.Router()
	router.Run(cfg.UploadServiceHost)
}

func main()  {
	// api 服务
	go startAPIService()
	// rpc 服务
	startRPCService()
}