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
)

// 处理文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 返回上传html页面
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internel server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		// 接收文件流及存储到本地目录
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("failed to get data, err:%s\n", err.Error())
			return
		}
		defer file.Close()

		tmpPath := cfg.TempLocalRootDir + head.Filename
		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: tmpPath,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Printf("Failed to create file, err:%s\n", err.Error())
			return
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Failed to save data into file, err:%s\n", err.Error())
			return
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)

		// 5.同步或异步将文件转移到Ceph/OSS
		newFile.Seek(0, 0) // 游标重新回到文件头部
		mergePath := cfg.MergeLocalRootDir + fileMeta.FileSha1
		if cfg.CurrentStoreType == cmn.StoreCeph {
			data, _ := ioutil.ReadAll(newFile)
			cephPath := "/ceph/" + fileMeta.FileSha1
			err = ceph.PutObject("userfile", cephPath, data)
			if err != nil {
				fmt.Println("upload ceph err: " + err.Error())
				w.Write([]byte("Upload failed!"))
				return
			}
			fileMeta.Location = cephPath
		} else if cfg.CurrentStoreType == cmn.StoreOSS {
			ossPath := "oss/" + fileMeta.FileSha1
			err = oss.Bucket().PutObject(ossPath, newFile)
			if err != nil {
				fmt.Println("upload oss err: " + err.Error())
				w.Write([]byte("upload failed"))
				return
			}
			fileMeta.Location = ossPath
		} else {
			fileMeta.Location = mergePath
		}

		err = os.Rename(tmpPath, mergePath) //移动文件
		if err != nil {
			fmt.Println("move local file err:", err.Error())
			w.Write([]byte("upload failed"))
			return
		}

		//meta.UpdateFileMeta(fileMeta)
		_ = meta.UpdateFileMetaDB(fileMeta)

		//更新用户文件表记录
		r.ParseForm()
		username := r.Form.Get("username")
		suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1,
			fileMeta.FileName, fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed"))
		}
	}
}

// 上传已完成
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished")
}

// 获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request)  {
	r.ParseForm()

	filehash := r.Form["filehash"][0]
	//fmeta := meta.GetFileMeta(filehash)
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.Write(data)
}

// 查询批量的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request)  {
	r.ParseForm()

	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	//fileMetas := meta.GetLastFileMetas(limitCnt)
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//w.Header().Set("Content-type", "application/json")
	w.Write(data)
}


// 更新元信息接口(重命名)
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request)  {
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")
	username := r.Form.Get("username")

	if opType != "0" || len(newFileName) < 1 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//curFileMeta := meta.GetFileMeta(fileSha1)
	//curFileMeta.FileName = newFileName
	//meta.UpdateFileMeta(curFileMeta)

	// 更新用户表tb_user_file中的文件名，tb_user_file文件名不用修改
	_ = dblayer.RenameFileName(username, fileSha1, newFileName)
	// 返回最新的文件信息

	userFile, err := dblayer.QueryUserFileMeta(username, fileSha1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	//w.Header().Set("Content-type", "application/json")
	w.Write(data)
}


// 删除文件及元信息
func FileDeleteHandler(w http.ResponseWriter, r *http.Request)  {
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")
	username := r.Form.Get("username")

	// 删除本地文件
	fm, err := meta.GetFileMetaDB(fileSha1)
	//fMeta := meta.GetFileMeta(fileSha1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	os.Remove(fm.Location)

	// 删除用户文件表的一条记录
	suc := dblayer.DeleteUserFile(username, fileSha1)
	if !suc {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//meta.RemoveFileMeta(fileSha1)
	w.WriteHeader(http.StatusOK)
}


// 尝试秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request)  {

	// 1.解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _:= strconv.Atoi(r.Form.Get("filesize"))

	// 2.从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3.查不到记录则返回秒传失败
	if fileMeta.FileSha1 == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg: "秒传失败， 请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
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
		w.Write(resp.JSONBytes())
		return
	}

	resp := util.RespMsg{
		Code: -2,
		Msg: "秒传失败,请重试",
	}
	w.Write(resp.JSONBytes())
}

// 生成文件下载地址
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	filehash := r.Form.Get("filehash")
	// 从文件表查找记录
	row, _ := dblayer.GetFileMeta(filehash)
	fmt.Println("fileAddr: "+ row.FileAddr.String)
	if strings.HasPrefix(row.FileAddr.String, cfg.MergeLocalRootDir) || strings.HasPrefix(row.FileAddr.String, "/ceph") {
		username := r.Form.Get("username")
		token := r.Form.Get("token")
		tmpURL := fmt.Sprintf(
			"http://%s/file/download?filehash=%s&username=%s&token=%s",
			r.Host, filehash, username, token)
		w.Write([]byte(tmpURL))
	} else if strings.HasPrefix(row.FileAddr.String, "oss/") {
		signedURL := oss.DownloadURL(row.FileAddr.String)
		w.Write([]byte(signedURL))
	} else {
		w.Write([]byte("Error: 下载链接暂时无法生成"))
	}

}

// 文件下载接口
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	username := r.Form.Get("username")

	fm, _ := meta.GetFileMetaDB(fsha1)
	userFile, err := dblayer.QueryUserFileMeta(username, fsha1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var fileData []byte
	if strings.HasPrefix(fm.Location, cfg.MergeLocalRootDir)  {
		f, err := os.Open(fm.Location)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer f.Close()

		fileData, err = ioutil.ReadAll(f)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}else if strings.HasPrefix(fm.Location, "/ceph") {
		fmt.Println("to download file from ceph...")
		bucket := ceph.GetCephBucket("userfile")
		fileData, err = bucket.Get(fm.Location)
		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}else {
		w.Write([]byte("file not found"))
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	// attachment表示文件将会提示下载到本地，而不是直接在浏览器中打开
	w.Header().Set("content-disposition", "attachment; filename=\""+userFile.FileName+"\"")
	w.Write(fileData)
}











