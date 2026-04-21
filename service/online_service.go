package service

import (
	"sync"
	"time"
)

type onlineService struct {
	users map[string]time.Time // userID -> last activity
	mutex sync.RWMutex
}

var OnlineService = &onlineService{
	users: make(map[string]time.Time),
}

// UserOnline 用户上线
func (s *onlineService) UserOnline(userID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.users[userID] = time.Now()
}

// UserOffline 用户离线
func (s *onlineService) UserOffline(userID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.users, userID)
}

// RefreshActivity 刷新用户活跃时间
func (s *onlineService) RefreshActivity(userID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.users[userID] = time.Now()
}

// GetOnlineCount 获取在线人数
func (s *onlineService) GetOnlineCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.users)
}

// GetOnlineUserIDs 获取所有在线用户ID（仅调试用）
func (s *onlineService) GetOnlineUserIDs() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	ids := make([]string, 0, len(s.users))
	for id := range s.users {
		ids = append(ids, id)
	}
	return ids
}
