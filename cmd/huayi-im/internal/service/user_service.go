package service

import (
	"context"

	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/api/response"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/model"
)

// UserService 用户业务逻辑接口
type UserService interface {
	GetByID(ctx context.Context, userID uint64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	Login(ctx context.Context, username string) (string, error) // 返回session ID
	Logout(ctx context.Context, username string) error
	SetOnlineStatus(ctx context.Context, username string, online bool) error
	GetOnlineUsers(ctx context.Context) (response.UserListResponse, error)
	GetAllUsers(ctx context.Context) (response.UserListResponse, error)
	GetUsernameBySessionID(sessionID string) (string, bool)
	SetNonResponseCount(ctx context.Context, username string, count int) error
	SetLastAckId(ctx context.Context, username string, ackID int64) error
}
