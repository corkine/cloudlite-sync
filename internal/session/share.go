package session

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

// ShareCode 分享码结构
type ShareCode struct {
	Code      string    `json:"code"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ShareCodeService 分享码服务
type ShareCodeService struct {
	codes         map[string]*ShareCode
	mutex         sync.RWMutex
	expireSeconds int
}

var (
	shareService *ShareCodeService
	once         sync.Once
)

// GetShareCodeService 获取分享码服务单例
func GetShareCodeService() *ShareCodeService {
	once.Do(func() {
		shareService = &ShareCodeService{
			codes:         make(map[string]*ShareCode),
			expireSeconds: 30, // 默认30秒
		}
		// 启动清理过期分享码的goroutine
		go shareService.cleanupExpiredCodes()
	})
	return shareService
}

// SetExpireSeconds 设置过期时间（秒）
func (s *ShareCodeService) SetExpireSeconds(seconds int) {
	s.mutex.Lock()
	s.expireSeconds = seconds
	s.mutex.Unlock()
}

// GenerateShareCode 生成6位数字分享码
func (s *ShareCodeService) GenerateShareCode(token string) (string, error) {
	// 生成6位随机数字
	code := ""
	for i := 0; i < 6; i++ {
		randomByte := make([]byte, 1)
		_, err := rand.Read(randomByte)
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", randomByte[0]%10)
	}

	// 创建分享码记录
	now := time.Now()
	shareCode := &ShareCode{
		Code:      code,
		Token:     token,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(s.expireSeconds) * time.Second),
	}

	// 存储到内存中
	s.mutex.Lock()
	s.codes[code] = shareCode
	s.mutex.Unlock()

	return code, nil
}

// GetTokenByCode 根据分享码获取令牌
func (s *ShareCodeService) GetTokenByCode(code string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	shareCode, exists := s.codes[code]
	if !exists {
		return "", false
	}

	// 检查是否过期
	if time.Now().After(shareCode.ExpiresAt) {
		// 异步删除过期的分享码
		go s.deleteCode(code)
		return "", false
	}

	return shareCode.Token, true
}

// DeleteCode 删除分享码
func (s *ShareCodeService) DeleteCode(code string) {
	s.mutex.Lock()
	delete(s.codes, code)
	s.mutex.Unlock()
}

// deleteCode 内部删除方法
func (s *ShareCodeService) deleteCode(code string) {
	s.mutex.Lock()
	delete(s.codes, code)
	s.mutex.Unlock()
}

// cleanupExpiredCodes 清理过期的分享码
func (s *ShareCodeService) cleanupExpiredCodes() {
	ticker := time.NewTicker(10 * time.Second) // 每10秒检查一次
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		s.mutex.Lock()

		for code, shareCode := range s.codes {
			if now.After(shareCode.ExpiresAt) {
				delete(s.codes, code)
			}
		}

		s.mutex.Unlock()
	}
}

// GetCodeInfo 获取分享码信息（用于前端显示）
func (s *ShareCodeService) GetCodeInfo(code string) (*ShareCode, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	shareCode, exists := s.codes[code]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(shareCode.ExpiresAt) {
		// 异步删除过期的分享码
		go s.deleteCode(code)
		return nil, false
	}

	return shareCode, true
}
