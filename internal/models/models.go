package models

import (
	"time"
)

// Project 项目模型
type Project struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Website     string    `json:"website" db:"website"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Credential 凭证模型
type Credential struct {
	ID        string    `json:"id" db:"id"`
	ProjectID string    `json:"project_id" db:"project_id"`
	Token     string    `json:"token" db:"token"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DatabaseVersion 数据库版本模型
type DatabaseVersion struct {
	ID          string    `json:"id" db:"id"`
	ProjectID   string    `json:"project_id" db:"project_id"`
	Version     string    `json:"version" db:"version"`
	FileHash    string    `json:"file_hash" db:"file_hash"`
	FileName    string    `json:"file_name" db:"file_name"`
	FileSize    int64     `json:"file_size" db:"file_size"`
	OSSKey      string    `json:"oss_key" db:"oss_key"`
	Description string    `json:"description" db:"description"`
	IsLatest    bool      `json:"is_latest" db:"is_latest"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Pagination 分页结构
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UploadRequest 上传请求
type UploadRequest struct {
	Token       string `json:"token"`
	Description string `json:"description"`
}

// DownloadRequest 下载请求
type DownloadRequest struct {
	Token   string `json:"token"`
	Version string `json:"version"` // 可选，不提供则下载最新版本
}

// JWTProject JWT项目模型
type JWTProject struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	PublicKey   string    `json:"public_key" db:"public_key"`
	PrivateKey  string    `json:"private_key" db:"private_key"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// JWTToken JWT令牌模型
type JWTToken struct {
	ID        string    `json:"id" db:"id"`
	ProjectID string    `json:"project_id" db:"project_id"`
	Purpose   string    `json:"purpose" db:"purpose"`
	Username  string    `json:"username" db:"username"`
	Role      string    `json:"role" db:"role"`
	Token     string    `json:"token" db:"token"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateJWTProjectRequest 创建JWT项目请求
type CreateJWTProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PublicKey   string `json:"public_key"`
	PrivateKey  string `json:"private_key"`
}

// CreateJWTTokenRequest 创建JWT令牌请求
type CreateJWTTokenRequest struct {
	ProjectID string    `json:"project_id"`
	Purpose   string    `json:"purpose"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
}

// UpdateJWTTokenRequest 更新JWT令牌请求
type UpdateJWTTokenRequest struct {
	ID        string    `json:"id"`
	Purpose   string    `json:"purpose"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	ExpiresAt time.Time `json:"expires_at"`
}
