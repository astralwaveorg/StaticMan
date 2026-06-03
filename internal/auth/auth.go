package auth

import (
	"crypto/subtle"
	"time"

	"github.com/athena/magichub/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

// Service 认证服务
type Service struct {
	cfg       *config.Config
	jwtSecret []byte
}

// New 创建认证服务
func New(cfg *config.Config, jwtSecret string) *Service {
	return &Service{
		cfg:       cfg,
		jwtSecret: []byte(jwtSecret),
	}
}

// VerifyPassword 验证密码（常量时间比较防止时序攻击）
func (s *Service) VerifyPassword(input string, currentPassword string) bool {
	if currentPassword == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(input), []byte(currentPassword)) == 1
}

// GenerateToken 生成 JWT token（7 天有效期）
func (s *Service) GenerateToken() (string, error) {
	claims := jwt.MapClaims{
		"authenticated": true,
		"exp":          time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":          time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken 验证 JWT token
func (s *Service) ValidateToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil {
		return false, err
	}
	return token.Valid, nil
}