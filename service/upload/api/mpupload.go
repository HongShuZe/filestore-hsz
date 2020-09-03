package handler

import (
	"os"
	"fmt"
	"net/http"
	"strconv"
	"filestore-hsz/util"
	cmn "filestore-hsz/common"
	rPool "filestore-hsz/cache/redis"
	"github.com/garyburd/redigo/redis"
	"time"
	"strings"
	"math"
	"path"

	cfg "filestore-hsz/config"
	"github.com/gin-gonic/gin"
	"log"
	dbCli "filestore-hsz/service/dbproxy/client"
	"encoding/json"
	"filestore-hsz/mq"
)

// 初始化信息
type MultipartUploadInfo struct {
	FileHash string
	FileSize int
	UploadID string
	ChunkSize int
	ChunkCount int
	// 已经上传完成的分块索引列表
	ChunkExists []int
}

const (
	// 上传分块所在位置
	ChunkDir = cfg.ChunckLocalRootDir
	// 合并后的文件所在目录
	MergeDir = cfg.MergeLocalRootDir
	// 分块信息对应的redis键前缀
	ChunkKeyPrefix = "MP_"
	// 文件hash映射uploadid对应的redis键前缀
	HashUpIDKeyPrefix = "HASH_UPID_"
)

func init()  {
	if err := os.MkdirAll(ChunkDir, 0744); err != nil {
		fmt.Println("无法指定目录用于存储分块文件:" + ChunkDir)
		os.Exit(1)
	}
	if err := os.MkdirAll(MergeDir, 0744); err != nil {
		fmt.Println("无法指定目录用于存储合并文件" + MergeDir)
		os.Exit(1)
	}
}

// 初始化分块上传
func InitialMultipartUploadHandler(c *gin.Context) {
	// 1.解析用户请求参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filesize, err := strconv.Atoi(c.Request.FormValue("filesize"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "params invalid",
		})
		return
	}

	// 判断文件是否存在
	if exists, _ := dbCli.IsUserFileUploaded(username, filehash); exists {
		c.JSON(http.StatusOK, gin.H{
			"code": int(cmn.FileAlreadExists),
			"msg":  "file exists",
		})
		return
	}

	// 2.获得redis的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3.通过文件hash判断是否断点续传, 并获取uploadID
	uploadID := ""
	keyExists, _ := redis.Bool(rConn.Do("EXISTS", HashUpIDKeyPrefix+filehash)) // 查询该值是否存在
	if keyExists {
		uploadID, err = redis.String(rConn.Do("GET", HashUpIDKeyPrefix+filehash))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -2,
				"msg":  "Upload part failed" + err.Error(),
			})
			return
		}
	}

	// 4.1 首次上传则新建uploadID
	// 4.2 断点续传则根据upload获取已上传的文件分块列表
	chunksExist := []int{}
	if uploadID == "" {
		uploadID = username + fmt.Sprintf("%x", time.Now().UnixNano())
	} else {
		chunks, err := redis.Values(rConn.Do("HGETALL", ChunkKeyPrefix+uploadID))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -3,
				"msg":  "Upload part failed" + err.Error(),
			})
			return
		}
		for i := 0; i < len(chunks); i += 2 {
			k := string(chunks[i].([]byte))
			v := string(chunks[i+1].([]byte))
			if strings.HasPrefix(k, "chkidx_") && v == "1" {
				// chkidx_6 -> 6
				chunkIdx, _ := strconv.Atoi(k[7:len(k)])
				chunksExist = append(chunksExist, chunkIdx)
			}
		}
	}

	// 5.生成分块上传的初始化信息
	upInfo := MultipartUploadInfo{
		FileHash:    filehash,
		FileSize:    filesize,
		UploadID:    uploadID,
		ChunkSize:   5 * 1024 * 1024, //5MB
		ChunkCount:  int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
		ChunkExists: chunksExist,
	}

	// 6.将初始化信息写入到redis缓存
	if len(upInfo.ChunkExists) <= 0 {
		hKey := ChunkKeyPrefix + upInfo.UploadID
		rConn.Do("HSET", hKey, "chunkcount", upInfo.ChunkCount) // 将哈希表hKey中的字段chunkcount的值设为upInfo.ChunkCount
		rConn.Do("HSET", hKey, "filehash", upInfo.FileHash)
		rConn.Do("HSET", hKey, "filesize", upInfo.FileSize)
		rConn.Do("EXPIRE", hKey, 43200)                                           // 设置键hKey有效期为半天
		rConn.Do("SET", HashUpIDKeyPrefix+filehash, upInfo.UploadID, "EX", 43200) // "EX", 43200设置有效期为半天
	}

	// 7.将响应初始化数据返回到客户端
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg": "ok",
		"data": upInfo,
	})
}

