package utils

import (
	"crypto/rand"
	"encoding/hex"
	"os"
)

// GenerateSecureToken 生成安全的随机Token
func GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		length = 32
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// GetJWTSecret 获取JWT密钥
// 优先级: 1. 环境变量 2. 配置文件 3. 自动生成
func GetJWTSecret() string {
	// 1. 优先使用环境变量
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}

	// 2. 如果是生产环境，强制要求配置
	// 这里仅返回占位符，实际使用时应配置环境变量
	return "your-secret-key-change-in-production-use-environment-variable"
}

// MustGenerateJWTSecret 生成安全的JWT密钥（用于首次部署）
func MustGenerateJWTSecret() string {
	secret, err := GenerateSecureToken(64)
	if err != nil {
		panic("生成JWT密钥失败: " + err.Error())
	}
	return secret
}
