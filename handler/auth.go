package handler

import (
	"net/http"
	"filestore-hsz/util"
	"filestore-hsz/common"
	"github.com/gin-gonic/gin"
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
			return
		}
		c.Next()
	}
}
