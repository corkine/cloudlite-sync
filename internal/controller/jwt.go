package controller

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"chchma.com/cloudlite-sync/internal/database"
	"chchma.com/cloudlite-sync/internal/models"
	"chchma.com/cloudlite-sync/internal/utils"
)

type JWTController struct {
	db *database.DB
}

func NewJWTController(db *database.DB) *JWTController {
	return &JWTController{db: db}
}

// JWT项目相关API

func (c *JWTController) CreateJWTProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	publicKey := r.FormValue("public_key")
	privateKey := r.FormValue("private_key")

	if name == "" {
		http.Redirect(w, r, "/jwt?error=项目名称不能为空", http.StatusSeeOther)
		return
	}

	// 如果公钥和私钥都为空，则自动生成密钥对
	if publicKey == "" && privateKey == "" {
		var err error
		privateKey, publicKey, err = utils.GenerateRSAKeyPair(2048)
		if err != nil {
			http.Redirect(w, r, "/jwt?error=自动生成密钥对失败: "+err.Error(), http.StatusSeeOther)
			return
		}
	} else if publicKey != "" && privateKey != "" {
		// 如果都提供了，则验证密钥对是否有效
		if err := utils.ValidateKeyPair(privateKey, publicKey); err != nil {
			http.Redirect(w, r, "/jwt?error=无效的密钥对: "+err.Error(), http.StatusSeeOther)
			return
		}
	} else {
		// 如果只提供了一个，则提示错误
		http.Redirect(w, r, "/jwt?error=请同时提供公钥和私钥，或者都留空让系统自动生成", http.StatusSeeOther)
		return
	}

	project := &models.JWTProject{
		ID:          database.GenerateProjectID(),
		Name:        name,
		Description: description,
		PublicKey:   publicKey,
		PrivateKey:  privateKey,
	}

	if err := c.db.CreateJWTProject(project); err != nil {
		http.Redirect(w, r, "/jwt?error=创建JWT项目失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/jwt", http.StatusSeeOther)
}

func (c *JWTController) GetJWTProject(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "缺少项目ID", http.StatusBadRequest)
		return
	}

	project, err := c.db.GetJWTProject(id)
	if err != nil {
		http.Error(w, "获取JWT项目失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    project,
	})
}

func (c *JWTController) ListJWTProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := c.db.ListJWTProjects()
	if err != nil {
		http.Error(w, "获取JWT项目列表失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    projects,
	})
}

func (c *JWTController) UpdateJWTProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	name := r.FormValue("name")
	description := r.FormValue("description")
	publicKey := r.FormValue("public_key")
	privateKey := r.FormValue("private_key")

	if id == "" || name == "" {
		http.Redirect(w, r, "/jwt?error=项目ID和名称不能为空", http.StatusSeeOther)
		return
	}

	// 验证密钥对是否有效
	if err := utils.ValidateKeyPair(privateKey, publicKey); err != nil {
		http.Redirect(w, r, "/jwt?error=无效的密钥对: "+err.Error(), http.StatusSeeOther)
		return
	}

	project := &models.JWTProject{
		ID:          id,
		Name:        name,
		Description: description,
		PublicKey:   publicKey,
		PrivateKey:  privateKey,
	}

	if err := c.db.UpdateJWTProject(project); err != nil {
		http.Redirect(w, r, "/jwt?error=更新JWT项目失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/jwt", http.StatusSeeOther)
}

func (c *JWTController) DeleteJWTProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Redirect(w, r, "/jwt?error=项目ID不能为空", http.StatusSeeOther)
		return
	}

	if err := c.db.DeleteJWTProject(id); err != nil {
		http.Redirect(w, r, "/jwt?error=删除JWT项目失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/jwt", http.StatusSeeOther)
}

// JWT令牌相关API

