package handler


import (
	"net/http"
	"filestore-hsz/util"

	dblayer "filestore-hsz/db"
	"fmt"
	"time"
	"github.com/gin-gonic/gin"
	"filestore-hsz/common"
)

const (
	// 用于加密的盐值
	pwdSalt = "#890"
)

// 响应注册页面
func SignupHandler(c *gin.Context) {
	// GET请求重定向到注册页面
	c.Redirect(http.StatusFound, "/static/view/signup.html")
}

// 处理注册post请求
func DoSignupHandler(c *gin.Context) {

	username := c.Request.FormValue("username")
	passwd := c.Request.FormValue("password")

	if len(username) < 3 || len(passwd) < 5 {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "请求参数无效",
			"code": common.StatusParamInvalid,
		})
		return
	}

	// 对密码进行加盐及sha1值加密
	encPasswd := util.Sha1([]byte(passwd + pwdSalt))
	// 将用户信息注册到用户表中
	suc := dblayer.UserSignup(username, encPasswd)
	if suc {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "注册成功",
			"code": common.StatusOK,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "注册失败",
			"code": common.StatusRegisterFailed,
		})
	}
}

// 响应登录页面
func SignInHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signin.html")
}

// 处理登录post请求
func DoSignInHandler(c *gin.Context) {

	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")
	encPasswd := util.Sha1([]byte(password + pwdSalt))

	// 1.校验用户名及密码
	pwdChecked := dblayer.UserSignin(username, encPasswd)
	if !pwdChecked {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "登录失败",
			"code": common.StatusLoginFailed,
		})
		return
	}
	// 2.生成访问凭证
	token := GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "登录失败",
			"code": common.StatusLoginFailed,
		})
		return
	}

	// 3.登录成功重定向到首页
	resp := util.RespMsg{
		Code: int(common.StatusOK),
		Msg:  "登录成功",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "/static/view/home.html",
			Username: username,
			Token: token,
		},
	}

	c.Data(http.StatusOK, "octet-stream", resp.JSONBytes())
	//
}

// 查询用户信息
func UserInfoHandler(c *gin.Context)  {
	// 1.解析请求参数

	username := c.Request.FormValue("username")

	// 2.查询用户信息
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	// 3.组装并响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg: "OK",
		Data: user,
	}
	c.Data(http.StatusOK, "application/json", resp.JSONBytes())
	//c.JSON(http.StatusOK, resp)
}

// 生成token
func GenToken(username string) string {
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

// token是否有效
func IsTokenValid(token, username string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO 判断token的时效性， 是否过期
	// 假设token的有效期为1天
	tokenTS := token[:8]
	if util.Hex2Dec(tokenTS) < time.Now().Unix()-86400 {
		return false
	}

	// TODO 从数据库表tbl_user_token查询username对应的token信息
	// TODO 对比两个token是否一致
	if dblayer.GetUserToken(username) != token {
		return false
	}

	return true
}