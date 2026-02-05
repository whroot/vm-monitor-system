package api

import (
	"net/http"
	"strconv"
	"time"

	"vm-monitoring-system/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VMHandler VM处理器
type VMHandler struct {
	db     *gorm.DB
	collector *services.VSphereCollector
}

// NewVMHandler 创建VM处理器
func NewVMHandler(db *gorm.DB) *VMHandler {
	return &VMHandler{db: db}
}

// SetCollector 设置vSphere采集器
func (h *VMHandler) SetCollector(collector *services.VSphereCollector) {
	h.collector = collector
}

// VMListResponse VM列表响应
type VMListResponse struct {
	List       []models.VM    `json:"list"`
	Pagination Pagination     `json:"pagination"`
	Summary    models.VMSummary `json:"summary"`
}

// List 获取VM列表
func (h *VMHandler) List(c *gin.Context) {
	var req models.VMListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ValidationError(c, err)
		return
	}

	// 默认值
	page, pageSize := PageParam(c)

	// 构建查询
	query := h.db.Model(&models.VM{}).Where("is_deleted = ?", false)

	// 筛选条件
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.OS != "" {
		query = query.Where("os_type = ?", req.OS)
	}
	if req.GroupID != "" {
		if groupID, err := uuid.Parse(req.GroupID); err == nil {
			query = query.Where("group_id = ?", groupID)
		}
	}
	if req.HostID != "" {
		query = query.Where("host_id = ?", req.HostID)
	}
	if req.ClusterID != "" {
		query = query.Where("cluster_id = ?", req.ClusterID)
	}
	if req.DatacenterID != "" {
		query = query.Where("datacenter_id = ?", req.DatacenterID)
	}
	if req.Keyword != "" {
		query = query.Where("name ILIKE ? OR ip::text ILIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 排序
	sortBy := "created_at"
	if req.SortBy != "" {
		allowedSortFields := map[string]bool{
			"created_at": true, "updated_at": true,
			"name": true, "status": true,
			"ip": true, "os": true,
		}
		if allowedSortFields[req.SortBy] {
			sortBy = req.SortBy
		}
	}
	sortOrder := "desc"
	if req.SortOrder == "asc" {
		sortOrder = "asc"
	}
	query = query.Order(sortBy + " " + sortOrder)

	// 查询总数
	var total int64
	query.Count(&total)

	// 分页查询
	var vms []models.VM
	offset := (req.Page - 1) * req.PageSize
	query.Preload("Group").Limit(req.PageSize).Offset(offset).Find(&vms)

	// 统计
	var summary models.VMSummary
	h.db.Model(&models.VM{}).Where("is_deleted = ?", false).Select(
		"COUNT(*) as total",
		"SUM(CASE WHEN status = 'online' THEN 1 ELSE 0 END) as online",
		"SUM(CASE WHEN status = 'offline' THEN 1 ELSE 0 END) as offline",
		"SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END) as error",
	).Scan(&summary)

	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "获取成功",
		Data: VMListResponse{
			List: vms,
			Pagination: BuildPagination(page, pageSize, int(total)),
			Summary: summary,
		},
	})
}

// Get 获取VM详情
func (h *VMHandler) Get(c *gin.Context) {
	id := c.Param("id")
	vmID, err := uuid.Parse(id)
	if err != nil {
		BadRequest(c, "ID格式错误")
		return
	}

	var vm models.VM
	if err := h.db.Preload("Group").Where("id = ? AND is_deleted = ?", vmID, false).First(&vm).Error; err != nil {
		NotFound(c, "VM不存在")
		return
	}

	Success(c, vm)
}

// CreateRequest 创建VM请求
type CreateRequest struct {
	Name        string   `json:"name" binding:"required"`
	IP          string   `json:"ip"`
	OS          string   `json:"os"`
	OSVersion   string   `json:"osVersion"`
	CPUCores    int      `json:"cpuCores"`
	MemoryGB    int      `json:"memoryGB"`
	DiskGB      int      `json:"diskGB"`
	GroupID     string   `json:"groupId"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
}

// Create 创建VM
func (h *VMHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, err)
		return
	}

	vm := models.VM{
		Name:        req.Name,
		Status:      "unknown",
		Description: &req.Description,
	}

	if req.IP != "" {
		vm.IP = &req.IP
	}
	if req.OS != "" {
		vm.OSType = &req.OS
	}
	if req.OSVersion != "" {
		vm.OSVersion = &req.OSVersion
	}
	if req.CPUCores > 0 {
		vm.CPUCores = &req.CPUCores
	}
	if req.MemoryGB > 0 {
		vm.MemoryGB = &req.MemoryGB
	}
	if req.DiskGB > 0 {
		vm.DiskGB = &req.DiskGB
	}
	if req.GroupID != "" {
		if groupID, err := uuid.Parse(req.GroupID); err == nil {
			vm.GroupID = &groupID
		}
	}
	if len(req.Tags) > 0 {
		tags := make(models.JSONMap, len(req.Tags))
		for i, tag := range req.Tags {
			tags[strconv.Itoa(i)] = tag
		}
		vm.Tags = tags
	}

	if user := GetUser(c); user != nil {
		vm.CreatedBy = &user.ID
		vm.UpdatedBy = &user.ID
	}

	if err := h.db.Create(&vm).Error; err != nil {
		InternalError(c, "创建失败", err)
		return
	}

	Created(c, vm)
}

// UpdateRequest 更新VM请求
type UpdateRequest struct {
	Name        string   `json:"name"`
	GroupID     string   `json:"groupId"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
}

