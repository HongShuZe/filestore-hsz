package route

import (
	"filestore-hsz/handler"
	"github.com/gin-gonic/gin"
)

// 路由表配置
func Router() *gin.Engine {
	router := gin.Default()

	// 静态资源处理
	router.Static("/static/", "./static")

	// 不需要token验证就能访问的接口
	router.GET("/user/signup", handler.SignupHandler)
	router.POST("/user/signup", handler.DoSignupHandler)

	router.GET("/user/signin", handler.SignInHandler)
	router.POST("/user/signin", handler.DoSignInHandler)

	// 加入中间件, 用于校验token的拦截器
	router.Use(handler.HTTPInterceptor())

	// Use之后的所有handler都会经过拦截器
	// 文件存取接口
	router.GET("/file/upload", handler.UploadHandler)
	router.POST("/file/upload", handler.DoUploadHandler)
	router.GET("/file/upload/suc", handler.UploadSucHandler)
	router.GET("/file/meta", handler.GetFileMetaHandler)
	router.POST("/file/query", handler.FileQueryHandler)
	router.GET("/file/download", handler.DownloadHandler)
	router.GET("/file/download/range", handler.RangeDownloadHandler)
	router.POST("/file/update", handler.FileMetaUpdateHandler)
	router.POST("/file/delete", handler.FileDeleteHandler)
	router.POST("/file/downloadurl", handler.DownloadURLHandler)

	// 秒传接口
	router.POST("/file/fastupload", handler.TryFastUploadHandler)

	// 分块上传接口
	router.POST("/file/mpupload/init", handler.InitialMultipartUploadHandler)
	router.POST("/file/mpupload/uppart", handler.UploadPartHandler)
	router.POST("/file/mpupload/complete", handler.CompleteUploadHandler)

	router.POST("/user/info", handler.UserInfoHandler)

	return router
}