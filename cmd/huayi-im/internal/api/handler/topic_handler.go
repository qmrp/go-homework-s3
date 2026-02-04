package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/api/response"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/manager"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/pkg/errno"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/service"
)

// TopicHandler 话题处理器
type TopicHandler struct {
	userService service.UserService
}

// NewTopicHandler 创建话题处理器实例
func NewTopicHandler(userService service.UserService) *TopicHandler {
	return &TopicHandler{
		userService: userService,
	}
}

// GetTopics 获取话题列表
// @Summary 获取话题列表
// @Description 获取所有话题列表
// @Tags 话题模块
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer session_id"
// @Success 200 {object} response.Response{data=[]model.Topic}
// @Failure 20001 {object} response.Response "未授权"
// @Router /api/topics [get]
func (h *TopicHandler) GetTopics(c *gin.Context) {
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

	// 3. 验证session并获取用户名
	_, exists := h.userService.GetUsernameBySessionID(sessionID)
	if !exists {
		response.AbortError(c, errno.Unauthorized.WithMsg("invalid session"))
		return
	}

	// 4. 获取所有话题
	topics := manager.TopicManager.GetAllTopics()

	// 5. 返回响应
	response.Success(c, topics)
}

// CreateTopic 创建话题
// @Summary 创建话题
// @Description 创建新话题
// @Tags 话题模块
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer session_id"
// @Param data body map[string]string true "话题信息，包含name字段"
// @Success 200 {object} response.Response{data=model.Topic}
// @Failure 10001 {object} response.Response "参数无效"
// @Failure 20001 {object} response.Response "未授权"
// @Router /api/topics [post]
func (h *TopicHandler) CreateTopic(c *gin.Context) {
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

	// 3. 验证session并获取用户名
	username, exists := h.userService.GetUsernameBySessionID(sessionID)
	if !exists {
		response.AbortError(c, errno.Unauthorized.WithMsg("invalid session"))
		return
	}

	// 4. 绑定并校验参数
	var req struct {
		Name string `json:"name" binding:"required,min=1,max=50"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortError(c, errno.ParamInvalid.WithMsg(err.Error()))
		return
	}

	// 5. 创建话题
	topic := manager.TopicManager.CreateTopic(req.Name)

	// 6. 将创建者加入话题
	manager.TopicManager.AddUserToTopic(req.Name, username)

	// 7. 返回响应
	response.Success(c, topic)
}

// DeleteTopic 删除话题
// @Summary 删除话题
// @Description 删除指定话题
// @Tags 话题模块
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer session_id"
// @Param topic path string true "话题名称"
// @Success 200 {object} response.Response
// @Failure 10001 {object} response.Response "参数无效"
// @Failure 20001 {object} response.Response "未授权"
// @Router /api/topics/{topic} [delete]
func (h *TopicHandler) DeleteTopic(c *gin.Context) {
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

	// 3. 验证session并获取用户名
	_, exists := h.userService.GetUsernameBySessionID(sessionID)
	if !exists {
		response.AbortError(c, errno.Unauthorized.WithMsg("invalid session"))
		return
	}

	// 4. 获取话题名称
	topicName := c.Param("topic")
	if topicName == "" {
		response.AbortError(c, errno.ParamInvalid.WithMsg("topic name is required"))
		return
	}

	// 5. 删除话题
	success := manager.TopicManager.DeleteTopic(topicName)
	if !success {
		response.AbortError(c, errno.NotFound.WithMsg("topic not found"))
		return
	}

	// 6. 返回响应
	response.Success(c, nil)
}

// JoinTopic 加入话题
// @Summary 加入话题
// @Description 显式加入指定话题
// @Tags 话题模块
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer session_id"
// @Param topic path string true "话题名称"
// @Success 200 {object} response.Response
// @Failure 10001 {object} response.Response "参数无效"
// @Failure 20001 {object} response.Response "未授权"
// @Router /api/topics/{topic}/actions/join [post]
func (h *TopicHandler) JoinTopic(c *gin.Context) {
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

	// 3. 验证session并获取用户名
	username, exists := h.userService.GetUsernameBySessionID(sessionID)
	if !exists {
		response.AbortError(c, errno.Unauthorized.WithMsg("invalid session"))
		return
	}

	// 4. 获取话题名称
	topicName := c.Param("topic")
	if topicName == "" {
		response.AbortError(c, errno.ParamInvalid.WithMsg("topic name is required"))
		return
	}

	// 5. 加入话题
	manager.TopicManager.AddUserToTopic(topicName, username)

	// 6. 返回响应
	response.Success(c, nil)
}

// QuitTopic 退出话题
// @Summary 退出话题
// @Description 显式退出指定话题
// @Tags 话题模块
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer session_id"
// @Param topic path string true "话题名称"
// @Success 200 {object} response.Response
// @Failure 10001 {object} response.Response "参数无效"
// @Failure 20001 {object} response.Response "未授权"
// @Router /api/topics/{topic}/actions/quit [post]
func (h *TopicHandler) QuitTopic(c *gin.Context) {
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

	// 3. 验证session并获取用户名
	username, exists := h.userService.GetUsernameBySessionID(sessionID)
	if !exists {
		response.AbortError(c, errno.Unauthorized.WithMsg("invalid session"))
		return
	}

	// 4. 获取话题名称
	topicName := c.Param("topic")
	if topicName == "" {
		response.AbortError(c, errno.ParamInvalid.WithMsg("topic name is required"))
		return
	}

	// 5. 退出话题
	manager.TopicManager.RemoveUserFromTopic(topicName, username)

	// 6. 返回响应
	response.Success(c, nil)
}
