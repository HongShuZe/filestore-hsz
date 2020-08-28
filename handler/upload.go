package handler

import (
	"net/http"
	"io/ioutil"
	"io"
	"fmt"
	"filestore-hsz/meta"
	"time"
	"os"
	"filestore-hsz/util"
	"encoding/json"
	"strconv"
	dblayer "filestore-hsz/db"
	cfg "filestore-hsz/config"
	cmn "filestore-hsz/common"
	"filestore-hsz/store/ceph"
	"strings"
	"filestore-hsz/store/oss"
	"filestore-hsz/mq"
	"github.com/gin-gonic/gin"
	"bytes"
)

func init()  {
	if err := os.MkdirAll(cfg.TempLocalRootDir, 0744); err != nil {
		fmt.Println("无法指定目录用于存储临时文件", cfg.TempLocalRootDir)
		os.Exit(1)
	}
	if err := os.MkdirAll(cfg.MergeLocalRootDir, 0744); err != nil {
		fmt.Println("无法指定目录用于存储临合并后文件", cfg.MergeLocalRootDir)
		os.Exit(1)
	}
}

// 处理用户注册请求
func UploadHandler(c *gin.Context)  {
	// 返回上传html页面
	data, err := ioutil.ReadFile("./static/view/index.html")
	if err != nil {
		c.String(404, `页面不存在`)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

// 处理文件上传
func DoUploadHandler(c *gin.Context) {

	errCode := 0
	defer func() {
		if errCode < 0 {
			c.JSON(http.StatusOK, gin.H{
				"code": errCode,
				"msg": "Upload failed",
			})
		}
	}()

	// 1.从form表单获取文件内热句柄
	file, head, err := c.Request.FormFile("file")
	if err != nil {
		fmt.Printf("failed to get data, err:%s\n", err.Error())
		errCode = -1
		return
	}
	defer file.Close()

	// 2.把文件内容转为[]byte
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		fmt.Printf("failed to get data, err:%s\n", err.Error())
		errCode = -2
		return
	}

	// 3.构建文件元信息
	fileMeta := meta.FileMeta{
		FileName: head.Filename,
		FileSha1: util.Sha1(buf.Bytes()),
		FileSize: int64(len(buf.Bytes())),
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	// 4.将文件写入临时存储位置
	fileMeta.Location = cfg.TempLocalRootDir + fileMeta.FileSha1
	newFile, err := os.Create(fileMeta.Location)
	if err != nil {
		fmt.Printf("Failed to create file, err:%s\n", err.Error())
		errCode = -3
		return
	}
	defer newFile.Close()

	nByte, err := newFile.Write(buf.Bytes())
	if int64(nByte) != fileMeta.FileSize || err != nil {
		fmt.Printf("Failed to save data into file, writenSize:%d, err:%s\n",nByte, err.Error() )
		errCode = -4
		return
	}
	newFile.Seek(0, 0)


	// 5.同步或异步将文件转移到Ceph/OSS
	newFile.Seek(0, 0) // 游标重新回到文件头部
	//mergePath := cfg.MergeLocalRootDir + fileMeta.FileSha1
	if cfg.CurrentStoreType == cmn.StoreCeph {
		data, _ := ioutil.ReadAll(newFile)
		cephPath := "/ceph/" + fileMeta.FileSha1
		err = ceph.PutObject("userfile", cephPath, data)
		if err != nil {
			fmt.Println("upload ceph err: " + err.Error())
			errCode = -5
			return
		}
		fileMeta.Location = cephPath
	} else if cfg.CurrentStoreType == cmn.StoreOSS {
		ossPath := "oss/" + fileMeta.FileSha1

		// 判断写入OSS为同步还是异步
		if !cfg.AsyncTransferEnable {
			// TODO: 设置文件oss中的文件名, 方便指定文件下载
			err = oss.Bucket().PutObject(ossPath, newFile)
			if err != nil {
				fmt.Println("upload oss err: " + err.Error())
				errCode = -5
				return
			}
			fileMeta.Location = ossPath
		} else {
			// 文件尚未转移, 暂存于本地mergePath
			// fileMeta.Location = mergePath
			// 写入异步转移任务队列
			data := mq.TransferData{
				FileHash: fileMeta.FileSha1,
				CurLocation: fileMeta.Location,
				DestLocation: ossPath,
				DestStoreType: cmn.StoreOSS,
			}
			fmt.Println(data)
			pubData, _ := json.Marshal(data)
			pubSuc := mq.Publish(
				cfg.TransExchangeName,
				cfg.TransOSSRoutingKey,
				pubData,
			)
			if !pubSuc {
				// TODO: 当前发送转移信息失败, 稍后重试
				errCode = -6
			}
		}
	}

	/*err = os.Rename(tmpPath, mergePath) //移动文件
	if err != nil {
		fmt.Println("move local file err:", err.Error())
		w.Write([]byte("upload failed"))
		return
	}*/

	// 6.更新文件表记录
	_ = meta.UpdateFileMetaDB(fileMeta)

	// 7.更新用户文件表记录
	username := c.Request.FormValue("username")
	suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1,
		fileMeta.FileName, fileMeta.FileSize)
	if suc {
		c.Redirect(http.StatusFound, "/static/view/home.html")
	} else {
		errCode = -6
	}
}

// 上传已完成
func UploadSucHandler(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg": "Upload Finish",
	})
}

