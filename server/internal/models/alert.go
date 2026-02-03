package models

import (
	"time"

	"github.com/google/uuid"
)

// AlertRule 告警规则
type AlertRule struct {
	ID                 uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name               string             `gorm:"type:varchar(200);not null" json:"name"`
	Description        *string            `gorm:"type:text" json:"description,omitempty"`
	Scope              string             `gorm:"type:varchar(20);not null" json:"scope"`
	ScopeID            *uuid.UUID         `gorm:"type:uuid" json:"scopeId,omitempty"`
	ScopeName          *string            `gorm:"type:varchar(200)" json:"scopeName,omitempty"`
	ConditionLogic     string             `gorm:"type:varchar(10);not null;default:'and'" json:"conditionLogic"`
	Enabled            bool               `gorm:"not null;default:true" json:"enabled"`
	Cooldown           int                `gorm:"not null;default:300" json:"cooldown"`
	Severity           string             `gorm:"type:varchar(20);not null" json:"severity"`
	NotificationConfig JSONMap            `gorm:"type:jsonb;not null" json:"notificationConfig"`
	TriggerCount       int                `gorm:"not null;default:0" json:"triggerCount"`
	LastTriggeredAt    *time.Time         `json:"lastTriggeredAt,omitempty"`
	IsDeleted          bool               `gorm:"not null;default:false" json:"-"`
	DeletedAt          *time.Time         `json:"-"`
	CreatedAt          time.Time          `json:"createdAt"`
	UpdatedAt          time.Time          `json:"updatedAt"`
	CreatedBy          *uuid.UUID         `gorm:"type:uuid" json:"-"`
	UpdatedBy          *uuid.UUID         `gorm:"type:uuid" json:"-"`

	// 关联
	Conditions []AlertCondition `gorm:"foreignkey:RuleID" json:"conditions,omitempty"`
}

// TableName 指定表名
func (AlertRule) TableName() string {
	return "alert_rules"
}

// AlertCondition 告警条件
type AlertCondition struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RuleID      uuid.UUID `gorm:"type:uuid;not null;index" json:"ruleId"`
	Metric      string    `gorm:"type:varchar(50);not null" json:"metric"`
	MetricType  string    `gorm:"type:varchar(100);not null" json:"metricType"`
	Operator    string    `gorm:"type:varchar(10);not null" json:"operator"`
	Threshold   float64   `gorm:"type:decimal(18,4);not null" json:"threshold"`
	ThresholdStr *string  `gorm:"type:varchar(255)" json:"thresholdStr,omitempty"`
	Duration    int       `gorm:"not null;default:60" json:"duration"`
	Aggregation string    `gorm:"type:varchar(20);default:'last'" json:"aggregation"`
	SortOrder   int       `gorm:"default:0" json:"sortOrder"`
	CreatedAt   time.Time `json:"createdAt"`
}

// TableName 指定表名
func (AlertCondition) TableName() string {
	return "alert_conditions"
}

// AlertRecord 告警记录
type AlertRecord struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RuleID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"ruleId"`
	RuleName          string     `gorm:"type:varchar(200);not null" json:"ruleName"`
	VMID              *uuid.UUID `gorm:"type:uuid;index" json:"vmId,omitempty"`
	VMName            *string    `gorm:"type:varchar(200)" json:"vmName,omitempty"`
	GroupID           *uuid.UUID `gorm:"type:uuid" json:"groupId,omitempty"`
	ClusterID         *string    `gorm:"type:varchar(100)" json:"clusterId,omitempty"`
	Metric            string     `gorm:"type:varchar(50);not null" json:"metric"`
	Severity          string     `gorm:"type:varchar(20);not null" json:"severity"`
	TriggerValue      float64    `gorm:"type:decimal(18,4);not null" json:"triggerValue"`
	Threshold         float64    `gorm:"type:decimal(18,4);not null" json:"threshold"`
	ConditionStr      *string    `gorm:"type:text" json:"conditionStr,omitempty"`
	TriggeredAt       time.Time  `gorm:"not null;index" json:"triggeredAt"`
	ResolvedAt        *time.Time `json:"resolvedAt,omitempty"`
	Duration          *int       `json:"duration,omitempty"`
	Status            string     `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`
	AcknowledgedBy    *uuid.UUID `gorm:"type:uuid" json:"-"`
	AcknowledgedByName *string   `gorm:"type:varchar(100)" json:"acknowledgedByName,omitempty"`
	AcknowledgedAt    *time.Time `json:"acknowledgedAt,omitempty"`
	AcknowledgeNote   *string    `gorm:"type:text" json:"acknowledgeNote,omitempty"`
	ResolvedBy        *uuid.UUID `gorm:"type:uuid" json:"-"`
	ResolvedByName    *string    `gorm:"type:varchar(100)" json:"resolvedByName,omitempty"`
	Resolution        *string    `gorm:"type:text" json:"resolution,omitempty"`
	Snapshot          JSONMap    `gorm:"type:jsonb" json:"snapshot,omitempty"`
	NotificationStatus JSONMap   `gorm:"type:jsonb;default:'[]'" json:"notificationStatus,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// TableName 指定表名
