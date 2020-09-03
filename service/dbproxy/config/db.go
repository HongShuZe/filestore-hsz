package config

const (
	// MySQLSource: 需要连接的数据库源
	// root:123456 是用户名和密码
	// 127.0.0.1:3306 是ip及端口
	// fileserver 是数据库名
	// charset=utf8 指定数据库以utf8字符编码进行传输
	MySQLSource = "root:123456@tcp(127.0.0.1:13306)/fileserver?charset=utf8"
)
