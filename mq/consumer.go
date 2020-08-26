package mq

import (
	"log"
	cfg "filestore-hsz/config"
	"fmt"
)
var done chan bool

// 接收消息
func StartConsume(qName, cName string, callback func(msg []byte) bool) {
	// 获取消息信道
	msgs, err := channel.Consume(
		qName,
		cName,
		true, // 自动答应
		false, // 非唯一的消费者
		false, // rabbitMQ只能设置为false
		false, // noWait, false表示会阻塞直到有消息过来
		nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	done = make(chan bool)

	go func() {
		// 循环读取channel的数据
		for d := range msgs {
			// 调用callback方法来处理新的消息
			processErr := callback(d.Body)
			if processErr {
				// TODO: 将任务写入错误队列, 待后续处理
				fmt.Println("发生错误, 将任务写入错误队列")
				Publish(
					cfg.TransExchangeName,
					cfg.TransOSSErrRoutingKey,
					d.Body)
			}
		}
	}()
	// 接收done信号, 没有信息过来则会一直阻塞, 避免该函数退出
	<- done
	// 关闭通道
	channel.Close()
}

// 停止监听队列
func StopConsume()  {
	done <- true
}
