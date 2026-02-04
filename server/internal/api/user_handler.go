package api

import (
	"net/http"
	"time"

	"vm-monitoring-system/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserHandler 用户处理器
type UserHandler struct {
	db *gorm.DB
}

// NewUserHandler 创建用户处理器
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// List 获取用户列表
func (h *UserHandler) List(c *gin.Context) {
	var users []User
	if err := h.db.Preload("Roles").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户列表失败",
		})
		return
	}

	// 转换响应数据
	var userList []gin.H
	for _, u := range users {
		userList = append(userList, gin.H{
			"id":        u.ID,
			"username":  u.Username,
			"email":     u.Email,
			"name":      u.Name,
			"phone":     u.Phone,
			"department": u.Department,
			"status":    u.Status,
			"roles":     u.Roles,
			"createdAt": u.CreatedAt,
			"updatedAt": u.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list": userList,
		},
	})
}

// Get 获取用户详情
func (h *UserHandler) Get(c *gin.Context) {
	userID := c.Param("id")
	
	var user User
	if err := h.db.Preload("Roles").Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户详情失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
			"message": "获取成功",
			"data": gin.H{
				"id":        user.ID,
				"username":  user.Username,
				"email":     user.Email,
				"name":      user.Name,
				"phone":     user.Phone,
				"department": user.Department,
				"status":    user.Status,
				"roles":     user.Roles,
				"preferences": user.Preferences,
				"createdAt": user.CreatedAt,
				"updatedAt": user.UpdatedAt,
			},
		})
}

