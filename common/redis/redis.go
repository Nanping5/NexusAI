package myredis

import (
	"NexusAi/config"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	RedisCli "github.com/redis/go-redis/v9"
)

var Rdb *RedisCli.Client

var ctx = context.Background()

func InitRedis() error {

	host := config.GetConfig().RedisConfig.Host
	port := config.GetConfig().RedisConfig.Port
	password := config.GetConfig().RedisConfig.Password
	db := config.GetConfig().RedisConfig.DB

	Rdb = RedisCli.NewClient(&RedisCli.Options{
		Addr:         host + ":" + strconv.Itoa(port),
		Password:     password,
		DB:           db,
		Protocol:     2,                   // 使用redis协议版本2，解决redis v6.0.0以上版本不兼容问题
		PoolSize:     100,                 // 连接池大小
		MinIdleConns: 10,                  // 最小空闲连接数
		MaxRetries:   3,                   // 最大重试次数
		DialTimeout:  5 * time.Second,     // 连接超时
		ReadTimeout:  3 * time.Second,     // 读取超时
		WriteTimeout: 3 * time.Second,     // 写入超时
	})
	_, err := Rdb.Ping(ctx).Result()
	return err
}

// ========== 验证码相关（兼容旧接口） ==========

// SetCaptchaForEmail 将验证码存储在Redis中，并设置过期时间
func SetCaptchaForEmail(email, captcha string) error {
	Key := CaptchaKey(email)
	return Rdb.Set(ctx, Key, captcha, TTLCaptcha).Err()
}

// CheckCaptchaRateLimit 检查验证码发送频率限制
// 返回值：canSend 是否可以发送，remainingSeconds 剩余等待秒数，err 错误信息
func CheckCaptchaRateLimit(email string) (canSend bool, remainingSeconds int, err error) {
	key := CaptchaRateKey(email)

	// 检查是否存在频率限制 key
	ttl, err := Rdb.TTL(ctx, key).Result()
	if err != nil {
		return false, 0, err
	}

	// key 存在且未过期
	if ttl > 0 {
		return false, int(ttl.Seconds()), nil
	}

	return true, 0, nil
}

// SetCaptchaRateLimit 设置验证码发送频率限制
// cooldown: 冷却时间（秒）
func SetCaptchaRateLimit(email string, cooldown int) error {
	key := CaptchaRateKey(email)
	return Rdb.Set(ctx, key, "1", time.Duration(cooldown)*time.Second).Err()
}

