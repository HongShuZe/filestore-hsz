package oss

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	cfg "filestore-hsz/config"
	"fmt"
)

var ossCli *oss.Client

// 创建oss client对象
func Client() *oss.Client {
	if ossCli != nil {
		return ossCli
	}
	ossCli, err := oss.New(cfg.OSSEndpoint,
		cfg.OSSAccesskeyID, cfg.OSSAccessSecret)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return ossCli
}

// 获取bucket存储空间
func Bucket() *oss.Bucket {
	cli := Client()
	if cli != nil {
		bucket, err := cli.Bucket(cfg.OSSBucket)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		return bucket
	}
	return nil
}

// 临时授权下载url
func DownloadURL(objName string) string {
	signedURL, err := Bucket().SignURL(objName, oss.HTTPGet, 3600)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return signedURL
}

// 针对指定bucket设置生命周期规则
func BuildLifecycleRule(bucketName string)  {
	// 表示前缀为test的对象(文件)距最后修改时间30天后过期
	ruleTest1 := oss.BuildLifecycleRuleByDays("rule1", "test/", true, 30)
	rules := []oss.LifecycleRule{ruleTest1}

	Client().SetBucketLifecycle(bucketName, rules)
}