// Create 创建用户
func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Username    string `json:"username" binding:"required,min=3,max=50,unique=users.username"`
		Email       string `json:"email" binding:"required,email,unique=users.email"`
		Password    string `json:"password" binding:"required,min=8,max=100"`
		Name        string `json:"name" binding:"required,max=100"`
		Phone       *string `json:"phone"`
		Department  *string `json:"department"`
		Status      string `json:"status" binding:"omitempty,oneof=active inactive locked"`
		RoleIDs     []uuid.UUID `json:"roleIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"errors": err.Error(),
		})
		return
	}

	// 检查用户名和邮箱是否已存在
	var existingUser User
	if h.db.Where("username = ?", req.Username).First(&existingUser).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户名已存在",
		})
		return
	}

	if h.db.Where("email = ?", req.Email).First(&existingUser).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "邮箱已存在",
		})
		return
	}

	// 创建用户
	user := User{
		Username:           req.Username,
		Email:              req.Email,
		PasswordHash:       hashPassword(req.Password),
		Name:               req.Name,
		Phone:              req.Phone,
		Department:         req.Department,
		Status:             req.Status,
		Preferences: UserPreferences{
			Language:   "zh-CN",
			Theme:      "dark",
			Timezone:   "Asia/Shanghai",
			DateFormat: "YYYY-MM-DD",
		},
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建用户
	if err := tx.Create(&user).Error; err != nil {
			tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建用户失败",
		})
		return
	}

	// 分配角色
	if len(req.RoleIDs) > 0 {
		var roles []Role
		if err := tx.Where("id IN ?", req.RoleIDs).Find(&roles).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "分配角色失败",
			})
			return
		}

		// 创建用户角色关联
		userRoles := make([]UserRole, len(roles))
		for i, role := range roles {
			userRoles[i] = UserRole{
				UserID: user.ID,
				RoleID: role.ID,
				CreatedAt: time.Now(),
			}
		}

		if err := tx.Create(&userRoles).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "分配角色失败",
			})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建用户失败",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建成功",
		"data": gin.H{
			"id": user.ID,
		},
	})
}

// Update 更新用户
func (h *UserHandler) Update(c *gin.Context) {
	userID := c.Param("id")
	
	var req struct {
		Name        string `json:"name" binding:"max=100"`
		Phone       *string `json:"phone"`
		Department  *string `json:"department"`
		Status      string `json:"status" binding:"omitempty,oneof=active inactive locked"`
		RoleIDs     []uuid.UUID `json:"roleIds"`
		Preferences UserPreferences `json:"preferences"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"errors": err.Error(),
		})
		return
	}

	// 查找用户
	var user User
	if err := h.db.Preload("Roles").Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户失败",
		})
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Phone != nil {
		updates["phone"] = req.Phone
	}
	if req.Department != nil {
		updates["department"] = req.Department
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Preferences.Language != "" || req.Preferences.Theme != "" || req.Preferences.Timezone != "" || req.Preferences.DateFormat != "" {
		updates["preferences"] = req.Preferences
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新用户
	if err := tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新用户失败",
		})
		return
	}

	// 更新角色
	if req.RoleIDs != nil {
		// 删除现有角色
		if err := tx.Where("user_id = ?", userID).Delete(&UserRole{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新角色失败",
			})
			return
		}

		// 分配新角色
		if len(req.RoleIDs) > 0 {
			var roles []Role
			if err := tx.Where("id IN ?", req.RoleIDs).Find(&roles).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "更新角色失败",
				})
				return
			}

			// 创建用户角色关联
			userRoles := make([]UserRole, len(roles))
			for i, role := range roles {
				userRoles[i] = UserRole{
					UserID: user.ID,
					RoleID: role.ID,
					CreatedAt: time.Now(),
				}
			}

			if err := tx.Create(&userRoles).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "更新角色失败",
				})
				return
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// Delete 删除用户
func (h *UserHandler) Delete(c *gin.Context) {
	userID := c.Param("id")
	
	// 检查用户是否存在
	var user User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户失败",
		})
		return
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除用户角色关联
	if err := tx.Where("user_id = ?", userID).Delete(&UserRole{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除用户失败",
		})
		return
	}

	// 删除用户
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除用户失败",
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// ResetPassword 重置密码
func (h *UserHandler) ResetPassword(c *gin.Context) {
	userID := c.Param("id")
	
	var req struct {
		NewPassword string `json:"newPassword" binding:"required,min=8,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"errors": err.Error(),
		})
		return
	}

	// 查找用户
	var user User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户失败",
		})
		return
	}

	// 更新密码
	if err := h.db.Model(&user).Update("password_hash", hashPassword(req.NewPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "重置密码失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码重置成功",
	})
}

// BatchUpdateStatus 批量更新状态
func (h *UserHandler) BatchUpdateStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "批量更新成功",
	})
}

// GetMyPermissions 获取当前用户权限
func (h *UserHandler) GetMyPermissions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    []string{},
	})
}

// GetUserPermissions 获取用户权限详情
func (h *UserHandler) GetUserPermissions(c *gin.Context) {
	userID := c.Param("id")
	_ = userID

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    gin.H{},
	})
}

// RoleHandler 角色处理器
type RoleHandler struct {
	db *gorm.DB
}

// NewRoleHandler 创建角色处理器
func NewRoleHandler(db *gorm.DB) *RoleHandler {
	return &RoleHandler{db: db}
}

// List 获取角色列表
func (h *RoleHandler) List(c *gin.Context) {
	var roles []Role
	if err := h.db.Preload("Permissions").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取角色列表失败",
		})
		return
	}

	// 转换响应数据
	var roleList []gin.H
	for _, r := range roles {
		roleList = append(roleList, gin.H{
			"id":          r.ID,
			"name":        r.Name,
			"description": r.Description,
			"level":      r.Level,
			"isSystem":   r.IsSystem,
			"permissions": r.Permissions,
			"createdAt":  r.CreatedAt,
			"updatedAt":  r.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    roleList,
	})
}

// Get 获取角色详情
func (h *RoleHandler) Get(c *gin.Context) {
	roleID := c.Param("id")
	
	var role Role
	if err := h.db.Preload("Permissions").Where("id = ?", roleID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "角色不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取角色详情失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"id":          role.ID,
			"name":        role.Name,
			"description": role.Description,
			"level":      role.Level,
			"isSystem":   role.IsSystem,
			"permissions": role.Permissions,
			"createdAt":  role.CreatedAt,
			"updatedAt":  role.UpdatedAt,
		},
	})
}

// Create 创建角色
func (h *RoleHandler) Create(c *gin.Context) {
	var req struct {
		Name         string   `json:"name" binding:"required,max=100"`
		Description  *string  `json:"description"`
		Level        int     `json:"level" binding:"omitempty,min=1,max=100"`
		PermissionIDs []string `json:"permissionIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"errors": err.Error(),
		})
		return
	}

	// 检查角色名是否已存在
	var existing Role
	if h.db.Where("name = ?", req.Name).First(&existing).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "角色名已存在",
		})
		return
	}

	// 创建角色
	role := Role{
		ID:     uuid.New(),
		Name:   req.Name,
		Level:  req.Level,
		Path:   "/" + req.Name,
	}

	if req.Description != nil {
		role.Description = req.Description
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建角色
	if err := tx.Create(&role).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建角色失败",
		})
		return
	}

	// 分配权限
	if len(req.PermissionIDs) > 0 {
		rolePermissions := make([]RolePermission, len(req.PermissionIDs))
		for i, permID := range req.PermissionIDs {
			rolePermissions[i] = RolePermission{
				ID:           uuid.New(),
				RoleID:       role.ID,
				PermissionID: permID,
				CreatedAt:    time.Now(),
			}
		}

		if err := tx.Create(&rolePermissions).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "分配权限失败",
			})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建角色失败",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建成功",
		"data": gin.H{
			"id": role.ID,
		},
	})
}

