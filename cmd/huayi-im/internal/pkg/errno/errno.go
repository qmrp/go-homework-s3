package errno

// Errno 错误码接口
type Errno interface {
	Code() int
	Message() string
	WithMsg(msg string) Errno
}

// errno 实现 Errno 接口
type errno struct {
	code    int
	message string
}

// 预定义错误码（按业务模块划分）
var (
	Success      = &errno{code: 0, message: "success"}
	ParamInvalid = &errno{code: 400, message: "参数无效"}
	ServerError  = &errno{code: 500, message: "服务器内部错误"}
	UserNotFound = &errno{code: 404, message: "用户不存在"}
	UserExists   = &errno{code: 20002, message: "用户已存在"}
	Unauthorized = &errno{code: 40001, message: "未授权"}
	NotFound     = &errno{code: 404, message: "资源不存在"}
)

// Code 获取错误码
func (e *errno) Code() int {
	return e.code
}

// Message 获取错误信息
func (e *errno) Message() string {
	return e.message
}

// WithMsg 自定义错误信息
func (e *errno) WithMsg(msg string) Errno {
	return &errno{code: e.code, message: msg}
}
