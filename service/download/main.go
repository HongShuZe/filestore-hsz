package main

import (
	"github.com/micro/go-micro"
	"time"
	cfg "filestore-hsz/service/download/config"
	"fmt"
	dlProto "filestore-hsz/service/download/proto"
	dlRpc "filestore-hsz/service/download/rpc"
	"filestore-hsz/service/download/route"
	dbproxy "filestore-hsz/service/dbproxy/client"
	"filestore-hsz/common"
	"filestore-hsz/config"
)


func startRPCService()  {
	service := micro.NewService(
		micro.Name("go.micro.service.download"), 	// 在注册中心中的服务名称
		micro.RegisterTTL(time.Second * 10), 		// TTL指定从上一次心跳间隔起，超过这个时间服务会被服务发现移除
		micro.RegisterInterval(time.Second * 5),	// 让服务在指定时间内重新注册, 保持TTL获取的注册时间有效
		micro.Registry(config.RegistryConsul()), 	// micro(v1.18) 需要显示指定consul
		micro.Flags(common.CustomFlags...),
	)
	service.Init()

	dbproxy.Init(service)

	dlProto.RegisterDownloadServiceHandler(service.Server(), new(dlRpc.Download))
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

func startAPIService()  {
	router := route.Router()
	router.Run(cfg.DownloadServiceHost)
}

func main()  {
	// 文件下载服务
	go startAPIService()

	// rpc服务
	startRPCService()
}
