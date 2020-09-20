package handler

import (
	"net/http"
	"io/ioutil"
	"io"
	"fmt"
	"time"
	"os"
	"filestore-hsz/util"
	"encoding/json"
	cfg "filestore-hsz/config"
	cmn "filestore-hsz/common"
	"filestore-hsz/store/ceph"
	"filestore-hsz/store/oss"
	"filestore-hsz/mq"
	"github.com/gin-gonic/gin"
	"bytes"
	dbCli "filestore-hsz/service/dbproxy/client"
	"log"
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

// 处理文件上传
func DoUploadHandler(c *gin.Context) {

	errCode := 0
	defer func() {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		if errCode < 0 {
			c.JSON(http.StatusOK, gin.H{
				"code": errCode,
				"msg":  "上传失败",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": errCode,
				"msg":  "上传成功",
			})
		}
	}()

	// 1.从form表单获取文件内热句柄
	file, head, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("failed to get data, err:%s\n", err.Error())
		errCode = -1
		return
	}
	defer file.Close()

	// 2.把文件内容转为[]byte
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		log.Printf("failed to get data, err:%s\n", err.Error())
		errCode = -2
		return
	}

	// 3.构建文件元信息
	fileMeta := dbCli.FileMeta{
		FileName: head.Filename,
		FileSha1: util.Sha1(buf.Bytes()),
		FileSize: int64(len(buf.Bytes())),
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	// 4.将文件写入临时存储位置
	fileMeta.Location = cfg.MergeLocalRootDir + fileMeta.FileSha1
	newFile, err := os.Create(fileMeta.Location)
	if err != nil {
		log.Printf("Failed to create file, err:%s\n", err.Error())
		errCode = -3
		return
	}
	defer newFile.Close()

	nByte, err := newFile.Write(buf.Bytes())
	if int64(nByte) != fileMeta.FileSize || err != nil {
		log.Printf("Failed to save data into file, writenSize:%d, err:%s\n",nByte, err.Error() )
		errCode = -4
		return
	}

	// 5.同步或异步将文件转移到Ceph/OSS
	newFile.Seek(0, 0) // 游标重新回到文件头部

	if cfg.CurrentStoreType == cmn.StoreCeph {
		data, _ := ioutil.ReadAll(newFile)
		cephPath := "/ceph/" + fileMeta.FileSha1
		err = ceph.PutObject("userfile", cephPath, data)
		if err != nil {
			log.Println("upload ceph err: " + err.Error())
			errCode = -5
			return
		}
		fileMeta.Location = cephPath
	} else if cfg.CurrentStoreType == cmn.StoreOSS {
		ossPath := cfg.OSSRootDir + fileMeta.FileName
		// 判断写入OSS为同步还是异步
		if !cfg.AsyncTransferEnable {
			// TODO: 设置文件oss中的文件名, 方便指定文件下载
			err = oss.Bucket().PutObject(ossPath, newFile)
			if err != nil {
				log.Println("upload oss err: " + err.Error())
				errCode = -5
				return
			}
			fileMeta.Location = ossPath
		} else {
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
				//errCode = -6
				fmt.Println("当前发送转移信息失败, 稍后重试")
			}
		}
	}

	// 6.更新文件表记录
	_, err = dbCli.OnFileUploadFinished(fileMeta)
	if err != nil {
		errCode = -6
		return
	}
	// 7.更新用户文件表记录
	username := c.Request.FormValue("username")
	upRes, err := dbCli.OnUserFileUploadFinished(username, fileMeta)
	if err == nil && upRes.Suc {
		errCode = 0
	} else {
		errCode = -7
	}
}


// 尝试秒传接口
func TryFastUploadHandler(c *gin.Context)  {

	// 1.解析请求参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	//filesize, _:= strconv.Atoi(c.Request.FormValue("filesize"))

	// 2.从文件表中查询相同hash的文件记录
	fileMetaResp, err := dbCli.GetFileMeta(filehash)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	// 3.查不到记录则返回秒传失败
	if !fileMetaResp.Suc || fileMetaResp.Data == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg: "秒传失败， 请访问普通上传接口",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	}

	// 4.上传过则将文件信息写入用户文件表
	tblFile := dbCli.ToTableFile(fileMetaResp.Data)
	fmeta := dbCli.TableFileToFileMeta(tblFile)
	fmeta.FileName = filename
	upRes, err := dbCli.OnUserFileUploadFinished(username, fmeta)
	if err == nil && upRes.Suc {
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

