package auth

import (
	"crypto/subtle"

	"github.com/athena/staticman/internal/config"
)

// Service 认证服务
type Service struct {
	cfg *config.Config
}

// New 创建认证服务
func New(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

// VerifyPassword 验证密码（常量时间比较防止时序攻击）
func (s *Service) VerifyPassword(input string, currentPassword string) bool {
	if currentPassword == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(input), []byte(currentPassword)) == 1
}
