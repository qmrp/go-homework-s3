package handler

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/api/response"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/manager"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/model"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/pkg/errno"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/pkg/utils"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/service"
)

// MessageHandler 消息处理器
type MessageHandler struct {
	userService service.UserService
}

// NewMessageHandler 创建消息处理器实例
func NewMessageHandler(userService service.UserService) *MessageHandler {
	return &MessageHandler{
		userService: userService,
	}
}

// SendMessage 发送消息
// @Summary 发送消息
// @Description 发送单聊或群聊消息
// @Tags 消息模块
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer session_id"
// @Param data body model.Message true "消息内容"
// @Success 200 {object} response.Response
// @Failure 10001 {object} response.Response "参数无效"
// @Failure 20001 {object} response.Response "未授权"
// @Router /api/messages [post]
func (h *MessageHandler) SendMessage(c *gin.Context) {
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
	var msg model.Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		response.AbortError(c, errno.ParamInvalid.WithMsg(err.Error()))
		return
	}

	// 5. 设置发送者和创建时间
	msg.From = username
	msg.CreatedAt = time.Now()
	msg.MessageType = "message"
	if msg.Topic != "" {
		topic, exists := manager.TopicManager.GetTopic(msg.Topic)
		if !exists { //  topic 不存在，创建并添加发送者和接收者
			manager.TopicManager.CreateTopic(msg.Topic)
			manager.TopicManager.AddUserToTopic(msg.Topic, msg.From)
			for _, user := range msg.To {
				exists, _ := h.userService.ExistsByUsername(c, user)
				if exists {
					manager.TopicManager.AddUserToTopic(msg.Topic, user)
				}
			}
			users, _ := manager.TopicManager.GetTopicUsers(msg.Topic)
			log.Printf("topic %s created, users: %v", msg.Topic, users)
		} else { // topic 存在，检查发送者是否在 topic 中
			users := topic.Users
			log.Printf("topic %s exists, users: %v", msg.Topic, users)
			userMap := utils.SliceToMap(users)
			for _, user := range msg.To {
				if !userMap[user] { // 检查接收者是否在 topic 中 如果不在 则添加
					manager.TopicManager.AddUserToTopic(msg.Topic, user)
				}
			}
		}
	}

	// 6. 使用消息管理器发送消息
	if err := manager.MessageManager.SendMessage(&msg); err != nil {
		response.AbortError(c, errno.ServerError.WithMsg(err.Error()))
		return
	}

	// 7. 返回成功响应
	response.Success(c, nil)
}
