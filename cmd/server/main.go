package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chchma.com/cloudlite-sync/config"
	"chchma.com/cloudlite-sync/internal/controller"
	"chchma.com/cloudlite-sync/internal/database"
	"chchma.com/cloudlite-sync/internal/oss"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 确保 data 目录存在
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// 初始化数据库，强制路径为 ./data/data.db
	db, err := database.New("./data/data.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 初始化OSS客户端
	var ossClient *oss.OSSClient
	if cfg.OSS.Endpoint != "" && cfg.OSS.AccessKeyID != "" && cfg.OSS.AccessKeySecret != "" {
		ossClient, err = oss.NewOSSClient(oss.OSSConfig{
			Endpoint:        cfg.OSS.Endpoint,
			AccessKeyID:     cfg.OSS.AccessKeyID,
			AccessKeySecret: cfg.OSS.AccessKeySecret,
			BucketName:      cfg.OSS.BucketName,
		})
		if err != nil {
			log.Fatalf("Failed to initialize OSS client: %v", err)
		}
	} else {
		log.Println("Warning: OSS configuration not provided, file operations will be disabled")
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      controller.NewRouter(cfg, db, ossClient),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		log.Printf("Starting server on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
