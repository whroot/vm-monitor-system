package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// AlertRecord 告警记录模型
type AlertRecord struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RuleID         uuid.UUID  `gorm:"type:uuid;not null"`
	RuleName       string     `gorm:"type:varchar(200);not null"`
	VMID           *uuid.UUID `gorm:"type:uuid"`
	VMName         *string    `gorm:"type:varchar(200)"`
	Metric         string     `gorm:"type:varchar(50);not null"`
	Severity       string     `gorm:"type:varchar(20);not null"`
	TriggerValue   float64    `gorm:"not null"`
	Threshold      float64    `gorm:"not null"`
	TriggeredAt    time.Time  `gorm:"not null"`
	Status         string     `gorm:"type:varchar(20);not null;default:'active'"`
	AcknowledgedBy *uuid.UUID `gorm:"type:uuid"`
	AcknowledgedAt *time.Time
	ResolvedBy     *uuid.UUID `gorm:"type:uuid"`
	ResolvedAt     *time.Time
	Resolution     *string `gorm:"type:text"`
	Duration       *int
	CreatedAt      time.Time
}

func (AlertRecord) TableName() string {
	return "alert_records"
}

// VM 虚拟机模型（仅用于查询）
type VM struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key"`
	Name string    `gorm:"type:varchar(200)"`
}

func (VM) TableName() string {
	return "vms"
}

