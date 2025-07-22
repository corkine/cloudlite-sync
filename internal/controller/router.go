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
	jwtCtrl   *JWTController
}

func NewRouter(cfg *config.Config, db *database.DB, ossClient *oss.OSSClient) *chi.Mux {
	session.Init(cfg.SessionSecret)

	handler := &Handler{
		config:    cfg,
		db:        db,
		ossClient: ossClient,
		tmpl:      template.New(),
		jwtCtrl:   NewJWTController(db),
	}

	r := chi.NewRouter()

	// 中间件
	r.Use(m.LoggingMiddleware)
	r.Use(cm.Recoverer)
	r.Use(m.CORSMiddleware)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// 分享路由（无需认证）
	r.Get("/s/{code}", handler.jwtCtrl.ShareAPI)

	// 认证路由
	r.Get("/login", handler.LoginPage)
	r.Post("/login", handler.Login)
	r.Get("/logout", handler.Logout)

	// 需要认证的路由
	r.Group(func(r chi.Router) {
		r.Use(m.AuthMiddleware)

		// 项目管理
		r.Get("/", handler.Dashboard)
		r.Get("/help", handler.HelpPage)
		r.Get("/jwt_help", handler.JWTHelpPage)

		r.Route("/project", func(r chi.Router) {
			r.Post("/create", handler.CreateProject)
			r.Post("/update", handler.UpdateProject)
			r.Post("/delete", handler.DeleteProject)
			r.Get("/detail", handler.ProjectDetail)
			r.Post("/upload_version", handler.UploadDatabaseVersion)
			r.Post("/delete_version", handler.DeleteDatabaseVersion)
			r.Get("/download", handler.ProjectDownload)
		})

		// 凭证管理
		r.Route("/credential", func(r chi.Router) {
			r.Post("/create", handler.CreateCredential)
			r.Post("/delete", handler.DeleteCredential)
			r.Post("/activate", handler.ActivateCredential)
			r.Post("/deactivate", handler.DeactivateCredential)
		})

		// JWT项目管理
		r.Route("/jwt", func(r chi.Router) {
			r.Get("/", handler.JWTDashboard)
			r.Get("/detail", handler.JWTProjectDetail)
			r.Post("/key/generate", handler.jwtCtrl.GenerateKeyPair)

			r.Route("/project", func(r chi.Router) {
				r.Post("/create", handler.jwtCtrl.CreateJWTProject)
				r.Get("/get", handler.jwtCtrl.GetJWTProject)
				r.Get("/list", handler.jwtCtrl.ListJWTProjects)
				r.Post("/update", handler.jwtCtrl.UpdateJWTProject)
				r.Post("/delete", handler.jwtCtrl.DeleteJWTProject)
			})

			r.Route("/token", func(r chi.Router) {
				r.Post("/create", handler.jwtCtrl.CreateJWTToken)
				r.Get("/get", handler.jwtCtrl.GetJWTToken)
				r.Get("/list", handler.jwtCtrl.ListJWTTokens)
				r.Post("/update", handler.jwtCtrl.UpdateJWTToken)
				r.Post("/delete", handler.jwtCtrl.DeleteJWTToken)
				r.Post("/delete_expired", handler.jwtCtrl.DeleteExpiredJWTTokens)
				r.Get("/verify", handler.jwtCtrl.VerifyJWTToken)
				r.Post("/share", handler.jwtCtrl.GenerateShareCode)
				r.Get("/share/info", handler.jwtCtrl.GetShareCodeInfo)
			})
		})
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
