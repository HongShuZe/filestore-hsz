#!/bin/bash

services="
dbproxy
upload
download
transfer
account
apiwg
"

stopContainers() {
  DOCKER_CONTAINER=$(docker ps -a | grep fileserver/$1 | awk '{print $1}')
  docker stop $DOCKER_CONTAINER
}

clearContainers() {
  DOCKER_CONTAINER=$(docker ps -a | grep fileserver/$1 | awk '{print $1}')
  docker rm $DOCKER_CONTAINER
}

removeUnwantedImage() {
  DOCKER_IMAGE=$(docker images | grep fileserver/$1 | awk '{print $3}')
  docker rmi $DOCKER_IMAGE
}

echo 关闭指定容器
for service in $services
do
    stopContainers $service
done

echo 删除指定容器
for service in $services
do
    clearContainers $service
done

echo 删除指定镜像
for service in $services
do
    removeUnwantedImage $service
done