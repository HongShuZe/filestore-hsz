package config

import (
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-plugins/registry/consul"
	"github.com/micro/go-micro/client/selector"
)


// 配置conusl
func RegistryConsul() registry.Registry {
	return consul.NewRegistry(
		//
		registry.Addrs("192.168.20.143:8500"),
	)
}

// 注册中心client
func RegistryClient(r registry.Registry) selector.Selector {
	return selector.NewSelector(
		selector.Registry(r), // 传入consul注册
		selector.SetStrategy(selector.RoundRobin),// 指定查询机制
	)
}