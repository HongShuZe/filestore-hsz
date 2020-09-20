package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"context"
	userProto "filestore-hsz/service/account/proto"
	"log"
)

// 查询批量的文件元信息
func FileQueryHandler(c *gin.Context)  {

	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	username := c.Request.FormValue("username")


	rpcResp, err := userCli.UserFiles(context.TODO(), &userProto.ReqUserFile{
		Username: username,
		Limit: int32(limitCnt),
	})
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	if len(rpcResp.FileData) <= 0 {
		rpcResp.FileData = []byte("[]")
	}

	c.Data(http.StatusOK, "application/json", rpcResp.FileData)
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

	rpcResp, err := userCli.UserFileRename(context.TODO(), &userProto.ReqUserFileRename{
		Username: username,
		Filehash: fileSha1,
		NewFileName: newFileName,
	})
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if len(rpcResp.FileData) <= 0 {
		rpcResp.FileData = []byte("[]")
	}
	c.Data(http.StatusOK, "application/json", rpcResp.FileData)
}

// 删除用户文件信息接口(重命名)
func FileDeleteHandler(c *gin.Context) {
	fileSha1 := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")
	// context.TODO返回一个非nil的空上下文。代码应该使用上下文。当不清楚要使用哪个上下文或者上下文还不可用(因为周围的函数还没有扩展到接受上下文参数)时，可以使用TODO。
	rpcResp, err := userCli.UserFileDelete(context.TODO(), &userProto.ReqUserFileDelete{
		Username: username,
		Filehash: fileSha1,
	})

	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if len(rpcResp.FileData) <= 0 {
		rpcResp.FileData = []byte("[]")
	}
	c.Data(http.StatusOK, "application/json", rpcResp.FileData)
}
