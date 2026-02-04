package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"vm-monitoring-system/internal/models"
)

// AlertEngine 告警引擎
type AlertEngine struct {
	db               *gorm.DB
	notifier         *NotificationService
	rules            map[uuid.UUID]*AlertRuleWithConditions
	rulesMutex       sync.RWMutex
	evalInterval     time.Duration
	stopChan         chan struct{}
	isRunning        bool
	runningMutex     sync.RWMutex
	triggerHistory   map[string]time.Time
	historyMutex     sync.RWMutex
}

// AlertRuleWithConditions 带条件的告警规则
type AlertRuleWithConditions struct {
	Rule       models.AlertRule
	Conditions []models.AlertCondition
}

// MetricData 指标数据
type MetricData struct {
	VMID      uuid.UUID
	VMName    string
	Metric    string
	Value     float64
	Timestamp time.Time
}

// NewAlertEngine 创建告警引擎
func NewAlertEngine(db *gorm.DB, notifier *NotificationService) *AlertEngine {
	return &AlertEngine{
		db:             db,
		notifier:       notifier,
		rules:          make(map[uuid.UUID]*AlertRuleWithConditions),
		evalInterval:   60 * time.Second, // 默认60秒评估一次
		stopChan:       make(chan struct{}),
		triggerHistory: make(map[string]time.Time),
	}
}

// Start 启动告警引擎
func (e *AlertEngine) Start() error {
	e.runningMutex.Lock()
	defer e.runningMutex.Unlock()

	if e.isRunning {
		return fmt.Errorf("告警引擎已经在运行")
	}

	// 加载所有启用的规则
	if err := e.loadRules(); err != nil {
		return fmt.Errorf("加载告警规则失败: %w", err)
	}

	e.isRunning = true
	e.stopChan = make(chan struct{})

	// 启动评估循环
	go e.evaluationLoop()

	log.Println("告警引擎已启动，评估间隔:", e.evalInterval)
	return nil
}

// Stop 停止告警引擎
func (e *AlertEngine) Stop() {
	e.runningMutex.Lock()
	defer e.runningMutex.Unlock()

	if !e.isRunning {
		return
	}

	close(e.stopChan)
	e.isRunning = false
	log.Println("告警引擎已停止")
}

// IsRunning 检查引擎是否运行中
func (e *AlertEngine) IsRunning() bool {
	e.runningMutex.RLock()
	defer e.runningMutex.RUnlock()
	return e.isRunning
}

// SetEvalInterval 设置评估间隔
func (e *AlertEngine) SetEvalInterval(interval time.Duration) {
	e.evalInterval = interval
}

// ReloadRules 重新加载规则
func (e *AlertEngine) ReloadRules() error {
	return e.loadRules()
}

// loadRules 从数据库加载所有启用的规则
func (e *AlertEngine) loadRules() error {
	var rules []models.AlertRule

	// 查询启用的、未删除的规则
	if err := e.db.Where("enabled = ? AND is_deleted = ?", true, false).Find(&rules).Error; err != nil {
		return err
	}

	newRules := make(map[uuid.UUID]*AlertRuleWithConditions)

	for _, rule := range rules {
		// 加载条件
		var conditions []models.AlertCondition
		if err := e.db.Where("rule_id = ?", rule.ID).Order("sort_order").Find(&conditions).Error; err != nil {
			log.Printf("加载规则 %s 的条件失败: %v", rule.ID, err)
			continue
		}

		newRules[rule.ID] = &AlertRuleWithConditions{
			Rule:       rule,
			Conditions: conditions,
		}
	}

	e.rulesMutex.Lock()
	e.rules = newRules
	e.rulesMutex.Unlock()

	log.Printf("已加载 %d 条告警规则", len(newRules))
	return nil
}

// evaluationLoop 评估循环
func (e *AlertEngine) evaluationLoop() {
	ticker := time.NewTicker(e.evalInterval)
	defer ticker.Stop()

	// 立即执行一次评估
	e.evaluateAllRules()

	for {
		select {
		case <-ticker.C:
			e.evaluateAllRules()
		case <-e.stopChan:
			return
		}
	}
}

