# 第一阶段：构建阶段
FROM golang:1.25.3-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o im-server ./cmd/huayi-im/cmd/api

# 第二阶段：运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 复制构建好的应用
COPY --from=builder /app/im-server .

# 复制配置文件
COPY cmd/huayi-im/configs/config.dev.yaml ./configs/

# 暴露端口（根据配置文件中的端口设置）
EXPOSE 8090

# 启动应用
CMD ["./im-server"]