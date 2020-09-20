package handler

import (
	"context"
	proto "filestore-hsz/service/account/proto"
	dbCli "filestore-hsz/service/dbproxy/client"
	"filestore-hsz/common"
	"encoding/json"
)

// 获取用户文件列表
func (u *User) UserFiles(ctx context.Context, req *proto.ReqUserFile, res *proto.RespUserFile) error {
	dbResp, err := dbCli.QueryUserFileMetas(req.Username, int(req.Limit))
	if err != nil || !dbResp.Suc {
		res.Code = common.StatusServerError
		return err
	}

	userFiles := dbCli.ToTableUserFiles(dbResp.Data)
	data, err := json.Marshal(userFiles)
	if err != nil {
		res.Code = common.StatusServerError
		return nil
	}

	res.FileData = data
	return nil
}


// 用户文件重命名
func (u *User) UserFileRename(ctx context.Context, req *proto.ReqUserFileRename, res *proto.RespUserFileRename) error {
	dbResp, err := dbCli.RenameFileName(req.Username, req.Filehash, req.NewFileName)
	if err != nil || !dbResp.Suc {
		res.Code = common.StatusServerError
		return err
	}

	userFiles := dbCli.ToTableUserFiles(dbResp.Data)
	data, err := json.Marshal(userFiles)
	if err != nil {
		res.Code = common.StatusServerError
		return nil
	}

	res.FileData = data
	return nil
}

// 删除用户文件
func (u *User) UserFileDelete(ctx context.Context, req *proto.ReqUserFileDelete, res *proto.RespUserFileDelete) error {
	dbResp, err := dbCli.DeleteUserFile(req.Username, req.Filehash)
	if err != nil || !dbResp.Suc {
		res.Code = common.StatusServerError
		return err
	}
	// 把dbResp.Data(interface{}) 转换为 []orm.TableUserFile{}
	userFiles := dbCli.ToTableUserFiles(dbResp.Data)
	data, err := json.Marshal(userFiles)
	if err != nil {
		res.Code = common.StatusServerError
		return nil
	}

	res.FileData = data
	return nil
}