// Update 更新角色
func (h *RoleHandler) Update(c *gin.Context) {
	roleID := c.Param("id")
	
	var req struct {
		Name        string   `json:"name" binding:"max=100"`
		Description *string  `json:"description"`
		Level       int      `json:"level" binding:"omitempty,min=1,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"errors": err.Error(),
		})
		return
	}

	// 查找角色
	var role Role
	if err := h.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "角色不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取角色失败",
		})
		return
	}

	// 检查是否为系统角色
	if role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "系统角色无法修改",
		})
		return
	}

	// 检查角色名是否重复
	if req.Name != "" && req.Name != role.Name {
		var existing Role
		if h.db.Where("name = ? AND id != ?", req.Name, roleID).First(&existing).Error == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "角色名已存在",
			})
			return
		}
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
		updates["path"] = "/" + req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Level > 0 {
		updates["level"] = req.Level
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新角色
	if err := tx.Model(&role).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新角色失败",
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新角色失败",
		})
		return
	}

	// 重新加载角色
	h.db.Preload("Permissions").Where("id = ?", roleID).First(&role)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data": gin.H{
			"id":          role.ID,
			"name":        role.Name,
			"description": role.Description,
			"level":      role.Level,
			"isSystem":   role.IsSystem,
			"permissions": role.Permissions,
			"updatedAt":  role.UpdatedAt,
		},
	})
}

// Delete 删除角色
func (h *RoleHandler) Delete(c *gin.Context) {
	roleID := c.Param("id")
	
	// 查找角色
	var role Role
	if err := h.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "角色不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取角色失败",
		})
		return
	}

	// 检查是否为系统角色
	if role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "系统角色无法删除",
		})
		return
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查是否有用户使用此角色
	var userCount int64
	if err := tx.Table("user_roles").Where("role_id = ?", roleID).Count(&userCount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "检查角色使用情况失败",
		})
		return
	}

	if userCount > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "该角色已被用户使用，无法删除",
		})
		return
	}

	// 删除角色权限关联
	if err := tx.Where("role_id = ?", roleID).Delete(&RolePermission{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除角色权限失败",
		})
		return
	}

	// 删除角色
	if err := tx.Delete(&role).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除角色失败",
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除角色失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// GetPermissions 获取角色权限
func (h *RoleHandler) GetPermissions(c *gin.Context) {
	roleID := c.Param("id")
	
	// 查找角色
	var role Role
	if err := h.db.Preload("Permissions").Where("id = ?", roleID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "角色不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取角色权限失败",
		})
		return
	}

	// 转换权限数据
	var permList []gin.H
	for _, p := range role.Permissions {
		permList = append(permList, gin.H{
			"id":          p.ID,
			"name":        p.Name,
			"resource":    p.Resource,
			"action":      p.Action,
			"description": p.Description,
			"createdAt":   p.CreatedAt,
			"updatedAt":   p.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    permList,
	})
}

// UpdatePermissions 更新角色权限
func (h *RoleHandler) UpdatePermissions(c *gin.Context) {
	roleID := c.Param("id")
	
	var req struct {
		PermissionIDs []string `json:"permissionIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"errors": err.Error(),
		})
		return
	}

	// 查找角色
	var role Role
	if err := h.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "角色不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取角色失败",
		})
		return
	}

	// 检查是否为系统角色
	if role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "系统角色无法修改权限",
		})
		return
	}

	// 检查权限是否存在
	var permCount int64
	if err := h.db.Table("permissions").Where("id IN ?", req.PermissionIDs).Count(&permCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "检查权限失败",
		})
		return
	}

	if permCount != int64(len(req.PermissionIDs)) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "部分权限不存在",
		})
		return
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除现有权限
	if err := tx.Where("role_id = ?", roleID).Delete(&RolePermission{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除现有权限失败",
		})
		return
	}

	// 分配新权限
	if len(req.PermissionIDs) > 0 {
		rolePermissions := make([]RolePermission, len(req.PermissionIDs))
		for i, permID := range req.PermissionIDs {
			rolePermissions[i] = RolePermission{
				ID:           uuid.New(),
				RoleID:       role.ID,
				PermissionID: permID,
				CreatedAt:    time.Now(),
			}
		}

		if err := tx.Create(&rolePermissions).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "分配权限失败",
			})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新权限失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "权限更新成功",
	})
}

