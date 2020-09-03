package route

import (
	"filestore-hsz/service/apiwg/handler"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"filestore-hsz/middleware"
)

// 网关api
func Router() *gin.Engine {
	router := gin.Default()

	// 静态资源处理
	router.Static("/static/", "./static")

	// 注册
	router.GET("/user/signup", handler.SignupHandler)
	router.POST("/user/signup", handler.DoSignupHandler)
	// 登录
	router.GET("/user/signin", handler.SigninHandler)
	router.POST("/user/signin", handler.DoSigninHandler)

	// 使用gin插件支持扩域请求
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"}, // []string{"http://localhost:8080"},
		AllowMethods:  []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Range", "x-requested-with", "content-Type"},
		ExposeHeaders: []string{"Content-Length", "Accept-Ranges", "Content-Range", "Content-Disposition"},
		// AllowCredentials: true,
	}))

	router.Use(middleware.HTTPInterceptor())

	// 用户查询
	router.POST("/user/info", handler.UserInfoHandler)
	// 用户文件查询
	router.POST("/file/query", handler.FileQueryHandler)
	// 用户文件修改(重命名)
	router.POST("/file/update", handler.FileMetaUpdateHandler)

	return router
}

