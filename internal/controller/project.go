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

	errorMsg := r.URL.Query().Get("error")
	if errorMsg != "" {
		data.SetError(errorMsg)
	}

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
	website := r.FormValue("website")

	if name == "" {
		data := template.NewPageData("创建项目", nil)
		data.SetUser(session.GetUsername(r))
		data.SetError("项目名称不能为空")
		http.Redirect(w, r, "/?error=项目名称不能为空", http.StatusSeeOther)
		return
	}

	// 校验自定义ID
	if id != "" {
		// 只允许字母、数字、下划线、短横线，长度1-32
		if len(id) > 32 {
			data := template.NewPageData("创建项目", nil)
			data.SetUser(session.GetUsername(r))
			data.SetError("项目ID不能超过32个字符")
			http.Redirect(w, r, "/?error=项目ID不能超过32个字符", http.StatusSeeOther)
			return
		}
		for _, c := range id {
			if !(c >= 'a' && c <= 'z') && !(c >= 'A' && c <= 'Z') && !(c >= '0' && c <= '9') && c != '_' && c != '-' {
				data := template.NewPageData("创建项目", nil)
				data.SetUser(session.GetUsername(r))
				data.SetError("项目ID只能包含字母、数字、下划线和短横线")
				http.Redirect(w, r, "/?error=项目ID只能包含字母、数字、下划线和短横线", http.StatusSeeOther)
				return
			}
		}
		// 检查ID是否已存在
		proj, err := h.db.GetProject(id)
		if err != nil {
			data := template.NewPageData("创建项目", nil)
			data.SetUser(session.GetUsername(r))
			data.SetError("检查项目ID时出错: " + err.Error())
			http.Redirect(w, r, "/?error=检查项目ID时出错: "+err.Error(), http.StatusSeeOther)
			return
		}
		if proj != nil {
			data := template.NewPageData("创建项目", nil)
			data.SetUser(session.GetUsername(r))
			data.SetError("项目ID已存在，请更换")
			http.Redirect(w, r, "/?error=项目ID已存在，请更换", http.StatusSeeOther)
			return
		}
	}

	if website != "" {
		if !utils.IsURL(website) {
			data := template.NewPageData("创建项目", nil)
			data.SetUser(session.GetUsername(r))
			data.SetError("网站URL格式不正确")
			http.Redirect(w, r, "/?error=网站URL格式不正确", http.StatusSeeOther)
			return
		}
	}

	project := &models.Project{
		ID:          id,
		Name:        name,
		Description: description,
		Website:     website,
	}
	if id == "" {
		project.ID = database.GenerateProjectID()
	}

	err := h.db.CreateProject(project)
	if err != nil {
		data := template.NewPageData("创建项目", nil)
		data.SetUser(session.GetUsername(r))
		data.SetError("创建项目失败: " + err.Error())
		http.Redirect(w, r, "/?error=创建项目失败: "+err.Error(), http.StatusSeeOther)
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
	website := r.FormValue("website")
	if projectID == "" || name == "" {
		http.Redirect(w, r, "/?error=项目ID和名称不能为空", http.StatusSeeOther)
		return
	}

	project, err := h.db.GetProject(projectID)
	if err != nil {
		http.Redirect(w, r, "/?error=获取项目失败", http.StatusSeeOther)
		return
	}

	if project == nil {
		http.Redirect(w, r, "/?error=项目不存在", http.StatusSeeOther)
		return
	}

	if website != "" {
		if !utils.IsURL(website) {
			http.Redirect(w, r, "/?error=网站URL格式不正确", http.StatusSeeOther)
			return
		}
	}

	project.Name = name
	project.Description = description
	project.Website = website
	err = h.db.UpdateProject(project)
	if err != nil {
		http.Redirect(w, r, "/?error=更新项目失败", http.StatusSeeOther)
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
		http.Redirect(w, r, "/?error=项目ID不能为空", http.StatusSeeOther)
		return
	}

	err := h.db.DeleteProject(projectID)
	if err != nil {
		http.Redirect(w, r, "/?error=删除项目失败", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ProjectDetail 显示项目详情
func (h *Handler) ProjectDetail(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("id")
	if projectID == "" {
		http.Redirect(w, r, "/?error=项目ID不能为空", http.StatusSeeOther)
		return
	}

	project, err := h.db.GetProject(projectID)
	if err != nil {
		http.Redirect(w, r, "/?error=获取项目失败", http.StatusSeeOther)
		return
	}

	if project == nil {
		http.Redirect(w, r, "/?error=项目不存在", http.StatusSeeOther)
		return
	}

	// 获取项目的凭证列表
	credPage, _ := strconv.Atoi(r.URL.Query().Get("cred_page"))
	if credPage <= 0 {
		credPage = 1
	}
	credentials, credTotal, err := h.db.ListCredentials(projectID, credPage, 10)
	if err != nil {
		http.Redirect(w, r, "/?error=获取凭证失败", http.StatusSeeOther)
		return
	}

	// 获取项目的数据库版本列表
	versionPage, _ := strconv.Atoi(r.URL.Query().Get("version_page"))
	if versionPage <= 0 {
		versionPage = 1
	}
	versions, versionTotal, err := h.db.ListDatabaseVersions(projectID, versionPage, 10)
	if err != nil {
		http.Redirect(w, r, "/?error=获取版本失败", http.StatusSeeOther)
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
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=项目ID不能为空", http.StatusSeeOther)
		return
	}

	// 校验项目是否存在
	project, err := h.db.GetProject(projectID)
	if err != nil || project == nil {
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=无法找到项目", http.StatusSeeOther)
		return
	}

	file, header, err := r.FormFile("database")
	if err != nil {
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=获取上传文件失败", http.StatusSeeOther)
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=读取文件失败", http.StatusSeeOther)
		return
	}

	fileHash := utils.GenerateHash(fileData)
	// 检查文件是否已存在
	existingVersion, err := h.db.GetVersionByHash(projectID, fileHash)
	if err != nil {
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=检查版本失败", http.StatusSeeOther)
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
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=上传文件失败", http.StatusSeeOther)
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
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=版本ID和项目ID不能为空", http.StatusSeeOther)
		return
	}
	// 获取版本信息
	version, err := h.db.GetDatabaseVersion(versionID)
	if err != nil || version == nil {
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=版本不存在", http.StatusSeeOther)
		return
	}
	// 删除OSS文件
	h.ossClient.DeleteFile(version.OSSKey)
	// 删除数据库记录
	err = h.db.DeleteDatabaseVersion(versionID)
	if err != nil {
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=删除版本失败", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/project/detail?id="+projectID, http.StatusSeeOther)
}

// ProjectDownload 数据库文件下载（session鉴权）
func (h *Handler) ProjectDownload(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	hash := r.URL.Query().Get("hash")
	if projectID == "" {
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=项目ID不能为空", http.StatusSeeOther)
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
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=获取版本失败", http.StatusSeeOther)
		return
	}
	if dbVersion == nil {
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=版本不存在", http.StatusSeeOther)
		return
	}
	fileData, err := h.ossClient.DownloadFile(dbVersion.OSSKey)
	if err != nil {
		http.Redirect(w, r, "/project/detail?id="+projectID+"&error=下载文件失败", http.StatusSeeOther)
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
