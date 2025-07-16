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