func (AlertRecord) TableName() string {
	return "alert_records"
}

// AlertStatistics 告警统计
type AlertStatistics struct {
	Overview struct {
		TotalRules          int `json:"totalRules"`
		ActiveRules         int `json:"activeRules"`
		TotalAlerts         int `json:"totalAlerts"`
		ActiveAlerts        int `json:"activeAlerts"`
		AcknowledgedAlerts  int `json:"acknowledgedAlerts"`
		ResolvedAlerts      int `json:"resolvedAlerts"`
	} `json:"overview"`
	BySeverity struct {
		Critical struct {
			Total  int `json:"total"`
			Active int `json:"active"`
		} `json:"critical"`
		High struct {
			Total  int `json:"total"`
			Active int `json:"active"`
		} `json:"high"`
		Medium struct {
			Total  int `json:"total"`
			Active int `json:"active"`
		} `json:"medium"`
		Low struct {
			Total  int `json:"total"`
			Active int `json:"active"`
		} `json:"low"`
	} `json:"bySeverity"`
	ByMetric []struct {
		Metric      string `json:"metric"`
		Count       int    `json:"count"`
		ActiveCount int    `json:"activeCount"`
	} `json:"byMetric"`
	ByVM []struct {
		VMID        string `json:"vmId"`
		VMName      string `json:"vmName"`
		Count       int    `json:"count"`
		ActiveCount int    `json:"activeCount"`
	} `json:"byVM"`
	ByRule []struct {
		RuleID       string `json:"ruleId"`
		RuleName     string `json:"ruleName"`
		TriggerCount int    `json:"triggerCount"`
	} `json:"byRule"`
	MTTR *struct {
		Avg          float64            `json:"avg"`
		BySeverity   map[string]float64 `json:"bySeverity"`
	} `json:"mttr,omitempty"`
	TimeRange struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	} `json:"timeRange"`
}

// NotificationConfig 通知配置结构
type NotificationConfig struct {
	Methods []string `json:"methods"` // email, sms, webhook, inApp
	Email   *struct {
		Enabled   bool     `json:"enabled"`
		Recipients []string `json:"recipients"`
		CC         []string `json:"cc,omitempty"`
		Template   string   `json:"template,omitempty"`
	} `json:"email,omitempty"`
	SMS *struct {
		Enabled      bool     `json:"enabled"`
		PhoneNumbers []string `json:"phoneNumbers"`
		Template     string   `json:"template,omitempty"`
	} `json:"sms,omitempty"`
	Webhook *struct {
		Enabled  bool              `json:"enabled"`
		URL      string            `json:"url"`
		Method   string            `json:"method"` // POST, PUT
		Headers  map[string]string `json:"headers,omitempty"`
		Secret   string            `json:"secret,omitempty"`
	} `json:"webhook,omitempty"`
	InApp *struct {
		Enabled bool     `json:"enabled"`
		Users   []string `json:"users,omitempty"` // 空表示全部管理员
	} `json:"inApp,omitempty"`
	Escalation *struct {
		Enabled bool `json:"enabled"`
		Levels  []struct {
			Delay     int      `json:"delay"` // 分钟
			Methods   []string `json:"methods"`
			Recipients []string `json:"recipients"`
		} `json:"levels"`
	} `json:"escalation,omitempty"`
}

// AuditLog 审计日志
type AuditLog struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Action       string    `gorm:"type:varchar(50);not null" json:"action"`
	ResourceType string    `gorm:"type:varchar(50);not null" json:"resourceType"`
	ResourceID   string    `gorm:"type:varchar(100);not null" json:"resourceId"`
	ResourceName *string   `gorm:"type:varchar(200)" json:"resourceName,omitempty"`
	Changes      JSONMap   `gorm:"type:jsonb" json:"changes,omitempty"`
	OperatorID   *uuid.UUID `gorm:"type:uuid" json:"operatorId,omitempty"`
	OperatorName *string   `gorm:"type:varchar(100)" json:"operatorName,omitempty"`
	OperatorIP   *string   `gorm:"type:inet" json:"operatorIp,omitempty"`
	UserAgent    *string   `gorm:"type:text" json:"userAgent,omitempty"`
	RequestID    *string   `gorm:"type:varchar(100)" json:"requestId,omitempty"`
	Note         *string   `gorm:"type:text" json:"note,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// Pagination 分页信息
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}