func main() {
	// 连接数据库
	dsn := "host=localhost user=postgres password=postgres dbname=vm_monitoring port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// 查询现有的VM
	var vms []VM
	if err := db.Find(&vms).Error; err != nil {
		panic("failed to fetch VMs: " + err.Error())
	}

	if len(vms) == 0 {
		fmt.Println("没有可用的VM，请先创建VM")
		return
	}

	fmt.Printf("找到 %d 台VM，开始生成告警数据...\n\n", len(vms))

	// 告警规则配置
	alertRules := []struct {
		ruleID    uuid.UUID
		name      string
		metric    string
		severity  string
		threshold float64
	}{
		{uuid.MustParse("79f71333-1bf0-4ff8-aa73-ee464b31f554"), "CPU使用率告警", "cpu.usage.percent", "warning", 80},
		{uuid.MustParse("5c44698f-66a7-4b0e-b1bb-ecc23f39db45"), "内存使用率告警", "memory.usage.percent", "warning", 80},
		{uuid.MustParse("3a3e9860-0315-427a-9237-4300191b5609"), "磁盘使用率告警", "disk.usage.percent", "critical", 90},
		{uuid.MustParse("b37820ea-ee6e-4c81-86b7-7522f631fece"), "CPU使用率严重告警", "cpu.usage.percent", "critical", 95},
	}

	// 生成告警记录
	alertCount := 0

	// 1. 生成一些活跃的告警（active）
	for i := 0; i < 5; i++ {
		vm := vms[rand.Intn(len(vms))]
		rule := alertRules[rand.Intn(len(alertRules))]

		triggerValue := rule.threshold + rand.Float64()*15 // 超过阈值15%以内

		alert := AlertRecord{
			ID:           uuid.New(),
			RuleID:       rule.ruleID,
			RuleName:     rule.name,
			VMID:         &vm.ID,
			VMName:       &vm.Name,
			Metric:       rule.metric,
			Severity:     rule.severity,
			TriggerValue: round(triggerValue),
			Threshold:    rule.threshold,
			TriggeredAt:  time.Now().Add(-time.Duration(rand.Intn(120)) * time.Minute), // 过去2小时内
			Status:       "active",
			CreatedAt:    time.Now(),
		}

		if err := db.Create(&alert).Error; err != nil {
			fmt.Printf("创建告警失败: %v\n", err)
		} else {
			alertCount++
			fmt.Printf("✓ 活跃告警: %s - %s (%.1f%%)\n", vm.Name, rule.name, triggerValue)
		}
	}

	// 2. 生成一些已确认的告警（acknowledged）
	for i := 0; i < 3; i++ {
		vm := vms[rand.Intn(len(vms))]
		rule := alertRules[rand.Intn(len(alertRules))]

		triggerValue := rule.threshold + rand.Float64()*10
		acknowledgedAt := time.Now().Add(-time.Duration(rand.Intn(60)) * time.Minute)

		alert := AlertRecord{
			ID:             uuid.New(),
			RuleID:         rule.ruleID,
			RuleName:       rule.name,
			VMID:           &vm.ID,
			VMName:         &vm.Name,
			Metric:         rule.metric,
			Severity:       rule.severity,
			TriggerValue:   round(triggerValue),
			Threshold:      rule.threshold,
			TriggeredAt:    time.Now().Add(-time.Duration(rand.Intn(180)) * time.Minute),
			Status:         "acknowledged",
			AcknowledgedBy: &uuid.UUID{},
			AcknowledgedAt: &acknowledgedAt,
			CreatedAt:      time.Now(),
		}

		if err := db.Create(&alert).Error; err != nil {
			fmt.Printf("创建告警失败: %v\n", err)
		} else {
			alertCount++
			fmt.Printf("✓ 已确认告警: %s - %s\n", vm.Name, rule.name)
		}
	}

	// 3. 生成一些已解决的告警（resolved）
	for i := 0; i < 4; i++ {
		vm := vms[rand.Intn(len(vms))]
		rule := alertRules[rand.Intn(len(alertRules))]

		triggerValue := rule.threshold + rand.Float64()*20
		triggeredAt := time.Now().Add(-time.Duration(rand.Intn(360)) * time.Minute) // 过去6小时内
		resolvedAt := triggeredAt.Add(time.Duration(rand.Intn(60)) * time.Minute)   // 1小时内解决
		duration := int(resolvedAt.Sub(triggeredAt).Minutes())
		resolution := "已自动恢复"

		alert := AlertRecord{
			ID:             uuid.New(),
			RuleID:         rule.ruleID,
			RuleName:       rule.name,
			VMID:           &vm.ID,
			VMName:         &vm.Name,
			Metric:         rule.metric,
			Severity:       rule.severity,
			TriggerValue:   round(triggerValue),
			Threshold:      rule.threshold,
			TriggeredAt:    triggeredAt,
			Status:         "resolved",
			AcknowledgedBy: &uuid.UUID{},
			AcknowledgedAt: &triggeredAt,
			ResolvedBy:     &uuid.UUID{},
			ResolvedAt:     &resolvedAt,
			Resolution:     &resolution,
			Duration:       &duration,
			CreatedAt:      time.Now(),
		}

		if err := db.Create(&alert).Error; err != nil {
			fmt.Printf("创建告警失败: %v\n", err)
		} else {
			alertCount++
			fmt.Printf("✓ 已解决告警: %s - %s (持续%d分钟)\n", vm.Name, rule.name, duration)
		}
	}

	// 4. 生成一些严重级别的告警（critical）
	for i := 0; i < 3; i++ {
		vm := vms[rand.Intn(len(vms))]
		triggerValue := 95.0 + rand.Float64()*5 // 95-100%

		alert := AlertRecord{
			ID:           uuid.New(),
			RuleID:       uuid.MustParse("b37820ea-ee6e-4c81-86b7-7522f631fece"),
			RuleName:     "CPU使用率严重告警",
			VMID:         &vm.ID,
			VMName:       &vm.Name,
			Metric:       "cpu.usage.percent",
			Severity:     "critical",
			TriggerValue: round(triggerValue),
			Threshold:    95,
			TriggeredAt:  time.Now().Add(-time.Duration(rand.Intn(30)) * time.Minute),
			Status:       "active",
			CreatedAt:    time.Now(),
		}

		if err := db.Create(&alert).Error; err != nil {
			fmt.Printf("创建告警失败: %v\n", err)
		} else {
			alertCount++
			fmt.Printf("✓ 严重告警: %s - CPU %.1f%%\n", vm.Name, triggerValue)
		}
	}

	fmt.Printf("\n✅ 告警数据生成完成！\n")
	fmt.Printf("共创建 %d 条告警记录\n", alertCount)

	// 统计
	var activeCount, acknowledgedCount, resolvedCount int64
	db.Model(&AlertRecord{}).Where("status = ?", "active").Count(&activeCount)
	db.Model(&AlertRecord{}).Where("status = ?", "acknowledged").Count(&acknowledgedCount)
	db.Model(&AlertRecord{}).Where("status = ?", "resolved").Count(&resolvedCount)

	fmt.Printf("\n统计:\n")
	fmt.Printf("- 活跃告警: %d\n", activeCount)
	fmt.Printf("- 已确认: %d\n", acknowledgedCount)
	fmt.Printf("- 已解决: %d\n", resolvedCount)
}

func round(val float64) float64 {
	return float64(int(val*10)) / 10
}
