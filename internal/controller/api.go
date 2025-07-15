package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"chchma.com/cloudlite-sync/internal/models"
	"chchma.com/cloudlite-sync/internal/utils"
	"github.com/go-chi/chi/v5"
)

// ApiUploadDatabase 上传数据库文件
func (h *Handler) ApiUploadDatabase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析multipart表单
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	token := r.FormValue("token")
	description := r.FormValue("description")

	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// 验证凭证
	credential, err := h.db.GetCredentialByToken(token)
	if err != nil {
		http.Error(w, "Failed to validate token", http.StatusInternalServerError)
		return
	}
	if credential == nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// 获取上传的文件
	file, header, err := r.FormFile("database")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 读取文件内容
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// 生成文件哈希
	fileHash := utils.GenerateHash(fileData)

	// 检查文件是否已存在
	existingVersion, err := h.db.GetVersionByHash(credential.ProjectID, fileHash)
	if err != nil {
		http.Error(w, "Failed to check existing version", http.StatusInternalServerError)
		return
	}
	if existingVersion != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "File already exists",
			"version": existingVersion,
		})
		return
	}

	// 生成版本信息
	version := utils.GenerateVersion()
	ossKey := utils.GenerateOSSKey(credential.ProjectID, version, header.Filename)

	// 上传到OSS
	err = h.ossClient.UploadFile(ossKey, fileData)
	if err != nil {
		http.Error(w, "Failed to upload file to OSS", http.StatusInternalServerError)
		return
	}

	// 创建数据库版本记录
	dbVersion := &models.DatabaseVersion{
		ID:          utils.GenerateUUID(),
		ProjectID:   credential.ProjectID,
		Version:     version,
		FileHash:    fileHash,
		FileName:    header.Filename,
		FileSize:    header.Size,
		OSSKey:      ossKey,
		Description: description,
		IsLatest:    true, // 新上传的版本设为最新
	}

	err = h.db.CreateDatabaseVersion(dbVersion)
	if err != nil {
		// 如果数据库操作失败，删除已上传的文件
		h.ossClient.DeleteFile(ossKey)
		http.Error(w, "Failed to save version record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Database uploaded successfully",
		"version": dbVersion,
	})
}

// ApiDownloadDatabase 下载数据库文件
func (h *Handler) ApiDownloadDatabase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	version := r.URL.Query().Get("version") // 可选，不提供则下载最新版本

	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// 验证凭证
	credential, err := h.db.GetCredentialByToken(token)
	if err != nil {
		http.Error(w, "Failed to validate token", http.StatusInternalServerError)
		return
	}
	if credential == nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var dbVersion *models.DatabaseVersion

	if version == "" {
		// 下载最新版本
		dbVersion, err = h.db.GetLatestVersion(credential.ProjectID)
		if err != nil {
			http.Error(w, "Failed to get latest version", http.StatusInternalServerError)
			return
		}
		if dbVersion == nil {
			http.Error(w, "No database version found", http.StatusNotFound)
			return
		}
	} else {
		// 下载指定版本
		dbVersion, err = h.db.GetDatabaseVersion(version)
		if err != nil {
			http.Error(w, "Failed to get version", http.StatusInternalServerError)
			return
		}
		if dbVersion == nil {
			http.Error(w, "Version not found", http.StatusNotFound)
			return
		}
		// 验证版本是否属于该项目
		if dbVersion.ProjectID != credential.ProjectID {
			http.Error(w, "Version not found", http.StatusNotFound)
			return
		}
	}

	// 从OSS下载文件
	fileData, err := h.ossClient.DownloadFile(dbVersion.OSSKey)
	if err != nil {
		http.Error(w, "Failed to download file from OSS", http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", dbVersion.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(dbVersion.FileSize, 10))

	// 写入文件内容
	w.Write(fileData)
}

// ApiListVersions 获取版本列表
func (h *Handler) ApiListVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}

	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// 验证凭证
	credential, err := h.db.GetCredentialByToken(token)
	if err != nil {
		http.Error(w, "Failed to validate token", http.StatusInternalServerError)
		return
	}
	if credential == nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// 获取版本列表
	versions, total, err := h.db.ListDatabaseVersions(credential.ProjectID, page, pageSize)
	if err != nil {
		http.Error(w, "Failed to get versions", http.StatusInternalServerError)
		return
	}

	response := models.PaginatedResponse{
		Data: versions,
		Pagination: models.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ApiGetVersionInfo 获取版本信息
func (h *Handler) ApiGetVersionInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	projectID := chi.URLParam(r, "projectID")
	hash := chi.URLParam(r, "hash")

	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// 验证凭证
	credential, err := h.db.GetCredentialByToken(token)
	if err != nil {
		http.Error(w, "Failed to validate token", http.StatusInternalServerError)
		return
	}
	if credential == nil || credential.ProjectID != projectID {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var dbVersion *models.DatabaseVersion

	if hash == "latest" {
		// 获取最新版本信息
		dbVersion, err = h.db.GetLatestVersion(credential.ProjectID)
	} else {
		// 获取指定版本信息
		dbVersion, err = h.db.GetVersionByHash(projectID, hash)
	}

	if err != nil {
		http.Error(w, "Failed to get version info", http.StatusInternalServerError)
		return
	}
	if dbVersion == nil {
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"version": dbVersion,
	})
}

