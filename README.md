# CloudLiteSync

一个基于 Go 语言开发的多功能数据管理平台，集成了 SQLite 数据库备份恢复系统和 JWT 令牌管理功能，支持阿里云OSS存储，提供Web管理界面和API接口。

## 功能特性

### 🔄 SQLite 数据管理
- 📁 项目管理（创建、编辑、删除）
- 🔑 凭证管理（生成、激活、停用、删除）
- 📊 数据库版本管理
- ☁️ 阿里云OSS文件存储
- 🔌 RESTful API接口
- 📄 分页显示
- 🔒 基于Token的第三方访问

### 🔐 JWT 令牌管理
- 🏗️ JWT项目创建和管理
- 🔑 RSA密钥对生成和管理
- 🎫 JWT令牌创建和验证
- 📤 分享码功能（6位数字码，可配置过期时间）
- 🔍 令牌状态监控和过期管理
- 📋 令牌信息查看和编辑

## 技术栈

- **后端**: Go 1.24.4
- **路由**: Chi Router
- **数据库**: SQLite
- **存储**: 阿里云OSS
- **前端**: HTML + TailwindCSS + Alpine.js
- **会话**: Gorilla Sessions
- **JWT**: golang-jwt/jwt/v5

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置系统

系统支持两种配置方式：

#### 方式一：使用 config.json 文件（推荐）

创建 `config.json` 文件：

```json
{
  "server": {
    "port": "8080",
    "host": "0.0.0.0"
  },
  "oss": {
    "endpoint": "https://oss-cn-hangzhou.aliyuncs.com",
    "access_key_id": "your_access_key_id",
    "access_key_secret": "your_access_key_secret",
    "bucket_name": "your_bucket_name"
  },
  "admin": {
    "username": "admin",
    "password": "admin123"
  },
  "session_secret": "your-session-secret-here",
  "share_code": {
    "expire_seconds": 30
  }
}
```

#### 方式二：使用环境变量

设置环境变量（会覆盖 config.json 中的配置）：

```bash
# 服务器配置
export PORT=8080
export HOST=0.0.0.0

# 阿里云OSS配置
export OSS_ENDPOINT=your-oss-endpoint
export OSS_ACCESS_KEY_ID=your-access-key-id
export OSS_ACCESS_KEY_SECRET=your-access-key-secret
export OSS_BUCKET_NAME=your-bucket-name

# 管理员配置
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=admin123

# 分享码配置
export SHARE_CODE_EXPIRE_SECONDS=30
```

### 3. 运行服务器

```bash
go run cmd/server/main.go
```

### 4. 访问系统

- 管理界面: http://localhost:8080
- 默认登录: admin / admin123

### 5. 编译 CSS

```bash
pnpm install
pnpm run build
```

## Docker 部署

### 1. 构建镜像

```bash
podman build -t corkine/cloudlite-sync:latest . --network=host
```

### 2. 运行容器

> 先创建 cloudlite 目录，然后在其中创建 config.json 配置，之后执行：

```bash
podman run -it --rm \
  -p 8080:8080 \
  -v $(pwd)/cloudlite/:/app/data/ \
  -v $(pwd)/cloudlite/config.json:/app/config.json \
  --name cloudlitesync \
  corkine/cloudlite-sync:latest
```

## API接口

### SQLite 数据管理 API

#### 上传数据库文件

```bash
curl -X POST http://localhost:8080/api/{PROJ_ID} \
  -F "token=YOUR_TOKEN" \
  -F "description=版本描述" \
  -F "database=@/path/to/database.db"
```

#### 下载数据库文件

```bash
# 下载最新版本
curl -O -J "http://localhost:8080/api/{PROJ_ID}/latest?token=YOUR_TOKEN"

# 下载指定版本
curl -O -J "http://localhost:8080/api/{PROJ_ID}/{HASH}?token=YOUR_TOKEN"
```

### JWT 令牌分享 API

#### 获取分享的令牌

```bash
curl http://localhost:8080/s/{SHARE_CODE}
```

**响应示例**：
```json
{
  "token": "eyJhbGciOiJSUzI1NiIs...",
  "success": true,
  "message": "密钥获取成功"
}
```

## 注意事项

- 确保阿里云OSS配置正确，否则文件上传功能将不可用
- 建议在生产环境中修改默认的管理员密码
- 定期备份SQLite数据库文件
- 分享码存储在内存中，服务重启后会丢失
- JWT私钥请妥善保管，不要泄露给他人
- 可以根据需要调整分页大小和文件上传限制