func (c *JWTController) CreateJWTToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectID := r.FormValue("project_id")
	purpose := r.FormValue("purpose")
	username := r.FormValue("username")
	role := r.FormValue("role")
	expiresAtStr := r.FormValue("expires_at")

	if projectID == "" || purpose == "" || username == "" || role == "" || expiresAtStr == "" {
		http.Redirect(w, r, "/jwt/detail?id="+projectID+"&error=所有字段都是必填的", http.StatusSeeOther)
		return
	}

	// 解析过期时间
	expiresAt, err := time.Parse("2006-01-02T15:04", expiresAtStr)
	if err != nil {
		// 尝试带秒的格式
		expiresAt, err = time.Parse("2006-01-02T15:04:05", expiresAtStr)
		if err != nil {
			http.Redirect(w, r, "/jwt/detail?id="+projectID+"&error=过期时间格式不正确", http.StatusSeeOther)
			return
		}
	}

	// 获取JWT项目
	project, err := c.db.GetJWTProject(projectID)
	if err != nil {
		http.Redirect(w, r, "/jwt/detail?id="+projectID+"&error=获取JWT项目失败", http.StatusSeeOther)
		return
	}
	if project == nil {
		http.Redirect(w, r, "/jwt/detail?id="+projectID+"&error=JWT项目不存在", http.StatusSeeOther)
		return
	}

	// 创建JWT管理器
	jwtManager, err := utils.NewJWTManager(project.PrivateKey, project.PublicKey)
	if err != nil {
		http.Redirect(w, r, "/jwt/detail?id="+projectID+"&error=创建JWT管理器失败", http.StatusSeeOther)
		return
	}

	// 生成真正的JWT令牌
	jwtToken, err := jwtManager.GenerateToken(username, role, purpose, expiresAt)
	if err != nil {
		http.Redirect(w, r, "/jwt/detail?id="+projectID+"&error=生成JWT令牌失败", http.StatusSeeOther)
		return
	}

	token := &models.JWTToken{
		ID:        database.GenerateJWTTokenID(),
		ProjectID: projectID,
		Purpose:   purpose,
		Username:  username,
		Role:      role,
		Token:     jwtToken,
		IsActive:  true,
		ExpiresAt: expiresAt,
	}

	if err := c.db.CreateJWTToken(token); err != nil {
		http.Redirect(w, r, "/jwt/detail?id="+projectID+"&error=创建JWT令牌失败", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/jwt/detail?id="+projectID, http.StatusSeeOther)
}

func (c *JWTController) GetJWTToken(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "缺少令牌ID", http.StatusBadRequest)
		return
	}

	token, err := c.db.GetJWTToken(id)
	if err != nil {
		http.Error(w, "获取JWT令牌失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    token,
	})
}

func (c *JWTController) ListJWTTokens(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		http.Error(w, "缺少项目ID", http.StatusBadRequest)
		return
	}

	tokens, err := c.db.ListJWTTokens(projectID)
	if err != nil {
		http.Error(w, "获取JWT令牌列表失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    tokens,
	})
}

func (c *JWTController) UpdateJWTToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 支持表单提交和JSON提交两种方式
	var token *models.JWTToken

	if r.Header.Get("Content-Type") == "application/json" {
		// JSON方式
		var req models.UpdateJWTTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "无效的请求数据", http.StatusBadRequest)
			return
		}

		token = &models.JWTToken{
			ID:        req.ID,
			Purpose:   req.Purpose,
			Username:  req.Username,
			Role:      req.Role,
			IsActive:  req.IsActive,
			ExpiresAt: req.ExpiresAt,
		}
	} else {
		// 表单方式
		id := r.FormValue("id")
		purpose := r.FormValue("purpose")
		username := r.FormValue("username")
		role := r.FormValue("role")
		isActiveStr := r.FormValue("is_active")
		expiresAtStr := r.FormValue("expires_at")

		if id == "" {
			http.Error(w, "缺少令牌ID", http.StatusBadRequest)
			return
		}

		// 解析is_active
		isActive := isActiveStr == "true"

		// 解析过期时间
		expiresAt, err := time.Parse("2006-01-02T15:04:05", expiresAtStr)
		if err != nil {
			http.Error(w, "过期时间格式不正确", http.StatusBadRequest)
			return
		}

		token = &models.JWTToken{
			ID:        id,
			Purpose:   purpose,
			Username:  username,
			Role:      role,
			IsActive:  isActive,
			ExpiresAt: expiresAt,
		}
	}

	if err := c.db.UpdateJWTToken(token); err != nil {
		if r.Header.Get("Content-Type") == "application/json" {
			http.Error(w, "更新JWT令牌失败: "+err.Error(), http.StatusInternalServerError)
		} else {
			// 获取令牌信息以获取项目ID用于重定向
			tokenInfo, _ := c.db.GetJWTToken(token.ID)
			projectID := ""
			if tokenInfo != nil {
				projectID = tokenInfo.ProjectID
			}
			http.Redirect(w, r, "/jwt/detail?id="+projectID+"&error=更新JWT令牌失败", http.StatusSeeOther)
		}
		return
	}

	if r.Header.Get("Content-Type") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "JWT令牌更新成功",
		})
	} else {
		// 获取令牌信息以获取项目ID用于重定向
		tokenInfo, _ := c.db.GetJWTToken(token.ID)
		projectID := ""
		if tokenInfo != nil {
			projectID = tokenInfo.ProjectID
		}
		http.Redirect(w, r, "/jwt/detail?id="+projectID+"&success=JWT令牌更新成功", http.StatusSeeOther)
	}
}

