package handler

import (
	"net/http"
	"filestore-hsz/util"

	"github.com/gin-gonic/gin"
	"filestore-hsz/common"
	userProto "filestore-hsz/service/account/proto"
	dlProto "filestore-hsz/service/download/proto"
	upProto "filestore-hsz/service/upload/proto"
	"github.com/micro/go-micro"
	"context"
	"log"
	cmn "filestore-hsz/common"
	"filestore-hsz/config"
	cfg"filestore-hsz/config"
	ratelimit2 "github.com/juju/ratelimit"
	"github.com/micro/go-plugins/wrapper/breaker/hystrix"
	"github.com/micro/go-plugins/wrapper/ratelimiter/ratelimit"
	_ "github.com/micro/go-plugins/registry/consul"
)

var (
	userCli userProto.UserService
	upCli   upProto.UploadService
	dlCli   dlProto.DownloadService
)

func init() {
	// 配置请求容量
	bRate := ratelimit2.NewBucketWithRate(100, 1000)
	service := micro.NewService(
		micro.Registry(config.RegistryConsul()),
		micro.Flags(common.CustomFlags...),
		micro.WrapClient(ratelimit.NewClientWrapper(bRate, false)), //加入限流功能, false为不等待(超限即返回请求失败)
		micro.WrapClient(hystrix.NewClientWrapper()), // 加入熔断功能, 处理rpc调用失败的情况(cirucuit breaker)
	)
	// 初始化, 解析命令行参数等
	service.Init()
	cli := service.Client()
	// 初始化一个account服务的客户端
	userCli = userProto.NewUserService("go.micro.service.user", cli)
	// 初始化一个upload服务的客户端
	upCli = upProto.NewUploadService("go.micro.service.upload", cli)
	// 初始化一个download服务的客户端
	dlCli = dlProto.NewDownloadService("go.micro.service.download", cli)
}

// 响应注册页面
func SignupHandler(c *gin.Context) {
	// GET请求重定向到注册页面
	c.Redirect(http.StatusFound, "/static/view/signup.html")
}

// 处理注册post请求
func DoSignupHandler(c *gin.Context) {

	username := c.Request.FormValue("username")
	passwd := c.Request.FormValue("password")

	resp, err := userCli.Signup(context.TODO(), &userProto.ReqSignup{
		Username: username,
		Password: passwd,
	}, cfg.RpcOpts)

	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  resp.Code,
		"msg": resp.Message,
	})
}

// 响应登录页面
func SigninHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signin.html")
}

// 处理登录post请求
func DoSigninHandler(c *gin.Context) {

	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")

	rpcResp, err := userCli.Signin(context.TODO(), &userProto.ReqSignin{
		Username: username,
		Password: password,
	}, cfg.RpcOpts)
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if rpcResp.Code != cmn.StatusOK {
		c.JSON(200, gin.H{
			"msg":  "登录失败",
			"code": rpcResp.Code,
		})
		return
	}

	// 动态获取上传入口地址
	upEntryResp, err := upCli.UploadEntry(context.TODO(), &upProto.ReqEntry{})
	if err != nil {
		log.Println(err.Error())
	} else if upEntryResp.Code != cmn.StatusOK {
		log.Println(upEntryResp.Message)
	}

	// 动态获取上下载口地址
	dlEntryResp, err := dlCli.DownloadEntry(context.TODO(), &dlProto.ReqEntry{})
	if err != nil {
		log.Println(err.Error())
	} else if dlEntryResp.Code != cmn.StatusOK {
		log.Println(dlEntryResp.Message)
	}


	// 3.登录成功重定向到首页
	resp := util.RespMsg{
		Code: int(common.StatusOK),
		Msg:  "登录成功",
		Data: struct {
			Location string
			Username string
			Token    string
			UploadEntry string
			DownloadEntry string
		}{
			Location: "/static/view/home.html",
			Username: username,
			Token: rpcResp.Token,
			UploadEntry: upEntryResp.Entry,
			DownloadEntry: dlEntryResp.Entry,
		},
	}

	c.Data(http.StatusOK, "application/json", resp.JSONBytes())
}
// octet-stream

// 查询用户信息
func UserInfoHandler(c *gin.Context)  {
	// 1.解析请求参数
	username := c.Request.FormValue("username")

	// 2.查询用户信息
	resp, err := userCli.UserInfo(context.TODO(), &userProto.ReqUserInfo{
		Username: username,
	}, cfg.RpcOpts)
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	// 3.组装并响应用户数据
	cliResp := util.RespMsg{
		Code: 0,
		Msg: "OK",
		Data: gin.H{
			"Username": username,
			"SignupAt": resp.SignupAt,
			"Phone": resp.Phone,
			"Email": resp.Email,
			"LastActive": resp.LastActiveAt,
		},
	}
	c.Data(http.StatusOK, "application/json", cliResp.JSONBytes())
}
