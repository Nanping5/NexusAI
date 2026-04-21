package aihelper

import (
	"context"
	"sync"
	"time"

	"NexusAi/config"
)

type AIHelperManager struct {
	helpers      map[string]map[string]*AIHelper // map[userID]map[sessionID]*AIHelper
	mutex        sync.RWMutex
	lastAccess   map[string]time.Time // 记录每个 helper 的最后访问时间
	cleanupTick  *time.Ticker
	stopCleanup  chan struct{}
}

// NewAIHelperManager 创建新的 AIHelperManager 实例
func NewAIHelperManager() *AIHelperManager {
	m := &AIHelperManager{
		helpers:     make(map[string]map[string]*AIHelper),
		lastAccess:  make(map[string]time.Time),
		stopCleanup: make(chan struct{}),
	}
	// 启动定期清理协程
	m.startCleanup()
	return m
}

// startCleanup 启动定期清理协程，清理长时间未访问的 helper
func (m *AIHelperManager) startCleanup() {
	cfg := config.GetConfig().AIConfig
	interval := cfg.HelperCleanupInterval
	if interval <= 0 {
		interval = 5 // 默认 5 分钟
	}
	m.cleanupTick = time.NewTicker(time.Duration(interval) * time.Minute)
	go func() {
		for {
			select {
			case <-m.cleanupTick.C:
				m.cleanup()
			case <-m.stopCleanup:
				return
			}
		}
	}()
}

// cleanup 清理超过配置的过期时间未访问的 helper
func (m *AIHelperManager) cleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cfg := config.GetConfig().AIConfig
	expireMinutes := cfg.HelperExpireDuration
	if expireMinutes <= 0 {
		expireMinutes = 30 // 默认 30 分钟
	}
	expireThreshold := time.Duration(expireMinutes) * time.Minute
	now := time.Now()

	for userID, userHelpers := range m.helpers {
		for sessionID := range userHelpers {
			key := userID + ":" + sessionID
			if lastAccess, exists := m.lastAccess[key]; exists {
				if now.Sub(lastAccess) > expireThreshold {
					delete(userHelpers, sessionID)
					delete(m.lastAccess, key)
				}
			}
		}
		// 如果用户没有任何 helper，删除用户条目
		if len(userHelpers) == 0 {
			delete(m.helpers, userID)
		}
	}
}

// Stop 停止清理协程
func (m *AIHelperManager) Stop() {
	if m.cleanupTick != nil {
		m.cleanupTick.Stop()
	}
	close(m.stopCleanup)
}

// GetOrCreateAIHelperByConfigID 根据配置 ID 获取或创建 AIHelper 实例
func (m *AIHelperManager) GetOrCreateAIHelperByConfigID(ctx context.Context, userID, sessionID, configID string) (*AIHelper, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	userHelpers, exists := m.helpers[userID]
	if !exists {
		userHelpers = make(map[string]*AIHelper)
		m.helpers[userID] = userHelpers
	}

	helper, exists := userHelpers[sessionID]
	if !exists {
		// 创建新的 AIHelper 实例
		factory := GetGlobalFactory()
		var err error
		helper, err = factory.CreateAIHelperByConfigID(ctx, configID, sessionID, userID)
		if err != nil {
			return nil, err
		}
		userHelpers[sessionID] = helper
		m.updateLastAccessLocked(userID, sessionID)
		return helper, nil
	}
	m.updateLastAccessLocked(userID, sessionID)
	return helper, nil
}

// GetOrCreateAIHelperWithAgentByConfigID 根据配置 ID 获取或创建带 Agent 的 AIHelper 实例
func (m *AIHelperManager) GetOrCreateAIHelperWithAgentByConfigID(ctx context.Context, userID, sessionID, configID string, agentConfig AgentConfig) (*AIHelper, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	userHelpers, exists := m.helpers[userID]
	if !exists {
		userHelpers = make(map[string]*AIHelper)
		m.helpers[userID] = userHelpers
	}

	helper, exists := userHelpers[sessionID]
	// 如果 helper 不存在，或者存在的 helper 模式与请求的模式不匹配，重新创建
	if !exists || helper.GetAgentMode() != agentConfig.Mode {
		// 创建带 Agent 的 AIHelper 实例
		factory := GetGlobalFactory()
		var err error
		helper, err = factory.CreateAIHelperWithAgentByConfigID(ctx, configID, sessionID, userID, agentConfig)
		if err != nil {
			return nil, err
		}
		userHelpers[sessionID] = helper
		m.updateLastAccessLocked(userID, sessionID)
		return helper, nil
	}
	m.updateLastAccessLocked(userID, sessionID)
	return helper, nil
}

// updateLastAccessLocked 更新最后访问时间（需要已持有锁）
func (m *AIHelperManager) updateLastAccessLocked(userID, sessionID string) {
	key := userID + ":" + sessionID
	m.lastAccess[key] = time.Now()
}

// GetAIHelper 获取指定的 AIHelper 实例
func (m *AIHelperManager) GetAIHelper(userID, sessionID string) (*AIHelper, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	userHelpers, exists := m.helpers[userID]
	if !exists {
		return nil, false
	}

	helper, exists := userHelpers[sessionID]
	if exists {
		m.updateLastAccessLocked(userID, sessionID)
	}
	return helper, exists
}

// RemoveAIHelper 删除指定的 AIHelper 实例
func (m *AIHelperManager) RemoveAIHelper(userID, sessionID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	userHelpers, exists := m.helpers[userID]
	if exists {
		if helper, ok := userHelpers[sessionID]; ok {
			// 停止队列 goroutine
			if helper.queueCancel != nil {
				helper.queueCancel()
			}
		}
		delete(userHelpers, sessionID)
		key := userID + ":" + sessionID
		delete(m.lastAccess, key)
		if len(userHelpers) == 0 {
			delete(m.helpers, userID)
		}
	}
}

// GetUserSessions 获取用户的所有会话 ID
func (m *AIHelperManager) GetUserSessions(userID string) []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	userHelpers, exists := m.helpers[userID]
	if !exists {
		return nil
	}
	sessions := make([]string, 0, len(userHelpers))
	for sessionID := range userHelpers {
		sessions = append(sessions, sessionID)
	}
	return sessions
}

var globalManager *AIHelperManager
var managerOnce sync.Once

// GetGlobalAIHelperManager 获取全局 AIHelperManager 实例，单例模式
func GetGlobalAIHelperManager() *AIHelperManager {
	managerOnce.Do(func() {
		globalManager = NewAIHelperManager()
	})
	return globalManager
}
