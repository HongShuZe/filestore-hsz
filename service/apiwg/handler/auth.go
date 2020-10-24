package handler

import (
	"net/http"
	"filestore-hsz/util"
	"filestore-hsz/common"
	"github.com/gin-gonic/gin"
	"time"
	"log"
	dbCli "filestore-hsz/service/dbproxy/client"
	"github.com/micro/go-micro"
	"filestore-hsz/config"
)

// http请求拦截器
func HTTPInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {

		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")

		// 验证登录token是否有效
		if len(username) < 3 || !IsTokenValid(token, username) {
			// token校验失败则跳转到登录界面
			c.Abort()
			resp := util.NewRespMsg(
				int(common.StatusInvalidToken),
				"token无效",
				nil,
			)
			c.JSON(http.StatusOK, resp)
		}
		c.Next()
	}
}

// token是否有效
func IsTokenValid(token string, username string) bool {
	if len(token) != 40 {
		log.Println("token invalid: " + token)
		return false
	}
	// 判断token的时效性， 是否过期
	// 假设token的有效期为1天
	tokenTS := token[32:40]
	if util.Hex2Dec(tokenTS) < time.Now().Unix()-86400 {
		log.Println("token expired: " + token)
		return false
	}

	// 先初始化dbCli
	service := micro.NewService(
		micro.Registry(config.RegistryConsul()),
	)
	service.Init()
	dbCli.Init(service)
	// 从数据库表tbl_user_token查询username对应的token信息
	// 对比两个token是否一致
	dbToken, err := dbCli.GetUserToken(username)
	if err != nil || dbToken != token {
		return false
	}

	return true
}

/*// 生成token
func GenToken(username string) string {
	// 40 位字符
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}*/
