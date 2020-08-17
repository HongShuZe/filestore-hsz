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
	// 10003 登录失败
	StatusLoginFailed
	// 10005 token无效
	StatusInvalidToken
)
