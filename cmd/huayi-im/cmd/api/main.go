package main

import (
	"github.com/gin-gonic/gin"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/api/router"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/config"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/manager"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/pkg/logger"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/pkg/middleware"
)

func main() {
	// 1. 加载配置
	cfg := config.Load()

	// 2. 初始化日志
	logger.Init(cfg.Log)
	logger.Info("日志初始化成功")

	// 3. 初始化全局管理器
	manager.Init()
	logger.Info("管理器初始化成功")

	// 4. 初始化 Gin 引擎
	r := gin.Default()
	// 注册跨域中间件（可选）
	r.Use(middleware.CorsMiddleware())
	// 5. 注册路由
	// 传递 nil 作为 dbInstance，因为我们使用基于内存的用户服务
	router.Register(r, nil)
	logger.Info("路由注册成功")

	// 6. 启动服务
	logger.Info("服务启动中，监听地址：", logger.Field("addr", cfg.Server.Addr))
	if err := r.Run(cfg.Server.Addr); err != nil {
		logger.Fatal("服务启动失败", logger.Field("error", err))
	}
}
