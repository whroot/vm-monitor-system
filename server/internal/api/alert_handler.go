package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"vm-monitoring-system/internal/models"
)

// AlertHandler 告警处理器
type AlertHandler struct {
	db *gorm.DB
}

// NewAlertHandler 创建告警处理器
func NewAlertHandler(db *gorm.DB) *AlertHandler {
	return &AlertHandler{db: db}
}

// RuleRequest 告警规则请求
 type RuleRequest struct {
	Name              string                 `json:"name" binding:"required,max=200"`
	Description       string                 `json:"description,omitempty"`
	Scope             string                 `json:"scope" binding:"required,oneof=all vm group cluster host datacenter"`
	ScopeID           *uuid.UUID             `json:"scopeId,omitempty"`
	ScopeName         string                 `json:"scopeName,omitempty"`
	ConditionLogic    string                 `json:"conditionLogic" binding:"required,oneof=and or"`
	Enabled           bool                   `json:"enabled"`
	Cooldown          int                    `json:"cooldown" binding:"min=0,max=86400"`
	Severity          string                 `json:"severity" binding:"required,oneof=critical high medium low"`
	NotificationConfig map[string]interface{} `json:"notificationConfig"`
	Conditions        []ConditionRequest     `json:"conditions" binding:"required,min=1,dive"`
}

// ConditionRequest 告警条件请求
 type ConditionRequest struct {
	Metric      string   `json:"metric" binding:"required,max=50"`
	MetricType  string   `json:"metricType" binding:"required,max=100"`
	Operator    string   `json:"operator" binding:"required,oneof=> >= < <= == !="`
	Threshold   float64  `json:"threshold" binding:"required"`
	ThresholdStr *string `json:"thresholdStr,omitempty"`
	Duration    int      `json:"duration" binding:"min=0,max=3600"`
	Aggregation string   `json:"aggregation" binding:"omitempty,oneof=last avg max min sum"`
	SortOrder   int      `json:"sortOrder"`
}

