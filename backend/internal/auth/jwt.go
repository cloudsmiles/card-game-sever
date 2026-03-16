package auth

import (
	"fmt"
	"time"

	"card-game-server/backend/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 自定义声明
type Claims struct {
	UserID    uint   `json:"user_id"`
	Nickname  string `json:"nickname"`
	LoginType string `json:"login_type"`
	jwt.RegisteredClaims
}

// GenerateToken 签发 JWT token
func GenerateToken(userID uint, nickname, loginType string) (string, error) {
	cfg := config.Global
	claims := Claims{
		UserID:    userID,
		Nickname:  nickname,
		LoginType: loginType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWT.ExpireHour) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ParseToken 解析并验证 JWT token
func ParseToken(tokenStr string) (*Claims, error) {
	cfg := config.Global
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
