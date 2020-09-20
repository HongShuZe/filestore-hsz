#!/bin/bash

ROOT_DIR=/home/zwx/go/src/filestore-hsz
#ROOT_DIR=/data/imooc/src/filestore-server
services="
dbproxy
upload
download
transfer
account
apiwg
"

# 替换为你的本地ip
hostIP="192.168.20.143"

registryAddr="--registry=consul --registry_address=${hostIP}:8500"
redisAddr="--cachehost=${hostIP}:6379"
mysqlAddr="--dbhost=${hostIP}:13306"
mqAddr="--mqhost=${hostIP}:5672"

# 强制删除已有的容器
# 生产环境不建议这么做, 后续用k8s可以实现服务平滑重启
echo -e "\033[31m检查并停止已有的容器... \033[0m"
for service in $services
do
    app=`sudo docker ps -a | grep "fileserver/${service}" | awk '{print $1}'`
    if [[ $app != "" ]];then
        echo $app | xargs sudo docker rm -f
    fi
done

echo -e "\033[32m启动微服务容器... \033[0m"
for service in $services
do
    volumes=""
    # 指定挂载目录
    if [[ $service == "upload" || $service == "download" || $service == "transfer" ]];then
        volumes="-v /home/zwx/data/fileserver_marge:/home/zwx/data/fileserver_marge -v /home/zwx/data/fileserver_chunk:/home/zwx/data/fileserver_chunk"
    fi

    # 启动容器
    sudo docker run -it -d \
      --net=host --privileged=true ${volumes} \
      -e PARAMS="${registryAddr} ${redisAddr} ${mysqlAddr} ${mqAddr}" \
      fileserver/${service}
done
