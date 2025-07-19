package config

import (
	"encoding/json"
	"os"
	"strconv"
)

type Config struct {
	Server        ServerConfig    `json:"server"`
	OSS           OSSConfig       `json:"oss"`
	Admin         AdminConfig     `json:"admin"`
	SessionSecret string          `json:"session_secret"`
	ShareCode     ShareCodeConfig `json:"share_code"`
}

type ServerConfig struct {
	Port string `json:"port"`
	Host string `json:"host"`
}

type OSSConfig struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	BucketName      string `json:"bucket_name"`
}

type AdminConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ShareCodeConfig struct {
	ExpireSeconds int `json:"expire_seconds"`
}

func Load() *Config {
	// 首先从 config.json 加载配置
	config := loadFromFile()

	// 然后用环境变量覆盖配置
	overrideWithEnv(config)

	return config
}

func loadFromFile() *Config {
	config := &Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "localhost",
		},
		OSS: OSSConfig{
			Endpoint:        "",
			AccessKeyID:     "",
			AccessKeySecret: "",
			BucketName:      "",
		},
		Admin: AdminConfig{
			Username: "admin",
			Password: "admin123",
		},
		SessionSecret: "",
		ShareCode: ShareCodeConfig{
			ExpireSeconds: 3600, // 默认过期时间为1小时
		},
	}

	data, err := os.ReadFile("config.json")

	if err != nil {
		// 如果文件不存在，返回默认配置
		return config
	}

	// 解析 JSON 配置
	if err := json.Unmarshal(data, config); err != nil {
		// 如果解析失败，返回默认配置
		return config
	}

	return config
}

func overrideWithEnv(config *Config) {
	// 服务器配置
	if value := os.Getenv("PORT"); value != "" {
		config.Server.Port = value
	}
	if value := os.Getenv("HOST"); value != "" {
		config.Server.Host = value
	}
	// OSS 配置
	if value := os.Getenv("OSS_ENDPOINT"); value != "" {
		config.OSS.Endpoint = value
	}
	if value := os.Getenv("OSS_ACCESS_KEY_ID"); value != "" {
		config.OSS.AccessKeyID = value
	}
	if value := os.Getenv("OSS_ACCESS_KEY_SECRET"); value != "" {
		config.OSS.AccessKeySecret = value
	}
	if value := os.Getenv("OSS_BUCKET_NAME"); value != "" {
		config.OSS.BucketName = value
	}
	// 管理员配置
	if value := os.Getenv("ADMIN_USERNAME"); value != "" {
		config.Admin.Username = value
	}
	if value := os.Getenv("ADMIN_PASSWORD"); value != "" {
		config.Admin.Password = value
	}
	// 分享码配置
	if value := os.Getenv("SHARE_CODE_EXPIRE_SECONDS"); value != "" {
		if expireSeconds, err := strconv.Atoi(value); err == nil {
			config.ShareCode.ExpireSeconds = expireSeconds
		}
	}
}
