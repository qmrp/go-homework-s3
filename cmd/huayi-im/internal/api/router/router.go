package router

import (
	"github.com/gin-gonic/gin"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/api/handler"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/service/impl"
	"gorm.io/gorm"
)

// Register 注册所有路由
func Register(r *gin.Engine, dbInstance *gorm.DB) {
	// 初始化依赖（实际项目建议用 wire 依赖注入）
	// 使用基于内存的用户服务，不依赖数据库
	userService := impl.NewInMemoryUserService()
	userHandler := handler.NewUserHandler(userService)

	// 初始化其他处理器
	messageHandler := handler.NewMessageHandler(userService)
	topicHandler := handler.NewTopicHandler(userService)

	wsHandler := handler.NewWSHandler(userService)

	// 健康检查路由
	r.GET("/api/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API 版本分组
	api := r.Group("/api/")
	{
		// 登录/登出
		api.POST("/login", userHandler.Login)   // 登录
		api.POST("/logout", userHandler.Logout) // 登出

		// 消息模块路由
		api.POST("/messages", messageHandler.SendMessage) // 发送消息

		// 话题模块路由
		topicGroup := api.Group("/topics")
		{
			topicGroup.GET("", topicHandler.GetTopics)                      // 获取topic列表
			topicGroup.POST("", topicHandler.CreateTopic)                   // 创建topic
			topicGroup.DELETE("/:topic", topicHandler.DeleteTopic)          // 删除topic
			topicGroup.POST("/:topic/actions/join", topicHandler.JoinTopic) // 显式加入topic
			topicGroup.POST("/:topic/actions/quit", topicHandler.QuitTopic) // 显式退出topic
		}

		// 用户模块路由
		userGroup := api.Group("/users")
		{
			userGroup.GET("", userHandler.GetUsers)             // 查询用户列表
			userGroup.GET("/:user_id", userHandler.GetUserByID) // 获取用户详情
		}

		// WebSocket 路由
		api.GET("/ws", wsHandler)
	}
}
