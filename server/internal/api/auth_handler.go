package api

import (
	"crypto/rsa"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"vm-monitoring-system/internal/config"
	"vm-monitoring-system/internal/models"
	"vm-monitoring-system/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	db         *gorm.DB
	config     *config.Config
	privateKey interface{} // RSA私钥，用于解密密码
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(db *gorm.DB, cfg *config.Config, privateKey interface{}) *AuthHandler {
	return &AuthHandler{
		db:         db,
		config:     cfg,
		privateKey: privateKey,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	IsEncrypted    bool   `json:"isEncrypted"`    // 密码是否已加密
	RememberMe     bool   `json:"rememberMe"`
	Language       string `json:"language"`
}

// GetPublicKeyResponse 公钥响应
type GetPublicKeyResponse struct {
	PublicKey string `json:"publicKey"`  // Base64编码的公钥
	ExpiresAt int64  `json:"expiresAt"`  // 公钥过期时间戳
}

// LoginResponse 登录响应
type LoginResponse struct {
	User          models.User `json:"user"`
	AccessToken   string      `json:"accessToken"`
	RefreshToken  string      `json:"refreshToken"`
	ExpiresIn     int         `json:"expiresIn"`
	Permissions   []string    `json:"permissions"`
}

// GetPublicKey 获取RSA公钥
func (h *AuthHandler) GetPublicKey(c *gin.Context) {
	// TODO: 从配置或密钥管理服务获取公钥
	// 这里需要实现公钥获取逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": GetPublicKeyResponse{
			PublicKey: "", // 从配置读取
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 解密密码（如果是加密传输）
	password := req.Password
	if req.IsEncrypted && h.privateKey != nil {
		// 解码Base64
		encryptedBytes, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "密码格式错误",
			})
			return
		}

		// 使用私钥解密
		if rsaPrivateKey, ok := h.privateKey.(*rsa.PrivateKey); ok {
			decryptedBytes, err := utils.DecryptWithPrivateKey(rsaPrivateKey, encryptedBytes)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "密码解密失败",
				})
				return
			}
			password = string(decryptedBytes)
		}
	}

	// 查找用户
	var user models.User
	if err := h.db.Preload("Roles").Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户名或密码错误",
		})
		return
	}

	// 检查账户状态
	if user.Status == "locked" {
		if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "账户已锁定，请稍后重试",
			})
			return
		}
		// 锁定时间已过，解锁
		user.Status = "active"
		user.LoginFailCount = 0
		h.db.Save(&user)
	}

	if user.Status == "expired" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "账户已过期",
		})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// 增加失败次数
		user.LoginFailCount++
		if user.LoginFailCount >= 5 {
			user.Status = "locked"
			lockTime := time.Now().Add(30 * time.Minute)
			user.LockedUntil = &lockTime
		}
		h.db.Save(&user)

		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户名或密码错误",
		})
		return
	}

	// 登录成功，重置失败次数
	user.LoginFailCount = 0
	now := time.Now()
	user.LastLoginAt = &now
	user.LastLoginIP = utils.StringPtr(c.ClientIP())

	// 更新语言偏好
	if req.Language != "" {
		user.Preferences.Language = req.Language
	}

	h.db.Save(&user)

	// 生成Token
	accessToken, refreshToken, err := h.generateTokens(user, req.RememberMe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成Token失败",
		})
		return
	}

	// 记录会话
	session := models.UserSession{
		UserID:           user.ID,
		AccessTokenHash:  utils.HashString(accessToken),
		RefreshTokenHash: utils.StringPtr(utils.HashString(refreshToken)),
		ExpiresAt:        time.Now().Add(h.config.JWT.AccessExpiry),
		RefreshExpiresAt: utils.TimePtr(time.Now().Add(h.getRefreshExpiry(req.RememberMe))),
		IPAddress:        utils.StringPtr(c.ClientIP()),
		UserAgent:        utils.StringPtr(c.GetHeader("User-Agent")),
		IsActive:         true,
	}
	h.db.Create(&session)

	// 获取权限
	permissions := getUserPermissionStrings(user)

	// 隐藏敏感字段
	user.PasswordHash = ""
	user.MFASecret = nil

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": LoginResponse{
			User:         user,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int(h.config.JWT.AccessExpiry.Seconds()),
			Permissions:  permissions,
		},
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从Header获取Token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "登出成功",
		})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "登出成功",
		})
		return
	}

	tokenString := parts[1]
	tokenHash := utils.HashString(tokenString)

	// 使Token失效
	var session models.UserSession
	if err := h.db.Where("access_token_hash = ?", tokenHash).First(&session).Error; err == nil {
		session.IsActive = false
		revokedAt := time.Now()
		session.RevokedAt = &revokedAt
		if user := GetUser(c); user != nil {
			session.RevokedBy = &user.ID
		}
		h.db.Save(&session)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登出成功",
	})
}

