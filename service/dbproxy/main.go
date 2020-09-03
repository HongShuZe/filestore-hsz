package main

import (
	"github.com/micro/go-micro"
	"time"
	"filestore-hsz/config"
	dbProxy "filestore-hsz/service/dbproxy/proto"
	dbRpc "filestore-hsz/service/dbproxy/rpc"
	log "log"
)

func startRPCService()  {
	service := micro.NewService(
		micro.Name("go.micro.service.dbproxy"), // 在注册中心中的服务名称
		micro.RegisterTTL(time.Second * 10), // 声明超时时间, 避免consul不主动删除掉已失去心跳的服务节点
		micro.RegisterInterval(time.Second * 5),
		micro.Registry(config.RegistryConsul()), // micro(v1.18) 需要显示指定consul
	)
	service.Init()

	dbProxy.RegisterDBProxyServiceHandler(service.Server(), new(dbRpc.DBProxy))
	if err := service.Run(); err != nil {
		log.Println(err)
	}
}

func main()  {
	startRPCService()
}