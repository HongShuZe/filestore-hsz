package common

// 错误码
type ErrorCode int32

const (
	_ int32 = iota + 9999

	// 10000 正常
	StatusOK
	// 10001 请求参数无效
	StatusParamInvalid
	// 10002 服务出错
	StatusServerError
	// 10003 注册失败
	StatusRegisterFailed
	// 10004 登录失败
	StatusLoginFailed
	// 10005 token无效
	StatusInvalidToken
	// 10006 文件已存在
	FileAlreadExists
	// 10007 用户不存在
	StatusUserNoExists
)