// RefreshTokenRequest Token刷新请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// RefreshToken 刷新Token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 验证Refresh Token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.config.JWT.Secret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Refresh Token无效或已过期",
		})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Token解析失败",
		})
		return
	}

	// 验证Token类型
	if tokenType, ok := claims["type"].(string); !ok || tokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Token类型错误",
		})
		return
	}

	// 查找用户
	userID, _ := claims["sub"].(string)
	var user models.User
	if err := h.db.Preload("Roles").First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户不存在",
		})
		return
	}

	// 验证Refresh Token是否匹配
	tokenHash := utils.HashString(req.RefreshToken)
	var session models.UserSession
	if err := h.db.Where("refresh_token_hash = ? AND is_active = ?", tokenHash, true).First(&session).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Refresh Token无效",
		})
		return
	}

	// 使旧Token失效
	session.IsActive = false
	h.db.Save(&session)

	// 生成新Token
	accessToken, refreshToken, err := h.generateTokens(user, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成Token失败",
		})
		return
	}

	// 创建新会话
	newSession := models.UserSession{
		UserID:           user.ID,
		AccessTokenHash:  utils.HashString(accessToken),
		RefreshTokenHash: utils.StringPtr(utils.HashString(refreshToken)),
		ExpiresAt:        time.Now().Add(h.config.JWT.AccessExpiry),
		RefreshExpiresAt: utils.TimePtr(time.Now().Add(h.config.JWT.RefreshExpiry)),
		IPAddress:        utils.StringPtr(c.ClientIP()),
		UserAgent:        utils.StringPtr(c.GetHeader("User-Agent")),
		IsActive:         true,
	}
	h.db.Create(&newSession)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "刷新成功",
		"data": gin.H{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
			"expiresIn":    int(h.config.JWT.AccessExpiry.Seconds()),
		},
	})
}

// GetMe 获取当前用户信息
func (h *AuthHandler) GetMe(c *gin.Context) {
	user := GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
		})
		return
	}

	// 重新加载用户，确保数据最新
	var freshUser models.User
	h.db.Preload("Roles").First(&freshUser, "id = ?", user.ID)

	// 隐藏敏感字段
	freshUser.PasswordHash = ""
	freshUser.MFASecret = nil

	// 获取权限
	permissions := getUserPermissionStrings(freshUser)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"user":        freshUser,
			"permissions": permissions,
		},
	})
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword     string `json:"oldPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	user := GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
		})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 验证两次输入的密码是否一致
	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "两次输入的密码不一致",
		})
		return
	}

	// 验证旧密码
	var fullUser models.User
	h.db.First(&fullUser, "id = ?", user.ID)

	if err := bcrypt.CompareHashAndPassword([]byte(fullUser.PasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "旧密码错误",
		})
		return
	}

	// 验证新密码复杂度
	if !utils.ValidatePasswordComplexity(req.NewPassword) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "密码复杂度不符合要求",
		})
		return
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "密码加密失败",
		})
		return
	}

	// 更新密码
	fullUser.PasswordHash = string(hashedPassword)
	fullUser.MustChangePassword = false
	now := time.Now()
	fullUser.PasswordExpiredAt = utils.TimePtr(now.AddDate(0, 0, 90)) // 90天后过期

	h.db.Save(&fullUser)

	// 记录审计日志
	h.createAuditLog(user.ID, "update", "user", user.ID.String(), "修改密码", nil)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码修改成功",
		"data": gin.H{
			"passwordChangedAt": now,
		},
	})
}

// CheckPermissionRequest 权限检查请求
type CheckPermissionRequest struct {
	Permission string `json:"permission" binding:"required"`
	Resource   string `json:"resource"`
}

// CheckPermission 检查权限
func (h *AuthHandler) CheckPermission(c *gin.Context) {
	user := GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
		})
		return
	}

	var req CheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	permissions := GetPermissions(c)
	allowed := false

	for _, perm := range permissions {
		if perm == req.Permission || perm == "*" {
			allowed = true
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "检查完成",
		"data": gin.H{
			"allowed":    allowed,
			"permission": req.Permission,
			"resource":   req.Resource,
		},
	})
}

// 辅助方法

func (h *AuthHandler) generateTokens(user models.User, rememberMe bool) (string, string, error) {
	now := time.Now()

	// Access Token
	accessClaims := jwt.MapClaims{
		"sub":      user.ID.String(),
		"username": user.Username,
		"type":     "access",
		"iat":      now.Unix(),
		"exp":      now.Add(h.config.JWT.AccessExpiry).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(h.config.JWT.Secret))
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	refreshExpiry := h.config.JWT.RefreshExpiry
	if rememberMe {
		refreshExpiry = h.config.JWT.RememberExpiry
	}

	refreshClaims := jwt.MapClaims{
		"sub": user.ID.String(),
		"type": "refresh",
		"iat": now.Unix(),
		"exp": now.Add(refreshExpiry).Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(h.config.JWT.Secret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (h *AuthHandler) getRefreshExpiry(rememberMe bool) time.Duration {
	if rememberMe {
		return h.config.JWT.RememberExpiry
	}
	return h.config.JWT.RefreshExpiry
}

func (h *AuthHandler) createAuditLog(operatorID uuid.UUID, action, resourceType, resourceID, note string, changes interface{}) {
	auditLog := models.AuditLog{
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		OperatorID:   &operatorID,
		Note:         &note,
	}

	if changes != nil {
		if c, ok := changes.(map[string]interface{}); ok {
			auditLog.Changes = models.JSONMap(c)
		}
	}

	h.db.Create(&auditLog)
}

func getUserPermissionStrings(user models.User) []string {
	var permissions []string
	seen := make(map[string]bool)

	for _, role := range user.Roles {
		var rolePerms []models.RolePermission
		models.DB.Where("role_id = ?", role.ID).Find(&rolePerms)
		for _, rp := range rolePerms {
			if !seen[rp.PermissionID] {
				seen[rp.PermissionID] = true
				permissions = append(permissions, rp.PermissionID)
			}
		}
	}

	return permissions
}
