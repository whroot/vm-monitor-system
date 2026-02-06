package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// StringPtr 返回字符串指针
func StringPtr(s string) *string {
	return &s
}

// TimePtr 返回时间指针
func TimePtr(t time.Time) *time.Time {
	return &t
}

// HashString 计算字符串哈希
func HashString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// HashPassword 计算密码哈希
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ValidatePasswordComplexity 验证密码复杂度
func ValidatePasswordComplexity(password string) bool {
	// 最少8位
	if len(password) < 8 {
		return false
	}

	// 包含大写字母
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	// 包含小写字母
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	// 包含数字
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	// 包含特殊字符
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}
