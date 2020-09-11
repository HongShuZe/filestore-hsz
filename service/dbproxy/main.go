package main

import (
	"github.com/micro/go-micro"
	"time"
	dbProxy "filestore-hsz/service/dbproxy/proto"
	dbRpc "filestore-hsz/service/dbproxy/rpc"
	"log"
	"filestore-hsz/common"
	"github.com/micro/cli"
	c1 "filestore-hsz/config"
	"filestore-hsz/service/dbproxy/config"
)

func startRPCService()  {
	service := micro.NewService(
		micro.Name("go.micro.service.dbproxy"), // 在注册中心中的服务名称
		micro.RegisterTTL(time.Second * 10), // 声明超时时间, 避免consul不主动删除掉已失去心跳的服务节点
		micro.RegisterInterval(time.Second * 5),
		micro.Registry(c1.RegistryConsul()), // micro(v1.18) 需要显示指定consul
		micro.Flags(common.CustomFlags...),
	)
	service.Init(
		micro.Action(func(c *cli.Context) {
			// 检查是否指定dbhost
			dbhost := c.String("dbhost")
			if len(dbhost) > 0 {
				log.Println("custom db address: " + dbhost)
				config.UpdataDBHost(dbhost)
			}
		}),
	)

	dbProxy.RegisterDBProxyServiceHandler(service.Server(), new(dbRpc.DBProxy))
	if err := service.Run(); err != nil {
		log.Println(err)
	}
}

func main()  {
	startRPCService()
}