// Update 更新VM
func (h *VMHandler) Update(c *gin.Context) {
	id := c.Param("id")
	vmID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "ID格式错误",
		})
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	var vm models.VM
	if err := h.db.Where("id = ? AND is_deleted = ?", vmID, false).First(&vm).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "VM不存在",
		})
		return
	}

	// 更新字段
	if req.Name != "" {
		vm.Name = req.Name
	}
	if req.GroupID != "" {
		if groupID, err := uuid.Parse(req.GroupID); err == nil {
			vm.GroupID = &groupID
		}
	}
	if len(req.Tags) > 0 {
		tags := make(models.JSONMap, len(req.Tags))
		for i, tag := range req.Tags {
			tags[strconv.Itoa(i)] = tag
		}
		vm.Tags = tags
	}
	if req.Description != "" {
		vm.Description = &req.Description
	}

	if user := GetUser(c); user != nil {
		vm.UpdatedBy = &user.ID
	}

	if err := h.db.Save(&vm).Error; err != nil {
		InternalError(c, "更新失败", err)
		return
	}

	Success(c, vm)
}

// Delete 删除VM
func (h *VMHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	vmID, err := uuid.Parse(id)
	if err != nil {
		BadRequest(c, "ID格式错误")
		return
	}

	var vm models.VM
	if err := h.db.Where("id = ? AND is_deleted = ?", vmID, false).First(&vm).Error; err != nil {
		NotFound(c, "VM不存在")
		return
	}

	// 软删除
	vm.IsDeleted = true
	if user := GetUser(c); user != nil {
		vm.DeletedBy = &user.ID
	}

	if err := h.db.Save(&vm).Error; err != nil {
		InternalError(c, "删除失败", err)
		return
	}

	Success(c, nil)
}

// Sync VM同步（占位符）
func (h *VMHandler) Sync(c *gin.Context) {
	var req models.VMSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// TODO: 实现实际的VMware同步逻辑

	c.JSON(http.StatusAccepted, gin.H{
		"code":    202,
		"message": "同步任务已创建",
		"data": models.VMSyncResponse{
			SyncID:    "sync_" + uuid.New().String(),
			Status:    "pending",
			StartedAt: time.Now(),
		},
	})
}

// Statistics 获取VM统计
func (h *VMHandler) Statistics(c *gin.Context) {
	var stats models.VMStatistics

	// 总体统计
	h.db.Model(&models.VM{}).Where("is_deleted = ?", false).Select(
		"COUNT(*) as total",
		"SUM(CASE WHEN status = 'online' THEN 1 ELSE 0 END) as online",
		"SUM(CASE WHEN status = 'offline' THEN 1 ELSE 0 END) as offline",
		"SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END) as error",
		"SUM(CASE WHEN status = 'unknown' THEN 1 ELSE 0 END) as unknown",
	).Scan(&stats.Overview)

	// TODO: 实现其他统计

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    stats,
	})
}

// ListGroups 获取分组列表
func (h *VMHandler) ListGroups(c *gin.Context) {
	var groups []models.VMGroup
	h.db.Where("is_deleted = ? OR is_deleted IS NULL", false).Find(&groups)

	Success(c, groups)
}

// CreateGroupRequest 创建分组请求
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type"`
	ParentID    string `json:"parentId"`
	Color       string `json:"color"`
}

// CreateGroup 创建分组
func (h *VMHandler) CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, err)
		return
	}

	group := models.VMGroup{
		Name:     req.Name,
		Type:     req.Type,
		Color:    req.Color,
		IsSystem: false,
	}

	if req.Description != "" {
		group.Description = &req.Description
	}
	if req.ParentID != "" {
		if parentID, err := uuid.Parse(req.ParentID); err == nil {
			group.ParentID = &parentID
		}
	}

	if err := h.db.Create(&group).Error; err != nil {
		InternalError(c, "创建失败", err)
		return
	}

	Created(c, group)
}