// evaluateAllRules 评估所有规则
func (e *AlertEngine) evaluateAllRules() {
	e.rulesMutex.RLock()
	rules := make(map[uuid.UUID]*AlertRuleWithConditions)
	for k, v := range e.rules {
		rules[k] = v
	}
	e.rulesMutex.RUnlock()

	for _, ruleWithCond := range rules {
		if err := e.evaluateRule(ruleWithCond); err != nil {
			log.Printf("评估规则 %s 失败: %v", ruleWithCond.Rule.ID, err)
		}
	}
}

// evaluateRule 评估单个规则
func (e *AlertEngine) evaluateRule(ruleWithCond *AlertRuleWithConditions) error {
	rule := ruleWithCond.Rule
	conditions := ruleWithCond.Conditions

	// 检查冷却期
	if !e.isCooldownExpired(rule.ID, rule.Cooldown) {
		return nil
	}

	// 根据范围获取目标VM
	vms, err := e.getTargetVMs(rule.Scope, rule.ScopeID)
	if err != nil {
		return fmt.Errorf("获取目标VM失败: %w", err)
	}

	// 对每个VM评估规则
	for _, vm := range vms {
		// 检查此VM是否在此规则下已触发且未恢复
		alertKey := fmt.Sprintf("%s:%s", rule.ID, vm.ID)
		if e.isActiveAlert(alertKey) {
			// 检查是否已恢复
			if e.checkRecovery(ruleWithCond, vm) {
				e.resolveAlert(alertKey, vm.ID, rule.ID)
			}
			continue
		}

		// 评估条件
		triggered, metricData, err := e.evaluateConditions(ruleWithCond, vm)
		if err != nil {
			log.Printf("评估条件失败 (规则:%s, VM:%s): %v", rule.ID, vm.ID, err)
			continue
		}

		if triggered {
			// 创建告警记录
			if err := e.createAlert(rule, vm, metricData); err != nil {
				log.Printf("创建告警失败: %v", err)
				continue
			}

			// 更新触发历史
			e.updateTriggerHistory(rule.ID)
		}
	}

	return nil
}

// getTargetVMs 获取目标VM列表
func (e *AlertEngine) getTargetVMs(scope string, scopeID *uuid.UUID) ([]models.VM, error) {
	var vms []models.VM

	switch scope {
	case "all":
		// 所有VM
		if err := e.db.Where("is_deleted = ? AND status != ?", false, "unknown").Find(&vms).Error; err != nil {
			return nil, err
		}

	case "vm":
		// 特定VM
		if scopeID == nil {
			return nil, fmt.Errorf("VM范围需要指定VM ID")
		}
		var vm models.VM
		if err := e.db.First(&vm, "id = ? AND is_deleted = ?", scopeID, false).Error; err != nil {
			return nil, err
		}
		vms = append(vms, vm)

	case "group":
		// VM分组
		if scopeID == nil {
			return nil, fmt.Errorf("分组范围需要指定分组 ID")
		}
		if err := e.db.Joins("JOIN vm_group_members ON vms.id = vm_group_members.vm_id").
			Where("vm_group_members.group_id = ? AND vms.is_deleted = ?", scopeID, false).
			Find(&vms).Error; err != nil {
			return nil, err
		}

	case "cluster":
		// 集群
		if scopeID == nil {
			return nil, fmt.Errorf("集群范围需要指定集群 ID")
		}
		if err := e.db.Where("cluster_id = ? AND is_deleted = ?", scopeID.String(), false).Find(&vms).Error; err != nil {
			return nil, err
		}

	case "host":
		// 主机
		if scopeID == nil {
			return nil, fmt.Errorf("主机范围需要指定主机 ID")
		}
		if err := e.db.Where("host_id = ? AND is_deleted = ?", scopeID.String(), false).Find(&vms).Error; err != nil {
			return nil, err
		}

	case "datacenter":
		// 数据中心
		if scopeID == nil {
			return nil, fmt.Errorf("数据中心范围需要指定数据中心 ID")
		}
		if err := e.db.Where("datacenter_id = ? AND is_deleted = ?", scopeID.String(), false).Find(&vms).Error; err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("未知的作用域类型: %s", scope)
	}

	return vms, nil
}

