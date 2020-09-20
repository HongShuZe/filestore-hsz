#!/bin/bash

services="
dbproxy
upload
download
transfer
account
apiwg
"

clearContainers() {
  DOCKER_CONTAINER=$(docker ps -a | grep fileserver/$1 | awk '{print $1}')
  docker rm $DOCKER_CONTAINER
}

removeUnwantedImage() {
  DOCKER_IMAGE=$(docker images | grep fileserver/$1 | awk '{print $3}')
  docker rmi $DOCKER_IMAGE
}

for service in $services
do
    clearContainers $service
done

for service in $services
do
    removeUnwantedImage $service
done