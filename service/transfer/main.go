package main

import (
	"log"
	"filestore-hsz/mq"
	"encoding/json"
	"os"
	"filestore-hsz/store/oss"
	"bufio"
	dbCli "filestore-hsz/service/dbproxy/client"
	"filestore-hsz/config"
	"github.com/micro/go-micro"
	"time"
	"fmt"
	"filestore-hsz/common"
	"github.com/micro/cli"
	dbproxy "filestore-hsz/service/dbproxy/client"
)

// 文件处理转移
func ProcessTransfer(msg []byte) bool {
	log.Println(string(msg))

	pubData := mq.TransferData{}
	err := json.Unmarshal(msg, &pubData)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// TODO docker部署时文件被锁, open打不开
	//fin, err := os.Open(pubData.CurLocation)
	fin, err := os.OpenFile(pubData.CurLocation,os.O_RDWR, 0777)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	err = oss.Bucket().PutObject(
		pubData.DestLocation,
		bufio.NewReader(fin))
	if err != nil {
		log.Println(err.Error())
		return false
	}

	resp, err := dbCli.UpdateFileLocation(
		pubData.FileHash,
		pubData.DestLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	if !resp.Suc {
		log.Println("更新数据库异常, 请检查:" + pubData.FileHash)
		return false
	}

	return true
}


func startRPCService()  {
	service := micro.NewService(
		micro.Name("go.micro.service.transfer"), 	// 在注册中心中的服务名称
		micro.RegisterTTL(time.Second * 10), 		// TTL指定从上一次心跳间隔起，超过这个时间服务会被服务发现移除
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

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

func startTranserService()  {
	if !config.AsyncTransferEnable {
		log.Println("异步转移文件功能目录禁用, 请检查相关配置")
		return
	}
	log.Println("文件转移服务启动中, 开始监听转移任务队列...")

	mq.Init()
	mq.StartConsume(
		config.TransOSSQueueName,
		"transfer_oss",
		ProcessTransfer)
}

func main()  {
	// 文件转移服务
	go startTranserService()

	// rpc服务
	startRPCService()
}
