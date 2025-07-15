package controller

import (
	"net/http"

	"chchma.com/cloudlite-sync/config"
	"chchma.com/cloudlite-sync/internal/database"
	m "chchma.com/cloudlite-sync/internal/middleware"
	"chchma.com/cloudlite-sync/internal/oss"
	"chchma.com/cloudlite-sync/internal/session"
	"chchma.com/cloudlite-sync/internal/template"
	"github.com/go-chi/chi/v5"
	cm "github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	config    *config.Config
	db        *database.DB
	ossClient *oss.OSSClient
	tmpl      *template.TemplateEngine
}

func NewRouter(cfg *config.Config, db *database.DB, ossClient *oss.OSSClient) *chi.Mux {
	session.Init(cfg.SessionSecret)

	handler := &Handler{
		config:    cfg,
		db:        db,
		ossClient: ossClient,
		tmpl:      template.New(),
	}

	r := chi.NewRouter()

	// 中间件
	r.Use(m.LoggingMiddleware)
	r.Use(cm.Recoverer)
	r.Use(m.CORSMiddleware)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// 认证路由
	r.Get("/login", handler.LoginPage)
	r.Post("/login", handler.Login)
	r.Get("/logout", handler.Logout)

	// 需要认证的路由
	r.Group(func(r chi.Router) {
		r.Use(m.AuthMiddleware)

		// 项目管理
		r.Get("/", handler.Dashboard)
		r.Post("/project/create", handler.CreateProject)
		r.Post("/project/update", handler.UpdateProject)
		r.Post("/project/delete", handler.DeleteProject)
		r.Get("/project/detail", handler.ProjectDetail)
		r.Post("/project/upload_version", handler.UploadDatabaseVersion)
		r.Post("/project/delete_version", handler.DeleteDatabaseVersion)
		r.Get("/project/download", handler.ProjectDownload)
		r.Get("/help", handler.HelpPage)

		// 凭证管理
		r.Post("/credential/create", handler.CreateCredential)
		r.Post("/credential/delete", handler.DeleteCredential)
		r.Post("/credential/activate", handler.ActivateCredential)
		r.Post("/credential/deactivate", handler.DeactivateCredential)
	})

	// API路由（第三方访问）
	r.Route("/api", func(r chi.Router) {
		r.Post("/{projectID}", handler.ApiUploadDatabase)
		r.Get("/{projectID}/latest", handler.ApiDownloadLatest)
		r.Get("/{projectID}/{hash}", handler.ApiDownloadByHash)
		r.Get("/{projectID}/versions", handler.ApiListVersions)
		r.Get("/{projectID}/info/{hash}", handler.ApiGetVersionInfo)
	})
	return r
}
