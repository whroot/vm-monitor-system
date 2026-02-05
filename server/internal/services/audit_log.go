package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLog 审计日志
type AuditLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Module      string    `gorm:"type:varchar(50);not null" json:"module"`       // 模块
	Action     string    `gorm:"type:varchar(50);not null" json:"action"`      // 操作类型
	UserID     *uuid.UUID `gorm:"type:uuid" json:"userId,omitempty"`            // 用户ID
	Username   string    `gorm:"type:varchar(100)" json:"username,omitempty"`    // 用户名
	Resource   string    `gorm:"type:varchar(200)" json:"resource"`            // 操作资源
	ResourceID string    `gorm:"type:varchar(100)" json:"resourceId,omitempty"` // 资源ID
	Details    string    `gorm:"type:text" json:"details,omitempty"`             // 详细信息
	IPAddress  string    `gorm:"type:inet" json:"ipAddress,omitempty"`           // IP地址
	UserAgent  string    `gorm:"type:text" json:"userAgent,omitempty"`           // User-Agent
	Status     string    `gorm:"type:varchar(20)" json:"status"`                // success, failed
	ErrorMsg   string    `gorm:"type:text" json:"errorMsg,omitempty"`             // 错误信息
	CreatedAt  time.Time `json:"createdAt"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// AuditAction 审计操作类型
type AuditAction string

const (
	// 认证相关
	ActionLogin         AuditAction = "LOGIN"
	ActionLogout        AuditAction = "LOGOUT"
	ActionTokenRefresh  AuditAction = "TOKEN_REFRESH"
	ActionPasswordChange AuditAction = "PASSWORD_CHANGE"

	// 用户管理
	ActionUserCreate    AuditAction = "USER_CREATE"
	ActionUserUpdate    AuditAction = "USER_UPDATE"
	ActionUserDelete    AuditAction = "USER_DELETE"
	ActionUserRoleChange AuditAction = "USER_ROLE_CHANGE"

	// 角色管理
	ActionRoleCreate   AuditAction = "ROLE_CREATE"
	ActionRoleUpdate   AuditAction = "ROLE_UPDATE"
	ActionRoleDelete   AuditAction = "ROLE_DELETE"
	ActionRolePermissionChange AuditAction = "ROLE_PERMISSION_CHANGE"

	// VM管理
	ActionVMCreate     AuditAction = "VM_CREATE"
	ActionVMUpdate     AuditAction = "VM_UPDATE"
	ActionVMDelete     AuditAction = "VM_DELETE"
	ActionVMSync       AuditAction = "VM_SYNC"

	// 告警管理
	ActionAlertRuleCreate AuditAction = "ALERT_RULE_CREATE"
	ActionAlertRuleUpdate AuditAction = "ALERT_RULE_UPDATE"
	ActionAlertRuleDelete AuditAction = "ALERT_RULE_DELETE"
	ActionAlertAcknowledge AuditAction = "ALERT_ACKNOWLEDGE"
	ActionAlertResolve   AuditAction = "ALERT_RESOLVE"

	// 系统配置
	ActionConfigUpdate AuditAction = "CONFIG_UPDATE"
	ActionExportData   AuditAction = "EXPORT_DATA"
)

// AuditLogService 审计日志服务
type AuditLogService struct {
	db *gorm.DB
}

// NewAuditLogService 创建审计日志服务
func NewAuditLogService(db *gorm.DB) *AuditLogService {
	return &AuditLogService{db: db}
}

// Create 创建审计日志
func (s *AuditLogService) Create(log *AuditLog) error {
	return s.db.Create(log).Error
}

// CreateWithContext 创建审计日志（带上下文）
func (s *AuditLogService) CreateWithContext(module string, action AuditAction, userID *uuid.UUID, username string, resource string, details string, ipAddress string, status string) error {
	log := &AuditLog{
		Module:     module,
		Action:     string(action),
		UserID:     userID,
		Username:   username,
		Resource:   resource,
		Details:    details,
		IPAddress:  ipAddress,
		Status:     status,
		CreatedAt:  time.Now(),
	}
	return s.db.Create(log).Error
}

// Query 查询审计日志
type AuditLogQuery struct {
	Module     string
	Action     string
	UserID     string
	Resource   string
	StartTime  *time.Time
	EndTime    *time.Time
	Status     string
	Page       int
	PageSize   int
}

// Query 查询审计日志
func (s *AuditLogService) Query(query *AuditLogQuery) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	db := s.db.Model(&AuditLog{})

	// 筛选条件
	if query.Module != "" {
		db = db.Where("module = ?", query.Module)
	}
	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.UserID != "" {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.Resource != "" {
		db = db.Where("resource LIKE ?", "%"+query.Resource+"%")
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	offset := (query.Page - 1) * query.PageSize
	db = db.Order("created_at DESC").Offset(offset).Limit(query.PageSize)

	if err := db.Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetStatistics 获取审计统计
func (s *AuditLogService) GetStatistics(startTime, endTime time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 按模块统计
	var moduleStats []struct {
		Module string
		Count  int64
	}
	if err := s.db.Model(&AuditLog{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("module, COUNT(*) as count").
		Group("module").
		Find(&moduleStats).Error; err != nil {
		return nil, err
	}
	stats["byModule"] = moduleStats

	// 按操作统计
	var actionStats []struct {
		Action string
		Count  int64
	}
	if err := s.db.Model(&AuditLog{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("action, COUNT(*) as count").
		Group("action").
		Find(&actionStats).Error; err != nil {
		return nil, err
	}
	stats["byAction"] = actionStats

	// 成功/失败统计
	var statusStats struct {
		Success int64
		Failed  int64
	}
	if err := s.db.Model(&AuditLog{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("COUNT(CASE WHEN status = 'success' THEN 1 END) as success, COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed").
		Find(&statusStats).Error; err != nil {
		return nil, err
	}
	stats["byStatus"] = statusStats

	// 总数
	var total int64
	if err := s.db.Model(&AuditLog{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	return stats, nil
}

// Cleanup 清理旧日志
func (s *AuditLogService) Cleanup(before time.Time) (int64, error) {
	result := s.db.Where("created_at < ?", before).Delete(&AuditLog{})
	return result.RowsAffected, result.Error
}

// AutoCleanup 自动清理（保留30天）
func (s *AuditLogService) AutoCleanup() error {
	cutoff := time.Now().AddDate(0, 0, -30)
	count, err := s.Cleanup(cutoff)
	if err != nil {
		return err
	}
	if count > 0 {
		fmt.Printf("已清理 %d 条审计日志\n", count)
	}
	return nil
}
