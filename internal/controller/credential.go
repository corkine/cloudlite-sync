package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"chchma.com/cloudlite-sync/internal/database"
	"chchma.com/cloudlite-sync/internal/models"
	"chchma.com/cloudlite-sync/internal/utils"
)

// CreateCredential 创建凭证
func (h *Handler) CreateCredential(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectID := r.FormValue("project_id")
	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	// 验证项目是否存在
	project, err := h.db.GetProject(projectID)
	if err != nil {
		http.Error(w, "Failed to get project", http.StatusInternalServerError)
		return
	}
	if project == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// 创建凭证
	credential := &models.Credential{
		ID:        utils.GenerateUUID(),
		ProjectID: projectID,
		Token:     database.GenerateToken(),
		IsActive:  true,
	}

	err = h.db.CreateCredential(credential)
	if err != nil {
		http.Error(w, "Failed to create credential", http.StatusInternalServerError)
		return
	}

	// 重定向回项目详情页面
	http.Redirect(w, r, "/project/detail?id="+projectID, http.StatusSeeOther)
}

// DeleteCredential 删除凭证
func (h *Handler) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	credentialID := r.FormValue("id")
	projectID := r.FormValue("project_id")

	if credentialID == "" {
		http.Error(w, "Credential ID is required", http.StatusBadRequest)
		return
	}

	err := h.db.DeleteCredential(credentialID)
	if err != nil {
		http.Error(w, "Failed to delete credential", http.StatusInternalServerError)
		return
	}

	// 重定向回项目详情页面
	http.Redirect(w, r, "/project/detail?id="+projectID, http.StatusSeeOther)
}

// DeactivateCredential 停用凭证
func (h *Handler) DeactivateCredential(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	credentialID := r.FormValue("id")
	projectID := r.FormValue("project_id")

	if credentialID == "" {
		http.Error(w, "Credential ID is required", http.StatusBadRequest)
		return
	}

	err := h.db.DeactivateCredential(credentialID)
	if err != nil {
		http.Error(w, "Failed to deactivate credential", http.StatusInternalServerError)
		return
	}

	// 重定向回项目详情页面
	http.Redirect(w, r, "/project/detail?id="+projectID, http.StatusSeeOther)
}

// ActivateCredential 激活凭证
func (h *Handler) ActivateCredential(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	credentialID := r.FormValue("id")
	projectID := r.FormValue("project_id")

	if credentialID == "" {
		http.Error(w, "Credential ID is required", http.StatusBadRequest)
		return
	}

	credential, err := h.db.GetCredential(credentialID)
	if err != nil {
		http.Error(w, "Failed to get credential", http.StatusInternalServerError)
		return
	}

	if credential == nil {
		http.Error(w, "Credential not found", http.StatusNotFound)
		return
	}

	credential.IsActive = true
	err = h.db.UpdateCredential(credential)
	if err != nil {
		http.Error(w, "Failed to activate credential", http.StatusInternalServerError)
		return
	}

	// 重定向回项目详情页面
	http.Redirect(w, r, "/project/detail?id="+projectID, http.StatusSeeOther)
}

// API获取凭证列表
func (h *Handler) APIListCredentials(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}

	credentials, total, err := h.db.ListCredentials(projectID, page, pageSize)
	if err != nil {
		http.Error(w, "Failed to get credentials", http.StatusInternalServerError)
		return
	}

	response := models.PaginatedResponse{
		Data: credentials,
		Pagination: models.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// API创建凭证
func (h *Handler) APICreateCredential(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProjectID string `json:"project_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ProjectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	// 验证项目是否存在
	project, err := h.db.GetProject(req.ProjectID)
	if err != nil {
		http.Error(w, "Failed to get project", http.StatusInternalServerError)
		return
	}
	if project == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// 创建凭证
	credential := &models.Credential{
		ID:        utils.GenerateUUID(),
		ProjectID: req.ProjectID,
		Token:     database.GenerateToken(),
		IsActive:  true,
	}

	err = h.db.CreateCredential(credential)
	if err != nil {
		http.Error(w, "Failed to create credential", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    credential,
	})
}

// API删除凭证
func (h *Handler) APIDeleteCredential(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	credentialID := r.URL.Query().Get("id")
	if credentialID == "" {
		http.Error(w, "Credential ID is required", http.StatusBadRequest)
		return
	}

	err := h.db.DeleteCredential(credentialID)
	if err != nil {
		http.Error(w, "Failed to delete credential", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Credential deleted successfully",
	})
}