// evaluateConditions 评估条件
func (e *AlertEngine) evaluateConditions(ruleWithCond *AlertRuleWithConditions, vm models.VM) (bool, *MetricData, error) {
	conditions := ruleWithCond.Conditions
	logic := ruleWithCond.Rule.ConditionLogic

	if len(conditions) == 0 {
		return false, nil, nil
	}

	results := make([]bool, len(conditions))
	var triggeredMetric *MetricData

	for i, cond := range conditions {
		// 获取指标值
		metricValue, err := e.getMetricValue(vm.ID, cond.Metric, cond.Aggregation, cond.Duration)
		if err != nil {
			// 如果无法获取指标值，认为条件不满足
			results[i] = false
			continue
		}

		// 评估单个条件
		result := e.evaluateSingleCondition(metricValue, cond.Operator, cond.Threshold)
		results[i] = result

		if result && triggeredMetric == nil {
			triggeredMetric = &MetricData{
				VMID:      vm.ID,
				VMName:    vm.Name,
				Metric:    cond.Metric,
				Value:     metricValue,
				Timestamp: time.Now(),
			}
		}
	}

	// 根据逻辑组合条件结果
	var triggered bool
	switch logic {
	case "and":
		triggered = true
		for _, r := range results {
			if !r {
				triggered = false
				break
			}
		}
	case "or":
		triggered = false
		for _, r := range results {
			if r {
				triggered = true
				break
			}
		}
	default:
		triggered = false
	}

	return triggered, triggeredMetric, nil
}

