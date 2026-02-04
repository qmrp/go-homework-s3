# chat-s3

基于 Go Gin 框架的项目模板

## 项目说明
    本项目是一个基于 Go Gin 框架的IM项目，用户数据基于内存存储，不支持持久化。

## 项目架构
```
github.com/qmrp/go-homework-s3/cmd/huayi-im/
├── cmd/                     # 程序入口
├── internal/                # 私有业务代码（分层架构）
├── pkg/                     # 公共库（对外暴露）
├── configs/                 # 配置文件
├── migrations/              # 数据库迁移文件
├── tests/                   # 测试目录
└── README.md                # 项目说明
```

## 重点说明
    manager 全局管理器实例负现管理topic和message的创建、删除、查询等操作,以mutex保护并发访问。

## 快速开始

### 1. 环境准备
- Go 1.21+


### 2. 配置文件
复制配置示例文件并修改：
```bash
cp configs/config.dev.yaml.example configs/config.dev.yaml
# 编辑数据库、Redis 等配置
```

### 3. 安装依赖
```bash
go mod tidy
```

### 4. 启动服务
```bash
go run ./cmd/huayi-im/cmd/api/main.go
```

### 5. 接口测试
- 健康检查：http://localhost:8090/health
- 用户登录：POST http://localhost:8090/api/login

## 开发规范
1. 分层架构：API 层 -> Service 层  -> Model 层
2. 依赖注入：使用 wire 管理依赖，避免硬编码
3. 错误处理：使用统一错误码（internal/pkg/errno）
4. 日志规范：使用 Zap 结构化日志
5. 测试覆盖：重点覆盖 Service 和 Repository 层