// CheckCaptchaForEmail 从Redis中获取验证码并进行验证，验证成功后删除验证码
func CheckCaptchaForEmail(email, InputCaptcha string) (bool, error) {
	Key := CaptchaKey(email)
	storedCaptcha, err := Rdb.Get(ctx, Key).Result()
	if err != nil {
		// key 不存在或已过期，视为验证码无效
		if errors.Is(err, RedisCli.Nil) {
			return false, nil
		}
		return false, err
	}
	// 使用严格比较，验证码区分大小写
	if storedCaptcha == InputCaptcha {
		// 验证成功后删除验证码
		if err := Rdb.Del(ctx, Key).Err(); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() error {
	if Rdb == nil {
		return nil
	}
	return Rdb.Close()
}

// ========== 新增缓存功能 ==========

// SetJSON 存储 JSON 对象
func SetJSON(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return Rdb.Set(ctx, key, string(data), ttl).Err()
}

// GetJSON 获取 JSON 对象
func GetJSON(key string, dest interface{}) error {
	data, err := Rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// ========== 会话列表缓存 ==========

// SessionListItem 会话列表项
type SessionListItem struct {
	SessionID string `json:"session_id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

// SetUserSessionList 缓存用户会话列表
func SetUserSessionList(userID string, sessions []SessionListItem) error {
	key := SessionsKey(userID)
	return SetJSON(key, sessions, TTLSessions)
}

// GetUserSessionList 获取用户会话列表缓存
func GetUserSessionList(userID string) ([]SessionListItem, error) {
	key := SessionsKey(userID)
	var sessions []SessionListItem
	err := GetJSON(key, &sessions)
	if err != nil {
		if errors.Is(err, RedisCli.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return sessions, nil
}

// InvalidateUserSessionList 使用户会话列表缓存失效
func InvalidateUserSessionList(userID string) error {
	key := SessionsKey(userID)
	return Rdb.Del(ctx, key).Err()
}

// ========== 消息历史缓存 ==========

// HistoryItem 消息历史项
type HistoryItem struct {
	Content string `json:"content"`
	IsUser  bool   `json:"is_user"`
}

// SetSessionHistory 缓存会话消息历史
func SetSessionHistory(sessionID string, history []HistoryItem) error {
	key := HistoryKey(sessionID)
	return SetJSON(key, history, TTLHistory)
}

// GetSessionHistory 获取会话消息历史缓存
func GetSessionHistory(sessionID string) ([]HistoryItem, error) {
	key := HistoryKey(sessionID)
	var history []HistoryItem
	err := GetJSON(key, &history)
	if err != nil {
		if errors.Is(err, RedisCli.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return history, nil
}

// AppendSessionHistory 追加消息到历史缓存
func AppendSessionHistory(sessionID string, item HistoryItem) error {
	key := HistoryKey(sessionID)
	// 使用 Redis List 存储，方便追加
	pipe := Rdb.TxPipeline()
	data, _ := json.Marshal(item)
	pipe.RPush(ctx, key, string(data))
	pipe.Expire(ctx, key, TTLHistory)
	_, err := pipe.Exec(ctx)
	return err
}

// InvalidateSessionHistory 使会话历史缓存失效
func InvalidateSessionHistory(sessionID string) error {
	key := HistoryKey(sessionID)
	return Rdb.Del(ctx, key).Err()
}

// ========== 用户信息缓存 ==========

// UserCacheInfo 用户缓存信息
type UserCacheInfo struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

// SetUserInfo 缓存用户信息
func SetUserInfo(userID string, user UserCacheInfo) error {
	key := UserKey(userID)
	return SetJSON(key, user, TTLUser)
}

// GetUserInfo 获取用户信息缓存
func GetUserInfo(userID string) (*UserCacheInfo, error) {
	key := UserKey(userID)
	var user UserCacheInfo
	err := GetJSON(key, &user)
	if err != nil {
		if errors.Is(err, RedisCli.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// InvalidateUserInfo 使用户信息缓存失效
func InvalidateUserInfo(userID string) error {
	key := UserKey(userID)
	return Rdb.Del(ctx, key).Err()
}

// ========== 会话管理 ==========

// SetSession 设置用户会话
func SetSession(sessionID, userID string) error {
	return Rdb.Set(ctx, SessionKey(sessionID), userID, TTLSession).Err()
}

// GetSession 获取用户会话
func GetSession(sessionID string) (string, error) {
	return Rdb.Get(ctx, SessionKey(sessionID)).Result()
}

// DeleteSession 删除用户会话
func DeleteSession(sessionID string) error {
	return Rdb.Del(ctx, SessionKey(sessionID)).Err()
}

// ========== 在线用户 ==========

// AddOnlineUser 添加在线用户
func AddOnlineUser(userID string) error {
	return Rdb.SAdd(ctx, OnlineKey(), userID).Err()
}

// RemoveOnlineUser 移除在线用户
func RemoveOnlineUser(userID string) error {
	return Rdb.SRem(ctx, OnlineKey(), userID).Err()
}

// GetOnlineUsers 获取所有在线用户
func GetOnlineUsers() ([]string, error) {
	return Rdb.SMembers(ctx, OnlineKey()).Result()
}

// IsUserOnline 检查用户是否在线
func IsUserOnline(userID string) (bool, error) {
	return Rdb.SIsMember(ctx, OnlineKey(), userID).Result()
}

// ========== 接口限流 ==========

// IncrLimit 增加限流计数
func IncrLimit(userID, api string) (int64, error) {
	key := LimitKey(userID, api)
	pipe := Rdb.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, TTLLimit)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Result()
}

// CheckLimit 检查是否超过限流
func CheckLimit(userID, api string, limit int64) (bool, error) {
	key := LimitKey(userID, api)
	count, err := Rdb.Get(ctx, key).Int64()
	if err != nil {
		if errors.Is(err, RedisCli.Nil) {
			return false, nil
		}
		return false, err
	}
	return count > limit, nil
}
