package config

// 配置上传入口地址,使用traefik时要改为"download.fileserver.com"
var DownloadEntry = "localhost:38080"

// 上传服务监听地址
var DownloadServiceHost = "0.0.0.0:38080"
