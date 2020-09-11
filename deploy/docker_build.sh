#!/bin/bash

#ROOT_DIR=$GOPATH/filestore-server
#ROOT_DIR=/data/go/work/src/filestore-server
ROOT_DIR=/home/zwx/go/src/filestore-hsz
services="
dbproxy
upload
download
transfer
account
apiwg
"

# 检查service进程
check_process() {
    sleep 1
    res=`ps aux | grep -v grep | grep "service/bin" | grep $1`
    if [[ $res != '' ]]; then
        echo -e "\033[32m已启动 \033[0m" "$1"
        return 1
    else
        echo -e "\033[31m启动失败 \033[0m" "$1"
        return 0
    fi
}

# 编译service可执行文件
build_service() {
    rm -f ${ROOT_DIR}/deploy/bin/$1
    go build -o ${ROOT_DIR}/deploy/bin/$1 ${ROOT_DIR}/service/$1/main.go
    echo -e "\033[32m编译完成: \033[0m ${ROOT_DIR}/deploy/bin/$1"
}

# 打包镜像
build_image() {
    # 替换(fileserver/$service), 自定义镜像名即可
    sudo docker build -t fileserver/$1 -f ./service/$1/Dockerfile .
    echo -e "\033[32m镜像打包完成: \033[0m fileserver/$1\n"
}

# 切换到工程根目录
cd ${ROOT_DIR}

# 打包静态资源
#mkdir ${ROOT_DIR}/assets -p && go-bindata-assetfs -pkg assets -o ${ROOT_DIR}/assets/asset.go static/...

# 执行编译service
mkdir -p ${ROOT_DIR}/deploy/bin && rm -f ${ROOT_DIR}/deploy/bin/*
for service in $services
do
    build_service $service
done

echo -e "\033[32m编译完毕, 开始构建docker镜像... \033[0m"

# 打包微服务镜像
cd ${ROOT_DIR}/deploy/
for service in $services
do
    build_image $service
done

echo -e "\033[32mdocker镜像构建完毕.\033[0m"

# 容器启动示例
# 启动account service
# docker run -it -e PARAMS="--registry=consul --registry_address=192.168.200.212:8500" fileserver/account