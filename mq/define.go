package mq

import (
	cmn "filestore-hsz/common"
)

// 将要写到rabbitmq的数据的结构体
type TransferData struct {
	FileHash string
	CurLocation string
	DestLocation string
	DestStoreType cmn.StoreType
}
