package route

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	 a "filestore-hsz/service/upload/api"
)


func Router() *gin.Engine {
	router := gin.Default()

	// 静态资源处理
	router.Static("/static/", "./static")

	// 使用gin插件支持扩域请求
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // []string{"http://localhost:8080"}
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Range", "x-requested-with", "content-Type"},
		ExposeHeaders: []string{"Content-Length", "Accept-Range", "Content-Range", "Content-Disposition"},
		// AllowCredentials: true,
	}))

	// 文件存取接口
	router.POST("/file/upload", a.DoUploadHandler)

	// 秒传接口
	router.POST("/file/fastupload", a.TryFastUploadHandler)

	// 分块上传接口
	router.POST("/file/mpupload/init", a.InitialMultipartUploadHandler)
	router.POST("/file/mpupload/uppart", a.UploadPartHandler)
	router.POST("/file/mpupload/complete", a.CompleteUploadHandler)

	return router
}

