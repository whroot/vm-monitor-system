package models

import (
	"time"

	"github.com/google/uuid"
)

// VM 虚拟机模型
type VM struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VMwareID          *string   `gorm:"type:varchar(100);uniqueIndex" json:"vmwareId,omitempty"`
	Name              string    `gorm:"type:varchar(200);not null" json:"name"`
	IP                *string   `gorm:"type:inet" json:"ip,omitempty"`
	OSType            *string   `gorm:"type:varchar(20)" json:"os,omitempty"`
	OSVersion         *string   `gorm:"type:varchar(100)" json:"osVersion,omitempty"`
	CPUCores          *int      `json:"cpuCores,omitempty"`
	MemoryGB          *int      `json:"memoryGB,omitempty"`
	DiskGB            *int      `json:"diskGB,omitempty"`
	NetworkAdapters   *int      `json:"networkAdapters,omitempty"`
	PowerState        *string   `gorm:"type:varchar(20)" json:"powerState,omitempty"`
	HostID            *string   `gorm:"type:varchar(100);index" json:"hostId,omitempty"`
	HostName          *string   `gorm:"type:varchar(200)" json:"hostName,omitempty"`
	DatacenterID      *string   `gorm:"type:varchar(100);index" json:"datacenterId,omitempty"`
	DatacenterName    *string   `gorm:"type:varchar(200)" json:"datacenterName,omitempty"`
	ClusterID         *string   `gorm:"type:varchar(100);index" json:"clusterId,omitempty"`
	ClusterName       *string   `gorm:"type:varchar(200)" json:"clusterName,omitempty"`
	GroupID           *uuid.UUID `gorm:"type:uuid;index" json:"groupId,omitempty"`
	Status            string    `gorm:"type:varchar(20);not null;default:'unknown'" json:"status"`
	LastSeen          *time.Time `json:"lastSeen,omitempty"`
	VMwareToolsStatus *string   `gorm:"type:varchar(20)" json:"vmwareToolsStatus,omitempty"`
	VMwareToolsVersion *string  `gorm:"type:varchar(50)" json:"vmwareToolsVersion,omitempty"`
	Tags              JSONMap   `gorm:"type:jsonb;default:'[]'" json:"tags,omitempty"`
	Description       *string   `gorm:"type:text" json:"description,omitempty"`
	Metadata          JSONMap   `gorm:"type:jsonb;default:'{}'" json:"metadata,omitempty"`
	IsDeleted         bool      `gorm:"not null;default:false" json:"-"`
	DeletedAt         *time.Time `json:"-"`
	DeletedBy         *uuid.UUID `gorm:"type:uuid" json:"-"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	CreatedBy         *uuid.UUID `gorm:"type:uuid" json:"-"`
	UpdatedBy         *uuid.UUID `gorm:"type:uuid" json:"-"`

	// 关联
	Group *VMGroup `gorm:"foreignkey:GroupID" json:"group,omitempty"`
}

// TableName 指定表名
func (VM) TableName() string {
	return "vms"
}

// VMGroup VM分组模型
type VMGroup struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name           string     `gorm:"type:varchar(200);not null" json:"name"`
	Description    *string    `gorm:"type:text" json:"description,omitempty"`
	Type           string     `gorm:"type:varchar(20);not null;default:'custom'" json:"type"`
	ParentID       *uuid.UUID `gorm:"type:uuid;index" json:"parentId,omitempty"`
	VMwareObjectID *string    `gorm:"type:varchar(100);index" json:"vmwareObjectId,omitempty"`
	Color          string     `gorm:"type:varchar(7);default:'#2196F3'" json:"color"`
	SortOrder      int        `gorm:"default:0" json:"sortOrder"`
	IsSystem       bool       `gorm:"not null;default:false" json:"isSystem"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	CreatedBy      *uuid.UUID `gorm:"type:uuid" json:"-"`
	UpdatedBy      *uuid.UUID `gorm:"type:uuid" json:"-"`

	// 关联
	Parent   *VMGroup `gorm:"foreignkey:ParentID" json:"parent,omitempty"`
	Children []VMGroup `gorm:"foreignkey:ParentID" json:"children,omitempty"`
	VMs      []VM      `gorm:"many2many:vm_group_members;" json:"-"`

	// 统计字段（非持久化）
	VMCount     int `gorm:"-" json:"vmCount"`
	OnlineCount int `gorm:"-" json:"onlineCount"`
	OfflineCount int `gorm:"-" json:"offlineCount"`
	ErrorCount  int `gorm:"-" json:"errorCount"`
}

