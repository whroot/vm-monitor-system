package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuditMiddleware 审计日志中间件
type AuditMiddleware struct {
	auditLogService *AuditLogService
}

// AuditLogService 审计日志服务（API层包装）
type AuditLogService struct {
	db *gorm.DB
}

// NewAuditLogService 创建审计日志服务
func NewAuditLogService(db *gorm.DB) *AuditLogService {
	return &AuditLogService{db: db}
}

// AuditLog 审计日志
type AuditLog struct {
	ID         string    `json:"id"`
	Module     string    `json:"module"`
	Action     string    `json:"action"`
	UserID     string    `json:"userId"`
	Username   string    `json:"username"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resourceId"`
	Details    string    `json:"details"`
	IPAddress  string    `json:"ipAddress"`
	Status     string    `json:"status"`
	ErrorMsg   string    `json:"errorMsg"`
	CreatedAt  time.Time `json:"createdAt"`
}

// AuditLogEntry 审计日志条目
type AuditLogEntry struct {
	Module    string
	Action    string
	UserID    string
	Username  string
	Resource  string
	Details   interface{}
	Status    string
	ErrorMsg  string
}

// NewAuditMiddleware 创建审计中间件
func NewAuditMiddleware(db *gorm.DB) *AuditMiddleware {
	return &AuditMiddleware{
		auditLogService: NewAuditLogService(db),
	}
}

// Audit 记录审计日志
func (m *AuditMiddleware) Audit(entry AuditLogEntry) {
	log := AuditLog{
		ID:        generateAuditID(),
		Module:    entry.Module,
		Action:    entry.Action,
		UserID:    entry.UserID,
		Username:  entry.Username,
		Resource: entry.Resource,
		Status:    entry.Status,
		ErrorMsg: entry.ErrorMsg,
		CreatedAt: time.Now(),
	}

	// 序列化详情
	if entry.Details != nil {
		if jsonData, err := json.Marshal(entry.Details); err == nil {
			log.Details = string(jsonData)
		}
	}

	// 从上下文中获取IP
	// 实际实现需要传入gin.Context

	log.Printf("审计日志: %s - %s - %s", entry.Module, entry.Action, entry.Resource)
}

// AuditMiddleware Gin中间件函数
func AuditMiddlewareFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 记录请求信息
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		
		// 继续处理请求
		c.Next()
		
		// 计算耗时
		duration := time.Since(start)
		
		// 获取响应状态
		status := "success"
		if c.Writer.Status() >= 400 {
			status = "failed"
		}
		
		// 记录审计日志
		log.Printf("审计: %s %s %s %d %s", 
			c.Request.Method,
			c.Request.URL.Path,
			status,
			c.Writer.Status(),
			duration)
	}
}

// generateAuditID 生成审计ID
func generateAuditID() string {
	return "audit_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

// randomString 生成随机字符串
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// AuditConfig 审计配置
type AuditConfig struct {
	Enabled     bool          // 是否启用
	ExcludePaths []string      // 排除的路径
	Modules     []string      // 需要审计的模块
}

// DefaultAuditConfig 默认审计配置
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		Enabled:     true,
		ExcludePaths: []string{"/health", "/metrics", "/ws"},
		Modules:     []string{"auth", "user", "role", "vm", "alert", "config"},
	}
}
