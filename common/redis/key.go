package myredis

import "time"

// Key 前缀常量
const (
	KeyPrefix = "nexus" // 项目前缀
	Separator = ":"      // 分隔符
)

// Key 类型
const (
	KeySession     = "session"      // 用户登录会话
	KeyUser        = "user"         // 用户信息缓存
	KeySessions    = "sessions"     // 用户会话列表
	KeyHistory     = "history"      // 聊天历史缓存
	KeyModel       = "model"        // 模型配置缓存
	KeyContext     = "context"      // AI 对话上下文
	KeyOnline      = "online"       // 在线用户集合
	KeyLimit       = "limit"        // 接口限流计数
	KeyCaptcha     = "captcha"      // 验证码
	KeyCaptchaRate = "captcha:rate" // 验证码频率限制
)

// TTL 常量
const (
	TTLSession  = 7 * 24 * time.Hour  // 用户登录态 7 天
	TTLUser     = 30 * time.Minute    // 用户信息 30 分钟
	TTLSessions = 5 * time.Minute     // 会话列表 5 分钟
	TTLHistory  = 1 * time.Hour       // 消息历史 1 小时
	TTLModel    = 5 * time.Minute     // 模型配置 5 分钟
	TTLContext  = 2 * time.Hour       // AI 对话上下文 2 小时
	TTLOnline   = 5 * time.Minute     // 在线用户心跳 5 分钟
	TTLLimit    = 1 * time.Minute     // 限流窗口 1 分钟
	TTLCaptcha  = 3 * time.Minute     // 验证码 3 分钟
)

// Session Key: nexus:session:{session_id}
func SessionKey(sessionID string) string {
	return KeyPrefix + Separator + KeySession + Separator + sessionID
}

// User Key: nexus:user:{user_id}
func UserKey(userID string) string {
	return KeyPrefix + Separator + KeyUser + Separator + userID
}

// Sessions Key: nexus:sessions:{user_id}
func SessionsKey(userID string) string {
	return KeyPrefix + Separator + KeySessions + Separator + userID
}

// History Key: nexus:history:{session_id}
func HistoryKey(sessionID string) string {
	return KeyPrefix + Separator + KeyHistory + Separator + sessionID
}

// Model Key: nexus:model:{config_id}
func ModelKey(configID string) string {
	return KeyPrefix + Separator + KeyModel + Separator + configID
}

// Context Key: nexus:context:{session_id}
func ContextKey(sessionID string) string {
	return KeyPrefix + Separator + KeyContext + Separator + sessionID
}

// Online Key: nexus:online
func OnlineKey() string {
	return KeyPrefix + Separator + KeyOnline
}

// Limit Key: nexus:limit:{user_id}:{api}
func LimitKey(userID, api string) string {
	return KeyPrefix + Separator + KeyLimit + Separator + userID + Separator + api
}

// Captcha Key: nexus:captcha:{email}
func CaptchaKey(email string) string {
	return KeyPrefix + Separator + KeyCaptcha + Separator + email
}

// CaptchaRate Key: nexus:captcha:rate:{email}
func CaptchaRateKey(email string) string {
	return KeyPrefix + Separator + KeyCaptchaRate + Separator + email
}

// ========== 兼容旧接口 ==========

// GenerateCaptchaKey 生成验证码 Key（兼容旧接口）
func GenerateCaptchaKey(email string) string {
	return CaptchaKey(email)
}

// GenerateCaptchaRateLimitKey 生成验证码频率限制 Key（兼容旧接口）
func GenerateCaptchaRateLimitKey(email string) string {
	return CaptchaRateKey(email)
}
