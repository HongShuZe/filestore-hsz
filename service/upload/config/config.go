package config

// 配置上传入口地址,使用traefik时要改为"upload.fileserver.com"
var UploadEntry = "localhost:28080"

// 上传服务监听的地址
var UploadServiceHost = "0.0.0.0:28080"
