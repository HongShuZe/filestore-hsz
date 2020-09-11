package config

const (
	// 是否开启文件异步转移
	AsyncTransferEnable = true
	// 用于文件transfer的交换机
	TransExchangeName = "uploadserver.trans"
	// oss转移队列名
	TransOSSQueueName = "uploadserver.trans.oss"
	// oss转移失败后写入另一个队列的队列名
	TransOSSErrQueueName = "uploadserver.trans.oss.err"
	// routingKey
	TransOSSRoutingKey = "oss"
	// routingKey
	TransOSSErrRoutingKey = "err"
)

var (
	// rabbitmq服务入口url
	RabbitURL = "amqp://guest:guest@127.0.0.1:5672/"
)


