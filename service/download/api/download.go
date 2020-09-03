package handler

import (
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	"fmt"
	"strings"
	"io/ioutil"
	"filestore-hsz/store/oss"
	dbCli "filestore-hsz/service/dbproxy/client"
	"filestore-hsz/common"
	cfg "filestore-hsz/config"
	"log"
	"filestore-hsz/store/ceph"
)


// 生成文件下载地址
func DownloadURLHandler(c *gin.Context) {
	filehash := c.Request.FormValue("filehash")
	// 从文件表查找记录
	dbResp, err := dbCli.GetFileMeta(filehash)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": common.StatusServerError,
			"msg": "server error",
		})
		return
	}

	tblFile := dbCli.ToTableFile(dbResp.Data)

	// 判断文件存在OSS, 还是Ceph, 还是本地
	if strings.HasPrefix(tblFile.FileAddr.String, cfg.MergeLocalRootDir) ||
		strings.HasPrefix(tblFile.FileAddr.String, cfg.CephRootDir) {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		tmpURL := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			c.Request.Host, filehash, username, token)
		c.Data(http.StatusOK, "application/octet-stream", []byte(tmpURL))
	} else if strings.HasPrefix(tblFile.FileAddr.String, cfg.OSSRootDir) {
		signedURL := oss.DownloadURL(tblFile.FileAddr.String)
		log.Println(tblFile.FileAddr.String)
		c.Data(http.StatusOK, "application/octet-stream", []byte(signedURL))
	} else {
		c.Data(http.StatusOK, "application/octet-stream", []byte("Error: 下载链接暂时无法生成"))
	}
}

// 文件下载接口
func DownloadHandler(c *gin.Context) {
	fsha1 := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")

	fResp, ferr := dbCli.GetFileMeta(fsha1)
	ufResp, uferr := dbCli.QueryUserFileMeta(username, fsha1)
	if ferr != nil || uferr != nil || !fResp.Suc || !ufResp.Suc{
		c.JSON(http.StatusOK, gin.H{
			"code": common.StatusServerError,
			"msg": "server error",
		})
		return
	}
	uniqFile := dbCli.ToTableFile(fResp.Data)
	userFile := dbCli.ToTableUserFile(ufResp.Data)

	if strings.HasPrefix(uniqFile.FileAddr.String, cfg.MergeLocalRootDir) {
		c.FileAttachment(uniqFile.FileAddr.String, userFile.FileName)
	} else if strings.HasPrefix(uniqFile.FileAddr.String, cfg.CephRootDir) {
		log.Println("to download file from ceph...")
		bucket := ceph.GetCephBucket("userfile")
		fileData, _ := bucket.Get(uniqFile.FileAddr.String)
		c.Header("Content-disposition", "attachment; filename=\""+userFile.FileName+"\"")
		c.Data(http.StatusOK, "application/octect-stream", fileData)
	} else if strings.HasPrefix(uniqFile.FileAddr.String, cfg.OSSRootDir) {
		log.Println("to download file from oss...")

		fd, err1 := oss.Bucket().GetObject(uniqFile.FileAddr.String)
		if err1 == nil {
			fileData, err2 := ioutil.ReadAll(fd)
			if err2 == nil {
				c.Header("Content-disposition", "attachment; filename=\""+userFile.FileName+"\"")
				c.Data(http.StatusOK, "application/octect-stream", fileData)
			}
		}
	} else {
		c.Data(http.StatusNotFound, "application/octect-stream", []byte("File not found."))
		return
	}
}


// 支持断点的文件下载接口
func RangeDownloadHandler(c *gin.Context) {

	fsha1 := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")

	fResp, ferr:= dbCli.GetFileMeta(fsha1)
	ufResp, uferr := dbCli.QueryUserFileMeta(username, fsha1)
	if ferr != nil || uferr != nil || !fResp.Suc || !ufResp.Suc{
		c.JSON(http.StatusOK, gin.H{
			"code": common.StatusServerError,
			"msg": "server error",
		})
		return
	}

	userFile := dbCli.ToTableUserFile(ufResp.Data)

	//使用本地目录
	fpath := cfg.MergeLocalRootDir + fsha1
	fmt.Println("range-download-fpath" + fpath)

	f, err := os.Open(fpath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": common.StatusServerError,
			"msg": "server error",
		})
		return
	}
	defer f.Close()

	c.Writer.Header().Set("Content-Type", "application/octect-stream")
	// attachment表示文件将会提示下载到本地，而不是直接在浏览器中打开
	c.Writer.Header().Set("Content-disposition", "attachment; filename=\""+userFile.FileName+"\"")
	//http.ServeFile(w, r, fm.Location)
	c.File(fpath)
}