// TableName 指定表名
func (VMGroup) TableName() string {
	return "vm_groups"
}

// VMGroupMember VM分组关联
type VMGroupMember struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VMID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_vm_group_member" json:"vmId"`
	GroupID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_vm_group_member" json:"groupId"`
	CreatedAt time.Time `json:"createdAt"`
}

// TableName 指定表名
func (VMGroupMember) TableName() string {
	return "vm_group_members"
}

// VMListRequest VM列表查询请求
type VMListRequest struct {
	Page       int      `json:"page" form:"page" binding:"min=1"`
	PageSize   int      `json:"pageSize" form:"pageSize" binding:"min=1,max=100"`
	Status     string   `json:"status" form:"status"`
	OS         string   `json:"os" form:"os"`
	GroupID    string   `json:"groupId" form:"groupId"`
	HostID     string   `json:"hostId" form:"hostId"`
	ClusterID  string   `json:"clusterId" form:"clusterId"`
	DatacenterID string `json:"datacenterId" form:"datacenterId"`
	Keyword    string   `json:"keyword" form:"keyword"`
	SortBy     string   `json:"sortBy" form:"sortBy"`
	SortOrder  string   `json:"sortOrder" form:"sortOrder"`
}

// VMListResponse VM列表响应
type VMListResponse struct {
	List       []VM       `json:"list"`
	Pagination Pagination `json:"pagination"`
	Summary    VMSummary  `json:"summary"`
}

// VMSummary VM统计摘要
type VMSummary struct {
	Total   int `json:"total"`
	Online  int `json:"online"`
	Offline int `json:"offline"`
	Error   int `json:"error"`
}

// VMStatistics VM统计
type VMStatistics struct {
	Overview struct {
		Total   int `json:"total"`
		Online  int `json:"online"`
		Offline int `json:"offline"`
		Error   int `json:"error"`
		Unknown int `json:"unknown"`
	} `json:"overview"`
	ByOS []struct {
		OS           string `json:"os"`
		Count        int    `json:"count"`
		OnlineCount  int    `json:"onlineCount"`
	} `json:"byOS"`
	ByGroup []struct {
		GroupID     string `json:"groupId"`
		GroupName   string `json:"groupName"`
		Count       int    `json:"count"`
		OnlineCount int    `json:"onlineCount"`
	} `json:"byGroup"`
	ByPowerState []struct {
		State string `json:"state"`
		Count int    `json:"count"`
	} `json:"byPowerState"`
	ByToolsStatus []struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	} `json:"byToolsStatus"`
}

// VMSyncRequest VM同步请求
type VMSyncRequest struct {
	Type         string `json:"type" binding:"required,oneof=full incremental"`
	DatacenterID string `json:"datacenterId"`
	ClusterID    string `json:"clusterId"`
	HostID       string `json:"hostId"`
}

// VMSyncResponse VM同步响应
type VMSyncResponse struct {
	SyncID      string     `json:"syncId"`
	Status      string     `json:"status"`
	Result      *SyncResult `json:"result,omitempty"`
	StartedAt   time.Time  `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

// SyncResult 同步结果
type SyncResult struct {
	TotalVMs int          `json:"totalVMs"`
	Added    int          `json:"added"`
	Updated  int          `json:"updated"`
	Removed  int          `json:"removed"`
	Failed   int          `json:"failed"`
	Errors   []SyncError  `json:"errors,omitempty"`
}

// SyncError 同步错误
type SyncError struct {
	VMwareID string `json:"vmwareId"`
	Error    string `json:"error"`
}