// API获取项目列表
func (h *Handler) APIListProjects(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}

	projects, total, err := h.db.ListProjects(page, pageSize)
	if err != nil {
		http.Error(w, "Failed to get projects", http.StatusInternalServerError)
		return
	}

	response := models.PaginatedResponse{
		Data: projects,
		Pagination: models.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 下载最新版本
func (h *Handler) ApiDownloadLatest(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	token := r.URL.Query().Get("token")
	if projectID == "" {
		http.Error(w, "ProjectID is required", http.StatusBadRequest)
		return
	}
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}
	// 验证凭证
	credential, err := h.db.GetCredentialByToken(token)
	if err != nil {
		http.Error(w, "Failed to validate token", http.StatusInternalServerError)
		return
	}
	if credential == nil || credential.ProjectID != projectID {
		http.Error(w, "Invalid token or project", http.StatusUnauthorized)
		return
	}
	// 获取最新版本
	dbVersion, err := h.db.GetLatestVersion(projectID)
	if err != nil {
		http.Error(w, "Failed to get latest version", http.StatusInternalServerError)
		return
	}
	if dbVersion == nil {
		http.Error(w, "No database version found", http.StatusNotFound)
		return
	}
	// 下载文件
	fileData, err := h.ossClient.DownloadFile(dbVersion.OSSKey)
	if err != nil {
		http.Error(w, "Failed to download file from OSS", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", dbVersion.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(dbVersion.FileSize, 10))
	w.Write(fileData)
}

// 下载指定 hash 版本
func (h *Handler) ApiDownloadByHash(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	hash := chi.URLParam(r, "hash")
	token := r.URL.Query().Get("token")
	if projectID == "" || hash == "" {
		http.Error(w, "ProjectID and hash are required", http.StatusBadRequest)
		return
	}
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}
	// 验证凭证
	credential, err := h.db.GetCredentialByToken(token)
	if err != nil {
		http.Error(w, "Failed to validate token", http.StatusInternalServerError)
		return
	}
	if credential == nil || credential.ProjectID != projectID {
		http.Error(w, "Invalid token or project", http.StatusUnauthorized)
		return
	}
	// 获取指定 hash 版本
	dbVersion, err := h.db.GetVersionByHash(projectID, hash)
	if err != nil {
		http.Error(w, "Failed to get version by hash", http.StatusInternalServerError)
		return
	}
	if dbVersion == nil {
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}
	// 下载文件
	fileData, err := h.ossClient.DownloadFile(dbVersion.OSSKey)
	if err != nil {
		http.Error(w, "Failed to download file from OSS", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", dbVersion.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(dbVersion.FileSize, 10))
	w.Write(fileData)
}