// GetUsers 获取角色下的用户
func (h *RoleHandler) GetUsers(c *gin.Context) {
	roleID := c.Param("id")
	
	// 检查角色是否存在
	var role Role
	if err := h.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "角色不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取角色失败",
		})
		return
	}

	// 获取使用该角色的用户
	var users []User
	if err := h.db.Table("users").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", roleID).
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户列表失败",
		})
		return
	}

	// 转换用户数据
	var userList []gin.H
	for _, u := range users {
		userList = append(userList, gin.H{
			"id":        u.ID,
			"username":  u.Username,
			"email":     u.Email,
			"name":      u.Name,
			"department": u.Department,
			"status":    u.Status,
			"createdAt": u.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    userList,
	})
}

// PermissionHandler 权限处理器
type PermissionHandler struct {
	db *gorm.DB
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(db *gorm.DB) *PermissionHandler {
	return &PermissionHandler{db: db}
}

// GetMatrix 获取权限矩阵
func (h *PermissionHandler) GetMatrix(c *gin.Context) {
	// 获取所有角色
	var roles []Role
	if err := h.db.Preload("Permissions").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取角色列表失败",
		})
		return
	}

	// 获取所有模块
	var modules []string
	if err := h.db.Table("permissions").Distinct("resource").Pluck("resource", &modules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取模块列表失败",
		})
		return
	}

	// 获取所有权限
	var permissions []Permission
	if err := h.db.Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取权限列表失败",
		})
		return
	}

	// 构建矩阵数据
	var matrix []gin.H
	for _, role := range roles {
		rolePerms := make(map[string][]string)
		for _, perm := range role.Permissions {
			if _, ok := rolePerms[perm.Resource]; !ok {
				rolePerms[perm.Resource] = []string{}
			}
			rolePerms[perm.Resource] = append(rolePerms[perm.Resource], perm.Action)
		}

		matrix = append(matrix, gin.H{
			"roleId":      role.ID,
			"roleName":    role.Name,
			"permissions": rolePerms,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"roles":   roles,
			"modules": modules,
			"matrix":  matrix,
		},
	})
}

// UpdateMatrix 批量设置权限
func (h *PermissionHandler) UpdateMatrix(c *gin.Context) {
	var req struct {
		RolePermissions []struct {
			RoleID       string   `json:"roleId" binding:"required"`
			PermissionIDs []string `json:"permissionIds" binding:"required"`
		} `json:"rolePermissions" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"errors": err.Error(),
		})
		return
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, rp := range req.RolePermissions {
		// 检查角色是否存在
		var role Role
		if err := tx.Where("id = ?", rp.RoleID).First(&role).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "角色不存在: " + rp.RoleID,
			})
			return
		}

		// 检查是否为系统角色
		if role.IsSystem {
			tx.Rollback()
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "系统角色无法修改: " + role.Name,
			})
			return
		}

		// 删除现有权限
		if err := tx.Where("role_id = ?", rp.RoleID).Delete(&RolePermission{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "删除现有权限失败",
			})
			return
		}

		// 分配新权限
		if len(rp.PermissionIDs) > 0 {
			rolePermissions := make([]RolePermission, len(rp.PermissionIDs))
			for i, permID := range rp.PermissionIDs {
				rolePermissions[i] = RolePermission{
					ID:           uuid.New(),
					RoleID:       uuid.MustParse(rp.RoleID),
					PermissionID: permID,
					CreatedAt:    time.Now(),
				}
			}

			if err := tx.Create(&rolePermissions).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "分配权限失败",
				})
				return
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "批量设置权限失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "权限设置成功",
	})
}

// CheckConflict 检查权限冲突
func (h *PermissionHandler) CheckConflict(c *gin.Context) {
	var req struct {
		RolePermissions []struct {
			RoleID       string   `json:"roleId" binding:"required"`
			PermissionIDs []string `json:"permissionIds" binding:"required"`
		} `json:"rolePermissions" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"errors": err.Error(),
		})
		return
	}

	var conflicts []gin.H

	for _, rp := range req.RolePermissions {
		// 检查是否有重复权限
		permCount := make(map[string]int)
		for _, permID := range rp.PermissionIDs {
			permCount[permID]++
			if permCount[permID] > 1 {
				conflicts = append(conflicts, gin.H{
					"type":    "duplicate",
					"roleId":  rp.RoleID,
					"message": "权限重复分配: " + permID,
				})
			}
		}
	}

	if len(conflicts) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "检查完成",
			"data": gin.H{
				"hasConflict": true,
				"conflicts":   conflicts,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "检查完成",
		"data": gin.H{
			"hasConflict": false,
			"conflicts":   []gin.H{},
		},
	})
}

