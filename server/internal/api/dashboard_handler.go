package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"vm-monitoring-system/internal/models"
)

// DashboardHandler 仪表盘处理器
type DashboardHandler struct {
	db *gorm.DB
}

// NewDashboardHandler 创建仪表盘处理器
func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// OverviewResponse 仪表盘概览响应
type OverviewResponse struct {
	HealthScore  float64          `json:"healthScore"`
	HealthTrend  string           `json:"healthTrend"`
	LastUpdated  time.Time        `json:"lastUpdated"`
	SystemStatus string           `json:"systemStatus"`
	Summary      VMGroupSummary   `json:"summary"`
	Metrics      DashboardMetrics `json:"metrics"`
	TopResources TopResources     `json:"topResources,omitempty"`
}

// VMGroupSummary VM分组统计
type VMGroupSummary struct {
	TotalVMs    int `json:"totalVMs"`
	OnlineVMs   int `json:"onlineVMs"`
	OfflineVMs  int `json:"offlineVMs"`
	WarningVMs  int `json:"warningVMs"`
	CriticalVMs int `json:"criticalVMs"`
}

// DashboardMetrics 仪表盘核心指标
type DashboardMetrics struct {
	CPU     MetricItem    `json:"cpu"`
	Memory  MetricItem    `json:"memory"`
	Disk    MetricItem    `json:"disk"`
	Network NetworkMetric `json:"network"`
}

// MetricItem 指标项
type MetricItem struct {
	UsagePercent float64 `json:"usagePercent"`
	Trend        string  `json:"trend"`
	TrendValue   float64 `json:"trendValue"`
}

// NetworkMetric 网络指标
type NetworkMetric struct {
	InboundMbps  float64 `json:"inboundMbps"`
	OutboundMbps float64 `json:"outboundMbps"`
	Trend        string  `json:"trend"`
	TrendValue   float64 `json:"trendValue"`
}

// TopResources 资源使用排行
type TopResources struct {
	ByCPU    []ResourceItem `json:"byCPU,omitempty"`
	ByMemory []ResourceItem `json:"byMemory,omitempty"`
}

// ResourceItem 资源项
type ResourceItem struct {
	VMID         string  `json:"vmId"`
	VMName       string  `json:"vmName"`
	UsagePercent float64 `json:"usagePercent"`
}

// VMStatusDistribution VM状态分布
type VMStatusDistribution struct {
	Status  string  `json:"status"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
	Color   string  `json:"color"`
}

// VMGroupStatus VM分组状态
type VMGroupStatus struct {
	GroupName string `json:"groupName"`
	Count     int    `json:"count"`
	Online    int    `json:"online"`
	Offline   int    `json:"offline"`
	Warning   int    `json:"warning"`
	Critical  int    `json:"critical"`
}

// OSStatus 操作系统分布
type OSStatus struct {
	OS      string  `json:"os"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}

// DashboardAlert 仪表盘告警
type DashboardAlert struct {
	ID           string    `json:"id"`
	VMID         string    `json:"vmId"`
	VMName       string    `json:"vmName"`
	VMIP         string    `json:"vmIP"`
	AlertType    string    `json:"alertType"`
	Severity     string    `json:"severity"`
	Message      string    `json:"message"`
	Value        string    `json:"value"`
	Threshold    string    `json:"threshold"`
	OccurredAt   time.Time `json:"occurredAt"`
	Status       string    `json:"status"`
	Acknowledged bool      `json:"acknowledged"`
}

// HealthTrendDataPoint 健康趋势数据点
type HealthTrendDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Score     float64   `json:"score"`
}

// ProblemVM 问题VM
type ProblemVM struct {
	VMID          string    `json:"vmId"`
	VMName        string    `json:"vmName"`
	VMIP          string    `json:"vmIP"`
	Group         string    `json:"group"`
	Severity      string    `json:"severity"`
	Issues        []VMIssue `json:"issues"`
	FirstDetected time.Time `json:"firstDetected"`
	Duration      string    `json:"duration"`
}

// VMIssue VM问题
type VMIssue struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Value   string `json:"value"`
}

