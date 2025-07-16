package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"io"

	"chchma.com/cloudlite-sync/internal/database"
	"chchma.com/cloudlite-sync/internal/models"
	"chchma.com/cloudlite-sync/internal/session"
	"chchma.com/cloudlite-sync/internal/template"
	"chchma.com/cloudlite-sync/internal/utils"
)

// Dashboard 显示仪表板
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize := 10

	projects, total, err := h.db.ListProjects(page, pageSize)
	if err != nil {
		http.Error(w, "Failed to get projects", http.StatusInternalServerError)
		return
	}

	data := template.NewPageData("仪表板", projects)
	data.SetUser(session.GetUsername(r))
	data.SetPagination(page, total, pageSize)

	h.tmpl.Render(w, "dashboard.html", data)
}

// CreateProject 创建项目
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	name := r.FormValue("name")
	description := r.FormValue("description")

	if name == "" {
		data := template.NewPageData("创建项目", nil)
		data.SetUser(session.GetUsername(r))
		data.SetError("项目名称不能为空")
		h.tmpl.Render(w, "create_project.html", data)
		return
	}

	// 校验自定义ID
	if id != "" {
		// 只允许字母、数字、下划线、短横线，长度1-32
		if len(id) > 32 {
			data := template.NewPageData("创建项目", nil)
			data.SetUser(session.GetUsername(r))
			data.SetError("项目ID不能超过32个字符")
			h.tmpl.Render(w, "create_project.html", data)
			return
		}
		for _, c := range id {
			if !(c >= 'a' && c <= 'z') && !(c >= 'A' && c <= 'Z') && !(c >= '0' && c <= '9') && c != '_' && c != '-' {
				data := template.NewPageData("创建项目", nil)
				data.SetUser(session.GetUsername(r))
				data.SetError("项目ID只能包含字母、数字、下划线和短横线")
				h.tmpl.Render(w, "create_project.html", data)
				return
			}
		}
		// 检查ID是否已存在
		proj, err := h.db.GetProject(id)
		if err != nil {
			data := template.NewPageData("创建项目", nil)
			data.SetUser(session.GetUsername(r))
			data.SetError("检查项目ID时出错: " + err.Error())
			h.tmpl.Render(w, "create_project.html", data)
			return
		}
		if proj != nil {
			data := template.NewPageData("创建项目", nil)
			data.SetUser(session.GetUsername(r))
			data.SetError("项目ID已存在，请更换")
			h.tmpl.Render(w, "create_project.html", data)
			return
		}
	}

	project := &models.Project{
		ID:          id,
		Name:        name,
		Description: description,
	}
	if id == "" {
		project.ID = database.GenerateProjectID()
	}

	err := h.db.CreateProject(project)
	if err != nil {
		data := template.NewPageData("创建项目", nil)
		data.SetUser(session.GetUsername(r))
		data.SetError("创建项目失败: " + err.Error())
		h.tmpl.Render(w, "create_project.html", data)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// UpdateProject 更新项目
func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectID := r.FormValue("id")
	name := r.FormValue("name")
	description := r.FormValue("description")

	if projectID == "" || name == "" {
		http.Error(w, "Project ID and name are required", http.StatusBadRequest)
		return
	}

	project, err := h.db.GetProject(projectID)
	if err != nil {
		http.Error(w, "Failed to get project", http.StatusInternalServerError)
		return
	}

	if project == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	project.Name = name
	project.Description = description

	err = h.db.UpdateProject(project)
	if err != nil {
		http.Error(w, "Failed to update project", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DeleteProject 删除项目
func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectID := r.FormValue("id")
	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	err := h.db.DeleteProject(projectID)
	if err != nil {
		http.Error(w, "Failed to delete project", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ProjectDetail 显示项目详情
func (h *Handler) ProjectDetail(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("id")
	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	project, err := h.db.GetProject(projectID)
	if err != nil {
		http.Error(w, "Failed to get project", http.StatusInternalServerError)
		return
	}

	if project == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// 获取项目的凭证列表
	credPage, _ := strconv.Atoi(r.URL.Query().Get("cred_page"))
	if credPage <= 0 {
		credPage = 1
	}
	credentials, credTotal, err := h.db.ListCredentials(projectID, credPage, 10)
	if err != nil {
		http.Error(w, "Failed to get credentials", http.StatusInternalServerError)
		return
	}

	// 获取项目的数据库版本列表
	versionPage, _ := strconv.Atoi(r.URL.Query().Get("version_page"))
	if versionPage <= 0 {
		versionPage = 1
	}
	versions, versionTotal, err := h.db.ListDatabaseVersions(projectID, versionPage, 10)
	if err != nil {
		http.Error(w, "Failed to get versions", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"project":      project,
		"credentials":  credentials,
		"versions":     versions,
		"credTotal":    credTotal,
		"versionTotal": versionTotal,
		"credPage":     credPage,
		"versionPage":  versionPage,
	}

	pageData := template.NewPageData("项目详情", data)
	pageData.SetUser(session.GetUsername(r))

	errorMsg := r.URL.Query().Get("error")
	if errorMsg != "" {
		pageData.SetError(errorMsg)
	}

	h.tmpl.Render(w, "project_detail.html", pageData)
}

// UploadDatabaseVersion 网页端上传数据库版本
func (h *Handler) UploadDatabaseVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectID := r.FormValue("project_id")
	description := r.FormValue("description")
	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	// 校验项目是否存在
	project, err := h.db.GetProject(projectID)
	if err != nil || project == nil {
		http.Error(w, "Project not found", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("database")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	fileHash := utils.GenerateHash(fileData)
	// 检查文件是否已存在
	existingVersion, err := h.db.GetVersionByHash(projectID, fileHash)
	if err != nil {
		http.Error(w, "Failed to check existing version", http.StatusInternalServerError)
		return
	}
	if existingVersion != nil {
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=文件已存在", http.StatusSeeOther)
		return
	}

	version := utils.GenerateVersion()
	ossKey := utils.GenerateOSSKey(projectID, version, header.Filename)
	// 上传到OSS
	err = h.ossClient.UploadFile(ossKey, fileData)
	if err != nil {
		http.Error(w, "Failed to upload file to OSS", http.StatusInternalServerError)
		return
	}

	dbVersion := &models.DatabaseVersion{
		ID:          utils.GenerateUUID(),
		ProjectID:   projectID,
		Version:     version,
		FileHash:    fileHash,
		FileName:    header.Filename,
		FileSize:    header.Size,
		OSSKey:      ossKey,
		Description: description,
		IsLatest:    true,
	}
	err = h.db.CreateDatabaseVersion(dbVersion)
	if err != nil {
		h.ossClient.DeleteFile(ossKey)
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=保存版本记录失败", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/project/detail?id="+projectID, http.StatusSeeOther)
}

// DeleteDatabaseVersion 网页端删除数据库版本
func (h *Handler) DeleteDatabaseVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	versionID := r.FormValue("id")
	projectID := r.FormValue("project_id")
	if versionID == "" || projectID == "" {
		http.Error(w, "Version ID and Project ID are required", http.StatusBadRequest)
		return
	}
	// 获取版本信息
	version, err := h.db.GetDatabaseVersion(versionID)
	if err != nil || version == nil {
		http.Error(w, "Version not found", http.StatusBadRequest)
		return
	}
	// 删除OSS文件
	h.ossClient.DeleteFile(version.OSSKey)
	// 删除数据库记录
	err = h.db.DeleteDatabaseVersion(versionID)
	if err != nil {
		http.Error(w, "Failed to delete version", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/project/detail?id="+projectID, http.StatusSeeOther)
}

// ProjectDownload 数据库文件下载（session鉴权）
func (h *Handler) ProjectDownload(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	hash := r.URL.Query().Get("hash")
	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}
	// session鉴权
	username := session.GetUsername(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	var dbVersion *models.DatabaseVersion
	var err error
	if hash == "" {
		dbVersion, err = h.db.GetLatestVersion(projectID)
	} else {
		dbVersion, err = h.db.GetVersionByHash(projectID, hash)
	}
	if err != nil {
		http.Error(w, "Failed to get version", http.StatusInternalServerError)
		return
	}
	if dbVersion == nil {
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}
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

func (h *Handler) HelpPage(w http.ResponseWriter, r *http.Request) {
	data := template.NewPageData("帮助", nil)
	data.SetUser(session.GetUsername(r))
	h.tmpl.Render(w, "help.html", data)
}
