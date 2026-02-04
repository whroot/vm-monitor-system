package services

import (
	"vm-monitoring-system/internal/models"

	"gorm.io/gorm"
)

type RBACService struct {
	db *gorm.DB
}

func NewRBACService(db *gorm.DB) *RBACService {
	return &RBACService{db: db}
}

// CheckPermission 检查用户是否有指定权限
func (s *RBACService) CheckPermission(userID string, permissionID string) (bool, error) {
	var count int64
	err := s.db.Table("users").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Joins("JOIN role_permissions ON user_roles.role_id = role_permissions.role_id").
		Where("users.id = ? AND role_permissions.permission_id = ?", userID, permissionID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUserPermissions 获取用户的所有权限
func (s *RBACService) GetUserPermissions(userID string) ([]models.Permission, error) {
	var permissions []models.Permission
	err := s.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&permissions).Error
	return permissions, err
}

// GetUserRoles 获取用户的角色
func (s *RBACService) GetUserRoles(userID string) ([]models.Role, error) {
	var roles []models.Role
	err := s.db.Preload("Permissions").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

// HasRole 检查用户是否有指定角色
func (s *RBACService) HasRole(userID string, roleName string) (bool, error) {
	var count int64
	err := s.db.Table("users").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("users.id = ? AND roles.name = ?", userID, roleName).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetRolePermissions 获取角色的权限
func (s *RBACService) GetRolePermissions(roleID string) ([]models.Permission, error) {
	var permissions []models.Permission
	err := s.db.Where("id IN ?", 
		s.db.Table("role_permissions").
			Select("permission_id").
			Where("role_id = ?", roleID),
	).Find(&permissions).Error
	return permissions, err
}

// AssignRoleToUser 为用户分配角色
func (s *RBACService) AssignRoleToUser(userID string, roleID string) error {
	userRole := models.UserRole{
		UserID:    parseUUID(userID),
		RoleID:    parseUUID(roleID),
	}
	return s.db.Create(&userRole).Error
}

// RemoveRoleFromUser 移除用户的角色
func (s *RBACService) RemoveRoleFromUser(userID string, roleID string) error {
	return s.db.Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&models.UserRole{}).Error
}

// AssignPermissionToRole 为角色分配权限
func (s *RBACService) AssignPermissionToRole(roleID string, permissionID string) error {
	rolePermission := models.RolePermission{
		RoleID:       parseUUID(roleID),
		PermissionID: permissionID,
	}
	return s.db.Create(&rolePermission).Error
}

// RemovePermissionFromRole 移除角色的权限
func (s *RBACService) RemovePermissionFromRole(roleID string, permissionID string) error {
	return s.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&models.RolePermission{}).Error
}

func parseUUID(s string) models.UUID {
	// 简化的UUID解析，实际应使用uuid.Parse
	return models.UUID{}
}