func (c *JWTController) DeleteJWTToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "缺少令牌ID", http.StatusBadRequest)
		return
	}

	// 获取令牌信息以获取项目ID用于重定向
	token, err := c.db.GetJWTToken(id)
	if err != nil {
		http.Error(w, "获取令牌信息失败", http.StatusInternalServerError)
		return
	}

	if err := c.db.DeleteJWTToken(id); err != nil {
		http.Redirect(w, r, "/jwt/detail?id="+token.ProjectID+"&error=删除JWT令牌失败", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/jwt/detail?id="+token.ProjectID, http.StatusSeeOther)
}

func (c *JWTController) DeleteExpiredJWTTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectID := r.FormValue("project_id")
	if projectID == "" {
		http.Error(w, "缺少项目ID", http.StatusBadRequest)
		return
	}

	count, err := c.db.DeleteExpiredJWTTokens(projectID)
	if err != nil {
		http.Redirect(w, r, "/jwt/detail?id="+projectID+"&error=删除过期JWT令牌失败", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/jwt/detail?id="+projectID+"&success=成功删除 "+string(rune(count))+" 个过期令牌", http.StatusSeeOther)
}

// VerifyJWTToken 验证JWT令牌（新增API）
func (c *JWTController) VerifyJWTToken(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "缺少令牌",
		})
		return
	}

	// 从数据库获取令牌信息
	token, err := c.db.GetJWTTokenByToken(tokenString)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "令牌不存在",
		})
		return
	}

	// 获取JWT项目
	project, err := c.db.GetJWTProject(token.ProjectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "获取JWT项目失败",
		})
		return
	}

	// 创建JWT管理器并验证令牌
	jwtManager, err := utils.NewJWTManager(project.PrivateKey, project.PublicKey)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "创建JWT管理器失败",
		})
		return
	}

	claims, err := jwtManager.VerifyToken(tokenString)

	// 检查是否是过期错误
	isExpiredError := err != nil && (strings.Contains(err.Error(), "token is expired") ||
		strings.Contains(err.Error(), "token has expired"))

	if err != nil && !isExpiredError {
		// 其他验证错误
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "令牌验证失败: " + err.Error(),
			"data": map[string]interface{}{
				"username":   token.Username,
				"role":       token.Role,
				"purpose":    token.Purpose,
				"expires_at": token.ExpiresAt,
				"is_active":  token.IsActive,
			},
		})
		return
	}

	// 如果是过期错误或验证成功，都返回成功状态
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if isExpiredError {
		// 过期令牌，返回数据库中的信息
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"username":   token.Username,
				"role":       token.Role,
				"purpose":    token.Purpose,
				"expires_at": token.ExpiresAt,
				"issued_at":  token.CreatedAt, // 使用创建时间作为签发时间
				"is_active":  token.IsActive,
			},
		})
	} else {
		// 验证成功，返回JWT解析的信息
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"username":   claims.Username,
				"role":       claims.Role,
				"purpose":    claims.Purpose,
				"expires_at": claims.ExpiresAt.Time,
				"issued_at":  claims.IssuedAt.Time,
				"is_active":  token.IsActive,
			},
		})
	}
}

// GenerateKeyPair 生成RSA密钥对
func (c *JWTController) GenerateKeyPair(w http.ResponseWriter, r *http.Request) {
	privateKeyPEM, publicKeyPEM, err := utils.GenerateRSAKeyPair(2048)
	if err != nil {
		http.Error(w, "生成密钥对失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"private_key": privateKeyPEM,
		"public_key":  publicKeyPEM,
		"message":     "密钥对生成成功",
	})
}