// 获取文件元信息
func GetFileMetaHandler(c *gin.Context)  {

	filehash := c.Request.FormValue("filehash")
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -2,
			"msg": "Upload failed",
		})
		return
	}

	fileMeta := meta.FileMeta{}
	if fMeta != fileMeta {
		data, err := json.Marshal(fMeta)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": -3,
				"msg": "Upload failed",
			})
			return
		}
		c.Data(http.StatusOK, "application/json", data)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": -4,
			"msg": "No such file",
		})
	}
}

// 查询批量的文件元信息
func FileQueryHandler(c *gin.Context)  {

	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	username := c.Request.FormValue("username")
	//fileMetas := meta.GetLastFileMetas(limitCnt)
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"msg": "Query failed",
		})
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -2,
			"msg": "Query failed",
		})
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}


// 更新元信息接口(重命名)
func FileMetaUpdateHandler(c *gin.Context)  {

	opType := c.Request.FormValue("op")
	fileSha1 := c.Request.FormValue("filehash")
	newFileName := c.Request.FormValue("filename")
	username := c.Request.FormValue("username")

	if opType != "0" || len(newFileName) < 1 {
		c.Status(http.StatusForbidden)
		return
	}


	// 更新用户表tb_user_file中的文件名，tb_user_file文件名不用修改
	_ = dblayer.RenameFileName(username, fileSha1, newFileName)
	// 返回最新的文件信息

	userFile, err := dblayer.QueryUserFileMeta(username, fileSha1)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFile)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, data)
}


// 删除文件及元信息
func FileDeleteHandler(c *gin.Context)  {

	fileSha1 := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")

	fm, err := meta.GetFileMetaDB(fileSha1)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	// 删除本地文件
	os.Remove(fm.Location)
	// TODO: 可考虑删除Ceph/OSS上传的文件
	// 可以不立即删除, 加个超时机制
	// 如该文件10天后没有用户上传,就可以真正删除

	// 删除用户文件表的一条记录
	suc := dblayer.DeleteUserFile(username, fileSha1)
	if !suc {
		c.Status(http.StatusInternalServerError)
		return
	}
	//meta.RemoveFileMeta(fileSha1)
	c.Status(http.StatusOK)
}


// 尝试秒传接口
func TryFastUploadHandler(c *gin.Context)  {

	// 1.解析请求参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	filesize, _:= strconv.Atoi(c.Request.FormValue("filesize"))

	// 2.从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	// 3.查不到记录则返回秒传失败
	if fileMeta.FileSha1 == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg: "秒传失败， 请访问普通上传接口",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	}

	// 4.上传过则将文件信息写入用户文件表
	suc := dblayer.OnUserFileUploadFinished(
		username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg: "秒传成功",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	}

	resp := util.RespMsg{
		Code: -2,
		Msg: "秒传失败,请重试",
	}
	c.Data(http.StatusOK, "application/json", resp.JSONBytes())
	return
}

// 生成文件下载地址
func DownloadURLHandler(c *gin.Context) {
	filehash := c.Request.FormValue("filehash")
	// 从文件表查找记录
	row, _ := dblayer.GetFileMeta(filehash)
	fmt.Println("fileAddr: "+ row.FileAddr.String)

	// 判断文件存在OSS, 还是Ceph, 还是本地
	if strings.HasPrefix(row.FileAddr.String, cfg.TempLocalRootDir) || strings.HasPrefix(row.FileAddr.String, "/ceph") {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		tmpURL := fmt.Sprintf(
			"http://%s/file/download?filehash=%s&username=%s&token=%s",
			c.Request.Host, filehash, username, token)
		c.Data(http.StatusOK, "octet-stream", []byte(tmpURL))
	} else if strings.HasPrefix(row.FileAddr.String, "oss/") {
		signedURL := oss.DownloadURL(row.FileAddr.String)
		c.Data(http.StatusOK, "octet-stream", []byte(signedURL))
	}
}

// 文件下载接口
func DownloadHandler(c *gin.Context) {

	fsha1 := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")

	fm, _ := meta.GetFileMetaDB(fsha1)
	userFile, err := dblayer.QueryUserFileMeta(username, fsha1)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	var fileData []byte
	if strings.HasPrefix(fm.Location, cfg.TempLocalRootDir)  {
		c.FileAttachment(fm.Location, userFile.FileName)
	}else if strings.HasPrefix(fm.Location, "/ceph") {
		fmt.Println("to download file from ceph...")
		bucket := ceph.GetCephBucket("userfile")
		fileData, err = bucket.Get(fm.Location)
		if err != nil {
			fmt.Println(err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}
	}else if strings.HasPrefix(fm.Location, "oss") {
		fmt.Println("to download file from oss...")

		fd, err := oss.Bucket().GetObject(fm.Location)
		if err != nil {
			fileData, err = ioutil.ReadAll(fd)
		}
		if err != nil {
			fmt.Println(err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}
	}else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"msg": "file not found",
		})
		return
	}

	c.Header("content-disposition", "attachment: filename=\""+userFile.FileName+"\"")
	c.Data(http.StatusOK, "application/octect-stream", fileData)
}











