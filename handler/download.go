package handler

import (
	"net/http"
	"os"
	dblayer "filestore-hsz/db"
	"filestore-hsz/meta"
)

// 支持断点的文件下载接口
func RangeDownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fsha1 := r.Form.Get("filehash")
	username := r.Form.Get("username")

	fm, _ := meta.GetFileMetaDB(fsha1)
	userFile, err := dblayer.QueryUserFileMeta(username, fsha1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/octect-stream")
	// attachment表示文件将会提示下载到本地，而不是直接在浏览器中打开
	w.Header().Set("content-disposition", "attachment; filename=\""+userFile.FileName+"\"")
	http.ServeFile(w, r, fm.Location)
}


