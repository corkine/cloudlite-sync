# 基于官方 Go 镜像
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git gcc musl-dev

# 开启 CGO
ENV CGO_ENABLED=1

# 复制 go.mod 和 go.sum 并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制项目源码
COPY . .

# 构建可执行文件
ARG APP_VERSION=latest
ENV APP_VERSION=${APP_VERSION}
RUN go build -o cloudlitesync ./cmd/server/main.go

# 生产环境镜像
FROM alpine:latest
WORKDIR /app

# 安装 ca-certificates 以支持 https
RUN apk add --no-cache ca-certificates

# 拷贝编译好的二进制文件
COPY --from=builder /app/cloudlitesync .

# 拷贝静态资源和模板
COPY static ./static
COPY templates ./templates

# 暴露端口
EXPOSE 8080

# 启动服务
CMD ["./cloudlitesync"] 