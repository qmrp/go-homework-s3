package model

import (
	"time"
)

// Message 消息模型
type Message struct {
	ID          uint64    `json:"id"`
	From        string    `json:"from"`
	To          []string  `json:"to,omitempty"`
	Topic       string    `json:"topic,omitempty"`
	ContentType string    `json:"content-type"`
	Content     string    `json:"content"`
	MessageType string    `json:"message-type"`
	CreatedAt   time.Time `json:"created-at"`
}

// OfflineMessage 离线消息模型
type OfflineMessage struct {
	UserID    string    `json:"user_id"`
	Message   *Message  `json:"message"`
	ExpiresAt time.Time `json:"expires_at"`
}
