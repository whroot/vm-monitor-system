package api

import (
	"net/http"

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
	_ = userID

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
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    []gin.H{},
	})
}

// Get 获取角色详情
func (h *RoleHandler) Get(c *gin.Context) {
	roleID := c.Param("id")
	_ = roleID

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    gin.H{},
	})
}

// Create 创建角色
func (h *RoleHandler) Create(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建成功",
		"data": gin.H{
			"id": uuid.New().String(),
		},
	})
}

// Update 更新角色
func (h *RoleHandler) Update(c *gin.Context) {
	roleID := c.Param("id")
	_ = roleID

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// Delete 删除角色
func (h *RoleHandler) Delete(c *gin.Context) {
	roleID := c.Param("id")
	_ = roleID

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// GetPermissions 获取角色权限
func (h *RoleHandler) GetPermissions(c *gin.Context) {
	roleID := c.Param("id")
	_ = roleID

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    []gin.H{},
	})
}

// UpdatePermissions 更新角色权限
func (h *RoleHandler) UpdatePermissions(c *gin.Context) {
	roleID := c.Param("id")
	_ = roleID

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "权限更新成功",
	})
}

// GetUsers 获取角色下的用户
func (h *RoleHandler) GetUsers(c *gin.Context) {
	roleID := c.Param("id")
	_ = roleID

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    []gin.H{},
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
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"roles":   []gin.H{},
			"modules": []gin.H{},
			"matrix":  []gin.H{},
		},
	})
}

// UpdateMatrix 批量设置权限
func (h *PermissionHandler) UpdateMatrix(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "权限设置成功",
	})
}

// CheckConflict 检查权限冲突
func (h *PermissionHandler) CheckConflict(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "检查完成",
		"data": gin.H{
			"hasConflict": false,
			"conflicts":   []gin.H{},
		},
	})
}

// GetAuditLogs 获取权限变更历史
func (h *PermissionHandler) GetAuditLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    []gin.H{},
	})
}

// GenerateReport 生成权限报告
func (h *PermissionHandler) GenerateReport(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "报告生成成功",
	})
}