// evaluateSingleCondition 评估单个条件
func (e *AlertEngine) evaluateSingleCondition(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "=", "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}

// checkRecovery 检查是否已恢复
func (e *AlertEngine) checkRecovery(ruleWithCond *AlertRuleWithConditions, vm models.VM) bool {
	// 如果所有条件都不满足，则认为已恢复
	triggered, _, _ := e.evaluateConditions(ruleWithCond, vm)
	return !triggered
}

// isCooldownExpired 检查冷却期是否已过
func (e *AlertEngine) isCooldownExpired(ruleID uuid.UUID, cooldownSeconds int) bool {
	e.historyMutex.RLock()
	lastTriggered, exists := e.triggerHistory[ruleID.String()]
	e.historyMutex.RUnlock()

	if !exists {
		return true
	}

	cooldown := time.Duration(cooldownSeconds) * time.Second
	return time.Since(lastTriggered) >= cooldown
}

// updateTriggerHistory 更新触发历史
func (e *AlertEngine) updateTriggerHistory(ruleID uuid.UUID) {
	e.historyMutex.Lock()
	defer e.historyMutex.Unlock()
	e.triggerHistory[ruleID.String()] = time.Now()
}

// isActiveAlert 检查是否存在活动告警
func (e *AlertEngine) isActiveAlert(alertKey string) bool {
	var count int64
	e.db.Model(&models.AlertRecord{}).
		Where("rule_id || ':' || vm_id = ? AND status = ?", alertKey, "active").
		Count(&count)
	return count > 0
}

// resolveAlert 解决告警
func (e *AlertEngine) resolveAlert(alertKey string, vmID uuid.UUID, ruleID uuid.UUID) {
	// 查找并更新活动告警
	var alert models.AlertRecord
	if err := e.db.Where("vm_id = ? AND rule_id = ? AND status = ?", vmID, ruleID, "active").
		Order("triggered_at DESC").
		First(&alert).Error; err != nil {
		return
	}

	now := time.Now()
	alert.Status = "resolved"
	alert.ResolvedAt = &now
	if alert.Duration != nil {
		duration := int(now.Sub(alert.TriggeredAt).Minutes())
		alert.Duration = &duration
	}

	if err := e.db.Save(&alert).Error; err != nil {
		log.Printf("解决告警失败: %v", err)
		return
	}

	log.Printf("告警已自动恢复: %s", alert.RuleName)
}

// createAlert 创建告警记录
func (e *AlertEngine) createAlert(rule models.AlertRule, vm models.VM, metricData *MetricData) error {
	// 检查是否已存在相同的活动告警
	var existingCount int64
	e.db.Model(&models.AlertRecord{}).
		Where("rule_id = ? AND vm_id = ? AND status = ?", rule.ID, vm.ID, "active").
		Count(&existingCount)

	if existingCount > 0 {
		return nil // 已存在活动告警，不重复创建
	}

	// 创建快照数据
	snapshot := map[string]interface{}{
		"vm": map[string]interface{}{
			"id":         vm.ID,
			"name":       vm.Name,
			"ip":         vm.IP,
			"os":         vm.OSType,
			"hostName":   vm.HostName,
			"clusterName": vm.ClusterName,
		},
		"triggeredAt": time.Now(),
	}

	snapshotJSON, _ := json.Marshal(snapshot)

	// 查找触发的具体条件
	var triggeredCondition models.AlertCondition
	for _, cond := range rule.Conditions {
		if cond.Metric == metricData.Metric {
			triggeredCondition = cond
			break
		}
	}

	// 构建条件字符串
	conditionStr := fmt.Sprintf("%s %s %.4f (实际值: %.4f)",
		metricData.Metric,
		triggeredCondition.Operator,
		triggeredCondition.Threshold,
		metricData.Value,
	)

	alert := models.AlertRecord{
		ID:            uuid.New(),
		RuleID:        rule.ID,
		RuleName:      rule.Name,
		VMID:          &vm.ID,
		VMName:        &vm.Name,
		ClusterID:     vm.ClusterID,
		Metric:        metricData.Metric,
		Severity:      rule.Severity,
		TriggerValue:  metricData.Value,
		Threshold:     triggeredCondition.Threshold,
		ConditionStr:  &conditionStr,
		TriggeredAt:   time.Now(),
		Status:        "active",
		Snapshot:      models.JSONMap(snapshot),
		NotificationStatus: models.JSONMap{},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 保存告警记录
	if err := e.db.Create(&alert).Error; err != nil {
		return fmt.Errorf("保存告警记录失败: %w", err)
	}

	// 更新规则触发计数
	e.db.Model(&models.AlertRule{}).
		Where("id = ?", rule.ID).
		Updates(map[string]interface{}{
			"trigger_count":    gorm.Expr("trigger_count + 1"),
			"last_triggered_at": time.Now(),
		})

	// 发送通知
	if e.notifier != nil {
		go e.notifier.SendAlert(context.Background(), alert, rule.NotificationConfig)
	}

	log.Printf("告警已触发: %s - %s", rule.Name, vm.Name)
	return nil
}

// getMetricValue 获取指标值（从数据库或缓存）
func (e *AlertEngine) getMetricValue(vmID uuid.UUID, metric string, aggregation string, duration int) (float64, error) {
	// 这里应该从时序数据库获取指标值
	// 目前先用模拟数据，后续需要接入实际的数据采集系统

	// TODO: 实现从TimescaleDB或Redis获取实时/历史指标数据
	// 查询逻辑示例:
	// 1. 优先从Redis缓存获取最新值
	// 2. 如果没有，从TimescaleDB查询聚合值
	// 3. 根据聚合类型(avg/max/min/last)计算结果

	// 临时返回模拟数据，实际实现需要接入数据采集
	return e.getSimulatedMetricValue(vmID, metric)
}

// getSimulatedMetricValue 获取模拟指标值（开发测试用）
func (e *AlertEngine) getSimulatedMetricValue(vmID uuid.UUID, metric string) (float64, error) {
	// 实际项目中应删除此函数，改为从真实数据源获取
	// 这里仅用于演示，返回随机但合理的值
	
	switch metric {
	case "cpu_usage":
		return 75.5, nil // 75.5% CPU使用率
	case "memory_usage":
		return 82.3, nil // 82.3% 内存使用率
	case "disk_usage":
		return 68.0, nil // 68% 磁盘使用率
	case "network_rx":
		return 1024.0, nil // 1MB/s 网络接收
	case "network_tx":
		return 512.0, nil // 512KB/s 网络发送
	case "disk_io_read":
		return 100.0, nil // 100 IOPS 磁盘读
	case "disk_io_write":
		return 80.0, nil // 80 IOPS 磁盘写
	default:
		return 0, fmt.Errorf("未知的指标类型: %s", metric)
	}
}
