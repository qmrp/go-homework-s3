package impl

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/api/response"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/manager"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/model"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/service"
)

// InMemoryUserService 基于内存的用户服务实现
type InMemoryUserService struct {
	users    map[string]*model.User
	sessions map[string]string
	mutex    sync.RWMutex
}

// NewInMemoryUserService 创建基于内存的用户服务实例
func NewInMemoryUserService() service.UserService {
	return &InMemoryUserService{
		users:    make(map[string]*model.User),
		sessions: make(map[string]string),
	}
}

// GetByID 根据ID获取用户
func (s *InMemoryUserService) GetByID(ctx context.Context, userID uint64) (*model.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, user := range s.users {
		if user.ID == userID {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

// GetByUsername 根据用户名获取用户
func (s *InMemoryUserService) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// ExistsByUsername 检查用户名是否存在
func (s *InMemoryUserService) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, exists := s.users[username]
	return exists, nil
}

// Login 登录
func (s *InMemoryUserService) Login(ctx context.Context, username string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 查找用户
	user, exists := s.users[username]
	if !exists {
		// 生成用户ID
		user = &model.User{
			ID:        uint64(len(s.users) + 1),
			Username:  username,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Online:    false,
		}

		// 保存用户
		s.users[user.Username] = user
	}

	// 生成session ID
	sessionID := fmt.Sprintf("%s_%d", username, time.Now().UnixNano())

	// 保存session
	s.sessions[sessionID] = username
	for sid, user := range s.sessions {
		if user == username {
			println(sid, user)
		}
	}

	// 设置用户在线状态
	user.Online = true
	user.UpdatedAt = time.Now()

	return sessionID, nil
}

// Logout 登出
func (s *InMemoryUserService) Logout(ctx context.Context, username string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 查找用户
	user, exists := s.users[username]
	if !exists {
		return nil // 用户不存在，忽略
	}

	// 设置用户离线状态
	user.Online = false
	user.UpdatedAt = time.Now()

	// 清除所有该用户的session
	for sessionID, user := range s.sessions {
		if user == username {
			delete(s.sessions, sessionID)
		}
	}
	manager.MessageManager.UnregisterConnection(username)
	return nil
}

// SetOnlineStatus 设置用户在线状态
func (s *InMemoryUserService) SetOnlineStatus(ctx context.Context, username string, online bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 查找用户
	user, exists := s.users[username]
	if !exists {
		return fmt.Errorf("user not found")
	}

	// 设置在线状态
	user.Online = online
	user.UpdatedAt = time.Now()

	return nil
}

// GetOnlineUsers 获取所有在线用户
func (s *InMemoryUserService) GetOnlineUsers(ctx context.Context) (response.UserListResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var onlineUsers []string
	for username, user := range s.users {
		if user.Online {
			onlineUsers = append(onlineUsers, username)
		}
	}

	return response.UserListResponse{
		List:  onlineUsers,
		Total: len(onlineUsers),
	}, nil
}

func (s *InMemoryUserService) GetAllUsers(ctx context.Context) (response.UserListResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var allUsers []string
	for username := range s.users {
		allUsers = append(allUsers, username)
	}

	return response.UserListResponse{
		List:  allUsers,
		Total: len(allUsers),
	}, nil
}

func (s *InMemoryUserService) GetUsernameBySessionID(sessionID string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	username, exists := s.sessions[sessionID]
	return username, exists
}

// SetNonResponseCount 设置用户非响应次数
func (s *InMemoryUserService) SetNonResponseCount(ctx context.Context, username string, count int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// 查找用户
	user, exists := s.users[username]
	if !exists {
		return fmt.Errorf("user not found")
	}

	// 设置非响应次数
	user.NonResponseCount = count
	user.UpdatedAt = time.Now()

	return nil
}

// SetLastAckId 设置用户最后一次收到的消息ID
func (s *InMemoryUserService) SetLastAckId(ctx context.Context, username string, ackID int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// 查找用户
	user, exists := s.users[username]
	if !exists {
		return fmt.Errorf("user not found")
	}
	// 设置最后一次收到的消息ID
	user.LastAskId = ackID
	user.UpdatedAt = time.Now()

	return nil
}