// 上传文件分块
func UploadPartHandler(c *gin.Context)  {
	// 1.解析用户请求参数

	uploadID := c.Request.FormValue("uploadid")
	chunkSha1 := c.Request.FormValue("chkhash")
	chunkIndex := c.Request.FormValue("index")

	// 2.获取redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3.获得文件句柄, 用于存储分块内容
	fpath := ChunkDir + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg": "Upload part failed" + err.Error(),
			"data": nil,
		})
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := c.Request.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	// 校验分块hash
	cmpSha1, err := util.ComputeSha1ByShell(fpath)
	if err != nil || cmpSha1 != chunkSha1 {
		log.Printf("Verify chunk sha1 failed, compare OK:%t, err:%+v\n",
			cmpSha1 == chunkSha1, err)

		c.JSON(http.StatusOK, gin.H{
			"code": -2,
			"msg": "Verify hash failed, chKIdx:"+chunkIndex,
			"data": nil,
		})
		return
	}

	// 4.更新redis缓存状态
	rConn.Do("HSET", ChunkKeyPrefix+uploadID, "chkidx_"+chunkIndex, 1)

	// 5.返回处理结果到客户端
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg": "ok",
		"data": nil,
	})
}

// 通知上传合并
func CompleteUploadHandler(c *gin.Context)  {
	// 1.解析用户请求参数

	uploadID := c.Request.FormValue("uploadid")
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filesize := c.Request.FormValue("filesize")
	filename := c.Request.FormValue("filename")
	// 2.获取redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3.通过uploadid查询redis并判断是否所有分块上传完成
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+uploadID))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg": "服务错误",
			"data": nil,
		})
		return
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		}else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		c.JSON(http.StatusOK, gin.H{
			"code": -2,
			"msg": "分块不完整",
			"data": nil,
		})
		return
	}

	// 4.合并上传(此合并逻辑非必须实现，因后期转移到ceph/oss时也可以通过分块方式上传)
	srcPath := cfg.ChunckLocalRootDir + uploadID + "/"
	destPath := cfg.MergeLocalRootDir + filehash
	cmd := fmt.Sprintf("cd %s && ls | sort -n | xargs cat > %s", srcPath, destPath)
	mergeRes, err := util.ExecLinuxShell(cmd)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -3,
			"msg": "合并失败",
			"data": nil,
		})
		return
	}
	log.Println(mergeRes)

	// 5.更新唯一文件表及用户文件表
	fsize, _ := strconv.Atoi(filesize)

	fileMeta := dbCli.FileMeta{
		FileSha1: filehash,
		FileName: filename,
		FileSize: int64(fsize),
		Location: destPath,
	}
	// 增加fileaddr参数的写入
	_, ferr := dbCli.OnFileUploadFinished(fileMeta)
	_, uferr :=  dbCli.OnUserFileUploadFinished(username, fileMeta)
	if ferr != nil || uferr != nil {
		errMsg := ""
		if ferr != nil {
			errMsg = ferr.Error()
		} else {
			errMsg = uferr.Error()
		}
		log.Println(errMsg)

		c.JSON(http.StatusOK, gin.H{
			"code": -4,
			"msg": "数据更新失败",
			"data": errMsg,
		})
		return
	}

	// 删除已上传的分块文件及redis分块信息
	_, delHashErr := rConn.Do("DEL", HashUpIDKeyPrefix+filehash)
	delUploadID, delUploadInfoErr := redis.Int64(rConn.Do("DEL", ChunkKeyPrefix+uploadID))
	if delUploadID != 1 || delUploadInfoErr != nil || delHashErr != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -5,
			"msg": "数据更新失败",
			"data": nil,
		})
		return
	}
	// 6.异步文件转移
	ossPath := cfg.OSSRootDir +fileMeta.FileSha1
	transMsg := mq.TransferData{
		FileHash: fileMeta.FileSha1,
		CurLocation: fileMeta.Location,
		DestLocation: ossPath,
		DestStoreType: cmn.StoreOSS,
	}
	fmt.Println(transMsg)
	pubData, _ := json.Marshal(data)
	pubSuc := mq.Publish(
		cfg.TransExchangeName,
		cfg.TransOSSRoutingKey,
		pubData,
	)
	if !pubSuc {
		// TODO: 当前发送转移信息失败, 稍后重试
		fmt.Println("当前发送转移信息失败, 稍后重试, data failed, sha1: " + fileMeta.FileSha1)
	}

	// 7.响应处理结果
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg": "ok",
		"data": nil,
	})
}



