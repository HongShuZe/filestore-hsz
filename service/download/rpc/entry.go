package rpc

import (
	"context"
	cfg "filestore-hsz/service/download/config"
	dlProto "filestore-hsz/service/download/proto"
)

type Download struct {}

func (u *Download) DownloadEntry (
	ctx context.Context,
	req *dlProto.ReqEntry,
	res *dlProto.RespEntry) error {

		res.Entry = cfg.DownloadEntry
		return nil
}



