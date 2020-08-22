package common

// 存储类型(表示文件存到哪里)
type StoreType int

const (
	_ StoreType = iota
	// 节点本地
	StoreLocal
	// Ceph集群
	StoreCeph
	// 阿里OSS
	StoreOSS
	// 混合(Ceph及阿里)
	StoreMix
	// 所有类型的存储都存在一份数据
	StoreAll
)
