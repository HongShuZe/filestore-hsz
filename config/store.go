package config

import cmn "filestore-hsz/common"

const (
	// 本地临时存储地址的路径
	TempLocalRootDir = "/home/zwx/data/fileserver_tmp/"
	// 本地存储地址的路径(包含普通上传及分块上传)
	MergeLocalRootDir = "/home/zwx/data/fileserver_marge/"
	// 分块存储地址的路径
	ChunckLocalRootDir = "/home/zwx/data/fileserver_chunk/"
	// Ceph的存储路径prefix
	CephRootDir = "/ceph"
	// OSS的存储路径prefix
	OSSRootDir = "oss/"
	// 设置当前文件的存储类型
	CurrentStoreType = cmn.StoreOSS
)