// UpdateGroup 更新分组
func (h *VMHandler) UpdateGroup(c *gin.Context) {
	id := c.Param("id")
	groupID, err := uuid.Parse(id)
	if err != nil {
		BadRequest(c, "ID格式错误")
		return
	}

	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, err)
		return
	}

	var group models.VMGroup
	if err := h.db.First(&group, "id = ?", groupID).Error; err != nil {
		NotFound(c, "分组不存在")
		return
	}

	group.Name = req.Name
	if req.Description != "" {
		group.Description = &req.Description
	}
	if req.Color != "" {
		group.Color = req.Color
	}

	if err := h.db.Save(&group).Error; err != nil {
		InternalError(c, "更新失败", err)
		return
	}

	Success(c, group)
}

// DeleteGroup 删除分组
func (h *VMHandler) DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	groupID, err := uuid.Parse(id)
	if err != nil {
		BadRequest(c, "ID格式错误")
		return
	}

	var group models.VMGroup
	if err := h.db.First(&group, "id = ?", groupID).Error; err != nil {
		NotFound(c, "分组不存在")
		return
	}

	if group.IsSystem {
		Forbidden(c, "系统分组不能删除")
		return
	}

	if err := h.db.Delete(&group).Error; err != nil {
		InternalError(c, "删除失败", err)
		return
	}

	Success(c, nil)
}

// Sync VM同步
func (h *VMHandler) Sync(c *gin.Context) {
	var req models.VMSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, err)
		return
	}

	// 验证同步类型
	if req.Type != "full" && req.Type != "incremental" {
		BadRequest(c, "同步类型必须是 full 或 incremental")
		return
	}

	// 检查采集器是否可用
	if h.collector == nil {
		Accepted(c, "同步任务已创建", gin.H{
			"syncId":    "sync_" + uuid.New().String(),
			"status":    "pending",
			"message":   "vSphere采集器未配置，仅记录同步请求",
			"startedAt": time.Now(),
		})
		return
	}

	// 如果指定了范围，验证范围ID
	if req.Type == "full" {
		// 全量同步
		if req.DatacenterID != "" {
			if _, err := uuid.Parse(req.DatacenterID); err != nil {
				BadRequest(c, "数据中心ID格式错误")
				return
			}
		}
	}

	// 创建同步任务
	syncID := "sync_" + uuid.New().String()

	// 异步执行同步（实际项目中应使用任务队列）
	go func() {
		var syncErr error
		var added, updated, removed int

		if h.collector != nil {
			// TODO: 实现实际的同步逻辑
			// 这里调用vsphereCollector进行数据同步
			// syncErr = h.collector.SyncVMs(req.Type, req.DatacenterID)
			added = 0
			updated = 0
			removed = 0
		}

		if syncErr != nil {
			logger.Error("VM同步失败", zap.Error(syncErr))
		} else {
			logger.Info("VM同步完成",
				zap.String("sync_type", req.Type),
				zap.Int("added", added),
				zap.Int("updated", updated),
				zap.Int("removed", removed),
			)
		}
	}()

	Accepted(c, "同步任务已创建", gin.H{
		"syncId":    syncID,
		"status":    "pending",
		"type":      req.Type,
		"startedAt": time.Now(),
	})
}

// Statistics 获取VM统计
func (h *VMHandler) Statistics(c *gin.Context) {
	var stats models.VMStatistics

	h.db.Model(&models.VM{}).Where("is_deleted = ?", false).Select(
		"COUNT(*) as total",
		"SUM(CASE WHEN status = 'online' THEN 1 ELSE 0 END) as online",
		"SUM(CASE WHEN status = 'offline' THEN 1 ELSE 0 END) as offline",
		"SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END) as error",
		"SUM(CASE WHEN status = 'unknown' THEN 1 ELSE 0 END) as unknown",
	).Scan(&stats.Overview)

	Success(c, stats)
}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    group,
	})
}

// DeleteGroup 删除分组
func (h *VMHandler) DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	groupID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "ID格式错误",
		})
		return
	}

	var group models.VMGroup
	if err := h.db.First(&group, "id = ?", groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "分组不存在",
		})
		return
	}

	if group.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "系统分组不能删除",
		})
		return
	}

	if err := h.db.Delete(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// BatchRequest 批量操作请求
type BatchRequest struct {
	Action string   `json:"action" binding:"required"`
	VMIDs  []string `json:"vmIds" binding:"required"`
	Force  bool     `json:"force"`
}

// Batch 批量操作
func (h *VMHandler) Batch(c *gin.Context) {
	var req BatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, err)
		return
	}

	Accepted(c, "批量任务已创建", gin.H{
		"taskId": "batch_" + uuid.New().String(),
		"status": "pending",
		"total":  len(req.VMIDs),
	})
}