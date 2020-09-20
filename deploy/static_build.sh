#!/bin/bash

ROOT_DIR=/home/zwx/go/src/filestore-hsz

# 切换到工程根目录
cd ${ROOT_DIR}

# 打包静态资源
mkdir ${ROOT_DIR}/assets -p && go-bindata-assetfs -pkg assets -o ${ROOT_DIR}/assets/asset.go static/...