# CloudLiteSync

一个基于 Go 语言开发的 SQLite 数据库备份和恢复系统，支持阿里云OSS存储，提供Web管理界面和API接口。

## 功能特性

- 🔐 管理员登录认证
- 📁 项目管理（创建、编辑、删除）
- 🔑 凭证管理（生成、激活、停用、删除）
- 📊 数据库版本管理
- ☁️ 阿里云OSS文件存储
- 🌐 Web管理界面（使用TailwindCSS）
- 🔌 RESTful API接口
- 📄 分页显示
- 🔒 基于Token的第三方访问

## 技术栈

- **后端**: Go 1.24.4
- **路由**: Chi Router
- **数据库**: SQLite
- **存储**: 阿里云OSS
- **前端**: HTML + TailwindCSS
- **会话**: Gorilla Sessions

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
```

### 3. 运行服务器

```bash
go run cmd/server/main.go
```

### 4. 访问系统

- 管理界面: http://localhost:8080
- 默认登录: admin / admin123

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

### 上传数据库文件

```bash
curl -X POST http://localhost:8080/api/{PROJ_ID} \
  -F "token=YOUR_TOKEN" \
  -F "description=版本描述" \
  -F "database=@/path/to/database.db"
```

### 下载数据库文件

```bash
# 下载最新版本
curl -O -J "http://localhost:8080/api/{PROJ_ID}/latest?token=YOUR_TOKEN"

# 下载指定版本
curl -O -J "http://localhost:8080/api/{PROJ_ID}/{HASH}?token=YOUR_TOKEN"
```

## 数据库表结构

### projects（项目表）
- id: 项目ID（8位字母）
- name: 项目名称
- description: 项目描述
- created_at: 创建时间
- updated_at: 更新时间

### credentials（凭证表）
- id: 凭证ID
- project_id: 项目ID
- token: 访问令牌
- is_active: 是否激活
- created_at: 创建时间
- updated_at: 更新时间

### database_versions（数据库版本表）
- id: 版本ID
- project_id: 项目ID
- version: 版本号
- file_hash: 文件哈希
- file_name: 文件名
- file_size: 文件大小
- oss_key: OSS存储键
- description: 版本描述
- is_latest: 是否最新版本
- created_at: 创建时间

## 使用流程

1. **管理员登录**: 使用配置的用户名和密码登录系统
2. **创建项目**: 在仪表板中创建新项目，系统会生成唯一的项目ID
3. **生成凭证**: 在项目详情页面为项目生成访问凭证
4. **第三方使用**: 第三方应用使用凭证通过API上传和下载数据库文件
5. **版本管理**: 系统自动管理数据库版本，支持下载最新版本或指定版本

## 注意事项

- 确保阿里云OSS配置正确，否则文件上传功能将不可用
- 建议在生产环境中修改默认的管理员密码
- 定期备份SQLite数据库文件
- 可以根据需要调整分页大小和文件上传限制