package main

import (
	"log"
	"filestore-hsz/mq"
	"encoding/json"
	"os"
	"filestore-hsz/store/oss"
	"bufio"
	dblayer "filestore-hsz/db"
	"filestore-hsz/config"
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

	fin, err := os.Open(pubData.CurLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	err = oss.Bucket().PutObject(
		pubData.DestLocation,
		bufio.NewReader(fin),
	)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	_ = dblayer.UpdateFileLocation(
		pubData.FileHash,
		pubData.DestLocation)

	return true
}

func main()  {
	if !config.AsyncTransferEnable {
		log.Println("异步转移文件功能目录禁用, 请检查相关配置")
		return
	}
	log.Println("文件转移服务启动中, 开始监听转移任务队列...")

	mq.StartConsume(
		config.TransOSSQueueName,
		"transfer_oss",
		ProcessTransfer)
}
