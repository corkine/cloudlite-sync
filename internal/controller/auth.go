package controller

import (
	"fmt"
	"net/http"

	"chchma.com/cloudlite-sync/internal/session"
	"chchma.com/cloudlite-sync/internal/template"
)

// LoginPage 显示登录页面
func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// 如果已经登录，重定向到首页
	if session.IsAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := template.NewPageData("登录", nil)
	fmt.Println("LoginPage")
	h.tmpl.Render(w, "login.html", data)
}

// Login 处理登录请求
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// 验证用户名和密码
	if username == h.config.Admin.Username && password == h.config.Admin.Password {
		// 设置认证状态
		err := session.SetAuthenticated(w, r, username)
		if err != nil {
			http.Error(w, "Failed to set session", http.StatusInternalServerError)
			return
		}

		// 重定向到首页
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// 登录失败
	data := template.NewPageData("登录", nil)
	data.SetError("用户名或密码错误")
	h.tmpl.Render(w, "login.html", data)
}

// Logout 处理登出请求
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	err := session.ClearSession(w, r)
	if err != nil {
		http.Error(w, "Failed to clear session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// API登录（用于API调用）
func (h *Handler) APILogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// 验证用户名和密码
	if username == h.config.Admin.Username && password == h.config.Admin.Password {
		// 设置认证状态
		err := session.SetAuthenticated(w, r, username)
		if err != nil {
			http.Error(w, "Failed to set session", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "message": "登录成功"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"success": false, "message": "用户名或密码错误"}`))
}