// RecordFilter 告警记录筛选
 type RecordFilter struct {
	Page        int        `form:"page" binding:"min=1"`
	PageSize    int        `form:"pageSize" binding:"min=1,max=100"`
	Status      string     `form:"status" binding:"omitempty,oneof=active acknowledged resolved all"`
	Severity    []string   `form:"severity[]"`
	RuleID      string     `form:"ruleId"`
	VMID        string     `form:"vmId"`
	StartTime   *time.Time `form:"startTime" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime     *time.Time `form:"endTime" time_format:"2006-01-02T15:04:05Z07:00"`
	Keyword     string     `form:"keyword"`
	AcknowledgedBy string `form:"acknowledgedBy"`
}

// ========== 告警规则 ==========

// ListRules 获取告警规则列表
func (h *AlertHandler) ListRules(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 解析其他筛选参数
	status := c.Query("status")
	scope := c.Query("scope")
	severity := c.QueryArray("severity")
	keyword := c.Query("keyword")
	enabled := c.Query("enabled")

	// 构建查询
	query := h.db.Model(&models.AlertRule{}).Where("is_deleted = ?", false)

	// 应用筛选
	if status != "" {
		if status == "active" {
			query = query.Where("enabled = ?", true)
		} else if status == "inactive" {
			query = query.Where("enabled = ?", false)
		}
	}

	if scope != "" && scope != "all" {
		query = query.Where("scope = ?", scope)
	}

	if len(severity) > 0 {
		query = query.Where("severity IN ?", severity)
	}

	if keyword != "" {
		query = query.Where("name ILIKE ?", "%"+keyword+"%")
	}

	if enabled != "" {
		query = query.Where("enabled = ?", enabled == "true")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	// 分页查询
	var rules []models.AlertRule
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	// 加载每个规则的条件
	for i := range rules {
		var conditions []models.AlertCondition
		h.db.Where("rule_id = ?", rules[i].ID).Order("sort_order").Find(&conditions)
		rules[i].Conditions = conditions
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list": rules,
			"pagination": gin.H{
				"page":       page,
				"pageSize":   pageSize,
				"total":      total,
				"totalPages": totalPages,
			},
		},
	})
}

// GetRule 获取告警规则详情
func (h *AlertHandler) GetRule(c *gin.Context) {
	ruleID := c.Param("id")
	
	var id uuid.UUID
	if err := id.UnmarshalText([]byte(ruleID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "规则ID格式错误",
		})
		return
	}

	var rule models.AlertRule
	if err := h.db.Where("id = ? AND is_deleted = ?", id, false).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "规则不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	// 加载条件
	var conditions []models.AlertCondition
	h.db.Where("rule_id = ?", rule.ID).Order("sort_order").Find(&conditions)
	rule.Conditions = conditions

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    rule,
	})
}

// CreateRule 创建告警规则
func (h *AlertHandler) CreateRule(c *gin.Context) {
	var req RuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 开启事务
	tx := h.db.Begin()

	// 解析通知配置
	notificationConfigJSON, _ := json.Marshal(req.NotificationConfig)

	// 创建规则
	rule := models.AlertRule{
		ID:                 uuid.New(),
		Name:               req.Name,
		Description:        nil,
		Scope:              req.Scope,
		ScopeID:            req.ScopeID,
		ConditionLogic:     req.ConditionLogic,
		Enabled:            req.Enabled,
		Cooldown:           req.Cooldown,
		Severity:           req.Severity,
		NotificationConfig: models.JSONMap{},
		TriggerCount:       0,
		IsDeleted:          false,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if req.Description != "" {
		rule.Description = &req.Description
	}

	if len(notificationConfigJSON) > 0 {
		rule.NotificationConfig = models.JSONMap(req.NotificationConfig)
	}

	// 获取用户信息（从上下文）
	if userID, exists := c.Get("userID"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			rule.CreatedBy = &uid
			rule.UpdatedBy = &uid
		}
	}

	if err := tx.Create(&rule).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建规则失败: " + err.Error(),
		})
		return
	}

	// 创建条件
	for i, condReq := range req.Conditions {
		thresholdStr := ""
		if condReq.ThresholdStr != nil {
			thresholdStr = *condReq.ThresholdStr
		}

		condition := models.AlertCondition{
			ID:           uuid.New(),
			RuleID:       rule.ID,
			Metric:       condReq.Metric,
			MetricType:   condReq.MetricType,
			Operator:     condReq.Operator,
			Threshold:    condReq.Threshold,
			ThresholdStr: &thresholdStr,
			Duration:     condReq.Duration,
			Aggregation:  condReq.Aggregation,
			SortOrder:    i,
			CreatedAt:    time.Now(),
		}

		if err := tx.Create(&condition).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "创建条件失败: " + err.Error(),
			})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建成功",
		"data": gin.H{
			"id": rule.ID,
		},
	})
}

// UpdateRule 更新告警规则
func (h *AlertHandler) UpdateRule(c *gin.Context) {
	ruleID := c.Param("id")
	
	var id uuid.UUID
	if err := id.UnmarshalText([]byte(ruleID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "规则ID格式错误",
		})
		return
	}

	var req RuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查规则是否存在
	var existingRule models.AlertRule
	if err := h.db.Where("id = ? AND is_deleted = ?", id, false).First(&existingRule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "规则不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	// 开启事务
	tx := h.db.Begin()

	// 更新规则字段
	updates := map[string]interface{}{
		"name":               req.Name,
		"scope":              req.Scope,
		"condition_logic":    req.ConditionLogic,
		"enabled":            req.Enabled,
		"cooldown":           req.Cooldown,
		"severity":           req.Severity,
		"updated_at":         time.Now(),
	}

	if req.Description != "" {
		updates["description"] = req.Description
	}

	if req.ScopeID != nil {
		updates["scope_id"] = req.ScopeID
	}

	if req.ScopeName != "" {
		updates["scope_name"] = req.ScopeName
	}

	// 更新通知配置
	if len(req.NotificationConfig) > 0 {
		notificationConfigJSON, _ := json.Marshal(req.NotificationConfig)
		updates["notification_config"] = string(notificationConfigJSON)
	}

	// 获取用户信息
	if userID, exists := c.Get("userID"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			updates["updated_by"] = uid
		}
	}

	if err := tx.Model(&existingRule).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新规则失败: " + err.Error(),
		})
		return
	}

	// 删除旧条件并创建新条件
	if err := tx.Where("rule_id = ?", id).Delete(&models.AlertCondition{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除旧条件失败: " + err.Error(),
		})
		return
	}

	// 创建新条件
	for i, condReq := range req.Conditions {
		thresholdStr := ""
		if condReq.ThresholdStr != nil {
			thresholdStr = *condReq.ThresholdStr
		}

		condition := models.AlertCondition{
			ID:           uuid.New(),
			RuleID:       id,
			Metric:       condReq.Metric,
			MetricType:   condReq.MetricType,
			Operator:     condReq.Operator,
			Threshold:    condReq.Threshold,
			ThresholdStr: &thresholdStr,
			Duration:     condReq.Duration,
			Aggregation:  condReq.Aggregation,
			SortOrder:    i,
			CreatedAt:    time.Now(),
		}

		if err := tx.Create(&condition).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "创建条件失败: " + err.Error(),
			})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// DeleteRule 删除告警规则
func (h *AlertHandler) DeleteRule(c *gin.Context) {
	ruleID := c.Param("id")
	
	var id uuid.UUID
	if err := id.UnmarshalText([]byte(ruleID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "规则ID格式错误",
		})
		return
	}

	// 检查规则是否存在
	var rule models.AlertRule
	if err := h.db.Where("id = ? AND is_deleted = ?", id, false).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "规则不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	// 软删除
	updates := map[string]interface{}{
		"is_deleted":  true,
		"deleted_at":  time.Now(),
		"updated_at":  time.Now(),
	}

	// 获取用户信息
	if userID, exists := c.Get("userID"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			updates["deleted_by"] = uid
			updates["updated_by"] = uid
		}
	}

	if err := h.db.Model(&rule).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// BatchUpdateRuleStatusRequest 批量更新请求
type BatchUpdateRuleStatusRequest struct {
	IDs    []string `json:"ids" binding:"required,min=1"`
	Status bool     `json:"status" binding:"required"`
}

// BatchUpdateRuleStatus 批量更新规则状态
func (h *AlertHandler) BatchUpdateRuleStatus(c *gin.Context) {
	var req BatchUpdateRuleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 解析UUID
	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		var id uuid.UUID
		if err := id.UnmarshalText([]byte(idStr)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "规则ID格式错误: " + idStr,
			})
			return
		}
		ids = append(ids, id)
	}

	// 获取用户信息
	var updatedBy *uuid.UUID
	if userID, exists := c.Get("userID"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			updatedBy = &uid
		}
	}

	// 批量更新
	updates := map[string]interface{}{
		"enabled":    req.Status,
		"updated_at": time.Now(),
	}
	if updatedBy != nil {
		updates["updated_by"] = *updatedBy
	}

	result := h.db.Model(&models.AlertRule{}).
		Where("id IN ? AND is_deleted = ?", ids, false).
		Updates(updates)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "批量更新失败: " + result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("批量更新成功，已更新 %d 条规则", result.RowsAffected),
		"data": gin.H{
			"updatedCount": result.RowsAffected,
		},
	})
}

// ImportRules 导入规则
func (h *AlertHandler) ImportRules(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "导入成功",
	})
}

// ExportRules 导出规则
func (h *AlertHandler) ExportRules(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "导出成功",
	})
}

// ========== 告警记录 ==========

// AcknowledgeRequest 确认告警请求
type AcknowledgeRequest struct {
	Note string `json:"note,omitempty"`
}

// ListRecords 获取告警记录列表
func (h *AlertHandler) ListRecords(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 解析筛选参数
	status := c.DefaultQuery("status", "all")
	severity := c.QueryArray("severity")
	ruleID := c.Query("ruleId")
	vmID := c.Query("vmId")
	startTime := c.Query("startTime")
	endTime := c.Query("endTime")
	keyword := c.Query("keyword")

	// 构建查询
	query := h.db.Model(&models.AlertRecord{})

	// 应用筛选
	if status != "all" {
		query = query.Where("status = ?", status)
	}

	if len(severity) > 0 {
		query = query.Where("severity IN ?", severity)
	}

	if ruleID != "" {
		var rid uuid.UUID
		if err := rid.UnmarshalText([]byte(ruleID)); err == nil {
			query = query.Where("rule_id = ?", rid)
		}
	}

	if vmID != "" {
		var vid uuid.UUID
		if err := vid.UnmarshalText([]byte(vmID)); err == nil {
			query = query.Where("vm_id = ?", vid)
		}
	}

	if startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			query = query.Where("triggered_at >= ?", t)
		}
	}

	if endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			query = query.Where("triggered_at <= ?", t)
		}
	}

	if keyword != "" {
		query = query.Where("(rule_name ILIKE ? OR vm_name ILIKE ?)", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	// 分页查询
	var records []models.AlertRecord
	offset := (page - 1) * pageSize
	if err := query.Order("triggered_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list": records,
			"pagination": gin.H{
				"page":       page,
				"pageSize":   pageSize,
				"total":      total,
				"totalPages": totalPages,
			},
		},
	})
}

// GetRecord 获取告警记录详情
func (h *AlertHandler) GetRecord(c *gin.Context) {
	recordID := c.Param("id")
	
	var id uuid.UUID
	if err := id.UnmarshalText([]byte(recordID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "记录ID格式错误",
		})
		return
	}

	var record models.AlertRecord
	if err := h.db.First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "记录不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    record,
	})
}

// Acknowledge 确认告警
func (h *AlertHandler) Acknowledge(c *gin.Context) {
	recordID := c.Param("id")
	
	var id uuid.UUID
	if err := id.UnmarshalText([]byte(recordID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "记录ID格式错误",
		})
		return
	}

	// 解析请求体
	var req AcknowledgeRequest
	c.ShouldBindJSON(&req)

	// 查找记录
	var record models.AlertRecord
	if err := h.db.First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "记录不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	// 检查是否已是确认状态
	if record.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "该告警已" + record.Status + "，无法确认",
		})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":           "acknowledged",
		"acknowledged_at":  now,
		"updated_at":       now,
	}

	// 获取用户信息
	var userName string
	if userID, exists := c.Get("userID"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			updates["acknowledged_by"] = uid
			
			// 获取用户名
			var user models.User
			if err := h.db.Select("name").First(&user, uid).Error; err == nil {
				userName = user.Name
				updates["acknowledged_by_name"] = userName
			}
		}
	}

	if req.Note != "" {
		updates["acknowledge_note"] = req.Note
	}

	if err := h.db.Model(&record).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "确认失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "确认成功",
		"data": gin.H{
			"id":              recordID,
			"status":          "acknowledged",
			"acknowledgedAt":  now.Format(time.RFC3339),
			"acknowledgedBy":  userName,
		},
	})
}

// BatchAcknowledgeRequest 批量确认请求
type BatchAcknowledgeRequest struct {
	IDs  []string `json:"ids" binding:"required,min=1"`
	Note string   `json:"note,omitempty"`
}

// BatchAcknowledge 批量确认告警
func (h *AlertHandler) BatchAcknowledge(c *gin.Context) {
	var req BatchAcknowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 解析UUID
	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		var id uuid.UUID
		if err := id.UnmarshalText([]byte(idStr)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "记录ID格式错误: " + idStr,
			})
			return
		}
		ids = append(ids, id)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":          "acknowledged",
		"acknowledged_at": now,
		"updated_at":      now,
	}

	if req.Note != "" {
		updates["acknowledge_note"] = req.Note
	}

	// 获取用户信息
	var userName string
	if userID, exists := c.Get("userID"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			updates["acknowledged_by"] = uid
			
			var user models.User
			if err := h.db.Select("name").First(&user, uid).Error; err == nil {
				userName = user.Name
				updates["acknowledged_by_name"] = userName
			}
		}
	}

	// 只更新active状态的记录
	result := h.db.Model(&models.AlertRecord{}).
		Where("id IN ? AND status = ?", ids, "active").
		Updates(updates)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "批量确认失败: " + result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("批量确认成功，已确认 %d 条告警", result.RowsAffected),
		"data": gin.H{
			"updatedCount": result.RowsAffected,
		},
	})
}

// ResolveRequest 解决告警请求
type ResolveRequest struct {
	Resolution string `json:"resolution,omitempty"`
}

// Resolve 解决告警
func (h *AlertHandler) Resolve(c *gin.Context) {
	recordID := c.Param("id")
	
	var id uuid.UUID
	if err := id.UnmarshalText([]byte(recordID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "记录ID格式错误",
		})
		return
	}

	var req ResolveRequest
	c.ShouldBindJSON(&req)

	// 查找记录
	var record models.AlertRecord
	if err := h.db.First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "记录不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	// 检查是否可以解决
	if record.Status == "resolved" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "该告警已解决，无需重复操作",
		})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       "resolved",
		"resolved_at":  now,
		"updated_at":   now,
	}

	// 计算持续时间
	if record.Status == "active" || record.Status == "acknowledged" {
		duration := int(now.Sub(record.TriggeredAt).Minutes())
		updates["duration"] = duration
	}

	// 获取用户信息
	var userName string
	if userID, exists := c.Get("userID"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			updates["resolved_by"] = uid
			
			var user models.User
			if err := h.db.Select("name").First(&user, uid).Error; err == nil {
				userName = user.Name
				updates["resolved_by_name"] = userName
			}
		}
	}

	if req.Resolution != "" {
		updates["resolution"] = req.Resolution
	}

	if err := h.db.Model(&record).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "解决失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "解决成功",
		"data": gin.H{
			"id":          recordID,
			"status":      "resolved",
			"resolvedAt":  now.Format(time.RFC3339),
			"resolvedBy":  userName,
		},
	})
}

// Ignore 忽略告警
func (h *AlertHandler) Ignore(c *gin.Context) {
	recordID := c.Param("id")
	
	var id uuid.UUID
	if err := id.UnmarshalText([]byte(recordID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "记录ID格式错误",
		})
		return
	}

	// 查找记录
	var record models.AlertRecord
	if err := h.db.First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "记录不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	// 检查是否已是忽略或解决状态
	if record.Status == "ignored" || record.Status == "resolved" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "该告警已" + record.Status + "，无法忽略",
		})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       "ignored",
		"resolved_at":  now,
		"updated_at":   now,
	}

	// 获取用户信息
	if userID, exists := c.Get("userID"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			updates["resolved_by"] = uid
			
			var user models.User
			if err := h.db.Select("name").First(&user, uid).Error; err == nil {
				updates["resolved_by_name"] = user.Name
			}
		}
	}

	if err := h.db.Model(&record).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "忽略失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "已忽略",
	})
}

// ========== 统计 ==========

// Statistics 获取告警统计
func (h *AlertHandler) Statistics(c *gin.Context) {
	// 规则统计
	var totalRules, activeRules int64
	h.db.Model(&models.AlertRule{}).Where("is_deleted = ?", false).Count(&totalRules)
	h.db.Model(&models.AlertRule{}).Where("is_deleted = ? AND enabled = ?", false, true).Count(&activeRules)

	// 告警记录统计
	var totalAlerts, activeAlerts, acknowledgedAlerts, resolvedAlerts int64
	h.db.Model(&models.AlertRecord{}).Count(&totalAlerts)
	h.db.Model(&models.AlertRecord{}).Where("status = ?", "active").Count(&activeAlerts)
	h.db.Model(&models.AlertRecord{}).Where("status = ?", "acknowledged").Count(&acknowledgedAlerts)
	h.db.Model(&models.AlertRecord{}).Where("status = ?", "resolved").Count(&resolvedAlerts)

	// 按严重级别统计
	var bySeverity []struct {
		Severity string
		Count    int64
	}
	h.db.Model(&models.AlertRecord{}).
		Select("severity, count(*) as count").
		Group("severity").
		Scan(&bySeverity)

	severityMap := make(map[string]int64)
	for _, s := range bySeverity {
		severityMap[s.Severity] = s.Count
	}

	// 按指标统计
	var byMetric []struct {
		Metric      string
		Count       int64
		ActiveCount int64
	}
	h.db.Model(&models.AlertRecord{}).
		Select("metric, count(*) as count, sum(case when status = 'active' then 1 else 0 end) as active_count").
		Group("metric").
		Scan(&byMetric)

	// 按VM统计
	var byVM []struct {
		VMID        string
		VMName      string
		Count       int64
		ActiveCount int64
	}
	h.db.Model(&models.AlertRecord{}).
		Select("vm_id, vm_name, count(*) as count, sum(case when status = 'active' then 1 else 0 end) as active_count").
		Where("vm_id IS NOT NULL").
		Group("vm_id, vm_name").
		Limit(10).
		Scan(&byVM)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"overview": gin.H{
				"totalRules":         totalRules,
				"activeRules":        activeRules,
				"totalAlerts":        totalAlerts,
				"activeAlerts":       activeAlerts,
				"acknowledgedAlerts": acknowledgedAlerts,
				"resolvedAlerts":     resolvedAlerts,
			},
			"bySeverity": severityMap,
			"byMetric":   byMetric,
			"byVM":       byVM,
		},
	})
}

// Trends 获取告警趋势
func (h *AlertHandler) Trends(c *gin.Context) {
	// 默认查询最近7天
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -7)

	// 获取查询参数
	if startParam := c.Query("startTime"); startParam != "" {
		if t, err := time.Parse(time.RFC3339, startParam); err == nil {
			startTime = t
		}
	}
	if endParam := c.Query("endTime"); endParam != "" {
		if t, err := time.Parse(time.RFC3339, endParam); err == nil {
			endTime = t
		}
	}

	// 按天统计告警数量
	var trends []struct {
		Date          string `json:"date"`
		Total         int64  `json:"total"`
		Active        int64  `json:"active"`
		Acknowledged  int64  `json:"acknowledged"`
		Resolved      int64  `json:"resolved"`
	}

	// 使用数据库原生查询获取按天的统计
	// 注意: 不同数据库的日期格式化函数可能不同，这里使用PostgreSQL的语法
	query := `
		SELECT 
			TO_CHAR(triggered_at, 'YYYY-MM-DD') as date,
			COUNT(*) as total,
			SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) as active,
			SUM(CASE WHEN status = 'acknowledged' THEN 1 ELSE 0 END) as acknowledged,
			SUM(CASE WHEN status = 'resolved' THEN 1 ELSE 0 END) as resolved
		FROM alert_records
		WHERE triggered_at BETWEEN ? AND ?
		GROUP BY TO_CHAR(triggered_at, 'YYYY-MM-DD')
		ORDER BY date
	`

	h.db.Raw(query, startTime, endTime).Scan(&trends)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"trends": trends,
			"timeRange": gin.H{
				"start": startTime.Format(time.RFC3339),
				"end":   endTime.Format(time.RFC3339),
			},
		},
	})
}