// PermissionAuditLog 权限变更日志
type PermissionAuditLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Action      string    `gorm:"type:varchar(50);not null" json:"action"`
	OperatorID  uuid.UUID `gorm:"type:uuid;not null" json:"operatorId"`
	OperatorName string   `gorm:"type:varchar(100)" json:"operatorName"`
	TargetType  string    `gorm:"type:varchar(50);not null" json:"targetType"`
	TargetID    string    `gorm:"type:varchar(100);not null" json:"targetId"`
	TargetName  string    `gorm:"type:varchar(200)" json:"targetName"`
	Changes     string    `gorm:"type:text" json:"changes"`
	IPAddress   string    `gorm:"type:inet" json:"ipAddress"`
	CreatedAt   time.Time `json:"createdAt"`
}

// TableName 指定表名
func (PermissionAuditLog) TableName() string {
	return "permission_audit_logs"
}

// GetAuditLogs 获取权限变更历史
func (h *PermissionHandler) GetAuditLogs(c *gin.Context) {
	var req struct {
		Page     int `form:"page" binding:"omitempty,min=1"`
		PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"`
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	var logs []PermissionAuditLog
	var total int64

	query := h.db.Model(&PermissionAuditLog{})

	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取日志数量失败",
		})
		return
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取日志列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  logs,
			"total": total,
			"page":  req.Page,
		},
	})
}

// GenerateReport 生成权限报告
func (h *PermissionHandler) GenerateReport(c *gin.Context) {
	// 获取统计信息
	var stats struct {
		TotalRoles       int `json:"totalRoles"`
		TotalPermissions int `json:"totalPermissions"`
		TotalUsers      int `json:"totalUsers"`
	}

	if err := h.db.Model(&Role{}).Count(&stats.TotalRoles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "统计角色数量失败",
		})
		return
	}

	if err := h.db.Model(&Permission{}).Count(&stats.TotalPermissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "统计权限数量失败",
		})
		return
	}

	if err := h.db.Model(&User{}).Count(&stats.TotalUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "统计用户数量失败",
		})
		return
	}

	// 获取角色权限分布
	var roleStats []gin.H
	var roles []Role
	if err := h.db.Preload("Permissions").Find(&roles).Error; err == nil {
		for _, role := range roles {
			roleStats = append(roleStats, gin.H{
				"roleId":        role.ID,
				"roleName":      role.Name,
				"permissionCount": len(role.Permissions),
				"isSystem":     role.IsSystem,
			})
		}
	}

	// 获取模块分布
	var moduleStats []gin.H
	var modules []string
	h.db.Table("permissions").Distinct("resource").Pluck("resource", &modules)
	for _, module := range modules {
		var permCount int64
		h.db.Table("permissions").Where("resource = ?", module).Count(&permCount)
		moduleStats = append(moduleStats, gin.H{
			"module":      module,
			"permCount":   permCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "报告生成成功",
		"data": gin.H{
			"generatedAt": time.Now(),
			"summary":     stats,
			"roles":       roleStats,
			"modules":     moduleStats,
		},
	})
}
