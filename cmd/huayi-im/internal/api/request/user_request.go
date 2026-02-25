package request

// GetUserReq 获取用户请求（路径参数）
type GetUserReq struct {
	Username string `uri:"username" binding:"required,min=3,max=50"`
}

// LoginReq 登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
}
