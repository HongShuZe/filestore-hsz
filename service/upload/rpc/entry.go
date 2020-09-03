package rpc

import (
	"context"
	upProto "filestore-hsz/service/upload/proto"
	cfg "filestore-hsz/service/upload/config"
)
// upload结构体
type Upload struct {
}

// 获取上传接口
func (u *Upload) UploadEntry (
	ctx context.Context,
	req *upProto.ReqEntry,
	res *upProto.RespEntry) error {

		res.Entry = cfg.UploadEntry
		return nil
}
