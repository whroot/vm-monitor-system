package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"vm-monitoring-system/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// 上下文键
type contextKey string

const (
	contextKeyUserID        contextKey = "user_id"
	contextKeyUser          contextKey = "user"
	contextKeyPermissions   contextKey = "permissions"
	contextKeyRequestID     contextKey = "request_id"
)

// JWTAuth JWT认证中间件
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header中获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供认证信息",
			})
			c.Abort()
			return
		}

		// 提取Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证格式错误",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token无效或已过期",
			})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token无效",
			})
			c.Abort()
			return
		}

		// 提取声明
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token解析失败",
			})
			c.Abort()
			return
		}

		// 验证Token类型
		if tokenType, ok := claims["type"].(string); !ok || tokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token类型错误",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		userID, _ := claims["sub"].(string)
		c.Set(string(contextKeyUserID), userID)

		// 从数据库获取完整用户信息
		var user models.User
		if err := models.DB.Preload("Roles").First(&user, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查用户状态
		if user.Status == "locked" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "账户已锁定",
			})
			c.Abort()
			return
		}

		if user.Status == "expired" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "账户已过期",
			})
			c.Abort()
			return
		}

		c.Set(string(contextKeyUser), user)

		// 获取用户权限
		permissions := getUserPermissions(user)
		c.Set(string(contextKeyPermissions), permissions)

		c.Next()
	}
}

// PermissionCheck 权限检查中间件
func PermissionCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 权限检查逻辑
		// 实际项目中根据具体业务需求实现
		c.Next()
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 生成请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Header("X-Request-ID", requestID)
		}
		c.Set(string(contextKeyRequestID), requestID)

		c.Next()

		// 记录日志
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// 这里可以使用日志库记录
		_ = latency
		_ = clientIP
		_ = method
		_ = statusCode
		_ = path
	}
}

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 处理错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			c.JSON(-1, gin.H{
				"code":    500,
				"message": err.Error(),
			})
		}
	}
}

// getUserPermissions 获取用户权限列表
func getUserPermissions(user models.User) []string {
	// 从Redis缓存获取
	cacheKey := fmt.Sprintf("user_permissions:%s", user.ID)
	ctx := context.Background()

	if models.Cache != nil {
		cached, err := models.Cache.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			// 解析缓存数据
			var permissions []string
			// TODO: 解析JSON
			_ = permissions
		}
	}

	// 从数据库获取
	var permissions []string
	for _, role := range user.Roles {
		var rolePerms []models.RolePermission
		models.DB.Where("role_id = ?", role.ID).Find(&rolePerms)
		for _, rp := range rolePerms {
			permissions = append(permissions, rp.PermissionID)
		}
	}

	// 去重
	permissions = uniqueStrings(permissions)

	// 缓存15分钟
	if models.Cache != nil {
		// TODO: 缓存权限列表
		_ = ctx
	}

	return permissions
}

// uniqueStrings 字符串去重
func uniqueStrings(strs []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, str := range strs {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}
	return result
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// GetUserID 从上下文中获取用户ID
func GetUserID(c *gin.Context) string {
	userID, _ := c.Get(string(contextKeyUserID))
	if id, ok := userID.(string); ok {
		return id
	}
	return ""
}

// GetUser 从上下文中获取用户
func GetUser(c *gin.Context) *models.User {
	user, exists := c.Get(string(contextKeyUser))
	if !exists {
		return nil
	}
	if u, ok := user.(models.User); ok {
		return &u
	}
	return nil
}

// GetPermissions 从上下文中获取权限
func GetPermissions(c *gin.Context) []string {
	perms, exists := c.Get(string(contextKeyPermissions))
	if !exists {
		return nil
	}
	if p, ok := perms.([]string); ok {
		return p
	}
	return nil
}

// HasPermission 检查是否有权限
func HasPermission(c *gin.Context, permission string) bool {
	perms := GetPermissions(c)
	for _, p := range perms {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}
