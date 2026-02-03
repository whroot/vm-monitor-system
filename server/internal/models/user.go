package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username           string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email              string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	PasswordHash       string         `gorm:"type:varchar(255);not null" json:"-"`
	Name               string         `gorm:"type:varchar(100);not null" json:"name"`
	Phone              *string        `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Department         *string        `gorm:"type:varchar(100)" json:"department,omitempty"`
	Status             string         `gorm:"type:vvarchar(20);not null;default:'active'" json:"status"`
	MustChangePassword bool           `gorm:"not null;default:false" json:"mustChangePassword"`
	MFAEnabled         bool           `gorm:"not null;default:false" json:"mfaEnabled"`
	MFASecret          *string        `gorm:"type:varchar(255)" json:"-"`
	LastLoginAt        *time.Time     `json:"lastLoginAt,omitempty"`
	LastLoginIP        *string        `gorm:"type:inet" json:"lastLoginIp,omitempty"`
	LoginFailCount     int            `gorm:"not null;default:0" json:"-"`
	LockedUntil        *time.Time     `json:"-"`
	PasswordExpiredAt  *time.Time     `json:"-"`
	Preferences        UserPreferences `gorm:"type:jsonb;not null" json:"preferences"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
	CreatedBy          *uuid.UUID     `gorm:"type:uuid" json:"-"`
	UpdatedBy          *uuid.UUID     `gorm:"type:uuid" json:"-"`

	// 关联
	Roles []Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// UserPreferences 用户偏好设置
type UserPreferences struct {
	Language   string `json:"language"`
	Theme      string `json:"theme"`
	Timezone   string `json:"timezone"`
	DateFormat string `json:"dateFormat"`
}

// Value 实现driver.Valuer接口
func (p UserPreferences) Value() (interface{}, error) {
	return JSONMap{
		"language":   p.Language,
		"theme":      p.Theme,
		"timezone":   p.Timezone,
		"dateFormat": p.DateFormat,
	}.Value()
}

// Scan 实现sql.Scanner接口
func (p *UserPreferences) Scan(value interface{}) error {
	var m JSONMap
	if err := m.Scan(value); err != nil {
		return err
	}

	if v, ok := m["language"].(string); ok {
		p.Language = v
	}
	if v, ok := m["theme"].(string); ok {
		p.Theme = v
	}
	if v, ok := m["timezone"].(string); ok {
		p.Timezone = v
	}
	if v, ok := m["dateFormat"].(string); ok {
		p.DateFormat = v
	}

	return nil
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Preferences.Language == "" {
		u.Preferences.Language = "zh-CN"
	}
	if u.Preferences.Theme == "" {
		u.Preferences.Theme = "dark"
	}
	if u.Preferences.Timezone == "" {
		u.Preferences.Timezone = "Asia/Shanghai"
	}
	if u.Preferences.DateFormat == "" {
		u.Preferences.DateFormat = "YYYY-MM-DD"
	}
	return nil
}

// Role 角色模型
type Role struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Description *string    `gorm:"type:text" json:"description,omitempty"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parentId,omitempty"`
	Level       int        `gorm:"not null;default:1" json:"level"`
	Path        string     `gorm:"type:varchar(500);not null" json:"path"`
	IsSystem    bool       `gorm:"not null;default:false" json:"isSystem"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	CreatedBy   *uuid.UUID `gorm:"type:uuid" json:"-"`
	UpdatedBy   *uuid.UUID `gorm:"type:uuid" json:"-"`

	// 关联
	Parent       *Role       `gorm:"foreignkey:ParentID" json:"parent,omitempty"`
	Children     []Role      `gorm:"foreignkey:ParentID" json:"children,omitempty"`
	Permissions  []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	UserCount    int         `gorm:"-" json:"userCount"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// Permission 权限模型
type Permission struct {
	ID          string    `gorm:"type:varchar(100);primary_key" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	Resource    string    `gorm:"type:varchar(50);not null" json:"resource"`
	Action      string    `gorm:"type:varchar(50);not null" json:"action"`
	Level       string    `gorm:"type:varchar(20);not null;default:'read'" json:"level"`
	Scope       string    `gorm:"type:varchar(20);default:'global'" json:"scope"`
	CreatedAt   time.Time `json:"createdAt"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// UserRole 用户角色关联
type UserRole struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_role" json:"userId"`
	RoleID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_role" json:"roleId"`
	CreatedAt time.Time `json:"createdAt"`
}

// TableName 指定表名
func (UserRole) TableName() string {
	return "user_roles"
}

// RolePermission 角色权限关联
type RolePermission struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RoleID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_role_permission" json:"roleId"`
	PermissionID string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_role_permission" json:"permissionId"`
	IsInherited  bool      `gorm:"not null;default:false" json:"isInherited"`
	CreatedAt    time.Time `json:"createdAt"`
}

// TableName 指定表名
func (RolePermission) TableName() string {
	return "role_permissions"
}

// UserSession 用户会话
type UserSession struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	AccessTokenHash  string     `gorm:"type:varchar(255);not null;index" json:"-"`
	RefreshTokenHash *string    `gorm:"type:varchar(255)" json:"-"`
	ExpiresAt        time.Time  `json:"expiresAt"`
	RefreshExpiresAt *time.Time `json:"-"`
	IPAddress        *string    `gorm:"type:inet" json:"ipAddress,omitempty"`
	UserAgent        *string    `gorm:"type:text" json:"userAgent,omitempty"`
	IsActive         bool       `gorm:"not null;default:true" json:"isActive"`
	RevokedAt        *time.Time `json:"-"`
	RevokedBy        *uuid.UUID `gorm:"type:uuid" json:"-"`
	CreatedAt        time.Time  `json:"createdAt"`
	LastActivityAt   time.Time  `json:"lastActivityAt"`
}

// TableName 指定表名
func (UserSession) TableName() string {
	return "user_sessions"
}
