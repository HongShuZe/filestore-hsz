package handler

import (
	"net/http"
	"os"
	dblayer "filestore-hsz/db"
	"filestore-hsz/meta"
	"github.com/gin-gonic/gin"
)

// 支持断点的文件下载接口
func RangeDownloadHandler(c *gin.Context) {

	fsha1 := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")

	fm, _ := meta.GetFileMetaDB(fsha1)
	userFile, err := dblayer.QueryUserFileMeta(username, fsha1)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	f, err := os.Open(fm.Location)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	c.Header("Content-Type", "application/octect-stream")
	// attachment表示文件将会提示下载到本地，而不是直接在浏览器中打开
	c.Header("content-disposition", "attachment; filename=\""+userFile.FileName+"\"")
	//http.ServeFile(w, r, fm.Location)
	c.File(fm.Location)
}


