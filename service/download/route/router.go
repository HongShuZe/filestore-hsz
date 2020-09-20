package route

import (
	"github.com/gin-gonic/gin"
	a "filestore-hsz/service/download/api"
	"github.com/gin-contrib/cors"
)


func Router() *gin.Engine {
	router := gin.Default()

	// 静态资源处理
	router.Static("/static/", "./static")

	// 使用gin插件支持扩域请求
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Range", "x-requested-with", "content-Type"},
		ExposeHeaders: []string{"Content-Length", "Accept-Ranges", "Content-Range", "Content-Disposition"},
		// AllowCredentials: true,
	}))
	//router.Use(CORS)

	// token验证中间件
	//router.Use(middleware.HTTPInterceptor())

	// 文件下载接口
	router.GET("/file/download", a.DownloadHandler)
	router.GET("/file/download/range", a.RangeDownloadHandler)
	router.POST("/file/downloadurl", a.DownloadURLHandler)

	return router
}

/*// 允许跨域
func CORS(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "*")

	if c.Request.Method == "OPTIONS" {
		c.String(http.StatusOK, "")
	}
	// 调用下个中间件
	c.Next()
}
*/
