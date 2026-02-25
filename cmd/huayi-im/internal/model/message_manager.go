package model

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/pkg/logger"
	"go.uber.org/zap"
)

// MessageManager 消息管理器
type MessageManager struct {
	connections     map[string]*websocket.Conn
	offlineMessages map[string][]*OfflineMessage
	topicManager    *TopicManager
	mutex           sync.RWMutex
	connMutex       sync.RWMutex
}

// NewMessageManager 创建消息管理器实例
func NewMessageManager(topicManager *TopicManager) *MessageManager {
	return &MessageManager{
		connections:     make(map[string]*websocket.Conn),
		offlineMessages: make(map[string][]*OfflineMessage),
		topicManager:    topicManager,
	}
}

// RegisterConnection 注册连接
func (mm *MessageManager) RegisterConnection(username string, conn *websocket.Conn) {
	mm.connMutex.Lock()
	mm.connections[username] = conn
	mm.connMutex.Unlock()

	// 推送离线消息
	mm.pushOfflineMessages(username, conn)
}

// UnregisterConnection 注销连接
func (mm *MessageManager) UnregisterConnection(username string) {
	mm.connMutex.Lock()
	conn := mm.connections[username]
	if conn != nil {
		// 关闭连接
		conn.Close()
	}
	delete(mm.connections, username)
	mm.connMutex.Unlock()
}

// SendMessage 发送消息
func (mm *MessageManager) SendMessage(msg *Message) error {
	// 单聊消息
	if msg.Topic == "" && len(msg.To) > 0 {
		return mm.sendPrivateMessage(msg)
	}

	// 群聊消息
	if msg.Topic != "" {
		return mm.sendTopicMessage(msg)
	}

	return nil
}

// sendPrivateMessage 发送单聊消息
func (mm *MessageManager) sendPrivateMessage(msg *Message) error {
	for _, recipient := range msg.To {
		// 检查接收者是否在线
		mm.connMutex.RLock()
		conn, exists := mm.connections[recipient]
		mm.connMutex.RUnlock()

		if exists {
			// 在线，直接发送
			if err := conn.WriteJSON(msg); err != nil {
				logger.Error("发送单聊消息失败:", zap.Error(err), zap.String("to", recipient))
				continue
			}
			logger.Info("发送单聊消息成功:", zap.String("from", msg.From), zap.String("to", recipient))
		} else {
			// 离线，保存离线消息
			mm.saveOfflineMessage(recipient, msg)
		}
	}

	return nil
}

// sendTopicMessage 发送群聊消息
func (mm *MessageManager) sendTopicMessage(msg *Message) error {
	// 获取Topic的所有用户
	users, exists := mm.topicManager.GetTopicUsers(msg.Topic)
	if !exists {
		// 如果Topic不存在，创建它
		mm.topicManager.CreateTopic(msg.Topic)
		// 发送者自动加入Topic
		mm.topicManager.AddUserToTopic(msg.Topic, msg.From)
		users = []string{msg.From}
	}

	for _, user := range users {
		if user == msg.From {
			continue // 跳过发送者自己
		}

		// 检查用户是否在线
		mm.connMutex.RLock()
		conn, exists := mm.connections[user]
		mm.connMutex.RUnlock()

		if exists {
			// 在线，直接发送
			if err := conn.WriteJSON(msg); err != nil {
				logger.Error("发送群聊消息失败:", zap.Error(err), zap.String("to", user), zap.String("topic", msg.Topic))
				continue
			}
			logger.Info("发送群聊消息成功:", zap.String("from", msg.From), zap.String("to", user), zap.String("topic", msg.Topic))
		} else {
			// 离线，保存离线消息
			mm.saveOfflineMessage(user, msg)
		}
	}

	return nil
}

// saveOfflineMessage 保存离线消息
func (mm *MessageManager) saveOfflineMessage(username string, msg *Message) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// 创建离线消息
	offlineMsg := &OfflineMessage{
		UserID:    username,
		Message:   msg,
		ExpiresAt: time.Now().Add(10 * time.Minute), // 10分钟过期
	}

	// 保存到用户的离线消息列表
	mm.offlineMessages[username] = append(mm.offlineMessages[username], offlineMsg)
	logger.Info("保存离线消息:", zap.String("to", username), zap.String("from", msg.From))

	// 清理过期离线消息
	mm.cleanupExpiredMessages()
}

// pushOfflineMessages 推送离线消息
func (mm *MessageManager) pushOfflineMessages(username string, conn *websocket.Conn) {
	mm.mutex.Lock()
	messages, exists := mm.offlineMessages[username]
	if !exists {
		mm.mutex.Unlock()
		return
	}

	// 清空离线消息列表
	delete(mm.offlineMessages, username)
	mm.mutex.Unlock()

	// 推送所有离线消息
	for _, offlineMsg := range messages {
		if time.Now().After(offlineMsg.ExpiresAt) {
			continue // 跳过过期消息
		}

		if err := conn.WriteJSON(offlineMsg.Message); err != nil {
			logger.Error("推送离线消息失败:", zap.Error(err), zap.String("to", username))
			continue
		}
		logger.Info("推送离线消息成功:", zap.String("to", username), zap.String("from", offlineMsg.Message.From))
	}
}

// cleanupExpiredMessages 清理过期离线消息
func (mm *MessageManager) cleanupExpiredMessages() {
	now := time.Now()

	for username, messages := range mm.offlineMessages {
		var validMessages []*OfflineMessage
		for _, msg := range messages {
			if now.Before(msg.ExpiresAt) {
				validMessages = append(validMessages, msg)
			}
		}

		if len(validMessages) == 0 {
			delete(mm.offlineMessages, username)
		} else {
			mm.offlineMessages[username] = validMessages
		}
	}
}

// CleanupExpiredMessages 定期清理过期离线消息（供外部调用）
func (mm *MessageManager) CleanupExpiredMessages() {
	mm.mutex.Lock()
	mm.cleanupExpiredMessages()
	mm.mutex.Unlock()
}

// GetConnection 获取连接
func (mm *MessageManager) GetConnection(username string) (*websocket.Conn, bool) {
	mm.connMutex.RLock()
	defer mm.connMutex.RUnlock()

	conn, exists := mm.connections[username]
	return conn, exists
}
