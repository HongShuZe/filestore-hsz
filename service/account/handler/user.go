package handler

import (
	"context"
	"fmt"
	"time"
	"filestore-hsz/util"
	proto "filestore-hsz/service/account/proto"
	"filestore-hsz/common"
	cfg "filestore-hsz/config"
	dbCli "filestore-hsz/service/dbproxy/client"
	"log"
)

// User用于实现UserServiceHandler接口的对象
type User struct{}

// 生成token
func GenToken(username string) string {
	// 40位字符
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

// 用户注册
func (u *User) Signup(ctx context.Context, req *proto.ReqSignup, res *proto.RespSignup) error {
	username := req.Username
	passwd := req.Password

	if len(username) < 3 || len(passwd) < 5 {
		res.Code = common.StatusParamInvalid
		res.Message = "请求参数无效"
		return nil
	}

	// 对密码进行加盐及sha1值加密
	encPasswd := util.Sha1([]byte(passwd + cfg.PasswordSalt))
	// 将用户信息注册到用户表中
	dbResp, err := dbCli.UserSignup(username, encPasswd)
	if err == nil && dbResp.Suc {
		res.Code = common.StatusOK
		res.Message = "注册成功"
	} else {
		res.Code = common.StatusRegisterFailed
		res.Message = "注册失败"
	}
	return nil
}

// 用户登录
func (u *User) Signin(ctx context.Context, req *proto.ReqSignin, res *proto.RespSignin) error {
	username := req.Username
	password := req.Password
	encPasswd := util.Sha1([]byte(password + cfg.PasswordSalt))

	log.Println("username:"+ username)
	log.Println("passwd:"+encPasswd)

	// 1.校验用户名及密码
	dbResp, err := dbCli.UserSignin(username, encPasswd)
	if err != nil || !dbResp.Suc {
		res.Code = common.StatusLoginFailed
		res.Message = "登录失败"
		return nil
	}
	// 2.生成访问凭证
	token := GenToken(username)
	upRes, err := dbCli.UpdateToken(username, token)
	if err != nil || !upRes.Suc {
		res.Code = common.StatusServerError
		return nil
	}

	// 3.登录成功重定向到首页
	res.Code = common.StatusOK
	res.Token = token
	return nil
}
// 获取用户信息
func (u *User) UserInfo(ctx context.Context, req *proto.ReqUserInfo, res *proto.RespUserInfo) error {
	// 1.解析请求参数
	username := req.Username
	// 2.查询用户信息
	dbRsep, err := dbCli.GetUserInfo(username)
	if err != nil {
		res.Code = common.StatusServerError
		res.Message = "服务错误"
		return nil
	}

	if !dbRsep.Suc {
		res.Code = common.StatusUserNoExists
		res.Message = "用户不存在"
		return nil
	}

	user := dbCli.ToTableUser(dbRsep.Data)
	// 3.组装并响应用户数据
	res.Code = common.StatusOK
	res.Username = user.Username
	res.SignupAt = user.SignupAt
	res.LastActiveAt = user.LastActiveAt
	res.Status = int32(user.Status)
	// TODO 需增加接口支持完善用户信息(email/phone等)
	res.Email = user.Email
	res.Phone = user.Phone
	return  nil
}