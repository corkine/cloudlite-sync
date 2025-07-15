package session

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

// Init 初始化会话存储
func Init(secretKey string) {
	key, err := base64.StdEncoding.DecodeString(secretKey)
	if err != nil {
		panic("session_secret 不是合法的 base64 字符串")
	}
	store = sessions.NewCookieStore(key)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7天
		HttpOnly: true,
		Secure:   false, // 在生产环境中应该设置为true
	}
}

// GetSession 获取会话
func GetSession(r *http.Request) (*sessions.Session, error) {
	return store.Get(r, "db-sync-session")
}

// SetAuthenticated 设置认证状态
func SetAuthenticated(w http.ResponseWriter, r *http.Request, username string) error {
	session, err := GetSession(r)
	if err != nil {
		log.Println("SetAuthenticated error:", err)
		return err
	}

	session.Values["authenticated"] = true
	session.Values["username"] = username
	session.Values["login_time"] = time.Now().Unix()

	return session.Save(r, w)
}

// IsAuthenticated 检查是否已认证
func IsAuthenticated(r *http.Request) bool {
	session, err := GetSession(r)
	if err != nil {
		return false
	}

	authenticated, ok := session.Values["authenticated"].(bool)
	return ok && authenticated
}

// GetUsername 获取用户名
func GetUsername(r *http.Request) string {
	session, err := GetSession(r)
	if err != nil {
		return ""
	}

	username, ok := session.Values["username"].(string)
	if !ok {
		return ""
	}

	return username
}

// ClearSession 清除会话
func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := GetSession(r)
	if err != nil {
		return err
	}

	session.Values["authenticated"] = false
	session.Values["username"] = ""
	session.Options.MaxAge = -1

	return session.Save(r, w)
}

// GenerateSecretKey 生成密钥
func GenerateSecretKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