// GetOverview 获取仪表盘概览数据
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	var vmSummary VMGroupSummary
	var totalVMs, onlineVMs, offlineVMs, warningVMs, criticalVMs int64

	// 获取VM总数
	h.db.Model(&models.VM{}).Count(&totalVMs)

	// 获取各状态VM数量
	h.db.Model(&models.VM{}).Where("status = ?", "online").Count(&onlineVMs)
	h.db.Model(&models.VM{}).Where("status = ?", "offline").Count(&offlineVMs)
	h.db.Model(&models.VM{}).Where("status = ?", "warning").Count(&warningVMs)
	h.db.Model(&models.VM{}).Where("status = ?", "critical").Count(&criticalVMs)

	vmSummary.TotalVMs = int(totalVMs)
	vmSummary.OnlineVMs = int(onlineVMs)
	vmSummary.OfflineVMs = int(offlineVMs)
	vmSummary.WarningVMs = int(warningVMs)
	vmSummary.CriticalVMs = int(criticalVMs)

	// 计算健康评分
	healthScore := calculateHealthScore(vmSummary)

	// 获取核心指标（从最新监控数据中获取）
	metrics := h.getCoreMetrics()

	// 获取资源使用排行
	topResources := h.getTopResources()

	// 确定系统状态
	systemStatus := "healthy"
	if healthScore >= 70 && healthScore < 90 {
		systemStatus = "warning"
	} else if healthScore < 70 {
		systemStatus = "critical"
	}

	response := OverviewResponse{
		HealthScore:  healthScore,
		HealthTrend:  "stable",
		LastUpdated:  time.Now(),
		SystemStatus: systemStatus,
		Summary:      vmSummary,
		Metrics:      metrics,
		TopResources: topResources,
	}

	Success(c, http.StatusOK, response)
}

// getCoreMetrics 获取核心指标数据
func (h *DashboardHandler) getCoreMetrics() DashboardMetrics {
	var metrics DashboardMetrics

	var latestData struct {
		CPUUsage  float64
		MemUsage  float64
		DiskUsage float64
		NetInBps  float64
		NetOutBps float64
	}

	// 这里应该从时序数据库或缓存中获取实际数据
	// 暂时返回模拟数据用于演示
	latestData.CPUUsage = 65.5
	latestData.MemUsage = 72.3
	latestData.DiskUsage = 58.2
	latestData.NetInBps = 125.5 * 1024 * 1024 / 8 // 转换为bps
	latestData.NetOutBps = 89.3 * 1024 * 1024 / 8

	metrics.CPU = MetricItem{
		UsagePercent: latestData.CPUUsage,
		Trend:        "stable",
		TrendValue:   2.5,
	}
	metrics.Memory = MetricItem{
		UsagePercent: latestData.MemUsage,
		Trend:        "up",
		TrendValue:   1.8,
	}
	metrics.Disk = MetricItem{
		UsagePercent: latestData.DiskUsage,
		Trend:        "stable",
		TrendValue:   0.5,
	}
	metrics.Network = NetworkMetric{
		InboundMbps:  latestData.NetInBps / 1024 / 1024 * 8,
		OutboundMbps: latestData.NetOutBps / 1024 / 1024 * 8,
		Trend:        "up",
		TrendValue:   5.2,
	}

	return metrics
}

// getTopResources 获取资源使用排行
func (h *DashboardHandler) getTopResources() TopResources {
	var resources TopResources

	var topCPU []struct {
		VMID   string
		Name   string
		CPUAvg float64
	}
	var topMemory []struct {
		VMID   string
		Name   string
		MemAvg float64
	}

	// 获取CPU使用率最高的VM
	h.db.Model(&models.VM{}).
		Select("id as vm_id, name as name, cpu_usage as cpu_avg").
		Where("status != ?", "offline").
		Order("cpu_avg DESC").
		Limit(5).
		Scan(&topCPU)

	// 获取内存使用率最高的VM
	h.db.Model(&models.VM{}).
		Select("id as vm_id, name as name, memory_usage as mem_avg").
		Where("status != ?", "offline").
		Order("mem_avg DESC").
		Limit(5).
		Scan(&topMemory)

	for _, item := range topCPU {
		resources.ByCPU = append(resources.ByCPU, ResourceItem{
			VMID:         item.VMID,
			VMName:       item.Name,
			UsagePercent: item.CPUAvg,
		})
	}

	for _, item := range topMemory {
		resources.ByMemory = append(resources.ByMemory, ResourceItem{
			VMID:         item.VMID,
			VMName:       item.Name,
			UsagePercent: item.MemAvg,
		})
	}

	return resources
}

