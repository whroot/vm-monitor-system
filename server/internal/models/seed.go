package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedPermissions 创建默认权限
func SeedPermissions(db *gorm.DB) error {
	permissions := []Permission{
		// VM相关权限
		{ID: "vm:read", Name: "查看虚拟机", Resource: "vm", Action: "read", Level: "read", Scope: "global"},
		{ID: "vm:write", Name: "管理虚拟机", Resource: "vm", Action: "write", Level: "write", Scope: "global"},
		{ID: "vm:delete", Name: "删除虚拟机", Resource: "vm", Action: "delete", Level: "write", Scope: "global"},
		{ID: "vm:metrics", Name: "查看监控指标", Resource: "vm", Action: "metrics", Level: "read", Scope: "global"},

		// 告警相关权限
		{ID: "alert:read", Name: "查看告警", Resource: "alert", Action: "read", Level: "read", Scope: "global"},
		{ID: "alert:write", Name: "管理告警规则", Resource: "alert", Action: "write", Level: "write", Scope: "global"},
		{ID: "alert:acknowledge", Name: "确认告警", Resource: "alert", Action: "acknowledge", Level: "write", Scope: "global"},
		{ID: "alert:resolve", Name: "解决告警", Resource: "alert", Action: "resolve", Level: "write", Scope: "global"},

		// 用户管理权限
		{ID: "user:read", Name: "查看用户", Resource: "user", Action: "read", Level: "read", Scope: "global"},
		{ID: "user:write", Name: "管理用户", Resource: "user", Action: "write", Level: "admin", Scope: "global"},
		{ID: "user:delete", Name: "删除用户", Resource: "user", Action: "delete", Level: "admin", Scope: "global"},

		// 角色权限
		{ID: "role:read", Name: "查看角色", Resource: "role", Action: "read", Level: "read", Scope: "global"},
		{ID: "role:write", Name: "管理角色", Resource: "role", Action: "write", Level: "admin", Scope: "global"},

		// 系统权限
		{ID: "system:settings", Name: "系统设置", Resource: "system", Action: "settings", Level: "admin", Scope: "global"},
		{ID: "system:logs", Name: "查看日志", Resource: "system", Action: "logs", Level: "admin", Scope: "global"},
	}

	for _, p := range permissions {
		var existing Permission
		if err := db.Where("id = ?", p.ID).First(&existing).Error; err == gorm.ErrRecordNotFound {
			p.CreatedAt = time.Now()
			if err := db.Create(&p).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// SeedRoles 创建默认角色
func SeedRoles(db *gorm.DB) error {
	adminDesc := "系统管理员，拥有所有权限"
	operatorDesc := "运维人员，可以管理VM和告警"
	viewerDesc := "只读用户，只能查看信息"

	roles := []Role{
		{
			ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			Name:        "系统管理员",
			Description: &adminDesc,
			Level:       1,
			Path:        "/admin",
			IsSystem:    true,
		},
		{
			ID:          uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:        "运维人员",
			Description: &operatorDesc,
			Level:       5,
			Path:        "/ops",
			IsSystem:    true,
		},
		{
			ID:          uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			Name:        "只读用户",
			Description: &viewerDesc,
			Level:       10,
			Path:        "/viewer",
			IsSystem:    true,
		},
	}

	for _, r := range roles {
		var existing Role
		if err := db.Where("id = ?", r.ID).First(&existing).Error; err == gorm.ErrRecordNotFound {
			r.CreatedAt = time.Now()
			if err := db.Create(&r).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// SeedRolePermissions 分配默认权限
func SeedRolePermissions(db *gorm.DB) error {
	// 系统管理员拥有所有权限
	adminRoleID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var adminRole Role
	if err := db.Preload("Permissions").First(&adminRole, adminRoleID).Error; err == nil {
		if len(adminRole.Permissions) == 0 {
			var permissions []Permission
			db.Find(&permissions)
			db.Model(&adminRole).Association("Permissions").Append(permissions)
		}
	}

	// 运维人员拥有VM和告警权限
	operatorRoleID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	var operatorRole Role
	if err := db.Preload("Permissions").First(&operatorRole, operatorRoleID).Error; err == nil {
		if len(operatorRole.Permissions) == 0 {
			var permissions []Permission
			db.Where("id IN ?", []string{
				"vm:read", "vm:write", "vm:delete", "vm:metrics",
				"alert:read", "alert:write", "alert:acknowledge", "alert:resolve",
			}).Find(&permissions)
			db.Model(&operatorRole).Association("Permissions").Append(permissions)
		}
	}

	// 只读用户只能查看
	viewerRoleID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	var viewerRole Role
	if err := db.Preload("Permissions").First(&viewerRole, viewerRoleID).Error; err == nil {
		if len(viewerRole.Permissions) == 0 {
			var permissions []Permission
			db.Where("id IN ?", []string{
				"vm:read", "vm:metrics",
				"alert:read",
			}).Find(&permissions)
			db.Model(&viewerRole).Association("Permissions").Append(permissions)
		}
	}

	return nil
}

// SeedDefaultData 执行所有初始化数据
func SeedDefaultData(db *gorm.DB) error {
	if err := SeedPermissions(db); err != nil {
		return err
	}
	if err := SeedRoles(db); err != nil {
		return err
	}
	if err := SeedRolePermissions(db); err != nil {
		return err
	}
	return nil
}

// AssignRoleToUser 为用户分配角色
func AssignRoleToUser(db *gorm.DB, userID, roleID uuid.UUID) error {
	userRole := UserRole{
		ID:     uuid.New(),
		UserID: userID,
		RoleID: roleID,
	}
	return db.Create(&userRole).Error
}
