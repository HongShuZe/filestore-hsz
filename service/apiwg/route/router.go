package route

import (
	"filestore-hsz/service/apiwg/handler"
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/elazarl/go-bindata-assetfs"
	"filestore-hsz/assets"
	"strings"
	"github.com/gin-contrib/cors"
)

type binaryFileSystem struct {
	fs http.FileSystem
}

func (b *binaryFileSystem) Open(name string) (http.File, error) {
	return b.fs.Open(name)
}

func (b *binaryFileSystem) Exists(prefix string, filepath string) bool {

	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		if _, err := b.fs.Open(p); err != nil {
			return false
		}
		return true
	}
	return false
}

func BinaryFileSystem(root string) *binaryFileSystem {
	fs := &assetfs.AssetFS{
		Asset:     assets.Asset,
		AssetDir:  assets.AssetDir,
		//AssetInfo: assets.AssetInfo,
		Prefix:    root,
	}
	return &binaryFileSystem{
		fs,
	}
}


// 网关api
func Router() *gin.Engine {
	router := gin.Default()

	// 静态资源处理
	router.Static("/static/", "./static")
	//router.Use(static.Serve("/static/", BinaryFileSystem("static")))

	// 注册
	router.GET("/user/signup", handler.SignupHandler)
	router.POST("/user/signup", handler.DoSignupHandler)
	// 登录
	router.GET("/user/signin", handler.SigninHandler)
	router.POST("/user/signin", handler.DoSigninHandler)

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // []string{"http://localhost:8080"}
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Range", "x-requested-with", "content-Type"},
		ExposeHeaders: []string{"Content-Length", "Accept-Range", "Content-Range", "Content-Disposition"},
		//AllowCredentials: true,
	}))

	// token验证中间件
	router.Use(handler.HTTPInterceptor())

	// 用户查询
	router.GET("/user/info", handler.UserInfoHandler)
	// 用户文件查询
	router.POST("/file/query", handler.FileQueryHandler)
	// 用户文件修改(重命名)
	router.POST("/file/update", handler.FileMetaUpdateHandler)
	// 用户文件删除
	router.POST("/file/delete", handler.FileDeleteHandler)

	return router
}

