package request

// GetUserReq 获取用户请求（路径参数）
type GetUserReq struct {
	UserID uint64 `uri:"user_id" binding:"required,min=1"`
}

// LoginReq 登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
}
