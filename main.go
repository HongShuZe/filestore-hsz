package main

import (
	"net/http"
	"filestore-hsz/handler"
	"fmt"
)

func main()  {
	// 静态资源处理
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("./static"))))

	// 动态接口路由设置
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)
	http.HandleFunc("/file/query", handler.FileQueryHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/update", handler.FileMetaUpdatehandler)
	http.HandleFunc("/file/delete", handler.FileDeleteHandler)

	//监听端口
	fmt.Println("上传服务器正在启动， 监听端口：8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("failed to start server, err:%s", err)
	}
}