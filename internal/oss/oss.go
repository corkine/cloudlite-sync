package oss

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OSSClient struct {
	client     *oss.Client
	bucket     *oss.Bucket
	bucketName string
}

type OSSConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
}

func NewOSSClient(config OSSConfig) (*OSSClient, error) {
	// 创建OSSClient实例
	client, err := oss.New(config.Endpoint, config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %w", err)
	}

	// 获取存储空间
	bucket, err := client.Bucket(config.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}

	return &OSSClient{
		client:     client,
		bucket:     bucket,
		bucketName: config.BucketName,
	}, nil
}

// UploadFile 上传文件到OSS
func (c *OSSClient) UploadFile(key string, data []byte) error {
	// 上传文件
	err := c.bucket.PutObject(key, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to upload file to OSS: %w", err)
	}
	return nil
}

// DownloadFile 从OSS下载文件
func (c *OSSClient) DownloadFile(key string) ([]byte, error) {
	// 下载文件
	object, err := c.bucket.GetObject(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get object from OSS: %w", err)
	}
	defer object.Close()

	// 读取文件内容
	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	return data, nil
}

// DeleteFile 从OSS删除文件
func (c *OSSClient) DeleteFile(key string) error {
	err := c.bucket.DeleteObject(key)
	if err != nil {
		return fmt.Errorf("failed to delete object from OSS: %w", err)
	}
	return nil
}

// FileExists 检查文件是否存在
func (c *OSSClient) FileExists(key string) (bool, error) {
	exist, err := c.bucket.IsObjectExist(key)
	if err != nil {
		return false, fmt.Errorf("failed to check if object exists: %w", err)
	}
	return exist, nil
}

// GetFileURL 获取文件的预签名URL（用于直接下载）
func (c *OSSClient) GetFileURL(key string, expires time.Duration) (string, error) {
	url, err := c.bucket.SignURL(key, oss.HTTPGet, int64(expires.Seconds()))
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}
	return url, nil
}

// GetFileInfo 获取文件信息
func (c *OSSClient) GetFileInfo(key string) (http.Header, error) {
	meta, err := c.bucket.GetObjectDetailedMeta(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get object meta: %w", err)
	}
	return meta, nil
}