// GetVMStatus 获取VM状态分布
func (h *DashboardHandler) GetVMStatus(c *gin.Context) {
	var distribution []VMStatusDistribution
	var totalCount int64

	h.db.Model(&models.VM{}).Count(&totalCount)

	if totalCount == 0 {
		Success(c, http.StatusOK, gin.H{
			"distribution": []VMStatusDistribution{},
			"byGroup":      []VMGroupStatus{},
			"byOS":         []OSStatus{},
		})
		return
	}

	// 按状态统计
	statusCounts := make(map[string]int64)
	statuses := []string{"online", "offline", "warning", "critical"}
	colors := map[string]string{
		"online":   "#00d4aa",
		"offline":  "#607d8b",
		"warning":  "#ff9800",
		"critical": "#f44336",
	}

	for _, status := range statuses {
		var count int64
		h.db.Model(&models.VM{}).Where("status = ?", status).Count(&count)
		statusCounts[status] = count

		distribution = append(distribution, VMStatusDistribution{
			Status:  status,
			Count:   int(count),
			Percent: float64(count) / float64(totalCount) * 100,
			Color:   colors[status],
		})
	}

	// 按分组统计（如果有分组信息）
	var byGroup []VMGroupStatus
	var groups []string
	h.db.Model(&models.VM{}).Distinct("group_name").Pluck("group_name", &groups)

	for _, group := range groups {
		var groupCount, online, offline, warning, critical int64
		h.db.Model(&models.VM{}).Where("group_name = ?", group).Count(&groupCount)
		h.db.Model(&models.VM{}).Where("group_name = ? AND status = ?", group, "online").Count(&online)
		h.db.Model(&models.VM{}).Where("group_name = ? AND status = ?", group, "offline").Count(&offline)
		h.db.Model(&models.VM{}).Where("group_name = ? AND status = ?", group, "warning").Count(&warning)
		h.db.Model(&models.VM{}).Where("group_name = ? AND status = ?", group, "critical").Count(&critical)

		byGroup = append(byGroup, VMGroupStatus{
			GroupName: group,
			Count:     int(groupCount),
			Online:    int(online),
			Offline:   int(offline),
			Warning:   int(warning),
			Critical:  int(critical),
		})
	}

	// 按操作系统统计
	var byOS []OSStatus
	var osCounts []struct {
		OS    string
		Count int64
	}

	h.db.Model(&models.VM{}).Select("os_name as os, count(*) as count").Group("os_name").Scan(&osCounts)

	for _, os := range osCounts {
		byOS = append(byOS, OSStatus{
			OS:      os.OS,
			Count:   int(os.Count),
			Percent: float64(os.Count) / float64(totalCount) * 100,
		})
	}

	Success(c, http.StatusOK, gin.H{
		"distribution": distribution,
		"byGroup":      byGroup,
		"byOS":         byOS,
	})
}

// GetAlerts 获取最新告警列表
func (h *DashboardHandler) GetAlerts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if limit < 1 || limit > 20 {
		limit = 5
	}

	var alerts []DashboardAlert
	var total, unread int64

	// 获取告警总数
	h.db.Model(&models.AlertRecord{}).Count(&total)

	// 获取未读告警数
	h.db.Model(&models.AlertRecord{}).Where("acknowledged = ?", false).Count(&unread)

	// 获取最新告警
	var records []models.AlertRecord
	h.db.Preload("VM").Order("occurred_at DESC").Limit(limit).Find(&records)

	for _, record := range records {
		vmIP := ""
		vmName := "Unknown VM"
		if record.VM.ID != "" {
			vmName = record.VM.Name
			vmIP = record.VM.IP
		}

		severity := "low"
		if record.Severity == "critical" || record.Severity == "high" {
			severity = record.Severity
		} else if record.Severity == "medium" {
			severity = "warning"
		}

		alerts = append(alerts, DashboardAlert{
			ID:           record.ID,
			VMID:         record.VMID,
			VMName:       vmName,
			VMIP:         vmIP,
			AlertType:    record.AlertType,
			Severity:     severity,
			Message:      record.Message,
			Value:        fmt.Sprintf("%.1f%%", record.CurrentValue),
			Threshold:    fmt.Sprintf("%.1f%%", record.Threshold),
			OccurredAt:   record.OccurredAt,
			Status:       string(record.Status),
			Acknowledged: record.Acknowledged,
		})
	}

	Success(c, http.StatusOK, gin.H{
		"alerts":      alerts,
		"total":       total,
		"unreadCount": unread,
	})
}

