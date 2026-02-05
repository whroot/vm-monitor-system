package services

import (
	"time"

	"vm-monitoring-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RBACInitializer struct {
	db *gorm.DB
}

func NewRBACInitializer(db *gorm.DB) *RBACInitializer {
	return &RBACInitializer{db: db}
}

// Initialize 初始化RBAC系统
func (i *RBACInitializer) Initialize() error {
	// 创建权限
	if err := i.createPermissions(); err != nil {
		return err
	}

	// 创建角色
	if err := i.createRoles(); err != nil {
		return err
	}

	// 创建默认管理员用户
	if err := i.createDefaultAdmin(); err != nil {
		return err
	}

	return nil
}

func (i *RBACInitializer) createPermissions() error {
	permissions := []models.Permission{
		// VM管理权限
		{ID: "vm:read", Name: "查看虚拟机", Resource: "vm", Action: "read", Level: "read", Scope: "global"},
		{ID: "vm:write", Name: "管理虚拟机", Resource: "vm", Action: "write", Level: "write", Scope: "global"},
		{ID: "vm:delete", Name: "删除虚拟机", Resource: "vm", Action: "delete", Level: "delete", Scope: "global"},
		{ID: "vm:export", Name: "导出虚拟机", Resource: "vm", Action: "export", Level: "write", Scope: "global"},
		
		// 告警权限
		{ID: "alert:read", Name: "查看告警", Resource: "alert", Action: "read", Level: "read", Scope: "global"},
		{ID: "alert:write", Name: "管理告警", Resource: "alert", Action: "write", Level: "write", Scope: "global"},
		{ID: "alert:acknowledge", Name: "确认告警", Resource: "alert", Action: "acknowledge", Level: "write", Scope: "global"},
		
		// 用户管理权限
		{ID: "user:read", Name: "查看用户", Resource: "user", Action: "read", Level: "read", Scope: "global"},
		{ID: "user:write", Name: "管理用户", Resource: "user", Action: "write", Level: "write", Scope: "global"},
		{ID: "user:delete", Name: "删除用户", Resource: "user", Action: "delete", Level: "delete", Scope: "global"},
		
		// 角色管理权限
		{ID: "role:read", Name: "查看角色", Resource: "role", Action: "read", Level: "read", Scope: "global"},
		{ID: "role:write", Name: "管理角色", Resource: "role", Action: "write", Level: "write", Scope: "global"},
		
		// 系统管理权限
		{ID: "system:read", Name: "查看系统", Resource: "system", Action: "read", Level: "read", Scope: "global"},
		{ID: "system:config", Name: "系统配置", Resource: "system", Action: "config", Level: "admin", Scope: "global"},
		{ID: "system:logs", Name: "系统日志", Resource: "system", Action: "logs", Level: "read", Scope: "global"},
		
		// 历史数据权限
		{ID: "history:read", Name: "查看历史", Resource: "history", Action: "read", Level: "read", Scope: "global"},
		{ID: "history:export", Name: "导出历史", Resource: "history", Action: "export", Level: "write", Scope: "global"},
	}

	for _, p := range permissions {
		var existing models.Permission
		if i.db.Where("id = ?", p.ID).First(&existing).Error == gorm.ErrRecordNotFound {
			if err := i.db.Create(&p).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func (i *RBACInitializer) createRoles() error {
	roles := []struct {
		Name        string
		Description string
		Level       int
		IsSystem    bool
		Permissions []string
	}{
		{
			Name:        "超级管理员",
			Description: "系统超级管理员，拥有所有权限",
			Level:       100,
			IsSystem:    true,
			Permissions: []string{
				"vm:read", "vm:write", "vm:delete", "vm:export",
				"alert:read", "alert:write", "alert:acknowledge",
				"user:read", "user:write", "user:delete",
				"role:read", "role:write",
				"system:read", "system:config", "system:logs",
				"history:read", "history:export",
			},
		},
		{
			Name:        "管理员",
			Description: "系统管理员，拥有大部分权限",
			Level:       50,
			IsSystem:    true,
			Permissions: []string{
				"vm:read", "vm:write", "vm:export",
				"alert:read", "alert:write", "alert:acknowledge",
				"user:read", "user:write",
				"role:read",
				"system:read", "system:logs",
				"history:read", "history:export",
			},
		},
		{
			Name:        "运维人员",
			Description: "负责日常运维操作",
			Level:       30,
			IsSystem:    true,
			Permissions: []string{
				"vm:read", "vm:write",
				"alert:read", "alert:acknowledge",
				"history:read", "history:export",
			},
		},
		{
			Name:        "只读用户",
			Description: "只读用户，只能查看数据",
			Level:       10,
			IsSystem:    true,
			Permissions: []string{
				"vm:read",
				"alert:read",
				"history:read",
			},
		},
	}

	for _, r := range roles {
		var existing models.Role
		if i.db.Where("name = ?", r.Name).First(&existing).Error == gorm.ErrRecordNotFound {
			role := models.Role{
				ID:          uuid.New(),
				Name:        r.Name,
				Level:       r.Level,
				IsSystem:    r.IsSystem,
				Path:        "/" + r.Name,
				CreatedAt:   time.Now(),
			}
			
			if r.Description != "" {
				role.Description = &r.Description
			}

			if err := i.db.Create(&role).Error; err != nil {
				return err
			}

			// 分配权限
			for _, permID := range r.Permissions {
				if err := i.db.Create(&models.RolePermission{
					ID:           uuid.New(),
					RoleID:       role.ID,
					PermissionID: permID,
					CreatedAt:    time.Now(),
				}).Error; err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (i *RBACInitializer) createDefaultAdmin() error {
	// 检查是否已存在admin用户
	var existing models.User
	if i.db.Where("username = ?", "admin").First(&existing).Error != gorm.ErrRecordNotFound {
		return nil // admin用户已存在
	}

	// 查找超级管理员角色
	var role models.Role
	if err := i.db.Where("name = ?", "超级管理员").First(&role).Error; err != nil {
		return err
	}

	// 创建admin用户
	// ⚠️ 首次登录必须修改密码
	passwordHash, _ := hashPassword("admin123")
	user := models.User{
		ID:                 uuid.New(),
		Username:           "admin",
		Email:              "admin@vm-monitor.com",
		PasswordHash:       passwordHash,
		Name:               "系统管理员",
		Status:             "active",
		MustChangePassword: true, // ⚠️ 首次登录必须修改密码
		Preferences: models.UserPreferences{
			Language:   "zh-CN",
			Theme:      "dark",
			Timezone:   "Asia/Shanghai",
			DateFormat: "YYYY-MM-DD",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := i.db.Create(&user).Error; err != nil {
		return err
	}

	// 分配超级管理员角色
	if err := i.db.Create(&models.UserRole{
		ID:        uuid.New(),
		UserID:    user.ID,
		RoleID:    role.ID,
		CreatedAt: time.Now(),
	}).Error; err != nil {
		return err
	}

	return nil
}
