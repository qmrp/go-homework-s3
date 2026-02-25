package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/api/response"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/manager"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/model"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/pkg/logger"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/service"
	"go.uber.org/zap"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// NewWSHandler 创建WebSocket处理器，接受UserService实例
func NewWSHandler(userService service.UserService) func(c *gin.Context) {
	return func(c *gin.Context) {
		// 从 query 参数获取 session ID
		sid := c.Query("sid")
		if sid == "" {
			c.JSON(402, gin.H{"error": "unauthorized"})
			return
		}

		// 验证session并获取用户名
		username, exists := userService.GetUsernameBySessionID(sid)
		if !exists {
			c.JSON(403, gin.H{"error": "invalid session"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("升级为 WebSocket 失败:", zap.Error(err))
			return
		}
		defer conn.Close()
		logger.Info("WebSocket 连接成功", zap.String("username", username), zap.String("sid", sid))

		// 注册连接到消息管理器
		manager.MessageManager.RegisterConnection(username, conn)
		defer func() {
			// 注销连接
			manager.MessageManager.UnregisterConnection(username)
			logger.Info("WebSocket 连接关闭", zap.String("username", username))
		}()

		done := make(chan struct{})

		go func() {
			defer close(done)
			for {
				conn.SetReadDeadline(time.Now().Add(65 * time.Second))
				messageType, message, err := conn.ReadMessage()

				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						logger.Error("WebSocket 读取消息错误:", zap.Error(err), zap.String("username", username))
					}
					return
				}

				switch messageType {
				case websocket.TextMessage:
					logger.Info("收到文本消息:", zap.String("username", username), zap.String("message", string(message)))

					var wsMsg response.WebSocketMessage
					if err := json.Unmarshal(message, &wsMsg); err != nil {
						logger.Error("解析WebSocket消息失败:", zap.Error(err), zap.String("username", username))
						break
					}

					switch wsMsg.MessageType {
					case "ping":
						ackMsg := response.WebSocketMessage{
							MessageType: "ack",
							AckID:       wsMsg.MessageID,
						}
						if err := conn.WriteJSON(ackMsg); err != nil {
							logger.Error("发送ack失败:", zap.Error(err), zap.String("username", username))
							return
						}
					case "ack":
						conn.SetReadDeadline(time.Now().Add(60 * time.Second))
						user, err := userService.GetByUsername(c, username)
						if err != nil {
							logger.Error("获取用户失败:", zap.Error(err), zap.String("username", username))
							return
						}
						if wsMsg.AckID == user.LastAskId {
							userService.SetNonResponseCount(c, username, 0)
						}
					case "message":
						logger.Info("收到普通消息:", zap.String("from", wsMsg.From), zap.String("content", wsMsg.Content))

						msg := &model.Message{
							From:        wsMsg.From,
							To:          wsMsg.To,
							Topic:       wsMsg.Topic,
							ContentType: wsMsg.ContentType,
							Content:     wsMsg.Content,
							MessageType: "message",
							CreatedAt:   time.Now(),
						}

						if err := manager.MessageManager.SendMessage(msg); err != nil {
							logger.Error("发送消息失败:", zap.Error(err), zap.String("from", wsMsg.From))
						}
					}
				}
			}
		}()

		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				pongMsg := response.WebSocketMessage{
					MessageID:   time.Now().UnixNano() / 1000000,
					From:        "server",
					To:          []string{username},
					MessageType: "pong",
				}
				if err := conn.WriteJSON(pongMsg); err != nil {
					logger.Error("发送心跳包失败:", zap.Error(err), zap.String("username", username))
					return
				}
				time.Sleep(500 * time.Microsecond) // 500ms 间隔
				user, err := userService.GetByUsername(c, username)
				if err != nil {
					logger.Error("获取用户失败:", zap.Error(err), zap.String("username", username))
					return
				}
				userService.SetNonResponseCount(c, username, user.NonResponseCount+1)
				userService.SetLastAckId(c, username, pongMsg.MessageID)
				if user.NonResponseCount >= 3 {
					logger.Error("用户无响应次数超过3次，已被标记为非在线状态", zap.String("username", username))
					userService.SetOnlineStatus(c, username, false)
					userService.SetNonResponseCount(c, username, 0)
					conn.Close()
					return
				}
			}
		}
	}
}
