package api

import (
	"fmt"
	"os"
	"strings"

	"vm-monitoring-system/internal/config"
	"vm-monitoring-system/internal/logger"
)

// ValidateConfig 验证配置文件，在启动时调用
func ValidateConfig(cfg *config.Config) error {
	var warnings []string

	// 1. 检查JWT密钥
	if cfg.JWT.Secret == "" || cfg.JWT.Secret == "your-secret-key-change-in-production" {
		// 尝试从环境变量获取
		envSecret := os.Getenv("JWT_SECRET")
		if envSecret == "" {
			warnings = append(warnings, "⚠️  JWT_SECRET未设置，使用自动生成的密钥")
			// 自动生成密钥
			cfg.JWT.Secret = generateFallbackSecret()
		} else {
			cfg.JWT.Secret = envSecret
		}
	}

	// 2. 验证JWT密钥强度
	if len(cfg.JWT.Secret) < 32 {
		warnings = append(warnings, "⚠️  JWT密钥长度不足32字符，建议使用64字符以上")
	}

	// 3. 检查密码强度配置
	if cfg.JWT.Secret == "" {
		warnings = append(warnings, "⚠️  未配置JWT密钥，生产环境必须配置")
	}

	// 记录警告
	if len(warnings) > 0 {
		for _, w := range warnings {
			logger.Warn(w)
		}
	}

	return nil
}

// generateFallbackSecret 生成回退密钥（仅用于开发环境）
func generateFallbackSecret() string {
	// 使用主机名和时间戳生成伪随机密钥
	hostname, _ := os.Hostname()
	return fmt.Sprintf("dev-secret-%s-%d", hostname, os.Getpid())
}

// CheckProductionSecurity 检查生产环境安全性
func CheckProductionSecurity() (bool, []string) {
	var issues []string

	// 1. 检查JWT密钥
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		issues = append(issues, "❌ JWT_SECRET环境变量未设置")
	} else if len(secret) < 32 {
		issues = append(issues, "❌ JWT_SECRET长度不足32字符")
	}

	// 2. 检查数据库密码
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		issues = append(issues, "❌ DB_PASSWORD环境变量未设置")
	}

	// 3. 检查Redis密码
	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		issues = append(issues, "⚠️  REDIS_PASSWORD环境变量未设置（建议设置）")
	}

	// 4. 检查是否在生产模式
	mode := strings.ToLower(os.Getenv("APP_MODE"))
	if mode != "production" && mode != "prod" {
		issues = append(issues, "⚠️  APP_MODE未设置为production")
	}

	return len(issues) == 0, issues
}

// PrintSecurityCheckResult 打印安全检查结果
func PrintSecurityCheckResult() {
	fmt.Println("\n🔒 安全检查...")
	passed, issues := CheckProductionSecurity()

	if passed {
		fmt.Println("✅ 所有安全检查通过")
	} else {
		fmt.Println("❌ 发现安全问题:")
		for _, issue := range issues {
			fmt.Println("   ", issue)
		}
	}
	fmt.Println()
}
