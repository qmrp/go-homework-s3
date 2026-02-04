package model

import (
	"time"
)

// User 用户模型（GORM）
type User struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username         string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	LastAskId        int64     `gorm:"default:0" json:"last_ask_id"`        // 最后一次请求ID
	NonResponseCount int       `gorm:"default:0" json:"non_response_count"` // 无响应次数
	Online           bool      `gorm:"default:false" json:"online"`         // 在线状态
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
