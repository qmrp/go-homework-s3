package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/api/request"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/api/response"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/pkg/errno"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/service"
)

// UserHandler 用户控制器
type UserHandler struct {
	userService service.UserService // 依赖注入业务层接口
}

// NewUserHandler 创建用户控制器实例
func NewUserHandler(s service.UserService) *UserHandler {
	return &UserHandler{userService: s}
}

/** GetUserByID 获取用户详情
 * @Summary 获取用户详情
 * @Description 根据用户ID查询用户信息
 * @Tags 用户模块
 * @Accept json
 * @Produce json
 * @Param user_id path uint64 true "用户ID"
 * @Success 200 {object} response.Response{data=model.User}
 * @Failure 10001 {object} response.Response "参数无效"
 * @Failure 20001 {object} response.Response "用户不存在"
 * @Router /api/v1/users/{user_id} [get]
 **/
func (h *UserHandler) GetUserByID(c *gin.Context) {
	// 1. 绑定路径参数
	var req request.GetUserReq
	if err := c.ShouldBindUri(&req); err != nil {
		response.AbortError(c, errno.ParamInvalid.WithMsg(err.Error()))
		return
	}

	// 2. 调用业务层
	user, err := h.userService.GetByUsername(context.Background(), req.Username)
	if err != nil {
		response.AbortError(c, errno.UserNotFound.WithMsg(err.Error()))
		return
	}

	// 3. 返回响应
	response.Success(c, user)
}

/** Login 登录
 * @Summary 用户登录
 * @Description 用户登录获取session ID
 * @Tags 用户模块
 * @Accept json
 * @Produce json
 * @Param data body request.LoginReq true "登录信息"
 * @Success 200 {object} response.Response{data=map[string]string}
 * @Failure 10001 {object} response.Response "参数无效"
 * @Failure 20001 {object} response.Response "用户名或密码错误"
 * @Router /api/login [post]
 **/
func (h *UserHandler) Login(c *gin.Context) {
	// 1. 绑定并校验参数
	var req request.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortError(c, errno.ParamInvalid.WithMsg(err.Error()))
		return
	}

	// 2. 调用业务层登录
	sessionID, err := h.userService.Login(context.Background(), req.Username)
	if err != nil {
		response.AbortError(c, errno.Unauthorized.WithMsg(err.Error()))
		return
	}

	// 3. 返回成功响应
	response.Success(c, gin.H{
		"sid":      sessionID,
		"username": req.Username,
	})
}

/** Logout 登出
 * @Summary 用户登出
 * @Description 用户登出，清除session
 * @Tags 用户模块
 * @Accept json
 * @Produce json
 * @Param Authorization header string true "Bearer session_id"
 * @Success 200 {object} response.Response
 * @Failure 10001 {object} response.Response "参数无效"
 * @Failure 20001 {object} response.Response "未授权"
 * @Router /api/logout [post]
 **/
func (h *UserHandler) Logout(c *gin.Context) {
	// 1. 从请求头获取session ID
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.AbortError(c, errno.Unauthorized.WithMsg("missing authorization header"))
		return
	}

	// 2. 简单处理：Bearer session_id
	sessionID := authHeader[7:]
	if sessionID == "" {
		response.AbortError(c, errno.Unauthorized.WithMsg("invalid authorization header"))
		return
	}

	// 3. 调用业务层登出
	// 注意：这里需要从session中获取用户名，后续需要完善session管理
	// 暂时简化处理，假设用户名在session中可以直接获取
	username, exists := h.userService.GetUsernameBySessionID(sessionID)
	if !exists {
		response.AbortError(c, errno.Unauthorized.WithMsg("invalid session"))
		return
	}
	err := h.userService.Logout(context.Background(), username)
	if err != nil {
		response.AbortError(c, errno.ServerError.WithMsg(err.Error()))
		return
	}

	// 4. 返回成功响应
	response.Success(c, nil)
}

// GetUsers 查询用户列表
// @Summary 查询用户列表
// @Description 查询用户列表，支持过滤在线用户
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param online query bool false "是否在线"
// @Success 200 {object} response.Response{data=[]model.User}
// @Failure 10001 {object} response.Response "参数无效"
// @Router /api/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	// 1. 获取查询参数
	online := c.Query("online")

	// 2. 调用业务层
	var users response.UserListResponse
	var err error

	if online == "true" {
		// 获取在线用户
		users, err = h.userService.GetOnlineUsers(context.Background())
	} else {
		users, err = h.userService.GetAllUsers(context.Background())
	}

	if err != nil {
		response.AbortError(c, errno.ServerError.WithMsg(err.Error()))
		return
	}

	// 3. 返回响应
	response.Success(c, users)
}
