package manager

import (
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/model"
)

// 全局管理器实例
var (
	TopicManager   *model.TopicManager
	MessageManager *model.MessageManager
)

// Init 初始化管理器
func Init() {
	TopicManager = model.NewTopicManager()
	MessageManager = model.NewMessageManager(TopicManager)
}