// GetHealthTrend 获取健康度历史趋势
func (h *DashboardHandler) GetHealthTrend(c *gin.Context) {
	period := c.DefaultQuery("period", "7d")

	var dataPoints []HealthTrendDataPoint

	switch period {
	case "24h":
		// 最近24小时，每小时一个数据点
		for i := 23; i >= 0; i-- {
			timestamp := time.Now().Add(-time.Duration(i) * time.Hour)
			score := 90.0 + float64(10-i%5) // 模拟数据
			dataPoints = append(dataPoints, HealthTrendDataPoint{
				Timestamp: timestamp,
				Score:     score,
			})
		}
	case "30d":
		// 最近30天，每天一个数据点
		for i := 29; i >= 0; i-- {
			timestamp := time.Now().AddDate(0, 0, -i)
			score := 88.0 + float64(i%7) // 模拟数据
			dataPoints = append(dataPoints, HealthTrendDataPoint{
				Timestamp: timestamp,
				Score:     score,
			})
		}
	case "7d":
		fallthrough
	default:
		// 最近7天，每天一个数据点
		for i := 6; i >= 0; i-- {
			timestamp := time.Now().AddDate(0, 0, -i)
			score := 88.0 + float64(i*2%5) // 模拟数据
			dataPoints = append(dataPoints, HealthTrendDataPoint{
				Timestamp: timestamp,
				Score:     score,
			})
		}
	}

	currentScore := 95.0
	if len(dataPoints) > 0 {
		currentScore = dataPoints[len(dataPoints)-1].Score
	}

	trend := "stable"
	if len(dataPoints) >= 2 {
		diff := dataPoints[len(dataPoints)-1].Score - dataPoints[len(dataPoints)-2].Score
		if diff > 1 {
			trend = "up"
		} else if diff < -1 {
			trend = "down"
		}
	}

	Success(c, http.StatusOK, gin.H{
		"period":       period,
		"currentScore": currentScore,
		"trend":        trend,
		"dataPoints":   dataPoints,
	})
}

// GetProblemVMs 获取问题VM列表
func (h *DashboardHandler) GetProblemVMs(c *gin.Context) {
	severityFilter := c.Query("severity")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var problemVMs []ProblemVM

	query := h.db.Model(&models.VM{}).Where("status IN ?", []string{"warning", "critical"})

	if severityFilter != "" {
		query = query.Where("status = ?", severityFilter)
	}

	query.Order("CASE WHEN status = 'critical' THEN 1 WHEN status = 'warning' THEN 2 ELSE 3 END").
		Order("updated_at DESC").
		Limit(limit).
		Scan(&problemVMs)

	total := len(problemVMs)

	Success(c, http.StatusOK, gin.H{
		"total": total,
		"vms":   problemVMs,
	})
}

// calculateHealthScore 计算健康评分
func calculateHealthScore(summary VMGroupSummary) float64 {
	if summary.TotalVMs == 0 {
		return 100.0
	}

	// 在线率权重 30%
	onlineRate := float64(summary.OnlineVMs) / float64(summary.TotalVMs) * 100

	// 性能指标权重 30%
	performanceScore := 100.0 - float64(summary.WarningVMs)*2 - float64(summary.CriticalVMs)*5

	// 告警扣分权重 25%
	alertPenalty := float64(summary.WarningVMs)*2 + float64(summary.CriticalVMs)*8

	// 计算总分
	score := onlineRate*0.3 + performanceScore*0.3 + (100-alertPenalty)*0.25 + 15 // 基础分15

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}
